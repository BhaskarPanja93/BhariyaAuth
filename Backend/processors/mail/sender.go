package mail

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	"time"

	graphmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
	graphusers "github.com/microsoftgraph/msgraph-sdk-go/users"
)

// sendMail sends an email using Microsoft Graph API with retry logic.
//
// - Constructs email message (HTML).
// - Sends via Graph API.
// - Retries on failure with credential refresh.
//
// Flow Summary:
//
//	build message → send → retry (on failure)
//
// Parameters:
// - mail: recipient email address.
// - subject: email subject.
// - content: HTML content.
// - attempts: number of retry attempts remaining.
//
// Returns:
// - true if mail sent successfully.
// - false if all retries fail.
//
// Retry Strategy:
// - Fixed delay (5 seconds).
// - Recursive retry until attempts exhausted.
//
// Security Considerations:
// - Uses application-level credentials (client secret).
// - Does not expose sensitive data externally.
func sendMail(mail, subject, content string, attempts uint8) bool {

	// Stop retrying if attempts exhausted
	if attempts <= 0 {
		return false
	}

	// Construct email body (HTML)
	body := graphmodels.NewItemBody()
	body.SetContentType(new(graphmodels.HTML_BODYTYPE))
	body.SetContent(&content)

	// Create message object
	message := graphmodels.NewMessage()
	message.SetSubject(&subject)
	message.SetBody(body)

	// Configure recipient
	recipient := graphmodels.NewRecipient()
	emailAddress := graphmodels.NewEmailAddress()
	emailAddress.SetAddress(&mail)
	recipient.SetEmailAddress(emailAddress)

	message.SetToRecipients([]graphmodels.Recipientable{recipient})

	// Build request payload
	requestBody := graphusers.NewItemSendMailPostRequestBody()
	requestBody.SetMessage(message)
	requestBody.SetSaveToSentItems(new(true))

	// Send email via Graph API
	err := client.
		Users().
		ByUserId(Secrets.MicrosoftMailId).
		SendMail().
		Post(Config.CtxBG, requestBody, nil)

	if err != nil {
		// On failure:
		// - wait
		// - refresh credentials
		// - retry
		time.Sleep(time.Second)
		refreshCredentials()

		return sendMail(mail, subject, content, attempts-1)
	}

	return true
}
