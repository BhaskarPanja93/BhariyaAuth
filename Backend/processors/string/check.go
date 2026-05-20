package string

import (
	"unicode"
)

func EmailIsValid(email string) bool {
	return len(email) > 5 &&
		len(email) <= 50 &&
		emailRegex.MatchString(email)
}

func NameIsValid(name string) bool {
	return len(name) > 2 && len(name) <= 50
}

func PasswordIsStrong(pw string) bool {

	if len(pw) < 8 || len(pw) > 72 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool

	for _, ch := range pw {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}

		if hasUpper && hasLower && hasDigit {
			return true
		}
	}

	return hasUpper && hasLower && hasDigit
}

func ParseUA(UA string) (string, string, string) {

	parsed := uaParser.Parse(UA)

	OS := parsed.OS().String()
	Device := parsed.Device().String()
	Browser := parsed.Browser().String()
	BrowserVersion := parsed.BrowserVersion()

	if OS == "" {
		OS = "Unknown"
	}
	if Device == "" {
		Device = "Unknown"
	}
	if Browser == "" {
		Browser = "Unknown"
	}
	if BrowserVersion != "" {
		Browser += " " + BrowserVersion
	}

	return OS, Device, Browser
}
