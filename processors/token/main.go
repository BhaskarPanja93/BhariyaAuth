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

func CreateFreshToken(userID uint32, refreshID uint16, userType UserTypes.T, remember bool, identifierType string) (TokenModels.NewTokenCombinedT, bool) {
	now := time.Now().UTC()
	csrf := Generators.SafeString(128)
	atUnEnc, err := json.Marshal(TokenModels.AccessTokenT{
		UserID:       userID,
		RefreshID:    refreshID,
		UserType:     userType,
		AccessExpiry: now.Add(Config.AccessTokenExpireDelta),
		RememberMe:   remember,
	})
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[FreshToken] Access Marshal failed for [UID-%d-RID-%d-TYP-%s] reason: %s", userID, refreshID, identifierType, err.Error()))
		return TokenModels.NewTokenCombinedT{}, false
	}
	atEnc, ok := StringProcessor.Encrypt(atUnEnc)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[FreshToken] Access Encrypt error for [UID-%d-RID-%d-TYP-%s]", userID, refreshID, identifierType))
		return TokenModels.NewTokenCombinedT{}, false
	}
	rtUnEnc, err := json.Marshal(TokenModels.RefreshTokenT{
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
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[FreshToken] Refresh Marshal error for [UID-%d-RID-%d-TYP-%s] reason: %s", userID, refreshID, identifierType, err.Error()))
		return TokenModels.NewTokenCombinedT{}, false
	}
	rtEnc, ok := StringProcessor.Encrypt(rtUnEnc)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[FreshToken] Refresh Encrypt error for [UID-%d-RID-%d-TYP-%s]", userID, refreshID, identifierType))
		return TokenModels.NewTokenCombinedT{}, false
	}
	return TokenModels.NewTokenCombinedT{
		AccessToken:  atEnc,
		RefreshToken: rtEnc,
		CSRF:         csrf,
		RememberMe:   remember,
	}, true
}

func CreateRenewToken(refresh TokenModels.RefreshTokenT) (TokenModels.NewTokenCombinedT, bool) {
	now := time.Now().UTC()
	csrf := Generators.SafeString(128)
	if refresh.RefreshIndex >= 60000 {
		refresh.RefreshIndex = 0
	}
	refresh.RefreshIndex++
	atUnEnc, err := json.Marshal(TokenModels.AccessTokenT{
		UserID:       refresh.UserID,
		RefreshID:    refresh.RefreshID,
		UserType:     refresh.UserType,
		AccessExpiry: now.Add(Config.AccessTokenExpireDelta),
		RememberMe:   refresh.RememberMe,
	})
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[CreateRenewToken] Access Marshal failed for [UID-%d] reason: %s", refresh.UserID, err.Error()))
		return TokenModels.NewTokenCombinedT{}, false
	}
	atEnc, ok := StringProcessor.Encrypt(atUnEnc)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[CreateRenewToken] Access Encrypt failed for [UID-%d]", refresh.UserID))
		return TokenModels.NewTokenCombinedT{}, false
	}
	rtUnEnc, err := json.Marshal(TokenModels.RefreshTokenT{
		UserID:         refresh.UserID,
		RefreshID:      refresh.RefreshID,
		RefreshIndex:   refresh.RefreshIndex,
		RefreshUpdated: now,
		RefreshCreated: refresh.RefreshCreated,
		RefreshExpiry:  now.Add(Config.RefreshTokenExpireDelta),
		UserType:       refresh.UserType,
		CSRF:           csrf,
		RememberMe:     refresh.RememberMe,
		IdentifierType: refresh.IdentifierType,
	})
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("[CreateRenewToken] Refresh Marshal failed for [UID-%d] reason: %s", refresh.UserID, err.Error()))
		return TokenModels.NewTokenCombinedT{}, false
	}
	rtEnc, ok := StringProcessor.Encrypt(rtUnEnc)
	if !ok {
		Logger.AccidentalFailure(fmt.Sprintf("[CreateRenewToken] Refresh Encrypt failed for [UID-%d]", refresh.UserID))
		return TokenModels.NewTokenCombinedT{}, false
	}
	return TokenModels.NewTokenCombinedT{
		AccessToken:  atEnc,
		RefreshToken: rtEnc,
		CSRF:         csrf,
		RememberMe:   refresh.RememberMe,
	}, true
}

func ReadAccessToken(ctx fiber.Ctx) TokenModels.AccessTokenT {
	var access TokenModels.AccessTokenT
	header := strings.TrimSpace(strings.TrimPrefix(ctx.Get(Config.AccessTokenInHeader), "Bearer "))
	dec, ok := StringProcessor.Decrypt(header)
	if ok {
		_ = json.Unmarshal(dec, &access)
	} else {
		Logger.AccidentalFailure(fmt.Sprintf("[ReadAccessToken] Decrypt failed length: %d", len(header)))
	}
	return access
}

func ReadRefreshToken(ctx fiber.Ctx) TokenModels.RefreshTokenT {
	var refresh TokenModels.RefreshTokenT
	cookie := strings.TrimSpace(ctx.Cookies(Config.RefreshTokenInCookie))
	dec, ok := StringProcessor.Decrypt(cookie)
	if ok {
		_ = json.Unmarshal(dec, &refresh)
	} else {
		Logger.AccidentalFailure(fmt.Sprintf("[ReadRefreshToken] Decrypt failed length: %d", len(cookie)))
	}
	return refresh
}

func MatchCSRF(ctx fiber.Ctx, refresh TokenModels.RefreshTokenT) bool {
	cookie := ctx.Cookies(Config.CSRFInCookie)
	header := ctx.Get(Config.CSRFInHeader)
	cookie = strings.TrimSpace(cookie)
	header = strings.TrimSpace(header)
	return refresh.CSRF == header && refresh.CSRF == cookie
}
