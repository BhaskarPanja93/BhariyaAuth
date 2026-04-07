package token

import "errors"

var (
	IncorrectAccessTokenTypeError        = errors.New("read access: token type not allowed")
	IncorrectRefreshTokenTypeError       = errors.New("read refresh: token type not allowed")
	IncorrectMFATokenTypeError           = errors.New("read mfa: token type not allowed")
	IncorrectSSOTokenTypeError           = errors.New("read sso: token type not allowed")
	IncorrectPasswordResetTokenTypeError = errors.New("read password reset: token type not allowed")
	IncorrectSignInTokenTypeError        = errors.New("read sign in: token type not allowed")
	IncorrectSignUpTokenTypeError        = errors.New("read sign up: token type not allowed")
)
