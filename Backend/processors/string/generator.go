package string

import (
	"crypto/rand"
	"io"
)

// Character sets used for random generation.
//
// letters:
// - URL-safe character set.
// - Includes alphanumeric + - . _ ~
// - Suitable for tokens, identifiers, and verification keys.
//
// numbers:
// - Numeric-only charset.
// - Used for OTP generation.
const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~"
	numbers = "0123456789"
)

// rejectionSample fills dst with cryptographically secure random characters
// from the provided charset using rejection sampling.
//
// This function ensures uniform distribution of characters by:
//  1. Reading random bytes from crypto/rand.
//  2. Rejecting values that would introduce modulo bias.
//  3. Mapping accepted bytes into the charset.
//
// Why Rejection Sampling:
// - Direct modulo (b % len(charset)) introduces bias when 256 is not divisible by charset length.
// - Rejection sampling avoids this by discarding values >= _max.
//
// Parameters:
// - dst: destination byte slice to fill.
// - charset: string of allowed characters.
//
// Behavior:
// - Continues reading random bytes until dst is fully populated.
// - Panics if crypto/rand fails (critical system failure).
//
// Security:
// - Uses crypto/rand (CSPRNG).
// - Ensures uniform distribution across charset.
func rejectionSample(dst []byte, charset string) {

	charsetLen := byte(len(charset))

	// Largest multiple of charsetLen less than 256
	// Ensures unbiased modulo operation
	_max := byte(256 - (256 % int(charsetLen)))

	buf := make([]byte, len(dst))
	n := 0

	for n < len(dst) {

		// Fill buffer with cryptographically secure random bytes
		if _, err := io.ReadFull(rand.Reader, buf); err != nil {
			panic(err) // critical failure: randomness unavailable
		}

		for _, b := range buf {

			// Reject values that would introduce bias
			if b < _max {
				dst[n] = charset[b%charsetLen]
				n++

				if n == len(dst) {
					return
				}
			}
		}
	}
}

// SafeString generates a cryptographically secure random string.
//
// - Uses rejection sampling with URL-safe charset.
// - Suitable for:
//   - tokens
//   - verification IDs
//   - session identifiers
//
// Parameters:
// - n: desired length (minimum enforced = 1).
//
// Returns:
// - Random string of length n.
//
// Security:
// - Uniform distribution.
// - CSPRNG-backed.
// - Safe for security-sensitive use cases.
func SafeString(n uint16) string {

	if n == 0 {
		n = 1
	}

	b := make([]byte, n)

	rejectionSample(b, letters)

	return string(b)
}

// SafeNumber generates a cryptographically secure numeric string.
//
// - Uses rejection sampling with numeric charset.
// - Typically used for OTPs.
//
// Parameters:
// - n: desired length (minimum enforced = 1).
//
// Returns:
// - Numeric string of length n.
//
// Security:
// - Uniform distribution across digits.
// - Resistant to bias and prediction.
func SafeNumber(n uint16) string {

	if n == 0 {
		n = 1
	}

	b := make([]byte, n)

	rejectionSample(b, numbers)

	return string(b)
}
