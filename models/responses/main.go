package responses

type GeneralT struct {
	Short   string `json:"short"`
	Long    string `json:"long"`
	Allowed bool   `json:"allowed"`
}

type AuthT struct {
	Allowed bool   `json:"allowed"`
	Change  bool   `json:"change"`
	Token   string `json:"token"`
}

type CombinedT struct {
	Auth          AuthT       `json:"auth"`
	Reply         interface{} `json:"reply"`
	Notifications []string    `json:"notifications"`
	Secret        interface{} `json:"secret"`
	Extra         interface{} `json:"extra"`
}

var (
	RefreshSucceeded = GeneralT{
		Short:   "REFRESH_SUCCEEDED",
		Long:    "Token refreshed",
		Allowed: true,
	}
	SignInIDVerified = GeneralT{
		Short:   "SIGN_IN_ID_VERIFIED",
		Long:    "Sign In ID Verified",
		Allowed: true,
	}
	SignedIn = GeneralT{
		Short:   "SIGNED_IN",
		Long:    "Sign In Successful",
		Allowed: true,
	}
	SignUpIDVerified = GeneralT{
		Short:   "SIGN_UP_ID_VERIFIED",
		Long:    "Sign Up ID Verified",
		Allowed: true,
	}
	SignedUp = GeneralT{
		Short:   "SIGNED_UP",
		Long:    "Sign Up Successful",
		Allowed: true,
	}
	PasswordResetIDVerified = GeneralT{
		Short:   "FORGOT_ID_VERIFIED",
		Long:    "Forgot ID Verified",
		Allowed: true,
	}
	SignedOut = GeneralT{
		Short:   "LOGGED_OUT",
		Long:    "Sign Out Successful",
		Allowed: true,
	}
	PasswordUpdated = GeneralT{
		Short:   "PASSWORD_UPDATED",
		Long:    "Password updated",
		Allowed: true,
	}
	PasswordNotUpdated = GeneralT{
		Short:   "PASSWORD_NOT_UPDATED",
		Long:    "Password not updated",
		Allowed: true,
	}
	InvalidCredentials = GeneralT{
		Short:   "INVALID_CREDENTIALS",
		Long:    "Invalid credentials",
		Allowed: false,
	}
	UserBlocked = GeneralT{
		Short:   "USER_BLOCKED",
		Long:    "User blocked, contact admin for more details",
		Allowed: false,
	}
	RefreshBlocked = GeneralT{
		Short:   "REFRESH_BLOCKED",
		Long:    "Refresh blocked, please login again",
		Allowed: false,
	}
	RefreshFailed = GeneralT{
		Short:   "REFRESH_FAILED",
		Long:    "Token refresh failed. Continue with Sign in",
		Allowed: false,
	}
	EmailAlreadyTaken = GeneralT{
		Short:   "EMAIL_ALREADY_TAKEN",
		Long:    "Email already has an account",
		Allowed: false,
	}
	EmailDoesntExist = GeneralT{
		Short:   "EMAIL_DOESNT_EXIST",
		Long:    "Email doesn't have an account",
		Allowed: false,
	}
	AccountDoesntExist = GeneralT{
		Short:   "ACCOUNT_DOESNT_EXIST",
		Long:    "Account not found",
		Allowed: false,
	}
	PasswordNotRegistered = GeneralT{
		Short:   "PASSWORD_NOT_REGISTERED",
		Long:    "Password not registered. Try SSO login.",
		Allowed: false,
	}
	PasswordTooSimple = GeneralT{
		Short:   "PASSWORD_TOO_SIMPLE",
		Long:    "Password too simple",
		Allowed: false,
	}
	InvalidOTP = GeneralT{
		Short:   "INVALID_OTP",
		Long:    "OTP doesn't match. Try again or request a new OTP.",
		Allowed: false,
	}
	InvalidEntries = GeneralT{
		Short:   "INVALID_ENTRIES",
		Long:    "Check entries",
		Allowed: false,
	}
	InvalidToken = GeneralT{
		Short:   "INVALID_TOKEN",
		Long:    "Invalid token. Try again or request a new token.",
		Allowed: false,
	}
	OtpSendFailed = GeneralT{
		Short:   "OTP_SEND_FAILED",
		Long:    "OTP send failed. Try again after some time.",
		Allowed: false,
	}
	Unknown = GeneralT{
		Short:   "UNKNOWN",
		Long:    "Some error occurred",
		Allowed: false,
	}
	RateLimited = GeneralT{
		Short:   "RATE_LIMITED",
		Long:    "Rate limited",
		Allowed: false,
	}

	DefaultAuth = AuthT{
		Allowed: true,
		Change:  false,
		Token:   "",
	}
)
