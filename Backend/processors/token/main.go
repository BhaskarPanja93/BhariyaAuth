package token

import (
	Config "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	Generators "BhariyaAuth/processors/generator"
	Logger "BhariyaAuth/processors/logs"
	StringProcessor "BhariyaAuth/processors/string"
	"fmt"
	"time"

	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

func encryptAccessToken(model TokenModels.AccessTokenT) (string, bool) {
	atUnEnc, err := json.Marshal(model)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[CreateRenewToken] Access Marshal failed for [UID-%d] reason: %s", model.UserID, err.Error()))
		return "", false
	}
	atEnc, ok := StringProcessor.Encrypt(atUnEnc)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[CreateRenewToken] Access Encrypt failed for [UID-%d]", model.UserID))
		return "", false
	}
	return atEnc, true
}

func encryptRefreshToken(model TokenModels.RefreshTokenT) (string, bool) {
	rtUnEnc, err := json.Marshal(model)
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[FreshToken] Refresh Marshal error for [UID-%d-RID-%d-TYP-%s] reason: %s", model.UserID, model.RefreshID, model.IdentifierType, err.Error()))
		return "", false
	}
	rtEnc, ok := StringProcessor.Encrypt(rtUnEnc)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[FreshToken] Refresh Encrypt error for [UID-%d-RID-%d-TYP-%s]", model.UserID, model.RefreshID, model.IdentifierType))
		return "", false
	}
	return rtEnc, true
}

func CreateFreshToken(userID uint32, refreshID uint16, userType UserTypes.T, remember bool, identifierType string, ctx fiber.Ctx) (TokenModels.NewTokenCombinedT, bool) {
	now := ctx.Locals("request-start").(time.Time)
	csrf := Generators.SafeString(128)
	atEnc, ok := encryptAccessToken(TokenModels.AccessTokenT{
		UserID:       userID,
		RefreshID:    refreshID,
		UserType:     userType,
		AccessExpiry: now.Add(Config.AccessTokenExpireDelta),
		RememberMe:   remember,
	})
	if !ok {
		return TokenModels.NewTokenCombinedT{}, false
	}
	rtEnc, ok := encryptRefreshToken(TokenModels.RefreshTokenT{
		UserID:         userID,
		RefreshID:      refreshID,
		RefreshIndex:   1,
		RefreshCreated: now,
		RefreshUpdated: now,
		RefreshExpiry:  now.Add(Config.RefreshTokenExpireDelta),
		UserType:       userType,
		CSRF:           csrf,
		RememberMe:     remember,
		IdentifierType: identifierType,
	})
	if !ok {
		return TokenModels.NewTokenCombinedT{}, false
	}
	return TokenModels.NewTokenCombinedT{
		AccessToken:  atEnc,
		RefreshToken: rtEnc,
		CSRF:         csrf,
		RememberMe:   remember,
	}, true
}

func CreateRenewToken(refresh TokenModels.RefreshTokenT, ctx fiber.Ctx) (TokenModels.NewTokenCombinedT, bool) {
	now := ctx.Locals("request-start").(time.Time)
	csrf := Generators.SafeString(128)
	atEnc, ok := encryptAccessToken(TokenModels.AccessTokenT{
		UserID:       refresh.UserID,
		RefreshID:    refresh.RefreshID,
		UserType:     refresh.UserType,
		AccessExpiry: now.Add(Config.AccessTokenExpireDelta),
		RememberMe:   refresh.RememberMe,
	})
	if !ok {
		return TokenModels.NewTokenCombinedT{}, false
	}
	refresh.CSRF = csrf
	refresh.RefreshExpiry = now.Add(Config.RefreshTokenExpireDelta)
	refresh.RefreshIndex %= 60000
	refresh.RefreshIndex++
	rtEnc, ok := encryptRefreshToken(refresh)
	if !ok {
		return TokenModels.NewTokenCombinedT{}, false
	}
	return TokenModels.NewTokenCombinedT{
		AccessToken:  atEnc,
		RefreshToken: rtEnc,
		CSRF:         csrf,
		RememberMe:   refresh.RememberMe,
	}, true
}

func ReadAccessToken(ctx fiber.Ctx) (TokenModels.AccessTokenT, bool) {
	var access TokenModels.AccessTokenT
	header := strings.TrimPrefix(ctx.Get(Config.AccessTokenInHeader), "Bearer ")
	dec, ok := StringProcessor.Decrypt(header)
	if ok && json.Unmarshal(dec, &access) == nil && access.UserID != 0 && access.RefreshID != 0 {
		return access, true
	}
	Logger.AccidentalFailure(fmt.Sprintf("[ReadAccessToken] Decrypt failed length: %d", len(header)))
	return access, false
}

func ReadRefreshToken(ctx fiber.Ctx) (TokenModels.RefreshTokenT, bool) {
	var refresh TokenModels.RefreshTokenT
	cookie := ctx.Cookies(Config.RefreshTokenInCookie)
	dec, ok := StringProcessor.Decrypt(cookie)
	if ok && json.Unmarshal(dec, &refresh) == nil && refresh.UserID != 0 && refresh.RefreshID != 0 {
		return refresh, true
	}
	Logger.AccidentalFailure(fmt.Sprintf("[ReadRefreshToken] Decrypt failed length: %d", len(cookie)))
	return refresh, false
}

func MatchCSRF(ctx fiber.Ctx, refresh TokenModels.RefreshTokenT) bool {
	cookie := ctx.Cookies(Config.CSRFInCookie)
	header := ctx.Get(Config.CSRFInHeader)
	return refresh.CSRF == header && refresh.CSRF == cookie
}
