package mail

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	MailTemplates "BhariyaAuth/models/mails/templates"
	Logger "BhariyaAuth/processors/logs"
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	graph "github.com/microsoftgraph/msgraph-sdk-go"
	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
)

var (
	cred        *azidentity.ClientSecretCredential
	graphClient *graph.GraphServiceClient
)

func init() {
	cred, _ = azidentity.NewClientSecretCredential(Secrets.MicrosoftTenantId, Secrets.MicrosoftClientId, Secrets.MicrosoftClientSecret, nil)
	refreshCredentials()
}

func refreshCredentials() {
	graphClient, _ = graph.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
}
func sendMail(mail, subject, content string, attempts uint8) bool {
	if attempts <= 0 {
		return false
	}

	body := graphmodels.NewItemBody()
	contentType := graphmodels.HTML_BODYTYPE
	body.SetContentType(&contentType)
	body.SetContent(&content)

	message := graphmodels.NewMessage()
	message.SetSubject(&subject)
	message.SetBody(body)

	recipient := graphmodels.NewRecipient()
	emailAddress := graphmodels.NewEmailAddress()
	emailAddress.SetAddress(&mail)
	recipient.SetEmailAddress(emailAddress)

	message.SetToRecipients([]graphmodels.Recipientable{recipient})

	requestBody := graphusers.NewItemSendMailPostRequestBody()
	requestBody.SetMessage(message)
	saveToSentItems := false
	requestBody.SetSaveToSentItems(&saveToSentItems)

	err := graphClient.Users().ByUserId(Secrets.MicrosoftMailId).SendMail().Post(context.Background(), requestBody, nil)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Mail] SendMail failed for [MAIL-%s]: %s", mail, err.Error()))
		time.Sleep(time.Second)
		refreshCredentials()
		return sendMail(mail, subject, content, attempts-1)
	}
	return true
}

func OTP(mail, otp string, subject string, header string, ignorable bool, attempts uint8) bool {
	return sendMail(
		mail,
		subject,
		MailTemplates.OTP(Config.FrontendURL, header, otp, ignorable),
		attempts,
	)
}

func NewLogin(mail string, subject string, IP string, OS string, device string, browser string, attempts uint8) bool {
	return sendMail(
		mail,
		subject,
		MailTemplates.NewLogin(Config.FrontendURL, OS, device, browser, IP),
		attempts)
}

func NewAccount(mail string, name string, subject string, IP string, OS string, device string, browser string, attempts uint8) bool {
	return sendMail(
		mail,
		subject,
		MailTemplates.NewAccount(Config.FrontendURL, name, OS, device, browser, IP),
		attempts)
}

func PasswordReset(mail string, subject string, IP string, OS string, device string, browser string, attempts uint8) bool {
	return sendMail(
		mail,
		subject,
		MailTemplates.PasswordReset(Config.FrontendURL, OS, device, browser, IP),
		attempts)
}

func AccountBlacklisted(mail string, attempts uint8) bool {
	content := `Your account has been flagged. All future actions will be blocked. Contact support ASAP if you think this is a mistake.`
	return sendMail(mail, "Account Blacklisted", content, attempts)
}
