package middlewares

import (
	ResponseModels "BhariyaAuth/models/responses"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v3"
)

type rateWindow struct {
	count   atomic.Uint32
	expires time.Time
}

func RouteRateLimiter(limit uint32, windowDuration time.Duration, cleanupInterval time.Duration, maxIdleDuration time.Duration) fiber.Handler {
	var mu sync.Mutex
	clientWindows := make(map[string]*rateWindow)

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()

			mu.Lock()
			for ip, window := range clientWindows {
				if now.After(window.expires.Add(maxIdleDuration)) {
					delete(clientWindows, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()
		mu.Lock()
		window, exists := clientWindows[ip]
		if !exists || now.After(window.expires) {
			window = &rateWindow{
				count:   atomic.Uint32{},
				expires: now.Add(windowDuration),
			}
			clientWindows[ip] = window
		}
		mu.Unlock()
		if window.count.Load() >= limit {
			retryAfter := int(window.expires.Sub(now).Seconds()) + 1
			return c.Status(fiber.StatusTooManyRequests).JSON(ResponseModels.APIResponseT{
				RetryAfter: retryAfter,
			})
		}
		err := c.Next()
		window.count.Add(RateLimitProcessor.Get(c))
		return err
	}
}
