package login

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UserModels "BhariyaAuth/models/users"
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

func Step2(ctx fiber.Ctx) error {
	form := new(Step2FormT)
	var SignInData TokenModels.SignInT
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			RateLimitProcessor.SetValue(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		Logger.IntentionalFailure("Login-2 Decrypt Failed")
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidToken,
				ResponseModels.DefaultAuth,
				[]string{"Failed to decrypt SignIn data, please contact support"},
				nil,
				nil,
			))
	}
	err := json.Unmarshal(data, &SignInData)
	if err != nil {
		Logger.AccidentalFailure("Login-2 Unmarshal Failed")
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to unmarshal SignIn data, please contact support"},
				nil,
				nil,
			))
	}
	if SignInData.TokenType != "SignIn" {
		Logger.IntentionalFailure("Login-2 Token not for Login")
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidToken,
				ResponseModels.DefaultAuth,
				[]string{"This token is not applicable for SignIn"},
				nil,
				nil,
			))
	}
	if SignInData.Step2Process == "password" && !AccountProcessor.PasswordMatches(SignInData.UserID, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("Login-2 Incorrect Password [%d-%s]", SignInData.UserID, SignInData.Mail))
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidCredentials,
				ResponseModels.DefaultAuth,
				[]string{"Incorrect Password"},
				nil,
				nil,
			))
	}
	if SignInData.Step2Process == "otp" && !Step2Processor.ValidateMailOTP(SignInData.Step2Code, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("Login-2 Incorrect OTP [%d-%s]", SignInData.UserID, SignInData.Mail))
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidOTP,
				ResponseModels.DefaultAuth,
				[]string{"Incorrect OTP"},
				nil,
				nil,
			))
	}
	if AccountProcessor.UserIsBlacklisted(SignInData.UserID) {
		Logger.IntentionalFailure(fmt.Sprintf("Login-2 Blacklisted account [%d-%s] attempted login", SignInData.UserID, SignInData.Mail))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.UserBlocked,
				ResponseModels.DefaultAuth,
				[]string{"Your account is disabled, please contact support"},
				nil,
				nil,
			))
	}
	refreshID := Generators.RefreshID()
	if !AccountProcessor.RecordReturningUser(refreshID, SignInData.UserID, SignInData.RememberMe) {
		Logger.AccidentalFailure(fmt.Sprintf("Login-2 Record Returning failed: [%d-%s]", SignInData.UserID, SignInData.Mail))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to login, please try again or contact support"},
				nil,
				nil,
			))
	}
	token := TokenProcessor.CreateFreshToken(
		SignInData.UserID,
		refreshID,
		UserModels.Find(AccountProcessor.GetUserType(SignInData.UserID)),
		SignInData.RememberMe,
		"email-login",
	)
	ResponseProcessor.AttachAuthCookies(ctx, token)
	Logger.Success(fmt.Sprintf("Login-2 Logged in: [%d-%s]", SignInData.UserID, SignInData.Mail))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseProcessor.CombineResponses(
			ResponseModels.SignedIn,
			ResponseModels.AuthT{
				Allowed: true,
				Change:  true,
				Token:   token.AccessToken,
			},
			[]string{"Logged In Successfully"},
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
		Logger.IntentionalFailure(fmt.Sprintf("Login-1 Invalid email: %s", form.MailAddress))
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
		RateLimitProcessor.SetValue(ctx)
		Logger.IntentionalFailure(fmt.Sprintf("Login-1 Account not found: %s", form.MailAddress))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.EmailDoesntExist,
				ResponseModels.DefaultAuth,
				[]string{"Account doesn't exist with the email"},
				nil,
				nil,
			))
	}
	process := ctx.Params("process")
	SignInData := TokenModels.SignInT{
		UserID:       userID,
		TokenType:    "SignIn",
		RememberMe:   form.RememberMe,
		Step2Process: process,
		Step2Code:    "",
		Mail:         form.MailAddress,
	}
	if process == "otp" {
		verification, retry := Step2Processor.SendMailOTP(ctx, form.MailAddress)
		if verification == "" {
			Logger.IntentionalFailure(fmt.Sprintf("Login-1 OTP failed: %s", form.MailAddress))
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseProcessor.CombineResponses(
					ResponseModels.OtpSendFailed,
					ResponseModels.DefaultAuth,
					[]string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
					nil,
					map[string]interface{}{"resend-after": retry.Seconds()},
				))
		}
		SignInData.Step2Code = verification
	} else if process == "password" {
		if !AccountProcessor.UserHasPassword(userID) {
			Logger.IntentionalFailure(fmt.Sprintf("Login-1 User doesnt have password: [%d-%s]", userID, form.MailAddress))
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseProcessor.CombineResponses(
					ResponseModels.PasswordNotRegistered,
					ResponseModels.DefaultAuth,
					[]string{"Password has not been set", "Please use OTP/SSO to login"},
					nil,
					nil,
				))
		}
		SignInData.Step2Code = ""
	} else {
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidEntries,
				ResponseModels.DefaultAuth,
				[]string{"Unknown process selected"},
				nil,
				nil,
			))
	}
	data, err := json.Marshal(SignInData)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Login-1 Marshal Failed: %s", err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to marshal SignIn data, please contact support"},
				nil,
				nil,
			))
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("Login-1 Excrypt Failed"))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to encrypt SignIn data, please contact support"},
				nil,
				nil,
			))
	}
	Logger.Success(fmt.Sprintf("Login-1 Token Created: %s", form.MailAddress))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseProcessor.CombineResponses(
			ResponseModels.SignInIDVerified,
			ResponseModels.DefaultAuth,
			[]string{fmt.Sprintf("Please enter the %s", process)},
			map[string]interface{}{"token": token},
			nil,
		))
}
