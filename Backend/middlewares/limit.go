package middlewares

import (
	ResponseModels "BhariyaAuth/models/responses"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
)

type rateWindow struct {
	count   atomic.Uint32 // Starts with 0 and holds the current value
	expires time.Time     // Keeps time for when the window expires (windowDuration)
	cleanup time.Time     // Keeps time for when the window should be cleaned if idle (2 x windowDuration)
}

// RouteRateLimiter provides a route-level fixed-window rate limiting middleware.
//
// This middleware enforces a per-IP rate limit using a fixed time window strategy.
// Each client IP is assigned a "window" that tracks:
//   - Number of requests made within the window.
//   - Expiry time of the current window.
//
// When the request count exceeds the configured limit:
//   - The request is rejected with HTTP 429 (Too Many Requests).
//   - A Retry-After header and response field indicate when the client can retry.
//
// Additionally:
// - A background cleaner periodically removes expired windows.
// - Atomic counters are used for lock-free increments per window.
// - RWMutex protects the shared map of client windows.
//
// Flow Summary:
//
//	identify client → fetch/create window → check expiry → enforce limit → update counter → process request
//
// Concurrency Model:
// - Map access is guarded by RWMutex.
// - Per-window counters use atomic operations (no lock needed for increments).
// - Window expiration reset is done opportunistically during request handling.
// - Cleanup runs in a background goroutine.
//
// Parameters:
// - limit: maximum allowed requests per window.
// - windowDuration: duration of each rate limit window.
// - cleanupInterval: interval for cleaning expired windows.
//
// Returns:
// - Middleware handler enforcing rate limits.
func RouteRateLimiter(limit uint32, windowDuration time.Duration, cleanupInterval time.Duration) fiber.Handler {

	// RWMutex to protect access to clientWindows map
	var mu sync.RWMutex

	// Map storing per-IP rate windows
	clientWindows := make(map[string]*rateWindow)

	// Periodically removes stale windows to prevent memory leaks
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			var toDelete []string

			// Identify expired windows (read lock)
			mu.RLock()
			for ip, window := range clientWindows {
				// Window is eligible for deletion only after:
				// expires + cleanupInterval
				// This delay allows reuse/reset instead of frequent reallocation
				if now.After(window.cleanup) {
					toDelete = append(toDelete, ip)
				}
			}
			mu.RUnlock()

			// Delete expired windows (write lock)
			if len(toDelete) > 0 {
				mu.Lock()
				for _, ip := range toDelete {
					// Double-check condition to avoid race with updates
					if window, ok := clientWindows[ip]; ok &&
						now.After(window.expires.Add(cleanupInterval)) {
						delete(clientWindows, ip)
					}
				}
				mu.Unlock()
			}
		}
	}()

	// Actual middleware function
	return func(c fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		// Fetch existing window (read lock)
		mu.RLock()
		window, exists := clientWindows[ip]
		mu.RUnlock()

		// Create new window if not exists
		if !exists {
			mu.Lock()
			window = &rateWindow{
				count:   atomic.Uint32{}, // starts at 0
				expires: now.Add(windowDuration),
				cleanup: now.Add(windowDuration + windowDuration),
			}
			clientWindows[ip] = window
			mu.Unlock()

			// Reset window if expired
		} else if now.After(window.expires) {
			window.count.Store(0)
			window.expires = now.Add(windowDuration)
			window.cleanup = now.Add(windowDuration + windowDuration)
		}

		// Enforce rate limit
		if window.count.Load() >= limit {

			// Calculate remaining time until reset
			retryAfter := int(window.expires.Sub(now).Seconds()) + 1

			// Set Retry-After header
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(retryAfter))

			// Return 429 response
			return c.Status(fiber.StatusTooManyRequests).JSON(ResponseModels.APIResponseT{
				RetryAfter: retryAfter,
			})
		}
		// Reserve capacity for this request to prevent overshoot
		// especially for long-running requests
		RateLimitProcessor.Init(c)
		window.count.Add(RateLimitProcessor.Get(c))

		// Execute downstream handler
		err := c.Next()

		// Add actual cost
		window.count.Add(RateLimitProcessor.Get(c))
		return err
	}
}
