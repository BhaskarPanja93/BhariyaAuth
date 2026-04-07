package otp

import (
	MailModels "BhariyaAuth/models/mails"
	MailNotifier "BhariyaAuth/processors/mail"
	StringProcessor "BhariyaAuth/processors/string"
	"fmt"
	"time"
)

// Send generates and dispatches an OTP to a user with rate limiting.
//
// Flow:
//
//	check rate limit → generate OTP → send mail → store OTP → return verification token
//
// Parameters:
// - address: recipient email.
// - model: mail template model.
// - identifier: additional key component (e.g., IP).
//
// Returns:
// - verification: token used later for validation.
// - delay: retry delay if rate-limited or after sending.
//
// Notes:
// - verification token is separate from OTP (prevents brute-force guessing).
// - OTP is stored server-side and never exposed directly.
func Send(address string, model MailModels.T, identifier string) (string, time.Duration) {
	key := fmt.Sprintf("%s:%s", address, identifier)

	// Check rate limit
	canSend, count, wait := checkCanSend(key)
	if !canSend {
		return "", wait
	}

	// Generate OTP (numeric)
	otp := StringProcessor.SafeNumber(6)

	// Generate verification token (lookup key)
	verification := StringProcessor.SafeString(12)

	// Record send + store OTP
	delay := recordSend(key, verification, otp, count)

	// Send email
	if ok := MailNotifier.OTP(address, otp, model, 2); !ok {
		return "", wait
	}

	return verification, delay
}

// Validate verifies the OTP against stored value.
//
// Flow:
//
//	lookup OTP → check expiry → compare → delete on success
//
// Parameters:
// - verification: lookup key.
// - otp: user-provided OTP.
//
// Returns:
// - true if valid.
// - false otherwise.
//
// Security:
// - OTP is single-use (deleted after successful validation).
// - Expired OTPs are rejected.
func Validate(verification, otp string) bool {
	val, ok := otpStore.Load(verification)
	if !ok {
		return false
	}

	entry := val.(*otpEntry)

	// Enforce single-use OTP
	if time.Now().After(entry.expires) || otp != entry.otp {
		return false
	}

	// Enforce single-use OTP
	otpStore.Delete(verification)
	return true
}
