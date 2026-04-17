package mail

import (
	Config "BhariyaAuth/constants/config"
	HTMLTemplates "BhariyaAuth/models/html"
	MailModels "BhariyaAuth/models/mails"
	"math/rand"
)

// OTP sends a one-time password email.
//
// Randomly selects a subject and injects OTP into HTML template.
func OTP(mail, otp string, model MailModels.T, attempts uint8) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.OTP(
			Config.FrontendRoute,
			model.Header,
			otp,
			model.Ignorable,
		),
		attempts,
	)
}

// SignIn sends a login notification email (new device/login alert).
func SignIn(mail string, model MailModels.T, IP, OS, device, browser string, attempts uint8) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.NewLogin(Config.FrontendRoute, OS, device, browser, IP),
		attempts,
	)
}

// SignUp sends a new account creation notification email.
func SignUp(mail, name string, model MailModels.T, IP, OS, device, browser string, attempts uint8) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.NewAccount(Config.FrontendRoute, name, OS, device, browser, IP),
		attempts,
	)
}

// PasswordReset sends a password reset confirmation email.
func PasswordReset(mail string, model MailModels.T, IP, OS, device, browser string, attempts uint8) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.PasswordReset(Config.FrontendRoute, OS, device, browser, IP),
		attempts,
	)
}

// AccountBlacklisted sends a security alert when account is blocked.
func AccountBlacklisted(mail string, attempts uint8) error {
	content := `Your account has been flagged. All future actions will be blocked. Contact support ASAP if you think this is a mistake.`

	return sendMail(mail, "Account Blacklisted", content, attempts)
}
