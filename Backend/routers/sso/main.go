package sso

import (
	Config "BhariyaAuth/constants/config"
	Secrets "BhariyaAuth/constants/secrets"
	Logs "BhariyaAuth/processors/logs"

	"github.com/gofiber/fiber/v3"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/discord"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/microsoftonline"
)

var (
	googleProvider = google.New(
		Secrets.GoogleClientId,
		Secrets.GoogleClientSecret,
		Config.ServerSSOCallbackURL+"/google",
		"profile", "email")

	discordProvider = discord.New(
		Secrets.DiscordClientId,
		Secrets.DiscordClientSecret,
		Config.ServerSSOCallbackURL+"/discord",
		"identify", "email", "openid")

	microsoftonlineProvider = microsoftonline.New(
		Secrets.MicrosoftClientId,
		Secrets.MicrosoftClientSecret,
		Config.ServerSSOCallbackURL+"/microsoftonline",
		"user.read")
)

func AttachProviders() {
	Logs.RootLogger.Add(Logs.Intent, "routers/sso/main", "", "Initializing goth providers")

	// Initialize all providers usable for SSO
	goth.UseProviders(googleProvider, discordProvider, microsoftonlineProvider)
}

func AttachRoutes(APIGroup fiber.Router) {
	Logs.RootLogger.Add(Logs.Intent, "routers/sso/main", "", "Attaching SSO Routes")

	SSORouter := APIGroup.Group("/sso")
	SSORouter.Get("/:"+ProviderParam, Step1)
	SSORouter.Get("/callback/:"+ProviderParam, Step2)
}
