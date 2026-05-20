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

func Step2(ctx fiber.Ctx) error {

	form := new(FormModels.SignInForm2)
	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Form read failed")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	data, err := TokenProcessor.ReadSignInToken(form.Token)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Token read failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+strconv.Itoa(int(data.UserID))+" "+data.Step2Process)

	var hash string
	var t string
	var blocked bool
	if data.Step2Process == PasswordProcess {

		err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT pw_hash, type, blocked FROM users WHERE user_id = $1 LIMIT 1`, data.UserID).Scan(&hash, &t, &blocked)
		if errors.Is(err, pgx.ErrNoRows) {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account does not exist")

			RequestProcessor.AddRateLimitWeight(ctx, 60_000)

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
		} else if err != nil {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed - SQL query: "+err.Error())

			RequestProcessor.AddRateLimitWeight(ctx, 1_000)

			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
		}

		if !StringProcessor.PasswordIsStrong(form.Verification) ||
			bcrypt.CompareHashAndPassword([]byte(hash), []byte(form.Verification)) != nil {
			Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Password incorrect")

			RequestProcessor.AddRateLimitWeight(ctx, 60_000)

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.PasswordIncorrect},
			})
		}
	} else if data.Step2Process == OTPProcess {
		if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
			Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "OTP incorrect")

			RequestProcessor.AddRateLimitWeight(ctx, 60_000)

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.OTPIncorrect},
			})
		}

		err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT type, blocked FROM users WHERE user_id = $1 LIMIT 1`, data.UserID).Scan(&t, &blocked)
		if errors.Is(err, pgx.ErrNoRows) {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account does not exist")

			RequestProcessor.AddRateLimitWeight(ctx, 60_000)

			return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.AccountNotFound},
			})
		} else if err != nil {
			Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account data fetch failed - SQL query: "+err.Error())

			RequestProcessor.AddRateLimitWeight(ctx, 1_000)

			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
				Notifications: []string{Notifications.DBReadError},
			})
		}
	}

	if blocked {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Account blocked")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountBlocked},
		})
	}

	deviceID, err := AccountProcessor.RecordReturningUser(ctx, data.MailAddress, data.UserID, data.Remember, true)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SignIn failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBWriteError},
		})
	}

	token, err := TokenProcessor.CreateFreshToken(ctx, data.UserID, deviceID, t, data.Remember, "email-signin")
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Access creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.UnknownError},
		})
	}

	CookieProcessor.DetachMFACookies(ctx)
	if data.Step2Process == OTPProcess {
		var mfaToken string
		mfaToken, err = TokenProcessor.CreateMFAToken(ctx, data.UserID, deviceID, data.Step2Code, true)
		if err != nil {
			Logs.RootLogger.Add(Logs.Warn, step2FileName, RequestProcessor.GetRequestId(ctx), "MFA creation failed: "+err.Error())
		}
		CookieProcessor.AttachMFACookie(ctx, mfaToken)
	}

	CookieProcessor.AttachAuthCookies(ctx, token)

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Request Complete: "+strconv.Itoa(int(deviceID)))

	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:    true,
		ModifyAuth: true,
		NewToken:   token.AccessToken,
		Reply:      token.AccessExpires,
	})
}
