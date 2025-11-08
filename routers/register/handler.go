package register

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	Generators "BhariyaAuth/processors/generator"
	Logger "BhariyaAuth/processors/logs"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
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
	Name        string `form:"name"`
	Password    string `form:"password"`
	RememberMe  bool   `form:"remember_me"`
}

type Step2FormT struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

func Step2(ctx fiber.Ctx) error {
	form := new(Step2FormT)
	var SignUpData TokenModels.SignUpT
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			RateLimitProcessor.SetValue(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		Logger.AccidentalFailure("Register-2 Decrypt Failed")
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidToken,
				ResponseModels.DefaultAuth,
				[]string{"Failed to decrypt SignUp data, please retry or contact support"},
				nil,
				nil,
			))
	}
	err := json.Unmarshal(data, &SignUpData)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Register-2 Unmarshal Failed: %s", err.Error()))
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidToken,
				ResponseModels.DefaultAuth,
				[]string{"Failed to unmarshal SignUp data, please retry or contact support"},
				nil,
				nil,
			))
	}
	if SignUpData.TokenType != "SignUp" {
		Logger.IntentionalFailure("Register-2 Token not for SignUp")
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidToken,
				ResponseModels.DefaultAuth,
				[]string{"This token is not applicable for SignUp"},
				nil,
				nil,
			))
	}
	userID, found := AccountProcessor.GetIDFromMail(SignUpData.Mail)
	if found {
		Logger.IntentionalFailure(fmt.Sprintf("Register-2 Account already exists: [%d-%s]", userID, SignUpData.Mail))
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.EmailAlreadyTaken,
				ResponseModels.DefaultAuth,
				[]string{"Account exists with the email"},
				nil,
				nil,
			))
	}
	if !Step2Processor.ValidateMailOTP(SignUpData.Step2Code, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("Register-2 Incorrect OTP: %s", SignUpData.Mail))
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
	userID = Generators.UserID()
	refreshID := Generators.RefreshID()
	if !AccountProcessor.RecordNewUser(userID, SignUpData.Password, SignUpData.Mail, SignUpData.Name) {
		Logger.AccidentalFailure(fmt.Sprintf("Register-2 RecordNew Failed: [%s]", SignUpData.Mail))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to create account, please contact support"},
				nil,
				nil,
			))
	}
	if !AccountProcessor.RecordReturningUser(refreshID, userID, SignUpData.RememberMe) {
		Logger.AccidentalFailure(fmt.Sprintf("Register-2 RecordReturning Failed: %s", SignUpData.Mail))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Account Created but failed to log you in with new account, please login manually"},
				nil,
				nil,
			))
	}
	token := TokenProcessor.CreateFreshToken(
		userID,
		refreshID,
		UserTypes.All.Viewer,
		SignUpData.RememberMe,
		"email-register",
	)
	ResponseProcessor.AttachAuthCookies(ctx, token)
	Logger.Success(fmt.Sprintf("Register-2 Created: [%d-%d-%s]", userID, refreshID, SignUpData.Mail))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseProcessor.CombineResponses(
			ResponseModels.SignedUp,
			ResponseModels.AuthT{
				Allowed: true,
				Change:  true,
				Token:   token.AccessToken,
			},
			[]string{"Account Created and Logged in Successfully"},
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
	if form.Name == "" {
		Logger.IntentionalFailure(fmt.Sprintf("Register-1 Incorrect Name: %s", form.Name))
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidEntries,
				ResponseModels.DefaultAuth,
				[]string{"Please enter your name"},
				nil,
				nil,
			))
	}
	if !StringProcessor.IsValidEmail(form.MailAddress) {
		Logger.IntentionalFailure(fmt.Sprintf("Register-1 Incorrect Email: %s", form.MailAddress))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidEntries,
				ResponseModels.DefaultAuth,
				[]string{"Please enter a valid email address"},
				nil,
				nil,
			))
	}
	if !AccountProcessor.PasswordIsStrong(form.Password) {
		Logger.IntentionalFailure(fmt.Sprintf("Register-1 Weak Password: %s", form.MailAddress))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.PasswordTooSimple,
				ResponseModels.DefaultAuth,
				[]string{"Please choose a stronger password"},
				nil,
				nil,
			))
	}
	userID, found := AccountProcessor.GetIDFromMail(form.MailAddress)
	if found {
		Logger.IntentionalFailure(fmt.Sprintf("Register-1 Account Exists: [%d-%s]", userID, form.Name))
		RateLimitProcessor.SetValue(ctx)
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.EmailAlreadyTaken,
				ResponseModels.DefaultAuth,
				[]string{"Account exists with the email", "Please continue with login"},
				nil,
				nil,
			))
	}
	SignUpData := TokenModels.SignUpT{
		TokenType:  "SignUp",
		Mail:       form.MailAddress,
		RememberMe: form.RememberMe,
		Name:       form.Name,
		Password:   form.Password,
		Step2Code:  "",
	}
	verification, retry := Step2Processor.SendMailOTP(ctx, form.MailAddress)
	if verification == "" {
		Logger.AccidentalFailure(fmt.Sprintf("Register-1 OTP failed: %s", form.MailAddress))
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.OtpSendFailed,
				ResponseModels.DefaultAuth,
				[]string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
				nil,
				map[string]interface{}{"resend-after": retry.Seconds()},
			))
	}
	SignUpData.Step2Code = verification
	data, err := json.Marshal(SignUpData)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Register-1 Marshal Failed: %s", err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to marshal SignUp data, please contact support"},
				nil,
				nil,
			))
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("Register-1 Encrypt Failed: %s", form.MailAddress))
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to encrypt SignUp data, please retry"},
				nil,
				nil,
			))
	}
	Logger.Success(fmt.Sprintf("Register-1 Token Generated: %s", form.MailAddress))
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseProcessor.CombineResponses(
			ResponseModels.SignUpIDVerified,
			ResponseModels.DefaultAuth,
			[]string{"Please enter the OTP sent to your mail"},
			map[string]interface{}{"token": token},
			nil,
		))
}
