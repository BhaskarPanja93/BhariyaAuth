package sso

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
)

func init() {
	// Initialize all providers usable for SSO
	goth.UseProviders(
		google.New(Secrets.GoogleClientId,
			Secrets.GoogleClientSecret,
			Config.ServerSSOCallbackURL+"/google", "profile", "email"),
		discord.New(Secrets.DiscordClientId,
			Secrets.DiscordClientSecret,
			Config.ServerSSOCallbackURL+"/discord", "identify", "email", "openid"),
		microsoftonline.New(Secrets.MicrosoftClientId,
			Secrets.MicrosoftClientSecret,
			Config.ServerSSOCallbackURL+"/microsoftonline", "user.read"))
}
