package Config

import (
	"context"
	"os"
	"strings"
	"time"
)

func getEnvString(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

var (
	BetaFrontend = getEnvBool("AUTH_BETA_FRONTEND", false)
	BetaAPI      = getEnvBool("AUTH_BETA_API", false)
	BetaWS       = getEnvBool("AUTH_BETA_WS", false)
)

var (
	Domain      = getEnvString("AUTH_DOMAIN", "bhariya.ddns.net")
	Origin      = getEnvString("AUTH_ORIGIN", "https://"+Domain)
	Purpose     = getEnvString("AUTH_PURPOSE", "/auth")
	PurposeFull = Origin + Purpose
)

const (
	MFATokenInCookie     = "mfa"
	MFATokenInHeader     = MFATokenInCookie
	CSRFInHeader         = "csrf"
	CSRFInCookie         = CSRFInHeader
	AccessTokenInHeader  = "authorization"
	RefreshTokenInCookie = "refresh"
	SSOStateInCookie     = "sso"

	AccessTokenExpireDelta  = time.Minute * 10
	RefreshTokenExpireDelta = time.Hour * 24 * 7
	SSOCookieExpireDelta    = time.Minute * 10

	RefreshBlocked                = "auth:blocked"
	AccountDetailsRequestChannel  = "auth:acc:req"
	AccountDetailsResponseChannel = "auth:acc:res"

	ServerID = "auth"
)

var (
	FrontendPrefix = ""
	FrontendSuffix = func() string {
		if BetaFrontend {
			return "/beta"
		}
		return ""
	}()

	APIPrefix = "/api"
	APISuffix = func() string {
		if BetaAPI {
			return "/beta"
		}
		return ""
	}()

	WSPrefix = "/ws"
	WSSuffix = func() string {
		if BetaWS {
			return "/beta"
		}
		return ""
	}()

	FrontendRoute = PurposeFull + FrontendPrefix + FrontendSuffix
	APIRoute      = PurposeFull + APIPrefix + APISuffix
	WSRoute       = PurposeFull + WSPrefix + WSSuffix

	ServerSSOCallbackURL = APIRoute + "/sso/callback"

	CtxBG = context.Background()
)
