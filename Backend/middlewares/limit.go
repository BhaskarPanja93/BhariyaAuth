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
	expires time.Time     // Keeps time for when the window expires (now() + windowDuration)
}

// RouteRateLimiter is a Route based Middleware that provides Fixed window rate limiting with added retry-after header and JSON response
func RouteRateLimiter(limit uint32, windowDuration time.Duration, cleanupInterval time.Duration) fiber.Handler {
	// Mutex for the entire window map
	var mu sync.Mutex
	// map for each IP to their current window
	clientWindows := make(map[string]*rateWindow)
	// Cleaner function to clean up expired windows
	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			now := time.Now()
			// Lock window creations and fetching (till current cleaning batch is complete)
			mu.Lock()
			for ip, window := range clientWindows {
				// Cleanup only after window has lasted for CreationTime + WindowDuration + CleanupInterval
				// The extra CleanupInterval will give room for the window to reset its values instead of completely deleting and
				// reassigning the same window
				if now.After(window.expires.Add(cleanupInterval)) {
					delete(clientWindows, ip)
				}
			}
			// Unlock window creations and fetching
			mu.Unlock()
		}
	}()

	// Actual middleware function
	return func(c fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()
		// Lock window creations and fetching (till a window is found or created)
		mu.Lock()
		window, exists := clientWindows[ip]
		if !exists {
			window = &rateWindow{
				count:   atomic.Uint32{},
				expires: now.Add(windowDuration),
			}
			clientWindows[ip] = window
		} else if now.After(window.expires) {
			window.count.Store(0)
			window.expires = now.Add(windowDuration)
		}
		// Unlock window creations and fetching
		mu.Unlock()
		if window.count.Load() >= limit {
			// Send the time when the window expires, which will in turn reset the old limit value as header: Retry-After and body: retry-after
			retryAfter := int(window.expires.Sub(now).Seconds()) + 1
			c.Set(fiber.HeaderRetryAfter, strconv.Itoa(retryAfter))
			return c.Status(fiber.StatusTooManyRequests).JSON(ResponseModels.APIResponseT{
				RetryAfter: retryAfter,
			})
		}
		// Add a temporary value until request is finished processing
		// This will prevent requests to cross limit (yet not count) if request takes a long time to complete
		// Because count is added only after the request is executed
		RateLimitProcessor.Set(c)
		window.count.Add(RateLimitProcessor.Get(c))
		// Execute current request
		err := c.Next()
		// Add actual value
		window.count.Add(RateLimitProcessor.Get(c))
		return err
	}
}
