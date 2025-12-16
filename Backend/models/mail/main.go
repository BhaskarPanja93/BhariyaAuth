package mail

type OtpT struct {
	Subject   []string
	Header    string
	Ignorable bool
}

var (
	MFAInitiated = OtpT{
		Subject: []string{
			"Just checking if its actually you",
			"MFA required",
			"MFA verification",
			"Approve MFA to proceed",
			"Verify your session for sensitive actions",
			"Confirm ownership to continue",
			"One-time code to strengthen your session",
			"Verify ownership of this session",
		},
		Header: "Enter the OTP below to verify your session and continue:",
	}

	PasswordResetInitiated = OtpT{
		Subject: []string{
			"Reset your password securely",
			"Your password reset code",
			"Let’s set a new password",
			"Password reset",
			"Finish resetting your password",
			"Your account recovery code",
			"Reset your password — one quick step left",
			"Secure your new password",
			"Confirm your password reset",
		},
		Header: "Enter the OTP below to reset password:",
	}

	LoginInitiated = OtpT{
		Subject: []string{
			"Confirm your login",
			"Logging in? Verify here",
			"Your login verification code",
			"Secure login confirmation",
			"Finish signing in",
			"Verify your login request",
			"Accessing your account?",
			"One step away from logging in",
		},
		Header: "Enter the OTP below to login:",
	}

	RegisterInitiated = OtpT{
		Subject: []string{
			"Welcome! Let’s verify your email",
			"Complete your registration",
			"Almost done — confirm your email",
			"Verify your email to get started",
			"Finish setting up your account",
			"Your registration verification code",
			"Confirm your email address",
			"Activate your new account",
			"One last step to join us",
			"Let’s get your account ready",
		},
		Header: "Enter the OTP below to verify your email:",
	}
)
