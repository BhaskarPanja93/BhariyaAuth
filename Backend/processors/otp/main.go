package otp

import (
	"math"
	"sync"
	"time"
)

type limiterEntry struct {
	mu        sync.RWMutex
	sentCount uint32
	lastSent  time.Time
}

type otpEntry struct {
	otp     string
	expires time.Time
}

var limiterStore sync.Map
var otpStore sync.Map

func init() {
	go cleanupLoop()
}

func calculateDelay(sentCount uint32) time.Duration {
	if sentCount == 0 {
		return 0
	}

	exp := math.Pow(2, float64(sentCount-1))
	delay := time.Duration(exp) * baseDelay

	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

func checkCanSend(key string) (bool, uint32, time.Duration) {
	val, ok := limiterStore.Load(key)
	if !ok {
		return true, 0, 0
	}

	entry := val.(*limiterEntry)

	entry.mu.RLock()
	defer entry.mu.RUnlock()

	delay := calculateDelay(entry.sentCount)
	elapsed := time.Since(entry.lastSent)

	if elapsed >= delay {
		return true, entry.sentCount, 0
	}

	return false, entry.sentCount, delay - elapsed
}

func recordSend(key, verification, otp string, prevCount uint32) time.Duration {
	now := time.Now()

	val, _ := limiterStore.LoadOrStore(key, &limiterEntry{})
	entry := val.(*limiterEntry)

	entry.mu.Lock()
	entry.sentCount = prevCount + 1
	entry.lastSent = now
	entry.mu.Unlock()
	newCount := entry.sentCount

	otpStore.Store(verification, &otpEntry{
		otp:     otp,
		expires: now.Add(otpTTL),
	})

	return calculateDelay(newCount)
}

func cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		otpStore.Range(func(key, value any) bool {
			entry := value.(*otpEntry)
			if now.After(entry.expires) {
				otpStore.Delete(key)
			}
			return true
		})

		limiterStore.Range(func(key, value any) bool {
			entry := value.(*limiterEntry)

			entry.mu.RLock()
			inactive := now.Sub(entry.lastSent) > maxDelay
			entry.mu.RUnlock()

			if inactive {
				limiterStore.Delete(key)
			}
			return true
		})
	}
}
