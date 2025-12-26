package sessions

import (
	FormModels "BhariyaAuth/models/requests"
	ResponseModels "BhariyaAuth/models/responses"
	AccountProcessor "BhariyaAuth/processors/account"
	Logger "BhariyaAuth/processors/logs"
	RateLimitProcessor "BhariyaAuth/processors/ratelimit"
	StringProcessor "BhariyaAuth/processors/string"
	TokenProcessor "BhariyaAuth/processors/token"
	Stores "BhariyaAuth/stores"
	"encoding/binary"

	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
)

func Fetch(ctx fiber.Ctx) error {
	now := ctx.Locals("request-start").(time.Time)
	access, ok := TokenProcessor.ReadAccessToken(ctx)
	if !ok || now.After(access.AccessExpiry) {
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	rows, err := Stores.MySQLClient.Query("SELECT refresh, count, remembered, creation, updated, os, device, browser FROM activities WHERE uid = ?", access.UserID)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[Sessions] Fetch error for [UID-%d] reason: %s", access.UserID, err.Error()))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Notifications: []string{"Failed to fetch data (DB-read issue)... Retrying"},
		})
	}
	defer rows.Close()

	var response ResponseModels.UserActivityResponseT

	userbuf := make([]byte, 4)
	binary.BigEndian.PutUint32(userbuf, access.UserID)
	response.UserID, ok = StringProcessor.Encrypt(userbuf)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[Sessions] UID Encrypt error for [UID-%d]", access.UserID))
		return ctx.Status(fiber.StatusInternalServerError).JSON(ResponseModels.APIResponseT{
			Success:       false,
			Notifications: []string{"Failed to fetch data (Encryptor issue)... Retrying"},
		})
	}
	var success, failure uint
	for rows.Next() {
		var RefreshID uint16
		var activity ResponseModels.SingleUserActivityT
		err = rows.Scan(&RefreshID, &activity.Count, &activity.Remembered, &activity.Creation, &activity.Updated, &activity.OS, &activity.Device, &activity.Browser)
		if err != nil {
			Logger.AccidentalFailure(fmt.Sprintf("[Sessions] Scan error for [UID-%d] reason: %s", access.UserID, err.Error()))
			failure++
			continue
		}
		ridbuf := make([]byte, 2)
		binary.BigEndian.PutUint16(ridbuf, RefreshID)
		activity.ID, ok = StringProcessor.Encrypt(ridbuf)
		if !ok {
			Logger.AccidentalFailure(fmt.Sprintf("[Sessions] RID Encrypt error for [UID-%d-RID-%d]", access.UserID, RefreshID))
			failure++
			continue
		}
		response.Activities = append(response.Activities, activity)
		success++
	}
	Logger.Success(fmt.Sprintf("[Sessions] Fetched for [UID-%d]", access.UserID))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success:       true,
		Reply:         response,
		Notifications: []string{fmt.Sprintf("Fetched data, successful: %d failed: %d", success, failure)},
	})
}

func Revoke(ctx fiber.Ctx) error {
	form := new(FormModels.DeviceRevokeForm)
	if err := ctx.Bind().Form(form); err != nil {
		if err = ctx.Bind().Body(form); err != nil {
			RateLimitProcessor.Set(ctx)
			return ctx.SendStatus(fiber.StatusUnprocessableEntity)
		}
	}
	access, _ := TokenProcessor.ReadAccessToken(ctx)
	uidbuf, ok := StringProcessor.Decrypt(form.UserID)
	if !ok {
		Logger.AccidentalFailure("[Revoke] UID Decrypt failed")
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	if access.UserID != binary.BigEndian.Uint32(uidbuf) {
		RateLimitProcessor.Set(ctx)
		return ctx.SendStatus(fiber.StatusUnauthorized)
	}
	if form.RevokeAll == "yes" {
		AccountProcessor.DeleteAllSessions(access.UserID)
		Logger.Success(fmt.Sprintf("Sessions RevokeAll succeeded for [%d]", access.UserID))
		return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
			Success: true,
		})
	}
	RefreshID, ok := StringProcessor.Decrypt(form.DeviceID)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("RevokeAll RID Decrypt failed [%s]", form.DeviceID))
		return ctx.SendStatus(fiber.StatusUnprocessableEntity)
	}
	AccountProcessor.DeleteSession(access.UserID, binary.BigEndian.Uint16(RefreshID))
	Logger.Success(fmt.Sprintf("Sessions Revoke succeeded for [%d-%d]", access.UserID, RefreshID))
	return ctx.Status(fiber.StatusOK).JSON(ResponseModels.APIResponseT{
		Success: true,
	})
}
