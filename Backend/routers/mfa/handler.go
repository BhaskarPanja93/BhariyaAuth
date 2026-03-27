package mfa

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"database/sql"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

func Step1(ctx fiber.Ctx) error {
	now := ctx.Locals("request-start").(time.Time)
	refresh, ok := TokenProcessor.ReadRefreshToken(ctx)
	// Refresh requires CSRF
	if !ok || !TokenProcessor.MatchCSRF(ctx, refresh) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	// Expiry should be in the future for a refresh token to be called active
	if now.After(refresh.Expiry) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionExpired},
			})
	}
	// Check if current session exists in DB
	var exists bool
	err := Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT EXISTS(SELECT 1 FROM devices WHERE user_id = $1 AND device_id = $2)", refresh.UserID, refresh.DeviceID).Scan(&exists)
	if err != nil { // Any DB error
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	// Reject request if session missing
	if !exists { // Row(device) not found
		CookieProcessor.DetachAuthCookies(ctx)
		CookieProcessor.DetachMFACookies(ctx)
		CookieProcessor.DetachSSOCookies(ctx)
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.SessionRevoked},
			})
	}
	// Fetch mail address to send OTP to
	var mail string
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT mail FROM users WHERE user_id = $1 LIMIT 1`, refresh.UserID).Scan(&mail)
	if errors.Is(err, sql.ErrNoRows) { // Row(account) not found
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusUnauthorized).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil { // Any DB error
		RateLimitProcessor.Add(ctx, 1_000) // 600 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	// OTP verification string to match in step 2
	step2code, retry := OTPProcessor.Send(mail, MailModels.MFAInitiated, ctx.IP())
	// OTP send failed
	if step2code == "" {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retry.Seconds(),
				Notifications: []string{Notifications.OTPSendFailed},
			})
	}
	// Create a struct with all data required for step 2
	// Client is responsible to bring the struct when requesting step2
	MFAToken := TokenModels.MFATokenT{
		TokenType: tokenType,
		Step2Code: step2code,
		UserID:    refresh.UserID,
		DeviceID:  refresh.DeviceID,
	}
	data, err := json.Marshal(MFAToken)
	if err != nil {
		// Struck marshaling failed
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		// Struck marshal encryption failed
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
	var MFAData TokenModels.MFATokenT
	err := json.Unmarshal(data, &MFAData)
	if err != nil {
		// Bytes Unmarshal failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	if MFAData.TokenType != tokenType {
		// Token belongs to some other purpose and not login
		RateLimitProcessor.Add(ctx, 30_000) // 20 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Check otp validity
	if !OTPProcessor.Validate(MFAData.Step2Code, form.Verification) {
		RateLimitProcessor.Add(ctx, 20_000) // 30 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
	}
	// Check if user is blocked from logging in
	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT blocked FROM users WHERE user_id = $1 LIMIT 1", MFAData.UserID).Scan(&blocked)
	if errors.Is(err, sql.ErrNoRows) { // Row(account) not found
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
	} else if err != nil { // Any DB error
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
	}
	// reject request if user is blocked
	if blocked {
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}
	// Reuse the same struct as the MFA cookie
	MFAData.Verified = true
	MFAData.Created = ctx.Locals("request-start").(time.Time)
	data, err = json.Marshal(MFAData)
	if err != nil {
		// Struck marshaling failed
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		// Struck marshal encryption failed
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}
	CookieProcessor.AttachMFACookie(ctx, token)
	// Return a success response with the new cookie attached
	return ctx.Status(fiber.StatusOK).JSON(
		ResponseModels.APIResponseT{
			Success: true,
		})
}
