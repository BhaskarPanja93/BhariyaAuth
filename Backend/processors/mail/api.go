package mail

import (
	Config "BhariyaAuth/constants/config"
	HTMLTemplates "BhariyaAuth/models/html"
	MailModels "BhariyaAuth/models/mails"
	"math/rand"
)

func OTP(mail, otp string, model MailModels.T) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.OTP(
			Config.FrontendRoute,
			model.Header,
			otp,
			model.Ignorable,
		),
		3,
	)
}

func SignIn(mail string, model MailModels.T, IP, OS, device, browser string) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.NewLogin(Config.FrontendRoute, OS, device, browser, IP),
		5,
	)
}

func SignUp(mail, name string, model MailModels.T, IP, OS, device, browser string) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.NewAccount(Config.FrontendRoute, name, OS, device, browser, IP),
		3,
	)
}

func PasswordReset(mail string, model MailModels.T, IP, OS, device, browser string) error {
	return sendMail(
		mail,
		model.Subjects[rand.Intn(len(model.Subjects))],
		HTMLTemplates.PasswordReset(Config.FrontendRoute, OS, device, browser, IP),
		5,
	)
}

func AccountBlacklisted(mail string) error {
	content := `Your account has been flagged. All future actions will be blocked. Contact support ASAP if you think this was a mistake.`

	return sendMail(mail, "Account Blacklisted", content, 5)
}

func Raw(mails []string, subject string, content string) {
	for _, mail := range mails {
		go func() {
			_ = sendMail(mail, subject, content, 3)
		}()
	}
}
