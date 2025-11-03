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
	CSRFInCookie         = fmt.Sprintf("%s_csrf", ServerFor)
	RefreshTokenInCookie = fmt.Sprintf("%s_refresh", ServerFor)
	SSOStateInCookie     = fmt.Sprintf("%s_sso_state", ServerFor)

	AccessTokenFreshnessExpireDelta = time.Minute * 5
	AccessTokenExpireDelta          = time.Minute * 15
	RefreshTokenExpireDelta         = time.Hour * 24 * 7

	RedisServerBase            = fmt.Sprintf("%s:%s", ServerFor, ServerRole)
	RedisServerOTPVerification = fmt.Sprintf("%s:verify", RedisServerBase)

	RedisServerRateLimits = fmt.Sprintf("%s:limit", RedisServerBase)
	RedisOTPRateLimit     = fmt.Sprintf("%s:otp", RedisServerRateLimits)

	RedisSharedBlacklist       = fmt.Sprintf("%s:ban", RedisServerBase)
	RedisUserIDBlacklist       = fmt.Sprintf("%s:user", RedisSharedBlacklist)
	RedisRefreshTokenBlacklist = fmt.Sprintf("%s:refresh", RedisSharedBlacklist)

	RedisZohoMailAccessToken = fmt.Sprintf("%s:token:zoho", ServerFor)
	RedisNewAccountChannel   = fmt.Sprintf("%s:new_account", ServerFor)
)
