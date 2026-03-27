package sso

const (
	UnknownProvider = "UnknownProvider"
	BeginFailed     = "BeginFailed"
	URLParseFailed  = "URLParseFailed"
	AuthoriseFailed = "AuthoriseFailed"
	FetchUserFailed = "FetchUserFailed"
	AuthURLNotFound = "AuthURLNotFound"

	StateEncryptFailed = "StateEncryptFailed"
	StateDecryptFailed = "StateDecryptFailed"

	StateMarshalFailed   = "StateMarshalFailed"
	StateUnMarshalFailed = "StateUnMarshalFailed"

	StateMissing = "StateMissing"

	SessionEncryptFailed = "SessionEncryptFailed"
	SessionDecryptFailed = "SessionDecryptFailed"

	SessionMarshalFailed   = "SessionMarshalFailed"
	SessionUnmarshalFailed = "SessionUnmarshalFailed"

	SessionMissing = "SessionMissing"
	SessionInvalid = "SessionInvalid"
	SessionExpired = "SessionExpired"

	AccountCreateFailed = "AccountCreateFailed"
	LoginFailed         = "LoginFailed"
	AccountBlocked      = "AccountBlocked"

	ProcessorParam = "processor"
)
