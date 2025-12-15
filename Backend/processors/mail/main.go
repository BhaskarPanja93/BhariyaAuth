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

func OTP(mail, subject string, heading string, otp string, ignorable bool, attempts uint8) bool {
	ignorableText := `
	<p style="margin: 0; font-size: 12px; color: #6b7280;">you can safely ignore this message</p>`
	if !ignorable {
		ignorableText = `<a href="https://bhariya.ddns.net/auth/passwordreset" style="margin: 0; font-size: 14px; color: #5865f2;" target="_blank"><b>change your password immediately</b></a>`
	}
	content := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8" />
    <title>BhariyaAuth OTP</title>
</head>
<body style="margin:0; padding:0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;">
<table width="100%%" cellpadding="0" cellspacing="0">
    <tr>
        <td align="center" style="padding: 40px 16px;">
            <table width="100%%" cellpadding="0" cellspacing="0" style="max-width: 520px; background-color: #0f1115; border-radius: 14px; box-shadow: 0 20px 40px rgba(0,0,0,0.5); overflow: hidden;">
                <tr>
                    <td align="center" style="padding: 32px 24px 16px;  background: linear-gradient(135deg, #4b5563, #1a1c20, #0b0d10);">
                        <img src="https://bhariya.ddns.net/auth/favicon-dark-mode.png" alt="Bhariya" width="120" style="display:block; margin-bottom: 12px;" />
                    </td>
                </tr>
                <tr>
                    <td style="padding: 28px 24px; text-align: center; background-color: #0f0f10">
                        <p style="margin: 0 0 12px; font-size: 15px; color: #cbd5f5;">
                            %s
                        </p>
                        <div style="margin: 24px auto; padding: 16px 24px; display: inline-block; background: linear-gradient(to right, #8b5cf6, #7c3aed); color: #ffffff; font-size: 28px; letter-spacing: 6px; font-weight: 700; border-radius: 10px; ">
                            %s
                        </div>
                        <p style="margin: 20px 0 0; font-size: 14px; color: #9ca3af;">
                            This OTP is valid for <strong>5 minutes</strong>.
                        </p>
                    </td>
                </tr>
                <tr>
                    <td style="padding: 20px 24px; text-align: center; background-color: #202227">
                        <p style="margin: 0; font-size: 12px; color: #6b7280;">
                            If you didn’t request this,
                        </p>
                        %s
                    </td>
                </tr>
            </table>
        </td>
    </tr>
</table>
</body>
</html>
`, heading, otp, ignorableText)
	return sendMail(mail, subject, content, attempts)
}

func NewLogin(mail string, attempts uint8) bool {
	subject := "New login"
	content := "A new device has logged in to your account."
	return sendMail(mail, subject, content, attempts)
}

func NewAccount(mail string, attempts uint8) bool {
	subject := "Welcome"
	content := "Your account for BhariyaAuth has been created. You will be using this account credentials for logging in to all our services."
	return sendMail(mail, subject, content, attempts)
}

func PasswordChange(mail string, attempts uint8) bool {
	subject := "Password Changed"
	content := "Your account password has been changed. Contact support if you think this is a mistake."
	return sendMail(mail, subject, content, attempts)
}

func AccountBlacklisted(mail string, attempts uint8) bool {
	subject := "Blacklisted"
	content := "Your account has been flagged. All future actions will be blocked. Contact support ASAP if you think this is a mistake."
	return sendMail(mail, subject, content, attempts)
}
