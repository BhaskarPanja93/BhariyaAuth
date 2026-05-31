package main

import (
	Middlewares "BhariyaAuth/middlewares"
	AccountProcessor "BhariyaAuth/processors/account"
	Logs "BhariyaAuth/processors/logs"
	AccountRouters "BhariyaAuth/routers/access"
	LogsRouters "BhariyaAuth/routers/logs"
	MailRouters "BhariyaAuth/routers/mail"
	MFARouters "BhariyaAuth/routers/mfa"
	PasswordResetRouters "BhariyaAuth/routers/passwordreset"
	SessionRouters "BhariyaAuth/routers/sessions"
	SignInRouters "BhariyaAuth/routers/signin"
	SignUpRouters "BhariyaAuth/routers/signup"
	SSORouters "BhariyaAuth/routers/sso"
	StatusRouters "BhariyaAuth/routers/status"
	Stores "BhariyaAuth/stores"
	"flag"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

func ReceiveCLIFlags(MainApp *fiber.App) {
	unixSocket := flag.String("bind", "", "Unix socket path (optional)")
	flag.Parse()

	Logs.RootLogger.Add(Logs.Info, "main", "", "Unix socket path received: "+*unixSocket)
	if *unixSocket == "" {
		Logs.RootLogger.Add(Logs.Warn, "main", "", "Unix socket path missing. Falling back to TCP")
		StartOnPort(MainApp)
	} else {
		StartOnSocket(MainApp, *unixSocket)
	}
}

func StartOnSocket(MainApp *fiber.App, unixSocket string) {
	Logs.RootLogger.Add(Logs.Intent, "main", "", "Attempting run on unix socket")
	err := MainApp.Listen(unixSocket, fiber.ListenConfig{
		ListenerNetwork:    fiber.NetworkUnix,
		UnixSocketFileMode: 0760,
	})
	if err != nil {
		Logs.RootLogger.Add(Logs.Error, "main", "", "Unix socket app start failed with error: "+err.Error())
		StartOnPort(MainApp)
	}
}

func StartOnPort(MainApp *fiber.App) {
	Logs.RootLogger.Add(Logs.Intent, "main", "", "Attempting run on port 3000")
	if err := MainApp.Listen(":3000"); err != nil {
		Logs.RootLogger.Add(Logs.Error, "main", "", "TCP app start failed with error: "+err.Error())
		Logs.RootLogger.Add(Logs.Error, "main", "", "No Other Methods remaining")
	}
}

func main() {
	Logs.RootLogger.Add(Logs.Info, "main", "", "Server startup")
	Stores.ConnectSQL()
	Stores.ConnectRedis()

	go AccountProcessor.ServeAccountDetails()
	go AccountProcessor.DatabaseAutoVacuum()
	SSORouters.AttachProviders()

	MainApp := fiber.New(fiber.Config{

		AppName: "BhariyaAuth",

		ProxyHeader: fiber.HeaderXForwardedFor,
		TrustProxy:  true,
		TrustProxyConfig: fiber.TrustProxyConfig{
			UnixSocket: true,
		},

		ReadBufferSize:  4 * 1024,
		WriteBufferSize: 4 * 1024,

		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,

		BodyLimit: 10 * 1024,

		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	APIGroup := MainApp.Group("/auth/api")

	Logs.RootLogger.Add(Logs.Intent, "main", "", "Attaching Profiling Middleware")
	APIGroup.Use(Middlewares.ProfilingMiddleware())

	Logs.RootLogger.Add(Logs.Intent, "main", "", "Attaching Routers")
	StatusRouters.AttachRoutes(APIGroup)
	AccountRouters.AttachRoutes(APIGroup)
	PasswordResetRouters.AttachRoutes(APIGroup)
	SignUpRouters.AttachRoutes(APIGroup)
	SignInRouters.AttachRoutes(APIGroup)
	SSORouters.AttachRoutes(APIGroup)
	SessionRouters.AttachRoutes(APIGroup)
	MFARouters.AttachRoutes(APIGroup)
	LogsRouters.AttachRoutes(APIGroup)
	MailRouters.AttachRoutes(APIGroup)

	Logs.RootLogger.Add(Logs.Intent, "main", "", "Reading app run parameters")
	ReceiveCLIFlags(MainApp)

	Logs.RootLogger.Add(Logs.Info, "main", "", "Shutting down server")
}
