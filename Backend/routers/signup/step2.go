package signup

import (
	Config "BhariyaAuth/constants/config"
	Notifications "BhariyaAuth/models/notifications"
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	UserTypes "BhariyaAuth/models/users"
	AccountProcessor "BhariyaAuth/processors/account"
	CookieProcessor "BhariyaAuth/processors/cookies"
	FormProcessor "BhariyaAuth/processors/form"
	Logs "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RequestProcessor "BhariyaAuth/processors/request"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const step2FileName = "routers/signup/step2"

func Step2(ctx fiber.Ctx) error {

	form := new(FormModels.SignUpForm2)

	if FormProcessor.ReadFormData(ctx, form) != nil {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Form read failed")

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}

	data, err := TokenProcessor.ReadSignUpToken(form.Token)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Token read failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.EncryptorError},
		})
	}

	Logs.RootLogger.Add(Logs.Intent, step2FileName, RequestProcessor.GetRequestId(ctx), "Requested for: "+data.MailAddress)

	var exists bool
	err = Stores.SQLClient.QueryRow(Config.CtxBG, `SELECT EXISTS(SELECT 1 FROM users WHERE mail = $1)`, data.MailAddress).Scan(&exists)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account existence check failed - SQL query: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 1_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBReadError},
		})
	}

	if exists {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Account exists during step2")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountPresent},
		})
	}

	if !OTPProcessor.Validate(data.Step2Code, form.Verification) {
		Logs.RootLogger.Add(Logs.Blocked, step2FileName, RequestProcessor.GetRequestId(ctx), "Incorrect OTP")

		RequestProcessor.AddRateLimitWeight(ctx, 60_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.OTPIncorrect},
		})
	}

	userType := UserTypes.All.Viewer.Short

	userID, err := AccountProcessor.RecordNewUser(ctx, userType, data.Password, data.MailAddress, data.Name)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SignUp failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.DBWriteError},
		})
	}
	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed Up: "+strconv.Itoa(int(userID)))

	deviceID, err := AccountProcessor.RecordReturningUser(ctx, data.MailAddress, userID, data.Remember, false)
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "SignIn failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}
	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Signed In: "+strconv.Itoa(int(userID)))

	token, err := TokenProcessor.CreateFreshToken(ctx, userID, deviceID, userType, data.Remember, "email-signup")
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, step2FileName, RequestProcessor.GetRequestId(ctx), "Access creation failed: "+err.Error())

		RequestProcessor.AddRateLimitWeight(ctx, 10_000)

		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Notifications: []string{Notifications.AccountCreated},
		})
	}

	CookieProcessor.AttachAuthCookies(ctx, token)

	mfaToken, err := TokenProcessor.CreateMFAToken(ctx, userID, deviceID, data.Step2Code, true)
	if err != nil {
		Logs.RootLogger.Add(Logs.Warn, step2FileName, RequestProcessor.GetRequestId(ctx), "MFA creation failed: "+err.Error())
	}
	CookieProcessor.AttachMFACookie(ctx, mfaToken)

	Logs.RootLogger.Add(Logs.Info, step2FileName, RequestProcessor.GetRequestId(ctx), "Request Complete: "+strconv.Itoa(int(userID))+" "+strconv.Itoa(int(deviceID)))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:    true,
		ModifyAuth: true,
		NewToken:   token.AccessToken,
		Reply:      token.AccessExpires,
	})
}
