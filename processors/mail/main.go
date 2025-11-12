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

func SendOTP(mail string, otp string, trial uint8) bool {
	if trial > 2 {
		return false
	}

	subject := "BhariyaAuth OTP"
	content := fmt.Sprintf("Your OTP (valid for 5 Minutes) for BhariyaAuth is: <b>%s</b>", otp)

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
		Logger.AccidentalFailure(fmt.Sprintf("[Step2Mail] sendInternal failed for [MAIL%s]: %s", mail, err.Error()))
		time.Sleep(time.Second)
		refreshCredentials()
		return SendOTP(mail, otp, trial+1)
	}
	return true
}

func SendText(mail string, text string, trial uint8) bool {

	return false
}

func SendHTML(mail string, html string, trial uint8) bool {

	return false
}
