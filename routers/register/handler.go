package register

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	UserModels "BhariyaAuth/models/users"
	"BhariyaAuth/processors/generator"

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
	Name        string `form:"name"`
	Password    string `form:"password"`
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

	if form.Name == "" {
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidEntries,
				ResponseModels.DefaultAuth,
				[]string{"Please enter a name"},
				nil,
				nil,
			))
	} else if !StringProcessor.IsValidEmail(form.MailAddress) {
		return ctx.Status(fiber.StatusOK).JSON(
			ResponseProcessor.CombineResponses(
				ResponseModels.InvalidEntries,
				ResponseModels.DefaultAuth,
				[]string{"Please enter a valid email address"},
				nil,
				nil,
			))
	} else if !AccountProcessor.PasswordIsStrong(form.Password) {
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.PasswordTooSimple,
			ResponseModels.DefaultAuth,
			[]string{"Please choose a stronger password"},
			nil,
			nil,
		))
	} else {
		// Check if account exists
		_, found := AccountProcessor.GetIDFromMail(form.MailAddress)
		if found {
			return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
				ResponseModels.EmailAlreadyTaken,
				ResponseModels.DefaultAuth,
				[]string{"Account exists with the email", "Please continue with login"},
				nil,
				nil,
			))
		} else {
			SignUpData := TokenModels.SignUpT{
				TokenType:  "SignUp",
				Mail:       form.MailAddress,
				RememberMe: form.RememberMe,
				Name:       form.Name,
				Password:   form.Password,
				Step2Code:  "",
			}
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
				SignUpData.Step2Code = verification
			}
			if data, err := json.Marshal(SignUpData); err == nil {
				if token, ok := StringProcessor.Encrypt(data); ok {
					return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.SignUpIDVerified,
						ResponseModels.DefaultAuth,
						[]string{"Please enter the OTP sent to your mail"},
						map[string]interface{}{"token": token},
						nil,
					))
				} else {
					return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.Unknown,
						ResponseModels.DefaultAuth,
						[]string{"Failed to encrypt SignUp data, please contact support"},
						nil,
						nil,
					))
				}
			} else {
				return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
					ResponseModels.Unknown,
					ResponseModels.DefaultAuth,
					[]string{"Failed to marshal SignUp data, please contact support"},
					nil,
					nil,
				))
			}
		}
	}
}

func Step2(ctx fiber.Ctx) error {
	form := new(Step2T)
	var SignUpData TokenModels.SignUpT

	// Parse the form
	if err := ctx.Bind().JSON(form); err != nil {
		if err = ctx.Bind().Form(form); err != nil {
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}

	if data, ok := StringProcessor.Decrypt(form.Token); ok {
		if err := json.Unmarshal(data, &SignUpData); err == nil {
			if SignUpData.TokenType == "SignUp" {
				_, found := AccountProcessor.GetIDFromMail(SignUpData.Mail)
				// Check if account exists
				if found {
					return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
						ResponseModels.EmailAlreadyTaken,
						ResponseModels.DefaultAuth,
						[]string{"Account exists with the email"},
						nil,
						nil,
					))
				} else {
					if !Step2Processor.ValidateMailOTP(SignUpData.Step2Code, form.Verification) {
						return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
							ResponseModels.InvalidOTP,
							ResponseModels.DefaultAuth,
							[]string{"Incorrect OTP"},
							nil,
							nil,
						))
					}
				}
			} else {
				return ctx.Status(fiber.StatusUnprocessableEntity).JSON(ResponseProcessor.CombineResponses(
					ResponseModels.InvalidToken,
					ResponseModels.DefaultAuth,
					[]string{"This token is not applicable for SignUp"},
					nil,
					nil,
				))
			}
		} else {
			return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
				ResponseModels.Unknown,
				ResponseModels.DefaultAuth,
				[]string{"Failed to unmarshal SignUp data, please contact support"},
				nil,
				nil,
			))
		}
	} else {
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.Unknown,
			ResponseModels.DefaultAuth,
			[]string{"Failed to decrypt SignUp data, please contact support"},
			nil,
			nil,
		))
	}
	// Process Step-2
	userID := generator.UserID()
	refreshID := generator.RefreshID()
	if !AccountProcessor.RecordNewUser(userID, SignUpData.Password, SignUpData.Mail, SignUpData.Name) {
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.Unknown,
			ResponseModels.DefaultAuth,
			[]string{"Failed to create account, please contact support"},
			nil,
			nil,
		))
	} else {
		token := TokenProcessor.CreateFreshToken(
			userID,
			refreshID,
			UserModels.All.Viewer,
			SignUpData.RememberMe,
			"email-register",
		)
		ResponseProcessor.AttachAuthCookies(ctx, ResponseProcessor.GenerateAuthCookies(token))
		return ctx.Status(fiber.StatusOK).JSON(ResponseProcessor.CombineResponses(
			ResponseModels.SignedUp,
			ResponseModels.AuthT{
				Allowed: true,
				Change:  true,
				Token:   token.AccessToken,
			},
			[]string{"Account Created Successfully"},
			nil,
			nil,
		))
	}
}
