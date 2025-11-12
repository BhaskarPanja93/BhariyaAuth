package step2

import (
	Config "BhariyaAuth/constants/config"
	Generators "BhariyaAuth/processors/generator"
	MailProcessor "BhariyaAuth/processors/mail"
	Stores "BhariyaAuth/stores"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

func SendMailOTP(ctx fiber.Ctx, mail string) (string, time.Duration) {
	rateLimitKey := fmt.Sprintf("%s:%s", ctx.IP(), mail)
	canSend, alreadySentCount, currentDelay := CheckCanSendOTP(rateLimitKey)
	if canSend {
		otp := Generators.SafeString(4)
		if success := MailProcessor.SendOTP(mail, otp, 0); !success {
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
