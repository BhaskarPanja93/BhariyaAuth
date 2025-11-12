package string

import (
	"regexp"

	"github.com/medama-io/go-useragent"
)

func IsValidEmail(email string) bool {
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func PasswordIsStrong(password string) bool {
	n := len(password)
	if n < 7 || n > 20 {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for i := 0; i < n; i++ {
		c := password[i]
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		}
		if hasUpper && hasLower && hasDigit {
			return true
		}
	}
	return hasUpper && hasLower && hasDigit
}

var UAParser = useragent.NewParser()
