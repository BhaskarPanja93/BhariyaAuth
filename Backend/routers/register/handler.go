package register

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	Generators "BhariyaAuth/processors/generator"
	Logger "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	ResponseProcessor "BhariyaAuth/processors/response"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	"time"

	"fmt"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

type Step1FormT struct {
	MailAddress string `form:"mail_address"`
	Name        string `form:"name"`
	Password    string `form:"password"`
	RememberMe  string `form:"remember_me"`
}

type Step2FormT struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

const tokenType = "Register"

func Step2(ctx fiber.Ctx) error {
	form := new(Step2FormT)
	var SignUpData TokenModels.SignUpT
	if err := ctx.Bind().Form(form); err != nil {
		if err = ctx.Bind().Body(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		Logger.AccidentalFailure("[Register2] Decrypt Failed")
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	err := json.Unmarshal(data, &SignUpData)
	if err != nil {
		Logger.AccidentalFailure("[Register2] Unmarshal Failed")
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to read token (Encryptor issue)... Retrying"},
		})
	}
	if SignUpData.TokenType != tokenType {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	userID, found := AccountProcessor.GetIDFromMail(SignUpData.Mail)
	if found {
		Logger.IntentionalFailure(fmt.Sprintf("[Register2] Attempted for [UID-%d]", userID))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Account exists with the email"},
		})
	}
	if !OTPProcessor.Validate(SignUpData.Step2Code, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("[Register2] Incorrect OTP for [MAIL-%s]", SignUpData.Mail))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Incorrect OTP"},
		})
	}
	userID = Generators.UserID()
	refreshID := Generators.RefreshID()
	if !AccountProcessor.RecordNewUser(userID, SignUpData.Password, SignUpData.Mail, SignUpData.Name) {
		Logger.AccidentalFailure(fmt.Sprintf("[Register2] Record New failed for [MAIL-%s]", SignUpData.Mail))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to register (DB-write issue)... Retrying"},
		})
	}
	if !AccountProcessor.RecordReturningUser(SignUpData.Mail, ctx.Get("User-Agent"), refreshID, userID, SignUpData.RememberMe) {
		Logger.AccidentalFailure(fmt.Sprintf("[Register2] Record Returning failed for [UID-%d]", userID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Account registered but failed to login. Please login manually"},
		})
	}
	token, ok := TokenProcessor.CreateFreshToken(userID, refreshID, UserTypes.All.Viewer, SignUpData.RememberMe, "email-register")
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Register2] CreateFreshToken failed for [UID-%d]", userID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Account registered but failed to login. Please login manually"},
		})
	}
	ResponseProcessor.AttachAuthCookies(ctx, token)
	var mfatoken string
	MFAToken := TokenModels.MFATokenT{
		Step2Code: SignUpData.Step2Code,
		UserID:    userID,
		Creation:  time.Now().UTC(),
		Verified:  true,
	}
	data, err = json.Marshal(MFAToken)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Register2MFA] Marshal Failed for [UID-%d] reason: %s", userID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	mfatoken, ok = StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Register2MFA] Encrypt Failed for [UID-%d]", userID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	ResponseProcessor.DetachMFACookies(ctx)
	ResponseProcessor.AttachMFACookie(ctx, mfatoken)
	Logger.Success(fmt.Sprintf("[Register2] Created: [UID-%d-RID-%d-MAIL-%s]", userID, refreshID, SignUpData.Mail))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:       true,
		ModifyAuth:    true,
		NewToken:      token.AccessToken,
		Notifications: []string{"Registered and logged in Successfully"},
	})
}

func Step1(ctx fiber.Ctx) error {
	form := new(Step1FormT)
	if err := ctx.Bind().Form(form); err != nil {
		if err = ctx.Bind().Body(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	if !StringProcessor.NameIsValid(form.Name) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if !StringProcessor.EmailIsValid(form.MailAddress) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if !StringProcessor.PasswordIsStrong(form.Password) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	userID, found := AccountProcessor.GetIDFromMail(form.MailAddress)
	if found {
		Logger.IntentionalFailure(fmt.Sprintf("[Register1] Attempted for [UID-%d]", userID))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Account exists with the email"},
		})
	}
	SignUpData := TokenModels.SignUpT{
		TokenType:  tokenType,
		Mail:       form.MailAddress,
		RememberMe: form.RememberMe == "yes",
		Name:       form.Name,
		Password:   form.Password,
	}
	verification, retry := OTPProcessor.Send(form.MailAddress, ctx.IP())
	if verification == "" {
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
		})
	}
	SignUpData.Step2Code = verification
	data, err := json.Marshal(SignUpData)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Register1] Marshal Failed for [MAIL-%s] reason: %s", form.MailAddress, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Parser issue)... Retrying"},
		})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Register1] Encrypt Failed for [MAIL-%s]", form.MailAddress))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	Logger.Success(fmt.Sprintf("[Register1] Token Generated for [MAIL-%s]", form.MailAddress))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:       true,
		Reply:         token,
		Notifications: []string{"Please enter the OTP sent to your mail"},
	})
}
