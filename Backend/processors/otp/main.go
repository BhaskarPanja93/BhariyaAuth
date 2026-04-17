package otp

import (
	"math"
	"sync"
	"time"
)

// limiterEntry tracks rate-limiting state per user (key).
//
// Fields:
// - sentCount: number of OTPs sent in current backoff progression.
// - lastSent: timestamp of last OTP dispatch.
// - mu: protects concurrent access to entry fields.
type limiterEntry struct {
	mu        sync.RWMutex
	sentCount uint32
	lastSent  time.Time
}

// otpEntry represents a single OTP instance.
//
// Fields:
// - otp: generated OTP value.
// - expires: expiration timestamp after which OTP becomes invalid.
type otpEntry struct {
	otp     string
	expires time.Time
}

// Global stores:
// - limiterStore: per-user rate limiting state.
// - otpStore: active OTPs indexed by verification token.
var limiterStore sync.Map // map[string]*limiterEntry
var otpStore sync.Map     // map[string]*otpEntry

// init starts background cleanup routine for expired OTPs and stale limiter entries.
func init() {
	go cleanupLoop()
}

// calculateDelay computes exponential backoff delay based on send count.
//
// Formula:
//
//	delay = baseDelay * 2^(sentCount-1)
//
// Behavior:
// - First send → no delay.
// - Subsequent sends → exponentially increasing delay.
// - Delay is capped at maxDelay.
//
// Purpose:
// - Prevent OTP abuse/spamming.
// - Gradually penalize repeated requests.
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

// checkCanSend determines whether an OTP can be sent for a given key.
//
// Flow:
//
//	load limiter entry → calculate delay → compare elapsed time
//
// Returns:
// - canSend: whether OTP can be sent now.
// - sentCount: previous send count.
// - wait: remaining delay before next allowed send.
//
// Concurrency:
// - Entry-level mutex ensures safe reads/writes.
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

// recordSend updates limiter state and stores OTP for verification.
//
// Flow:
//
//	update limiter → increment count → store OTP with TTL
//
// Parameters:
// - key: rate limit identifier (user/IP/etc).
// - verification: unique token for OTP lookup.
// - otp: generated OTP value.
// - prevCount: previous send count.
//
// Returns:
// - next delay duration based on updated count.
func recordSend(key, verification, otp string, prevCount uint32) time.Duration {
	now := time.Now()

	// Ensure limiter entry exists
	val, _ := limiterStore.LoadOrStore(key, &limiterEntry{})
	entry := val.(*limiterEntry)

	// Ensure limiter entry exists
	entry.mu.Lock()
	entry.sentCount = prevCount + 1
	entry.lastSent = now
	entry.mu.Unlock()
	newCount := entry.sentCount

	// Store OTP with expiration
	otpStore.Store(verification, &otpEntry{
		otp:     otp,
		expires: now.Add(otpTTL),
	})

	return calculateDelay(newCount)
}

// cleanupLoop periodically removes expired OTPs and inactive limiter entries.
//
// Behavior:
// - Runs every cleanupInterval.
// - Deletes:
//   - expired OTP entries.
//   - limiter entries inactive longer than maxDelay.
//
// Purpose:
// - Prevent memory leaks.
// - Keep stores bounded over time.
func cleanupLoop() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()

		// Clean expired OTPs
		otpStore.Range(func(key, value any) bool {
			entry := value.(*otpEntry)
			if now.After(entry.expires) {
				otpStore.Delete(key)
			}
			return true
		})

		// Clean inactive limiter entries
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
