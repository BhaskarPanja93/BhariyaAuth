package notifications

const (
	SessionExpired = "Current Session Expired. Please re-login"
	SessionRevoked = "Current Session Revoked. Please re-login"

	DBReadError    = "DB Read error. Retrying"
	DBWriteError   = "DB Write error. Retrying"
	MarshalError   = "Marshal error. Retrying"
	EncryptorError = "Encryption error. Retrying"

	AccountNotFound = "Account not found"
	AccountCreated  = "Account created"
	AccountBlocked  = "Account blocked"
	AccountPresent  = "Account present"

	OTPSendFailed     = "OTP send failed"
	OTPIncorrect      = "OTP incorrect"
	PasswordNotSet    = "Password not set"
	PasswordIncorrect = "Password incorrect"
	PasswordChanged   = "Password changed"

	RevokeFailed = "Revoke failed"
)
