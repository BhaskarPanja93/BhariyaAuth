package otp

import "time"

// Constants defining OTP and rate-limiting behavior.
//
// baseDelay: initial delay applied after first OTP send.
// maxDelay: maximum delay cap to prevent unbounded backoff.
// otpTTL: validity duration of each OTP.
// cleanupInterval: frequency of background cleanup process.
const (
	baseDelay       = 5 * time.Second
	maxDelay        = 30 * time.Minute
	otpTTL          = 5 * time.Minute
	cleanupInterval = 5 * time.Minute
)
