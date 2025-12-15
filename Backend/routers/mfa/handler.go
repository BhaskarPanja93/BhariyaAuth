package mfa

import (
	ResponseModels "BhariyaAuth/models/responses"
	TokenModels "BhariyaAuth/models/tokens"
	AccountProcessor "BhariyaAuth/processors/account"
	Logger "BhariyaAuth/processors/logs"
	OTPProcessor "BhariyaAuth/processors/otp"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	ResponseProcessor "BhariyaAuth/processors/response"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

const tokenType = "mfa"

type Step2FormT struct {
	Token        string `form:"token"`
	Verification string `form:"verification"`
}

func Step2(ctx fiber.Ctx) error {
	form := new(Step2FormT)
	var MFAData TokenModels.MFATokenT
	if err := ctx.Bind().Form(form); err != nil {
		if err = ctx.Bind().Body(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	data, ok := StringProcessor.Decrypt(form.Token)
	if !ok {
		Logger.AccidentalFailure("[MFA2] Decrypt Failed")
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	err := json.Unmarshal(data, &MFAData)
	if err != nil {
		Logger.AccidentalFailure("[MFA2] Unmarshal Failed")
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to read token (Encryptor issue)... Retrying"},
		})
	}
	if MFAData.TokenType != tokenType {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if !OTPProcessor.Validate(MFAData.Step2Code, form.Verification) {
		Logger.IntentionalFailure(fmt.Sprintf("[MFA2] Incorrect OTP for [UID-%d]", MFAData.UserID))
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Incorrect OTP"},
		})
	}
	if AccountProcessor.CheckUserIsBlacklisted(MFAData.UserID) {
		Logger.IntentionalFailure(fmt.Sprintf("[MFA2] Blacklisted account [UID-%d] attempted Mfa", MFAData.UserID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Your account is disabled, please contact support"},
		})
	}
	MFAData.Verified = true
	MFAData.Creation = time.Now().UTC()
	data, err = json.Marshal(MFAData)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[MFA2] Marshal Failed for [UID-%d] reason: %s", MFAData.UserID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[MFA2] Encrypt Failed for [UID-%d]", MFAData.UserID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	ResponseProcessor.AttachMFACookie(ctx, token)
	Logger.Success(fmt.Sprintf("[MFA2] Successful for [UID-%d]", MFAData.UserID))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
	})
}

func Step1(ctx fiber.Ctx) error {
	now := time.Now().UTC()
	refresh := TokenProcessor.ReadRefreshToken(ctx)
	if refresh.UserID == 0 {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	if !TokenProcessor.MatchCSRF(ctx, refresh) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if now.After(refresh.RefreshExpiry) {
		Logger.IntentionalFailure(fmt.Sprintf("[ProcessRefresh] Expired session [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
		ResponseProcessor.DetachAuthCookies(ctx)
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusUnauthorized).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"This session has expired... Please login again"},
		})
	}
	if !AccountProcessor.CheckSessionExists(refresh.UserID, refresh.RefreshID) {
		Logger.IntentionalFailure(fmt.Sprintf("[ProcessRefresh] Revoked session [UID-%d-RID-%d]", refresh.UserID, refresh.RefreshID))
		ResponseProcessor.DetachAuthCookies(ctx)
		RateLimitProcessor.Set(ctx)
		return ctx.Status(fiber.StatusUnauthorized).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"This session has been revoked... Please login again"},
		})
	}
	mail, found := AccountProcessor.GetMailFromID(refresh.UserID)
	if !found {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusInternalServerError)
	}
	verification, retry := OTPProcessor.Send(mail, "Multi-Factor Verification", "Enter the OTP below to complete MFA verification:", false, fmt.Sprintf("%s:verified", ctx.IP()))
	if verification == "" {
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Reply:         retry.Seconds(),
			Notifications: []string{fmt.Sprintf("Unable to send OTP, please try again after %.1f seconds", retry.Seconds())},
		})
	}
	MFAToken := TokenModels.MFATokenT{
		TokenType: tokenType,
		Step2Code: verification,
		UserID:    refresh.UserID,
		Creation:  time.Now().UTC(),
		Verified:  false,
	}
	data, err := json.Marshal(MFAToken)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[MFA1] Marshal Failed for [UID-%d] reason: %s", refresh.UserID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	token, ok := StringProcessor.Encrypt(data)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[MFA1] Encrypt Failed for [UID-%d]", refresh.UserID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to acquire token (Encryptor issue)... Retrying"},
		})
	}
	Logger.Success(fmt.Sprintf("[MFA1] Token Created for [UID-%d]", refresh.UserID))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
		Reply:   token,
	})
}
