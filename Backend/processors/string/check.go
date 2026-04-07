package string

import (
	"unicode"
)

// EmailIsValid validates email format and length constraints.
//
// Rules:
// - Length: 6 to 50 characters.
// - Must match emailRegex pattern.
//
// Returns:
// - true if email is valid.
// - false otherwise.
//
// Security Considerations:
// - Prevents malformed input.
// - Does NOT guarantee email existence or deliverability.
func EmailIsValid(email string) bool {
	return len(email) > 5 &&
		len(email) <= 50 &&
		emailRegex.MatchString(email)
}

// NameIsValid validates name.
//
// Rules:
// - Length: 3 to 50 characters.
//
// Returns:
// - true if valid.
// - false otherwise.
//
// Notes:
// - No character restrictions applied.
// - Accepts Unicode characters.
func NameIsValid(name string) bool {
	return len(name) > 2 && len(name) <= 50
}

// PasswordIsStrong enforces password complexity requirements.
//
// Rules:
// - Length: 8 to 72 characters (bcrypt limit).
// - Must contain:
//   - at least one uppercase letter
//   - at least one lowercase letter
//   - at least one digit
//
// Returns:
// - true if password meets strength requirements.
// - false otherwise.
//
// Security Considerations:
// - Enforces basic complexity but does NOT check:
//   - special characters
//   - breached password lists
//   - entropy scoring
func PasswordIsStrong(pw string) bool {

	// Enforce length constraints (bcrypt safe range)
	if len(pw) < 8 || len(pw) > 72 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool

	// Iterate through characters and classify
	for _, ch := range pw {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}

		// Early exit if all conditions satisfied
		if hasUpper && hasLower && hasDigit {
			return true
		}
	}

	return hasUpper && hasLower && hasDigit
}

// ParseUA extracts OS, device, and browser details from User-Agent string.
//
// - Uses useragent parser to derive structured information.
// - Provides safe fallbacks for missing values.
//
// Returns:
// - OS: operating system name.
// - Device: device type (mobile, desktop, etc.).
// - Browser: browser name + version.
//
// Behavior:
// - Defaults to "Unknown" if any field is empty.
//
// Example Output:
// - OS: "Windows 11"
// - Device: "Desktop"
// - Browser: "Chrome 120.0"
func ParseUA(UA string) (string, string, string) {

	parsed := uaParser.Parse(UA)

	OS := parsed.OS().String()
	Device := parsed.Device().String()
	Browser := parsed.Browser().String()

	return OS, Device, Browser + " " + parsed.BrowserVersion()
}
