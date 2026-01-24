package config

import (
	"fmt"
	"time"
)

var (
	ServerFor            = "bhariya"
	ServerRole           = "auth"
	CookieDomain         = "bhariya.ddns.net"
	FrontendURL          = fmt.Sprintf("https://%s/auth", CookieDomain)
	ServerSSOCallbackURL = fmt.Sprintf("https://%s/auth-backend/sso/callback", CookieDomain)
)

var (
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

	RedisServerBase            = fmt.Sprintf("%s:%s", ServerFor, ServerRole)
	RedisServerOTPVerification = fmt.Sprintf("%s:verify", RedisServerBase)

	AccountDetailsRequestChannel  = "auth:account:request"
	AccountDetailsResponseChannel = "auth:account:response"

	MaxUserSessions = 50
)
