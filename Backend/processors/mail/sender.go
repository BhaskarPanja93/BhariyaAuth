package mail

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	Logs "BhariyaAuth/processors/logs"
	"errors"
	"time"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
)

func sendMail(mail, subject, content string, attempts uint8) error {
	if attempts == 0 {
		return errors.New("send mail: retries exhausted")
	}

	if client == nil {
		refreshCredentials()
		if client == nil {
			time.Sleep(time.Second)
			return sendMail(mail, subject, content, attempts-1)
		}
	}

	body := graphmodels.NewItemBody()
	body.SetContentType(new(graphmodels.HTML_BODYTYPE))
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
	requestBody.SetSaveToSentItems(new(true))

	err := client.
		Users().
		ByUserId(Secrets.MicrosoftMailId).
		SendMail().
		Post(Config.CtxBG, requestBody, nil)

	if err == nil {
		return nil
	}

	Logs.RootLogger.Add(Logs.Error, "processors/mail/sender", "", "Send mail failed: "+err.Error())
	time.Sleep(time.Second)
	refreshCredentials()
	return sendMail(mail, subject, content, attempts-1)
}
