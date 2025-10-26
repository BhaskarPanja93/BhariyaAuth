package token

import (
	Important "BhariyaAuth/constants/config"
	TokenModels "BhariyaAuth/models/tokens"
	UserTypes "BhariyaAuth/models/users"
	StringProcessor "BhariyaAuth/processors/string"
	Stores "BhariyaAuth/stores"
	"time"

	"fmt"
	"strings"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

func BlacklistRefresh(userID uint32, refreshID uint16, deep bool) {
	Stores.RedisClient.Set(Stores.Ctx, fmt.Sprintf("%s:%d:%d", Important.RedisRefreshTokenBlacklist, userID, refreshID), 1, Important.AccessTokenExpireDelta)
	if deep {
		_, err := Stores.MySQLClient.Exec("UPDATE activities SET blocked = ? WHERE uid = ? AND refresh = ?", true, userID, refreshID)
		if err != nil {
			return
		}
	}
}

func RefreshIsBlacklisted(userID uint32, refreshID uint16, fromDB bool) bool {
	key := fmt.Sprintf("%s:%d:%d", Important.RedisRefreshTokenBlacklist, userID, refreshID)
	if !fromDB {
		exists, err := Stores.RedisClient.Exists(Stores.Ctx, key).Result()
		if err != nil {
			return true
		}
		return exists > 0
	}
	var blocked bool
	err := Stores.MySQLClient.QueryRow("SELECT blocked FROM activities WHERE uid = ? AND refresh = ?", userID, refreshID).Scan(&blocked)
	if err != nil {
		return true
	}
	return blocked
}

func CreateFreshToken(userID uint32, refreshID uint16, userType UserTypes.T, remember bool, identifier string, identifierType string) TokenModels.NewTokenT {
	csrf := StringProcessor.GenerateSafeString(256)
	atUnEnc, err := json.Marshal(TokenModels.AccessTokenT{
		UserID:         userID,
		RefreshID:      refreshID,
		RefreshIndex:   1,
		UserType:       userType,
		AccessCreated:  time.Now(),
		RefreshCreated: time.Now(),
		RememberMe:     remember,
	})
	if err != nil {
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
		IdentifierUsed: identifier,
		IdentifierType: identifierType,
	})
	if err != nil {
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
	csrf := StringProcessor.GenerateSafeString(256)
	atUnEnc, err := json.Marshal(TokenModels.AccessTokenT{
		UserID:         refresh.UserID,
		RefreshID:      refresh.RefreshID,
		RefreshIndex:   refresh.RefreshIndex + 1,
		UserType:       refresh.UserType,
		AccessCreated:  time.Now(),
		RefreshCreated: refresh.RefreshCreated,
		RememberMe:     refresh.RememberMe,
	})
	if err != nil {
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
		IdentifierUsed: refresh.IdentifierUsed,
		IdentifierType: refresh.IdentifierType,
	})
	if err != nil {
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
	header := ctx.Get(Important.AccessTokenInHeader)
	header = strings.TrimPrefix(header, "Bearer ")
	header = strings.TrimSpace(header)
	access := TokenModels.AccessTokenT{}
	dec, ok := StringProcessor.Decrypt(header)
	if ok {
		if err := json.Unmarshal(dec, &access); err != nil {
		}
	}
	return access
}

func ReadRefreshToken(ctx fiber.Ctx) TokenModels.RefreshTokenT {
	cookie := ctx.Cookies(Important.RefreshTokenInCookie)
	cookie = strings.TrimSpace(cookie)
	refresh := TokenModels.RefreshTokenT{}
	dec, ok := StringProcessor.Decrypt(cookie)
	if ok {
		if err := json.Unmarshal(dec, &refresh); err != nil {
		}
	}
	return refresh
}

func MatchCSRF(ctx fiber.Ctx, refresh TokenModels.RefreshTokenT) bool {
	cookie := ctx.Cookies(Important.CSRFInCookie)
	header := ctx.Get(Important.CSRFInHeader)
	cookie = strings.TrimSpace(cookie)
	header = strings.TrimSpace(header)
	return refresh.CSRF == header && refresh.CSRF == cookie
}
