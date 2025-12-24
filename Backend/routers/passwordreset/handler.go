package passwordreset

import (
	FormModels "BhariyaAuth/models/forms"
	MailModels "BhariyaAuth/models/mails"
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	AccountProcessor "BhariyaAuth/processors/account"
	Logger "BhariyaAuth/processors/logs"
	MailNotifier "BhariyaAuth/processors/mail"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	"fmt"
	"math/rand"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

const tokenType = "PasswordReset"

func Step1(ctx fiber.Ctx) error {
	form := new(FormModels.PasswordResetForm1)
	if err := ctx.Bind().Form(form); err != nil {
		if err = ctx.Bind().Body(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	if !StringProcessor.EmailIsValid(form.MailAddress) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	userID, found := AccountProcessor.GetIDFromMail(form.MailAddress)
	if !found {
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Account doesn't exist with the email"},
		})
	}
	PasswordReset := TokenModels.PasswordResetT{
		TokenType: tokenType,
		Mail:      form.MailAddress,
		UserID:    userID,
		Step2Code: "",
	}
	mailModel := MailModels.PasswordResetInitiated
	verification, retry := OTPProcessor.Send(form.MailAddress, mailModel.Subjects[rand.Intn(len(mailModel.Subjects))], mailModel.Header, mailModel.Ignorable, ctx.IP())
	if verification == "" {
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Reply:         retry.Seconds(),
			Notifications: []string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
		})
	}
	PasswordReset.Step2Code = verification
	data, err := json.Marshal(PasswordReset)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Reset1] Marshal Failed for [UID-%d] reason: %s", userID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Reset1] Encrypt Failed for [UID-%d]", userID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	Logger.Success(fmt.Sprintf("[Reset1] Token Created for [UID-%d]", userID))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   token,
	})
}

func Step2(ctx fiber.Ctx) error {
	form := new(FormModels.PasswordResetForm2)
	var ResetData TokenModels.PasswordResetT
	if err := ctx.Bind().Form(form); err != nil {
		if err = ctx.Bind().Body(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		Logger.AccidentalFailure("[Reset2] Decrypt Failed")
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	err := json.Unmarshal(data, &ResetData)
	if err != nil {
		Logger.AccidentalFailure("[Reset2] Unmarshal Failed")
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to read token (Encryptor issue)... Retrying"},
		})
	}
	if ResetData.TokenType != tokenType {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if !OTPProcessor.Validate(ResetData.Step2Code, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("[Reset2] Incorrect OTP for [UID-%d]", ResetData.UserID))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Incorrect OTP"},
		})
	}
	if AccountProcessor.CheckUserIsBlacklisted(ResetData.UserID) {
		Logger.IntentionalFailure(fmt.Sprintf("[Reset2] Blacklisted account [UID-%d] attempted login", ResetData.UserID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Your account is disabled, please contact support"},
		})
	}
	if !AccountProcessor.UpdatePassword(ResetData.UserID, form.NewPassword) {
		Logger.AccidentalFailure(fmt.Sprintf("[Reset2] Update failed for [UID-%d]", ResetData.UserID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to reset (DB-write issue)... Retrying"},
		})
	}
	Logger.Success(fmt.Sprintf("[Reset2] Password changed for [UID-%d]", ResetData.UserID))
	UA := StringProcessor.UAParser.Parse(ctx.Get("User-Agent"))
	browser := UA.Browser().String()
	if browser == "" {
		browser = "Unknown"
	}
	device := UA.Device().String()
	if device == "" {
		device = "Unknown"
	}
	mailModel := MailModels.PasswordChanged
	MailNotifier.PasswordReset(ResetData.Mail, mailModel.Subjects[rand.Intn(len(mailModel.Subjects))], ctx.IP(), device, browser, 2)
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
	})
}
