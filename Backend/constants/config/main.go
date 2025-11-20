package config

import (
	"fmt"
	"time"
)

var (
	ServerFor            = "bhariya"
	ServerRole           = "auth"
	CookieDomain         = "bhariya-hidden.ddns.net"
	ServerSSOCallbackURL = "https://bhariya-hidden.ddns.net/auth/sso/callback"
)

var (
	CSRFInHeader         = "CSRF"
	AccessTokenInHeader  = "Authorization"
	ServerName           = fmt.Sprintf("%s_%s", ServerFor, ServerRole)
	MFATokenInCookie     = fmt.Sprintf("%s_mfa", ServerName)
	CSRFInCookie         = fmt.Sprintf("%s_csrf", ServerName)
	RefreshTokenInCookie = fmt.Sprintf("%s_refresh", ServerName)
	SSOStateInCookie     = fmt.Sprintf("%s_sso", ServerName)

	AccessTokenExpireDelta  = time.Minute * 10
	RefreshTokenExpireDelta = time.Hour * 24 * 7
	SSOCookieExpireDelta    = time.Minute * 10

	RedisServerBase            = fmt.Sprintf("%s:%s", ServerFor, ServerRole)
	RedisServerOTPVerification = fmt.Sprintf("%s:verify", RedisServerBase)

	AccountDetailsRequestChannel  = "auth:account:request"
	AccountDetailsResponseChannel = "auth:account:response"

	MaxUserSessions = 10
)
