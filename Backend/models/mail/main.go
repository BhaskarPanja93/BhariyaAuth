package mail

type OtpT struct {
	Subject   string
	Header    string
	Ignorable bool
}

var (
	MFAInitiated = OtpT{
		Subject:   "Multi-Factor started",
		Header:    "Enter the OTP below to complete MFA verification:",
		Ignorable: false,
	}
	PasswordResetInitiated = OtpT{
		Subject:   "Password reset started",
		Header:    "Enter the OTP below to reset password:",
		Ignorable: true,
	}
	LoginInitiated = OtpT{
		Subject:   "New login started",
		Header:    "Enter the OTP below to login:",
		Ignorable: true,
	}
	RegisterInitiated = OtpT{
		Subject:   "Registration started",
		Header:    "Enter the OTP below to verify your email:",
		Ignorable: true,
	}
)
