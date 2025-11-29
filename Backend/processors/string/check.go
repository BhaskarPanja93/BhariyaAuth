package string

import (
	"regexp"

	"github.com/medama-io/go-useragent"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

var passwordRegex = regexp.MustCompile(`^(?=.*[A-Z])(?=.*[a-z])(?=.*\d).{7,20}$`)

func PasswordIsStrong(password string) bool {
	return passwordRegex.MatchString(password)
}

var UAParser = useragent.NewParser()
