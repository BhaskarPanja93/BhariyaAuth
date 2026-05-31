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

const batchSize = 500

func sendMail(mails []string, subject, content string, attempts uint8) error {
	if attempts == 0 {
		return errors.New("send mail: retries exhausted")
	}

	if client == nil {
		refreshCredentials()
		if client == nil {
			time.Sleep(time.Second)
			return sendMail(mails, subject, content, attempts-1)
		}
	}

	body := graphmodels.NewItemBody()
	body.SetContentType(new(graphmodels.HTML_BODYTYPE))
	body.SetContent(&content)

	message := graphmodels.NewMessage()
	message.SetSubject(&subject)
	message.SetBody(body)

	if len(mails) <= batchSize {
		recipients := make([]graphmodels.Recipientable, 0, len(mails))

		for _, mail := range mails {
			emailAddress := graphmodels.NewEmailAddress()
			emailAddress.SetAddress(&mail)

			recipient := graphmodels.NewRecipient()
			recipient.SetEmailAddress(emailAddress)

			recipients = append(recipients, recipient)
		}

		message.SetToRecipients(recipients)

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
		return sendMail(mails, subject, content, attempts-1)
	}

	for i := 0; i < len(mails); i += batchSize {
		end := i + batchSize
		if end > len(mails) {
			end = len(mails)
		}
		go func() { _ = sendMail(mails[i:end], subject, content, attempts) }()
	}
	return nil
}
