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

// Step1 takes in a form with a valid name, mail address, password and remember (as 'yes' or 'no').
// Check the provided email if it already exists in self database, and if so, reject the request.
// If the entry is new, a struct with all the values is serialized and returned to the user that the user needs to bring in
// while submitting the form in Step2 along with OTP for verifying the email actually exists, is active and owned by the user.
func Step1(ctx fiber.Ctx) error {
	// Read the form else return 422
	form := new(FormModels.RegisterForm1)
	if !FormProcessor.ReadFormData(ctx, form) || !StringProcessor.NameIsValid(form.Name) || !StringProcessor.EmailIsValid(form.Mail) || !StringProcessor.PasswordIsStrong(form.Password) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Check if account exists with the provided email
	var exists bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE mail = $1)`, form.Mail).Scan(&exists)
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
	step2code, retryAfter := OTPProcessor.Send(form.Mail, MailModels.RegisterInitiated, ctx.IP())
	// OTP send failed
	if step2code == "" {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retryAfter.Seconds(),
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
		Step2Code:   step2code,
	}
	data, err := json.Marshal(SignUpData)
	if err != nil {
		// Struck marshaling failed
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		// Struck marshal encryption failed
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
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

// Step2 requires a form with the token received from Step1 along with verification which would be the OTP received in their email.
// The token is decrypted, unmarshalled and that would provide Step1's form data including email.
// The email will be rechecked for account in self database (to prevent someone creating the account after completing Step1 and retrying Step2)
// OTP gets validated and on success, new account gets created, also the user device gets logged in at the same time.
// On success, new Auth and Refresh tokens are created and sent back as response and cookie respectively.
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
		// Token belongs to some other purpose and not registration
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Check if account exists with the provided email
	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1)`, SignUpData.MailAddress).Scan(&exists)
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
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
	}
	userType := UserTypes.All.Viewer.Short
	// Register new account into database
	userID, ok := AccountProcessor.RecordNewUser(ctx, userType, SignUpData.Password, SignUpData.MailAddress, SignUpData.Name)
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
	token, ok := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, userType, SignUpData.RememberMe, "email-register")
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
