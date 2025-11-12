package passwordreset

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	AccountProcessor "BhariyaAuth/processors/account"
	Logger "BhariyaAuth/processors/logs"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	Step2Processor "BhariyaAuth/processors/step2"
	StringProcessor "BhariyaAuth/processors/string"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

type Step1FormT struct {
	MailAddress string `form:"mail_address"`
}

type Step2FormT struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
	NewPassword  string `form:"new_password"`
}

const tokenType = "PasswordReset"

func Step2(ctx fiber.Ctx) error {
	form := new(Step2FormT)
	var ResetData TokenModels.SignInT
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to read token (Encryptor issue)... Retrying"},
			})
	}
	if ResetData.TokenType != tokenType {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if !Step2Processor.ValidateMailOTP(ResetData.Step2Code, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("[Reset2] Incorrect OTP for [UID-%d]", ResetData.UserID))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Incorrect OTP"},
			})
	}
	if AccountProcessor.CheckUserIsBlacklisted(ResetData.UserID) {
		Logger.IntentionalFailure(fmt.Sprintf("[Reset2] Blacklisted account [UID-%d] attempted login", ResetData.UserID))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Your account is disabled, please contact support"},
			})
	}
	if !AccountProcessor.UpdatePassword(ResetData.UserID, form.NewPassword) {
		Logger.AccidentalFailure(fmt.Sprintf("[Reset2] Update failed for [UID-%d]", ResetData.UserID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to reset (DB-write issue)... Retrying"},
			})
	}
	Logger.Success(fmt.Sprintf("[Reset2] Password changed for [UID-%d]", ResetData.UserID))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:       true,
			Reply:         true,
			Notifications: []string{"Password changed successfully"},
		})
}

func Step1(ctx fiber.Ctx) error {
	form := new(Step1FormT)
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	if !StringProcessor.IsValidEmail(form.MailAddress) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	userID, found := AccountProcessor.GetIDFromMail(form.MailAddress)
	if !found {
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Account doesn't exist with the email"},
			})
	}
	PasswordReset := TokenModels.PasswordResetT{
		TokenType: tokenType,
		UserID:    userID,
		Step2Code: "",
	}
	verification, retry := Step2Processor.SendMailOTP(ctx, form.MailAddress)
	if verification == "" {
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
			})
	}
	PasswordReset.Step2Code = verification
	data, err := json.Marshal(PasswordReset)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Reset1] Marshal Failed for [UID-%d] reason: %s", userID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
			})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Reset1] Encrypt Failed for [UID-%d]", userID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
			})
	}
	Logger.Success(fmt.Sprintf("[Reset1] Token Created for [UID-%d]", userID))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:       true,
			Reply:         token,
			Notifications: []string{"Please enter the OTP sent to your mail"},
		})
}
