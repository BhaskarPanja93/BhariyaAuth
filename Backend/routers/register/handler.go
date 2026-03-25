package register

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

func Step1(ctx fiber.Ctx) error {
	// Read the form else return 422
	form := new(FormModels.RegisterForm1)
	if !FormProcessor.ReadFormData(ctx, form) || !StringProcessor.NameIsValid(form.Name) || !StringProcessor.EmailIsValid(form.Mail) || !StringProcessor.PasswordIsStrong(form.Password) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Check if account exists with the provided email
	var exists bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 )`, form.Mail).Scan(&exists)
	if err != nil { // Any DB error
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	// Deny processing if account exists
	if exists {
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountPresent},
			})
	}
	// OTP verification string for step 2
	verification, retry := OTPProcessor.Send(form.Mail, MailModels.RegisterInitiated, ctx.IP())
	// OTP send failed
	if verification == "" {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retry.Seconds(),
				Notifications: []string{Notifications.OTPSendFailed},
			})
	}
	// Create a struct with all data required for step 2
	// Client is responsible to bring the struct when requesting step2
	SignUpData := TokenModels.SignUpT{
		TokenType:   tokenType,
		MailAddress: form.Mail,
		RememberMe:  form.Remember == "yes",
		Name:        form.Name,
		Password:    form.Password,
		Step2Code:   verification,
	}
	data, err := json.Marshal(SignUpData)
	if err != nil {
		// Struck marshaling failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		// Struck marshal encryption failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}
	// Return the processed encrypted string that user needs to bring for step 2 along with otp
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
			Reply:   token,
		})
}

func Step2(ctx fiber.Ctx) error {
	// Read the form else return 422
	form := new(FormModels.RegisterForm2)
	if !FormProcessor.ReadFormData(ctx, form) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		// Decrypt failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}
	var SignUpData TokenModels.SignUpT
	err := json.Unmarshal(data, &SignUpData)
	if err != nil {
		// Bytes Unmarshal failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	if SignUpData.TokenType != tokenType {
		// Token belongs to some other purpose and not login
		RateLimitProcessor.Add(ctx, 30_000) // 20 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Check if account exists with the provided email
	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 )`, SignUpData.MailAddress).Scan(&exists)
	if err != nil { // Any DB error
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	// Deny processing if account exists
	if exists {
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountPresent},
			})
	}
	// Check otp validity
	if !OTPProcessor.Validate(SignUpData.Step2Code, form.Verification) {
		RateLimitProcessor.Add(ctx, 20_000) // 30 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
	}
	// Register new account into database
	userID, ok := AccountProcessor.RecordNewUser(ctx, SignUpData.Password, SignUpData.MailAddress, SignUpData.Name)
	if !ok {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}
	// Register new device into database
	deviceID, ok := AccountProcessor.RecordReturningUser(ctx, SignUpData.MailAddress, userID, SignUpData.RememberMe, false)
	if !ok {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}
	// Create their access and refresh tokens
	token, ok := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, UserTypes.All.Viewer.Short, SignUpData.RememberMe, "email-register")
	if !ok {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}
	CookieProcessor.AttachAuthCookies(ctx, token)
	// For registrations providing mfa is compulsory as user used otp to create account
	MFAToken := TokenModels.MFATokenT{
		Step2Code: SignUpData.Step2Code,
		UserID:    userID,
		DeviceID:  deviceID,
		Created:   ctx.Locals("request-start").(time.Time),
		Verified:  true,
	}
	data, err = json.Marshal(MFAToken)
	if err != nil {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountCreated},
			})
	}
	mfaToken, ok := StringProcessor.Encrypt(data)
	if !ok {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountCreated},
			})
	}
	CookieProcessor.AttachMFACookie(ctx, mfaToken)
	// Return the access token and its expiry
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
			NewToken:   token.AccessToken,
			Reply:      token.AccessExpires,
		})
}
