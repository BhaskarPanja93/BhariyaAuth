package login

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	Generators "BhariyaAuth/processors/generator"
	Logger "BhariyaAuth/processors/logs"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"

	AccountProcessor "BhariyaAuth/processors/account"
	ResponseProcessor "BhariyaAuth/processors/response"
	Step2Processor "BhariyaAuth/processors/step2"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"

	"fmt"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

type Step1FormT struct {
	MailAddress string `form:"mail_address"`
	RememberMe  bool   `form:"remember_me"`
}

type Step2FormT struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

const tokenType = "Login"

func Step2(ctx fiber.Ctx) error {
	form := new(Step2FormT)
	var SignInData TokenModels.SignInT
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		Logger.AccidentalFailure("[Login2] Decrypt Failed")
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	err := json.Unmarshal(data, &SignInData)
	if err != nil {
		Logger.AccidentalFailure("[Login2] Unmarshal Failed")
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to read token (Encryptor issue)... Retrying"},
			})
	}
	if SignInData.TokenType != tokenType {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if SignInData.Step2Process == "password" && !AccountProcessor.CheckPasswordMatches(SignInData.UserID, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("[Login2] Incorrect Password for [UID-%d]", SignInData.UserID))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Incorrect Password"},
			})
	}
	if SignInData.Step2Process == "otp" && !Step2Processor.ValidateMailOTP(SignInData.Step2Code, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("[Login2] Incorrect OTP for [UID-%d]", SignInData.UserID))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Incorrect OTP"},
			})
	}
	if AccountProcessor.CheckUserIsBlacklisted(SignInData.UserID) {
		Logger.IntentionalFailure(fmt.Sprintf("[Login2] Blacklisted account [UID-%d] attempted login", SignInData.UserID))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Your account is disabled, please contact support"},
			})
	}
	refreshID := Generators.RefreshID()
	if !AccountProcessor.RecordReturningUser(ctx.Get("User-Agent"), refreshID, SignInData.UserID, SignInData.RememberMe) {
		Logger.AccidentalFailure(fmt.Sprintf("[Login2] Record Returning failed for [UID-%d]", SignInData.UserID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to login (DB-write issue)... Retrying"},
			})
	}
	token, ok := TokenProcessor.CreateFreshToken(
		SignInData.UserID,
		refreshID,
		AccountProcessor.GetUserType(SignInData.UserID),
		SignInData.RememberMe,
		"email-login",
	)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Login2] CreateFreshToken failed for [UID-%d]", SignInData.UserID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to acquire session (Encryptor issue)... Retrying"},
			})
	}
	ResponseProcessor.AttachAuthCookies(ctx, token)
	Logger.Success(fmt.Sprintf("[Login2] Logged in: [UID-%d]", SignInData.UserID))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:       true,
			Reply:         token.AccessToken,
			Notifications: []string{"Logged In Successfully"},
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
	process := ctx.Params("process")
	SignInData := TokenModels.SignInT{
		UserID:       userID,
		TokenType:    tokenType,
		RememberMe:   form.RememberMe,
		Step2Process: process,
		Mail:         form.MailAddress,
	}
	if process == "otp" {
		verification, retry := Step2Processor.SendMailOTP(ctx, form.MailAddress)
		if verification == "" {
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Success:       false,
					Notifications: []string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
				})
		}
		SignInData.Step2Code = verification
	} else if process == "password" {
		if !AccountProcessor.CheckUserHasPassword(userID) {
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Success:       false,
					Notifications: []string{"Password has not been set", "Please use OTP/SSO to login"},
				})
		}
		SignInData.Step2Code = ""
	} else {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	data, err := json.Marshal(SignInData)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Login1] Marshal Failed for [UID-%d] reason: %s", userID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to acquire token (Parser issue)... Retrying"},
			})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Login1] Encrypt Failed for [UID-%d]", userID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Success:       false,
				Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
			})
	}
	Logger.Success(fmt.Sprintf("[Login1] Token Created for [UID-%d]", userID))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:       true,
			Reply:         token,
			Notifications: []string{fmt.Sprintf("Please enter the %s", process)},
		})
}
