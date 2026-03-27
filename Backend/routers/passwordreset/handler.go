package passwordreset

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	AccountProcessor "BhariyaAuth/processors/account"
	FormProcessor "BhariyaAuth/processors/form"
	MailNotifier "BhariyaAuth/processors/mail"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	Stores "BhariyaAuth/stores"
	"errors"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func Step1(ctx fiber.Ctx) error {
	// Read the form else return 422
	form := new(FormModels.PasswordResetForm1)
	if !FormProcessor.ReadFormData(ctx, form) || !StringProcessor.EmailIsValid(form.Mail) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Fetch the associated account
	var userID int32
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT user_id FROM users WHERE mail = $1 LIMIT 1`, form.Mail).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) { // Row(account) not found
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
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
	// OTP verification string for step 2
	step2code, retry := OTPProcessor.Send(form.Mail, MailModels.PasswordResetInitiated, ctx.IP())
	// OTP send failed
	if step2code == "" {
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Reply:         retry.Seconds(),
				Notifications: []string{Notifications.OTPSendFailed},
			})
	}
	// Create a struct with all data required for step 2
	// Client is responsible to bring the struct when requesting step2
	PasswordReset := TokenModels.PasswordResetT{
		TokenType:   tokenType,
		MailAddress: form.Mail,
		UserID:      userID,
		Step2Code:   step2code,
	}
	data, err := json.Marshal(PasswordReset)
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
	form := new(FormModels.PasswordResetForm2)
	if !FormProcessor.ReadFormData(ctx, form) || !StringProcessor.PasswordIsStrong(form.Password) {
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		// Decrypt failed
		RateLimitProcessor.Add(ctx, 10_000) // 60 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	var ResetData TokenModels.PasswordResetT
	err := json.Unmarshal(data, &ResetData)
	if err != nil {
		// Bytes Unmarshal failed
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.MarshalError},
			})
	}
	if ResetData.TokenType != tokenType {
		// Token belongs to some other purpose and not password reset
		RateLimitProcessor.Add(ctx, 30_000) // 20 mistakes allowed / minute
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	// Check otp validity
	if !OTPProcessor.Validate(ResetData.Step2Code, form.Verification) {
		RateLimitProcessor.Add(ctx, 20_000) // 30 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
	}
	var blocked bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, "SELECT blocked FROM users WHERE user_id = $1 LIMIT 1", ResetData.UserID).Scan(&blocked)
	if errors.Is(err, pgx.ErrNoRows) { // Account not found
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
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
	// Reject request if account is blocked from logging in
	if blocked {
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountBlocked},
			})
	}
	// Hash the new password
	hash, ok := StringProcessor.HashPassword(form.Password)
	if !ok {
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.EncryptorError},
			})
	}
	// Update the new hash into DB
	_, err = Stores.SQLClient.Exec(Config.CtxBG, `UPDATE users SET pw_hash = $1 WHERE user_id = $2`, hash, ResetData.UserID)
	if err != nil {
		RateLimitProcessor.Add(ctx, 60_000) // 10 mistakes allowed / minute
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBWriteError},
			})
	}
	// Delete all old sessions for this account
	if !AccountProcessor.DeleteAllSessions(ResetData.UserID) {
		return ctx.Status(fiber.StatusInternalServerError).JSON(
			ResponseModels.APIResponseT{
				Notifications: []string{Notifications.PasswordChanged, Notifications.RevokeFailed},
			})
	}
	os, device, browser := StringProcessor.ParseUA(ctx.Get("UserID-Agent"))
	// Send mail for successful password change
	MailNotifier.PasswordReset(ResetData.MailAddress, MailModels.PasswordChanged, ctx.IP(), os, device, browser, 2)
	// Return nothing but a success response
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
	})
}
