package middlewares

import (
	ResponseModels "BhariyaAuth/models/responses"
	RateLimitProcessor "BhariyaAuth/processors/request"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

type rateWindow struct {
	mu      sync.Mutex
	count   uint32
	expires time.Time
	cleanup time.Time
}

func RouteRateLimiter(limit uint32, windowDuration time.Duration, cleanupInterval time.Duration) fiber.Handler {
	var mu sync.RWMutex
	clientWindows := make(map[string]*rateWindow)

	go func() {
		ticker := time.NewTicker(cleanupInterval)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			var toDelete []string

			mu.RLock()
			for ip, window := range clientWindows {
				window.mu.Lock()
				shouldDelete := now.After(window.cleanup)
				window.mu.Unlock()

				if shouldDelete {
					toDelete = append(toDelete, ip)
				}
			}
			mu.RUnlock()

			if len(toDelete) == 0 {
				continue
			}

			mu.Lock()
			for _, ip := range toDelete {
				window, ok := clientWindows[ip]
				if !ok {
					continue
				}

				window.mu.Lock()
				stillExpired := now.After(window.expires.Add(cleanupInterval))
				window.mu.Unlock()

				if stillExpired {
					delete(clientWindows, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(ctx fiber.Ctx) error {
		ip := ctx.IP()
		now := time.Now()

		mu.RLock()
		window, exists := clientWindows[ip]
		mu.RUnlock()

		if !exists {
			mu.Lock()
			window, exists = clientWindows[ip]
			if !exists {
				window = &rateWindow{
					count:   0,
					expires: now.Add(windowDuration),
					cleanup: now.Add(windowDuration + cleanupInterval),
				}
				clientWindows[ip] = window
			}
			mu.Unlock()
		}

		window.mu.Lock()
		if now.After(window.expires) {
			window.count = 0
			window.expires = now.Add(windowDuration)
			window.cleanup = now.Add(windowDuration + cleanupInterval)
		}

		if window.count >= limit {
			retryAfter := int(window.expires.Sub(now).Seconds()) + 1
			window.mu.Unlock()

			ctx.Set(fiber.HeaderRetryAfter, strconv.Itoa(retryAfter))
			return ctx.Status(fiber.StatusTooManyRequests).JSON(ResponseModels.APIResponseT{
				RetryAfter: retryAfter,
			})
		}

		window.count++
		window.mu.Unlock()

		err := ctx.Next()

		weight := RateLimitProcessor.GetRateLimitWeight(ctx)
		if weight > 1 {
			window.mu.Lock()
			window.count += weight - 1
			window.mu.Unlock()
		}

		return err
	}
}
