package mail

import (
	Secrets "BhariyaAuth/constants/secrets"
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
func sendMail(mail, subject, content string, trial uint8) bool {
	if trial > 2 {
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
		return sendMail(mail, subject, content, trial+1)
	}
	return true
}

func OTP(mail, otp string, trial uint8) bool {
	subject := "BhariyaAuth OTP"
	content := fmt.Sprintf("Your OTP (valid for 5 Minutes) for BhariyaAuth is: <b>%s</b>", otp)
	return sendMail(mail, subject, content, trial)
}

func NewLogin(mail string, trial uint8) bool {
	subject := "New login"
	content := "A new device has logged in to your account."
	return sendMail(mail, subject, content, trial)
}

func NewAccount(mail string, trial uint8) bool {
	subject := "Welcome"
	content := "Your account for BhariyaAuth has been created. You will be using this account credentials for logging in to all our services."
	return sendMail(mail, subject, content, trial)
}

func PasswordChange(mail string, trial uint8) bool {
	subject := "Password Changed"
	content := "Your account password has been changed."
	return sendMail(mail, subject, content, trial)
}
