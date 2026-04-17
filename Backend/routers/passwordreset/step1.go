package passwordreset

import (
	Config "BhariyaAuth/constants/config"
	MailModels "BhariyaAuth/models/mails"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

const step1FileName = "routers/passwordreset/step1"

// Step1 initiates the password reset flow by verifying the user's email
// and issuing an OTP along with a secure, encrypted reset token.
//
// This function serves as the entry point for password recovery. It performs:
//  1. Validation of email.
//  2. Lookup of the user account associated with the email.
//  3. OTP generation and delivery to the user's email.
//  4. Construction of a secure payload containing reset metadata.
//  5. Encryption of the payload into a token returned to the client.
//
// The returned token must be provided in Step2 along with the OTP
// to complete password reset.
//
// Flow Summary:
//
//	validate input → fetch user → send OTP → build state → encrypt → return token
//
// Security Considerations:
// - Rate limiting is applied to prevent abuse and email enumeration attacks.
// - OTP ensures that only the email owner can proceed.
// - Sensitive data (userID, OTP reference) is encrypted before being returned.
// - Stateless design: no server-side session is stored; client carries encrypted state.
//
// Request Body:
// - mail (string)
//
// Returns:
// - 200 OK with encrypted token (on success)
// - 200 OK with notification (account not found, OTP failure)
// - 422 Unprocessable Entity (invalid input)
// - 500 Internal Server Error (database or system failure)
func Step1(ctx fiber.Ctx) error {

	// Parse and validate incoming form data
	form := new(FormModels.PasswordResetForm1)

	// Validate:
	// - form parsing success
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Invalid form")

		// Penalize invalid input attempts
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Validate:
	// - email format correctness
	if !StringProcessor.EmailIsValid(form.Mail) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Invalid email")

		// Penalize invalid input attempts
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	Logs.RootLogger.Add(Logs.Intent, step1FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+form.Mail)

	// Fetch user account associated with the provided email
	var userID int32
	err := Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT user_id FROM users WHERE mail = $1 LIMIT 1`, form.Mail).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		Logs.RootLogger.Add(Logs.Blocked, step1FileName, RequestProcessor.GetRequestId(ctx), "Account not found")

		// Account does not exist (high penalty to prevent enumeration)
		RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountNotFound},
		})
	} else if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Account find failed - SQL query: "+err.Error())

		// Unexpected database error
		RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	// Send OTP to user's email for password reset verification
	step2code, retry, err := OTPProcessor.Send(form.Mail, MailModels.PasswordResetStarted, ctx.IP())

	// OTP dispatch failed (rate-limited or system issue)
	if step2code == "" {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Step2 code empty: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Reply:         retry.Seconds(), // Inform client when retry is allowed
			Notifications: []string{Notifications.OTPSendFailed},
		})
	}

	// Construct payload required for Step2
	// Contains identity + OTP reference for verification
	token, err := TokenProcessor.CreatePasswordResetToken(form.Mail, userID, step2code)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step1FileName, RequestProcessor.GetRequestId(ctx), "Token creation failed: "+err.Error())

		// Encryption failure (critical security issue)
		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Info, step1FileName, RequestProcessor.GetRequestId(ctx), "Request complete")
	// Return encrypted token to client.
	// Client must include this token in Step2 along with OTP
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   token,
	})
}
