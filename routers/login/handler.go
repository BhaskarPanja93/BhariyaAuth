package login

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UserModels "BhariyaAuth/models/users"

	AccountProcessor "BhariyaAuth/processors/account"
	ResponseProcessor "BhariyaAuth/processors/response"
	Step2Processor "BhariyaAuth/processors/step2"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"

	"fmt"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

type Step1T struct {
	MailAddress string `form:"mail_address"`
	RememberMe  bool   `form:"remember_me"`
}

type Step2T struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

func Step1(ctx fiber.Ctx) error {
	form := new(Step1T)

	// Parse the form
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}

	// Check for valid mail
	if !StringProcessor.IsValidEmail(form.MailAddress) {
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidEntries,
				ResponseModels.DefaultAuth,
				[]string{"Please enter a valid email address"},
				nil,
				nil,
			))
	} else {
		// Check if account exists
		userID, found := AccountProcessor.GetIDFromMail(form.MailAddress)
		if !found {
			return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
				ResponseModels.EmailDoesntExist,
				ResponseModels.DefaultAuth,
				[]string{"Account doesn't exist with the email"},
				nil,
				nil,
			))
		} else {
			process := ctx.Params("process")
			SignInData := TokenModels.SignInT{
				UserID:       userID,
				TokenType:    "SignIn",
				RememberMe:   form.RememberMe,
				Step2Process: process,
				Step2Code:    "",
				Mail:         form.MailAddress,
			}
			if process == "otp" {
				// Check if under limits
				verification, retry := Step2Processor.SendMailOTP(form.MailAddress)
				if verification == "" {
					return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.OtpSendFailed,
						ResponseModels.DefaultAuth,
						[]string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
						nil,
						map[string]interface{}{"resend-after": retry.Seconds()},
					))
				} else {
					SignInData.Step2Code = verification
				}
			} else if process == "password" {
				// Check if password exists
				if !AccountProcessor.UserHasPassword(userID) {
					return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.PasswordNotRegistered,
						ResponseModels.DefaultAuth,
						[]string{"Password has not been set", "Please use OTP/SSO to login"},
						nil,
						nil,
					))
				} else {
					SignInData.Step2Code = ""
				}
			} else {
				return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ResponseProcessor.CombineResponses(
					ResponseModels.InvalidEntries,
					ResponseModels.DefaultAuth,
					[]string{"Unknown process selected"},
					nil,
					nil,
				))
			}
			if data, err := json.Marshal(SignInData); err == nil {
				if token, ok := StringProcessor.Encrypt(data); ok {
					return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.SignInIDVerified,
						ResponseModels.DefaultAuth,
						[]string{fmt.Sprintf("Please enter the %s", process)},
						map[string]interface{}{"token": token},
						nil,
					))
				} else {
					return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.Unknown,
						ResponseModels.DefaultAuth,
						[]string{"Failed to encrypt SignIn data, please contact support"},
						nil,
						nil,
					))
				}
			} else {
				return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
					ResponseModels.Unknown,
					ResponseModels.DefaultAuth,
					[]string{"Failed to marshal SignIn data, please contact support"},
					nil,
					nil,
				))
			}
		}
	}
}

func Step2(ctx fiber.Ctx) error {
	form := new(Step2T)
	var SignInData TokenModels.SignInT

	// Parse the form
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}

	if data, ok := StringProcessor.Decrypt(form.Token); ok {
		if err := json.Unmarshal(data, &SignInData); err == nil {
			if SignInData.TokenType == "SignIn" {
				if SignInData.Step2Process == "password" {
					if !AccountProcessor.PasswordMatches(SignInData.UserID, form.Verification) {
						return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
							ResponseModels.InvalidCredentials,
							ResponseModels.DefaultAuth,
							[]string{"Incorrect Password"},
							nil,
							nil,
						))
					}
				} else if SignInData.Step2Process == "otp" {
					if !Step2Processor.ValidateMailOTP(SignInData.Step2Code, form.Verification) {
						return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
							ResponseModels.InvalidOTP,
							ResponseModels.DefaultAuth,
							[]string{"Incorrect OTP"},
							nil,
							nil,
						))
					}
				} else {
					return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.InvalidEntries,
						ResponseModels.DefaultAuth,
						[]string{"Unknown process"},
						nil,
						nil,
					))
				}
			} else {
				return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ResponseProcessor.CombineResponses(
					ResponseModels.InvalidToken,
					ResponseModels.DefaultAuth,
					[]string{"This token is not applicable for SignIn"},
					nil,
					nil,
				))
			}
		} else {
			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to unmarshal SignIn data, please contact support"},
				nil,
				nil,
			))
		}
	} else {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.Unknown,
			ResponseModels.DefaultAuth,
			[]string{"Failed to decrypt SignIn data, please contact support"},
			nil,
			nil,
		))
	}
	// Process Step-2
	if AccountProcessor.UserIsBlacklisted(SignInData.UserID) {
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.UserBlocked,
			ResponseModels.DefaultAuth,
			[]string{"Your account is disabled, please contact support"},
			nil,
			nil,
		))
	} else {
		refreshID := StringProcessor.GenerateRefreshID()
		if !AccountProcessor.RecordReturningUser(refreshID, SignInData) {
			return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to login, please try again or contact support"},
				nil,
				nil,
			))
		} else {
			token := TokenProcessor.CreateFreshToken(
				SignInData.UserID,
				refreshID,
				UserModels.Find(AccountProcessor.GetUserType(SignInData.UserID)),
				SignInData.RememberMe,
				SignInData.Mail,
				"email-login",
			)
			ResponseProcessor.AttachCookies(ctx, ResponseProcessor.GenerateCookies(token))
			return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
				ResponseModels.SignedIn,
				ResponseModels.AuthT{
					Allowed: true,
					Change:  true,
					Token:   token.AccessToken,
				},
				[]string{"Logged In Successfully"},
				nil,
				nil,
			))
		}
	}
}
