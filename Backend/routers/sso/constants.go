package sso

const (
	UnknownProvider     = "UnknownProvider"
	SessionCreateFailed = "SessionCreateFailed"
	URLParseFailed      = "URLParseFailed"
	AuthoriseFailed     = "AuthoriseFailed"
	FetchUserFailed     = "FetchUserFailed"
	AuthURLNotFound     = "AuthURLNotFound"

	StateInvalid       = "StateInvalid"
	StateEncryptFailed = "StateEncryptFailed"

	SessionEncryptFailed = "SessionEncryptFailed"
	SessionDecryptFailed = "SessionDecryptFailed"

	SessionInvalid = "SessionInvalid"
	SessionExpired = "SessionExpired"

	AccountCreateFailed = "AccountCreateFailed"
	SignInFailed        = "SignInFailed"
	AccountBlocked      = "AccountBlocked"

	ProviderParam = "processor"
)
