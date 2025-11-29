package string

import (
	"regexp"
	"unicode"

	"github.com/medama-io/go-useragent"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func NameIsValid(name string) bool {
	return len(name) > 0 && len(name) < 51
}

func EmailIsValid(email string) bool {
	return len(email) > 0 && len(email) < 51 && emailRegex.MatchString(email)
}

func PasswordIsStrong(pw string) bool {
	if len(pw) < 8 || len(pw) > 19 {
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
