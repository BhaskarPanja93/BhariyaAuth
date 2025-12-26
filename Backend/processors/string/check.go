package string

import (
	"regexp"
	"unicode"

	"github.com/medama-io/go-useragent"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func EmailIsValid(email string) bool {
	return len(email) > 0 && len(email) < 51 && emailRegex.MatchString(email)
}

func NameIsValid(name string) bool {
	return len(name) > 0 && len(name) < 51
}

func PasswordIsStrong(pw string) bool {
	if len(pw) < 8 || len(pw) > 72 {
		return false
	}
	hasUpper := false
	hasLower := false
	hasDigit := false
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

var UAParser = useragent.NewParser()

func ParseUA(UA string) (string, string, string) {
	parsed := UAParser.Parse(UA)
	OS := parsed.OS().String()
	Device := parsed.Device().String()
	Browser := parsed.Browser().String()
	if OS == "" {
		OS = "Unknown"
	}
	if Device == "" {
		Device = "Unknown"
	}
	if Browser == "" {
		Browser = "Unknown"
	}
	return OS, Device, Browser + " " + parsed.BrowserVersion()
}
