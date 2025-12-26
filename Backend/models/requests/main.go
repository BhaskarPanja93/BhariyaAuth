package requests

type PasswordResetForm1 struct {
	MailAddress string `form:"mail_address"`
}
type PasswordResetForm2 struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
	NewPassword  string `form:"new_password"`
}

type LoginForm1 struct {
	MailAddress string `form:"mail_address"`
	RememberMe  string `form:"remember_me"`
}
type LoginForm2 struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

type RegisterForm1 struct {
	MailAddress string `form:"mail_address"`
	Name        string `form:"name"`
	Password    string `form:"password"`
	RememberMe  string `form:"remember_me"`
}
type RegisterForm2 struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

type DeviceRevokeForm struct {
	UserID    string `form:"user_id"`
	RevokeAll string `form:"revoke_all"`
	DeviceID  string `form:"device_id"`
}

type MFAForm struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}
