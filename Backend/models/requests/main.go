package requests

// All form types

type PasswordResetForm1 struct {
	Mail string `form:"mail"`
}

type PasswordResetForm2 struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
	Password     string `form:"password"`
}

type SignInForm1 struct {
	Mail     string `form:"mail"`
	Remember string `form:"remember"`
}

type SignInForm2 struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

type SignUpForm1 struct {
	Mail     string `form:"mail"`
	Name     string `form:"name"`
	Password string `form:"password"`
	Remember string `form:"remember"`
}

type SignUpForm2 struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

type DeviceRevokeForm struct {
	User    string `form:"uid"`
	All     string `form:"all"`
	Refresh string `form:"refresh"`
}

type MFAForm struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}
