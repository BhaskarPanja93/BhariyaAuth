package login

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UsersTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func Step1(ctx fiber.Ctx) error {
	// Read the form else return 422
	form := new(FormModels.LoginForm1)
	if !FormProcessor.ReadFormData(ctx, form) || !StringProcessor.EmailIsValid(form.Mail) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Fetch all data at once to prevent DB overloading with multiple requests
	var userID int32
	var blocked bool
	var hasPassword bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT user_id, blocked, pw_hash IS NOT NULL AND pw_hash <> '' FROM users WHERE mail = $1 LIMIT 1`, form.Mail).Scan(&userID, &blocked, &hasPassword)
	if errors.Is(err, pgx.ErrNoRows) { // Row(account) not found
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil { // Any other DB error
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	// User account is blocked
	if blocked {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}
	// Process can be otp or password
	process := ctx.Params(ProcessParam)
	// OTP verification string to match in step 2
	step2code := ""
	if process == OTPProcess {
		step2code, retry := OTPProcessor.Send(form.Mail, MailModels.LoginInitiated, ctx.IP())
		// OTP send failed
		if step2code == "" {
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Reply:         retry.Seconds(),
					Notifications: []string{Notifications.OTPSendFailed},
				})
		}
	} else if process == PasswordProcess {
		// Prevent login if password was never set
		if !hasPassword {
			RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.PasswordNotSet},
				})
		}
	} else {
		// No valid process name was provided so send 422
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Create a struct with all data required for step 2
	// Client is responsible to bring the struct when requesting step2
	SignInData := TokenModels.SignInT{
		UserID:       userID,
		TokenType:    tokenType,
		RememberMe:   form.Remember == "yes",
		Step2Process: process,
		MailAddress:  form.Mail,
		Step2Code:    step2code,
	}
	data, err := json.Marshal(SignInData)
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
	// Return the processed encrypted string that user needs to bring for step 2 along
	// with password or otp based on their choice in step 1
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
			Reply:   token,
		})
}

func Step2(ctx fiber.Ctx) error {
	// Read the form else return 422
	form := new(FormModels.LoginForm2)
	if !FormProcessor.ReadFormData(ctx, form) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Decrypt the string provided by the client
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		// Decrypt failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}
	var SignInData TokenModels.SignInT
	// Unmarshal the decrypted bytes into the actual struct created in step 1
	err := json.Unmarshal(data, &SignInData)
	if err != nil {
		// Bytes Unmarshal failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	// Match token type, so other token cant be used in current context as tokens might have same keys
	if SignInData.TokenType != tokenType {
		// Token belongs to some other purpose and not login
		RateLimitProcessor.Add(ctx, 30_000) // 20 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	var hash string
	var t string
	var blocked bool
	if SignInData.Step2Process == PasswordProcess {
		// Fetch all data at once to prevent DB overloading with multiple requests
		err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT pw_hash, type, blocked FROM users WHERE user_id = $1 LIMIT 1`, SignInData.UserID).Scan(&hash, &t, &blocked)
		if errors.Is(err, pgx.ErrNoRows) { // Row(account) not found (extremely rare as step 1 has already passed)
			RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.AccountNotFound},
				})
		} else if err != nil { // Any other DB error
			RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.DBReadError},
				})
		}
		// For password process check password validity
		if !StringProcessor.PasswordIsStrong(form.Verification) || bcrypt.CompareHashAndPassword([]byte(hash), []byte(form.Verification)) != nil {
			RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.PasswordIncorrect},
				})
		}
	} else if SignInData.Step2Process == OTPProcess {
		// For otp process check otp validity
		if !OTPProcessor.Validate(SignInData.Step2Code, form.Verification) {
			RateLimitProcessor.Add(ctx, 20_000) // 30 mistakes allowed / minute
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.OTPIncorrect},
				})
		}
		// Fetch all data at once to prevent DB overloading with multiple requests
		err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT type, blocked FROM users WHERE user_id = $1 LIMIT 1`, SignInData.UserID).Scan(&t, &blocked)
		if errors.Is(err, pgx.ErrNoRows) {
			return ctx.Status(fiber.StatusOK).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.AccountNotFound},
				})
		} else if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(
				ResponseModels.APIResponseT{
					Notifications: []string{Notifications.DBReadError},
				})
		}
	}
	// User account is blocked
	if blocked {
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}
	// Register new device into database
	deviceID, ok := AccountProcessor.RecordReturningUser(ctx, SignInData.MailAddress, SignInData.UserID, SignInData.RememberMe, true)
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}
	// Create their access and refresh tokens
	token, ok := TokenProcessor.CreateFreshToken(ctx, SignInData.UserID, deviceID, UsersTypes.Find(t), SignInData.RememberMe, "email-login")
	if !ok {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.UnknownError},
		})
	}
	// If otp process was chosen, automatically provide mfa cookie
	if SignInData.Step2Process == OTPProcess {
		var mfaToken string
		MFAToken := TokenModels.MFATokenT{
			Step2Code: SignInData.Step2Code,
			UserID:    SignInData.UserID,
			DeviceID:  deviceID,
			Created:   ctx.Locals("request-start").(time.Time),
			Verified:  true,
		}
		data, err = json.Marshal(MFAToken)
		if err == nil {
			mfaToken, ok = StringProcessor.Encrypt(data)
			if ok {
				CookieProcessor.AttachMFACookie(ctx, mfaToken)
			}
		}
	} else {
		// If process wasn't otp, remove any older mfa cookies
		CookieProcessor.DetachMFACookies(ctx)
	}
	// Refresh and CSRF cookie is attached
	CookieProcessor.AttachAuthCookies(ctx, token)
	// Return the access token and its expiry
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success:    true,
			ModifyAuth: true,
			NewToken:   token.AccessToken,
			Reply:      token.AccessExpires.Format(time.RFC3339),
		})
}
