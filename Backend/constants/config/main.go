package Config

import (
	"context"
	"time"
)

const (
	BetaFrontend = false
	BetaAPI      = false
	BetaWS       = false
)

const (
	Domain      = "bhariya.ddns.net"
	Origin      = "https://" + Domain
	Purpose     = "/auth"
	PurposeFull = Origin + Purpose
)

const (
	CSRFInHeader         = "csrf"
	MFATokenInHeader     = "mfa"
	CSRFInCookie         = CSRFInHeader
	MFATokenInCookie     = MFATokenInHeader
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
