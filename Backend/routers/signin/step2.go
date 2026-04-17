package signin

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const step2FileName = "routers/signin/step2"

// Step2 completes the signin process by verifying user credentials
// (either OTP or password) and issuing authentication tokens.
//
// This function finalizes authentication by:
//  1. Validating and decrypting the Step1-issued token.
//  2. Ensuring the token belongs to the signin flow.
//  3. Verifying user credentials based on selected method:
//     - Password-based authentication
//     - OTP-based authentication
//  4. Validating account status (exists, not blocked).
//  5. Registering the device/session.
//  6. Generating access and refresh tokens.
//  7. Optionally issuing MFA cookie for OTP-based signin.
//
// Flow Summary:
//
//	validate form → decrypt token → validate token → verify credentials → check account → create session → issue tokens → attach cookies
//
// Authentication Modes:
//   - Password:
//     Requires password verification using stored hash.
//   - OTP:
//     Requires OTP validation and automatically upgrades session to MFA-verified.
//
// Security Considerations:
// - Encrypted token prevents tampering with signin state.
// - TokenType validation prevents cross-flow token reuse.
// - Password is verified using bcrypt (secure hash comparison).
// - OTP validation ensures email ownership.
// - Rate limiting prevents brute-force attacks.
// - Session is recorded in DB to allow revocation.
// - MFA cookie is conditionally issued for OTP-based signin.
//
// Request Body:
// - token (string) → encrypted payload from Step1
// - verification (string) → password or OTP depending on flow
//
// Returns:
// - 200 OK with auth tokens on success
// - 200 OK with notifications (invalid credentials, account issues)
// - 422 Unprocessable Entity (invalid input)
// - 500 Internal Server Error (system failures)
func Step2(ctx fiber.Ctx) error {

	// Parse incoming form data
	form := new(FormModels.SignInForm2)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Form read failed")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	// Decrypt token received from Step1
	data, err := TokenProcessor.ReadSignInToken(form.Token)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Token read failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(data.UserID))+" "+data.Step2Process)

	var hash string  // Stored password hash (for password flow)
	var t string     // User type/role
	var blocked bool // Account status
	if data.Step2Process == PasswordProcess {

		// Fetch user credentials and status
		err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT pw_hash, type, blocked FROM users WHERE user_id = $1 LIMIT 1`, data.UserID).Scan(&hash, &t, &blocked)
		if errors.Is(err, pgx.ErrNoRows) {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account does not exist")

			// Account missing (edge case)
			RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
		} else if err != nil { // Any other DB error
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed - SQL query: "+err.Error())

			RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
		}

		// Validate password:
		// - Ensure password meets minimum format
		// - Compare with stored bcrypt hash
		if !StringProcessor.PasswordIsStrong(form.Verification) ||
			bcrypt.CompareHashAndPassword([]byte(hash), []byte(form.Verification)) != nil {
			Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Password incorrect")

			RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.PasswordIncorrect},
			})
		}
	} else if data.Step2Process == OTPProcess {
		// Validate OTP provided by user
		if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
			Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "OTP incorrect")

			RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
		}

		// Fetch user type and status
		err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT type, blocked FROM users WHERE user_id = $1 LIMIT 1`, data.UserID).Scan(&t, &blocked)
		if errors.Is(err, pgx.ErrNoRows) {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account does not exist")

			RequestProcessor.AddRateLimitWeight(ctx, 60_000) // 10 invalid attempts/minute

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
		} else if err != nil {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed - SQL query: "+err.Error())

			RequestProcessor.AddRateLimitWeight(ctx, 1_000) // 600 invalid attempts/minute

			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
		}
	}

	// Reject signin if account is blocked
	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountBlocked},
		})
	}

	// Register device/session in database
	deviceID, err := AccountProcessor.RecordReturningUser(ctx, data.MailAddress, data.UserID, data.Remember, true)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SignIn failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBWriteError},
		})
	}

	// Generate access + refresh tokens
	token, err := TokenProcessor.CreateFreshToken(ctx, data.UserID, deviceID, t, data.Remember, "email-signin")
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Access creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000) // 60 invalid attempts/minute

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.UnknownError},
		})
	}

	CookieProcessor.DetachMFACookies(ctx)
	// If OTP signin → automatically mark MFA as verified
	if data.Step2Process == OTPProcess {
		var mfaToken string
		mfaToken, err = TokenProcessor.CreateMFAToken(ctx, data.UserID, deviceID, data.Step2Code)
		if err != nil {
			Logs.RootLogger.Add(Logs.Warn, step2FileName, RequestProcessor.GetRequestId(ctx), "MFA creation failed: "+err.Error())
		}
		CookieProcessor.AttachMFACookie(ctx, mfaToken)
	}

	// Attach authentication cookies (refresh + CSRF)
	CookieProcessor.AttachAuthCookies(ctx, token)

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Request Complete: "+strconv.Itoa(int(deviceID)))

	// Return access token and expiry
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:    true,
		ModifyAuth: true,
		NewToken:   token.AccessToken,
		Reply:      token.AccessExpires,
	})
}
