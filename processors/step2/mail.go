package step2

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	Generators "BhariyaAuth/processors/generator"
	Logger "BhariyaAuth/processors/logs"
	Stores "BhariyaAuth/stores"
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/gofiber/fiber/v3"
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
	_RefreshCredentials()
}

func _RefreshCredentials() {
	graphClient, _ = graph.NewGraphServiceClientWithCredentials(cred, []string{"https://graph.microsoft.com/.default"})
}

func sendInternal(mail string, otp string, trial uint8) bool {
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
		Logger.AccidentalFailure(fmt.Sprintf("OTP send failed: %s", err.Error()))
		time.Sleep(time.Second)
		_RefreshCredentials()
		return sendInternal(mail, otp, trial+1)
	}
	return true
}

func SendMailOTP(ctx fiber.Ctx, mail string) (string, time.Duration) {
	rateLimitKey := fmt.Sprintf("%s:%s", ctx.IP(), mail)
	canSend, alreadySentCount, currentDelay := CheckCanSendOTP(rateLimitKey)
	if canSend {
		otp := Generators.SafeString(4)
		if success := sendInternal(mail, otp, 0); !success {
			return "", currentDelay
		}
		verification := Generators.UnsafeString(10)
		key := fmt.Sprintf("%s:%s", Config.RedisServerOTPVerification, verification)
		Stores.RedisClient.Set(Stores.Ctx, key, otp, 5*time.Minute)
		currentDelay = RecordSendOTP(rateLimitKey, alreadySentCount+1)
		return verification, currentDelay
	} else {
		return "", currentDelay
	}
}

func ValidateMailOTP(verification, otp string) bool {
	key := fmt.Sprintf("%s:%s", Config.RedisServerOTPVerification, verification)
	value, _ := Stores.RedisClient.Get(Stores.Ctx, key).Result()
	if value == otp && otp != "" {
		Stores.RedisClient.Del(Stores.Ctx, key)
		return true
	}
	return false
}
