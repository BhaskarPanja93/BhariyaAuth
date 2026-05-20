package otp

import "time"

const (
	baseDelay       = 5 * time.Second
	maxDelay        = 30 * time.Minute
	otpTTL          = 5 * time.Minute
	cleanupInterval = 5 * time.Minute
)
