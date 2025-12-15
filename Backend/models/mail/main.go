package mail

type T struct {
	Subject   string
	Header    string
	Ignorable bool
}

type allT struct {
	MFAInitiate             T
	PasswordResetInitiate   T
	LoginInitiate           T
	RegisterInitiate        T
	MFASuccessful           T
	PasswordResetSuccessful T
	LoginSuccessful         T
	RegisterSuccessful      T
}

var All = allT{
	MFAInitiate: T{
		Subject:   "Your verification code",
		Header:    "Enter the OTP below to complete MFA verification:",
		Ignorable: false,
	},
	MFASuccessful: T{
		Subject:   "Verification successful",
		Header:    "Multi-Factor Authentication Completed",
		Ignorable: true,
	},
	PasswordResetInitiate: T{
		Subject:   "Password reset request",
		Header:    "Enter the OTP below to reset password:",
		Ignorable: true,
	},
	PasswordResetSuccessful: T{
		Subject:   "Password successfully reset",
		Header:    "Your Password Has Been Changed",
		Ignorable: true,
	},
	LoginInitiate: T{
		Subject:   "New login attempt detected",
		Header:    "Enter the OTP below to login:",
		Ignorable: true,
	},
	LoginSuccessful: T{
		Subject:   "Login successful",
		Header:    "Successful Login",
		Ignorable: false,
	},
	RegisterInitiate: T{
		Subject:   "Registration attempt",
		Header:    "Enter the OTP below to verify your email:",
		Ignorable: true,
	},
	RegisterSuccessful: T{
		Subject:   "Welcome! Your account is ready",
		Header:    "Registration Complete",
		Ignorable: false,
	},
}
