package middlewares

import (
	ResponseModels "BhariyaAuth/models/responses"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
)

type window struct {
	count uint16
	end   time.Time
}

func RouteRateLimiter(limit uint16, period time.Duration, autoVacuumInterval, maxIdleFor time.Duration) fiber.Handler {
	var mutex sync.Mutex
	clients := make(map[string]*window)

	go func() {
		ticker := time.NewTicker(autoVacuumInterval)
		defer ticker.Stop()
		for range ticker.C {
			mutex.Lock()
			now := time.Now()
			for i, v := range clients {
				if now.Sub(v.end) > maxIdleFor {
					delete(clients, i)
				}
			}
			mutex.Unlock()
		}
	}()

	return func(ctx fiber.Ctx) error {
		key := ctx.IP()
		now := time.Now()

		mutex.Lock()
		entry, exists := clients[key]
		mutex.Unlock()

		if !exists || now.After(entry.end) {
			err := ctx.Next()
			mutex.Lock()
			if e, ok := clients[key]; ok && now.Before(e.end) {
				entry = e
				entry.count += RateLimitProcessor.Get(ctx)
			} else {
				clients[key] = &window{count: 1, end: now.Add(period)}
			}
			mutex.Unlock()
			return err
		}

		if entry.count < limit*100 {
			err := ctx.Next()
			mutex.Lock()
			entry.count += RateLimitProcessor.Get(ctx)
			mutex.Unlock()
			return err
		}

		retryAfter := int(entry.end.Sub(now).Seconds()) + 1
		return ctx.Status(fiber.StatusTooManyRequests).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{fmt.Sprintf("Too many requests, retrying automatically after %d seconds", retryAfter)},
				RetryAfter:    retryAfter,
			})
	}
}
