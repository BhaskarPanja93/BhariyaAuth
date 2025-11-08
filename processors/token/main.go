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

func CreateFreshToken(userID uint32, refreshID uint16, userType UserTypes.T, remember bool, identifierType string) TokenModels.NewTokenT {
	csrf := Generators.SafeString(128)
	atUnEnc, err := json.Marshal(TokenModels.AccessTokenT{
		UserID:         userID,
		RefreshID:      refreshID,
		UserType:       userType,
		AccessCreated:  time.Now(),
		RefreshCreated: time.Now(),
		RememberMe:     remember,
	})
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Access Marshal failed: %s", err.Error()))
		return TokenModels.NewTokenT{}
	}
	atEnc, _ := StringProcessor.Encrypt(atUnEnc)
	rtUnEnc, err := json.Marshal(TokenModels.RefreshTokenT{
		UserID:         userID,
		RefreshID:      refreshID,
		RefreshIndex:   1,
		RefreshCreated: time.Now(),
		RefreshUpdated: time.Now(),
		UserType:       userType,
		CSRF:           csrf,
		RememberMe:     remember,
		IdentifierType: identifierType,
	})
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Refresh Marshal failed: %s", err.Error()))
		return TokenModels.NewTokenT{}
	}
	rtEnc, _ := StringProcessor.Encrypt(rtUnEnc)
	return TokenModels.NewTokenT{
		AccessToken:  atEnc,
		RefreshToken: rtEnc,
		CSRF:         csrf,
		RefreshID:    refreshID,
		RefreshIndex: 1,
		RememberMe:   remember,
	}
}

func CreateRenewToken(refresh TokenModels.RefreshTokenT) TokenModels.NewTokenT {
	csrf := Generators.SafeString(128)
	atUnEnc, err := json.Marshal(TokenModels.AccessTokenT{
		UserID:         refresh.UserID,
		RefreshID:      refresh.RefreshID,
		UserType:       refresh.UserType,
		AccessCreated:  time.Now(),
		RefreshCreated: refresh.RefreshCreated,
		RememberMe:     refresh.RememberMe,
	})
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Access Marshal failed: %s", err.Error()))
		return TokenModels.NewTokenT{}
	}
	atEnc, _ := StringProcessor.Encrypt(atUnEnc)
	rtUnEnc, err := json.Marshal(TokenModels.RefreshTokenT{
		UserID:         refresh.UserID,
		RefreshID:      refresh.RefreshID,
		RefreshIndex:   refresh.RefreshIndex + 1,
		RefreshUpdated: time.Now(),
		RefreshCreated: refresh.RefreshCreated,
		UserType:       refresh.UserType,
		CSRF:           csrf,
		RememberMe:     refresh.RememberMe,
		IdentifierType: refresh.IdentifierType,
	})
	if err != nil {
		Logger.AccidentalFailure(fmt.Sprintf("Refresh Marshal failed: %s", err.Error()))
		return TokenModels.NewTokenT{}
	}
	rtEnc, _ := StringProcessor.Encrypt(rtUnEnc)
	return TokenModels.NewTokenT{
		AccessToken:  atEnc,
		RefreshToken: rtEnc,
		CSRF:         csrf,
		RefreshID:    refresh.RefreshID,
		RefreshIndex: refresh.RefreshIndex + 1,
		RememberMe:   refresh.RememberMe,
	}
}

func ReadAccessToken(ctx fiber.Ctx) TokenModels.AccessTokenT {
	access := TokenModels.AccessTokenT{}
	header := strings.TrimSpace(strings.TrimPrefix(ctx.Get(Config.AccessTokenInHeader), "Bearer "))
	dec, ok := StringProcessor.Decrypt(header)
	if ok {
		_ = json.Unmarshal(dec, &access)
	} else {
		Logger.AccidentalFailure(fmt.Sprintf("Access Decrypt failed length: %d", len(header)))
	}
	return access
}

func ReadRefreshToken(ctx fiber.Ctx) TokenModels.RefreshTokenT {
	refresh := TokenModels.RefreshTokenT{}
	cookie := strings.TrimSpace(ctx.Cookies(Config.RefreshTokenInCookie))
	dec, ok := StringProcessor.Decrypt(cookie)
	if ok {
		_ = json.Unmarshal(dec, &refresh)
	} else {
		Logger.AccidentalFailure(fmt.Sprintf("Refresh Decrypt failed length: %d", len(cookie)))
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
