package passwordreset

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	AccountProcessor "BhariyaAuth/processors/account"
	Logger "BhariyaAuth/processors/logs"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	ResponseProcessor "BhariyaAuth/processors/response"
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

func Step2(ctx fiber.Ctx) error {
	form := new(Step2FormT)
	var PasswordResetData TokenModels.PasswordResetT
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			RateLimitProcessor.SetValue(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		RateLimitProcessor.SetValue(ctx)
		Logger.AccidentalFailure("Reset-2 Decrypt Failed")
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.Unknown,
			ResponseModels.DefaultAuth,
			[]string{"Failed to unmarshal SignIn data, please contact support"},
			nil,
			nil,
		))
	}
	err := json.Unmarshal(data, &PasswordResetData)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Reset-2 Unmarshal Failed: %s", err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.Unknown,
			ResponseModels.DefaultAuth,
			[]string{"Failed to unmarshal SignIn data, please contact support"},
			nil,
			nil,
		))
	}
	if PasswordResetData.TokenType != "PasswordReset" {
		RateLimitProcessor.SetValue(ctx)
		Logger.IntentionalFailure("Reset-2 Token not for SignUp")
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.InvalidToken,
			ResponseModels.DefaultAuth,
			[]string{"This token is not applicable for Password Reset"},
			nil,
			nil,
		))
	}
	if !Step2Processor.ValidateMailOTP(PasswordResetData.Step2Code, form.Verification) {
		RateLimitProcessor.SetValue(ctx)
		Logger.IntentionalFailure(fmt.Sprintf("Reset-2 Incorrect OTP [%d]", PasswordResetData.UserID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.InvalidOTP,
			ResponseModels.DefaultAuth,
			[]string{"Incorrect OTP"},
			nil,
			nil,
		))
	}
	if !AccountProcessor.IDExists(PasswordResetData.UserID) {
		RateLimitProcessor.SetValue(ctx)
		Logger.IntentionalFailure(fmt.Sprintf("Reset-2 Account not found [%d]", PasswordResetData.UserID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.AccountDoesntExist,
			ResponseModels.DefaultAuth,
			[]string{"Account doesn't exist"},
			nil,
			nil,
		))
	}
	if !AccountProcessor.UpdatePassword(PasswordResetData.UserID, form.NewPassword) {
		Logger.AccidentalFailure(fmt.Sprintf("Reset-2 Update failed [%d]", PasswordResetData.UserID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.PasswordNotUpdated,
			ResponseModels.DefaultAuth,
			[]string{"Password not changed"},
			nil,
			nil,
		))
	}
	Logger.Success(fmt.Sprintf("Reset-2 Password changed [%d]", PasswordResetData.UserID))
	return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
		ResponseModels.PasswordUpdated,
		ResponseModels.DefaultAuth,
		[]string{"Password changed"},
		nil,
		nil,
	))
}

func Step1(ctx fiber.Ctx) error {
	form := new(Step1FormT)
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			RateLimitProcessor.SetValue(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	if !StringProcessor.IsValidEmail(form.MailAddress) {
		Logger.IntentionalFailure(fmt.Sprintf("Reset-1 Incorrect Mail: %s", form.MailAddress))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidEntries,
				ResponseModels.DefaultAuth,
				[]string{"Please enter a valid email address"},
				nil,
				nil,
			))
	}
	userID, found := AccountProcessor.GetIDFromMail(form.MailAddress)
	if !found {
		Logger.IntentionalFailure(fmt.Sprintf("Reset-1 Account not found: %s", form.MailAddress))
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.EmailDoesntExist,
				ResponseModels.DefaultAuth,
				[]string{"Account doesn't exist with the email"},
				nil,
				nil,
			))
	}
	PasswordReset := TokenModels.PasswordResetT{
		TokenType: "PasswordReset",
		UserID:    userID,
		Step2Code: "",
	}
	verification, retry := Step2Processor.SendMailOTP(ctx, form.MailAddress)
	if verification == "" {
		Logger.AccidentalFailure(fmt.Sprintf("Reset-1 OTP failed: %s", form.MailAddress))
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.OtpSendFailed,
			ResponseModels.DefaultAuth,
			[]string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
			nil,
			map[string]interface{}{"resend-after": retry.Seconds()},
		))
	}
	PasswordReset.Step2Code = verification
	data, err := json.Marshal(PasswordReset)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Reset-1 Marshal Failed: %s", err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.Unknown,
			ResponseModels.DefaultAuth,
			[]string{"Failed to marshal PasswordReset data, please contact support"},
			nil,
			nil,
		))
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("Reset-1 Encrypt Failed: %s", form.MailAddress))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.Unknown,
			ResponseModels.DefaultAuth,
			[]string{"Failed to encrypt PasswordReset data, please contact support"},
			nil,
			nil,
		))
	}
	Logger.Success(fmt.Sprintf("Reset-1 Token Created: %s", form.MailAddress))
	return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
		ResponseModels.PasswordResetIDVerified,
		ResponseModels.DefaultAuth,
		[]string{"Please enter the OTP sent to your mail"},
		map[string]interface{}{"token": token},
		nil,
	))
}
