package mails

type T struct {
	Subjects  []string
	Header    string
	Ignorable bool
}

var (
	MFAInitiated = T{
		Subjects: []string{
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

	PasswordResetInitiated = T{
		Subjects: []string{
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

	LoginInitiated = T{
		Subjects: []string{
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

	RegisterInitiated = T{
		Subjects: []string{
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

	NewLogin = T{
		Subjects: []string{
			"New device signed in to your account",
			"A new device just accessed your account",
			"New login",
			"Your account was accessed from a new device",
			"New device activity detected",
			"Was this you? New device login",
			"Security alert: new device sign-in",
			"New device connected to your account",
			"Account access from a new device",
		},
	}

	RegistrationCompleted = T{
		Subjects: []string{
			"Welcome! Your account has been created",
			"Your new account is ready",
			"Account successfully created",
			"Welcome to our platform",
			"Your account has been set up",
			"New account confirmation",
			"Thanks for signing up",
		},
	}

	PasswordChanged = T{
		Subjects: []string{
			"Your password was changed successfully",
			"Password updated",
			"Your account password has been updated",
			"Password changed successfully",
			"Security update: password changed",
			"Your new password is now set",
			"Your account is now secure",
			"Password successfully updated",
			"Your password has been changed",
		},
	}
)
