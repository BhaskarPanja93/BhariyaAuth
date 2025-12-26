package main

import (
	Middlewares "BhariyaAuth/middlewares"
	AccountProcessor "BhariyaAuth/processors/account"
	AccountRouters "BhariyaAuth/routers/account"
	LoginRouters "BhariyaAuth/routers/login"
	MFARouters "BhariyaAuth/routers/mfa"
	PasswordResetRouters "BhariyaAuth/routers/passwordreset"
	RegisterRouters "BhariyaAuth/routers/register"
	SessionRouters "BhariyaAuth/routers/sessions"
	SSORouters "BhariyaAuth/routers/sso"
	StatusRouters "BhariyaAuth/routers/status"
	Stores "BhariyaAuth/stores"
	"flag"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

func ReceiveCLIFlags(MainApp *fiber.App) {
	unixSocket := flag.String("bind", "", "Unix socket path (optional)")
	flag.Parse()
	if *unixSocket == "" {
		fmt.Println("No unix socket path provided. Fallback to port 3000.")
		StartOnPort(MainApp)
	} else {
		StartOnSocket(MainApp, *unixSocket)
	}
}

func StartOnSocket(MainApp *fiber.App, unixSocket string) {
	err := MainApp.Listen(unixSocket, fiber.ListenConfig{
		ListenerNetwork:    fiber.NetworkUnix,
		UnixSocketFileMode: 0760,
	})
	if err != nil {
		fmt.Println("Unable to start on network socket:", err.Error())
		StartOnPort(MainApp)
	}
}

func StartOnPort(MainApp *fiber.App) {
	fmt.Println("Attempt listen on port 3000")
	if err := MainApp.Listen(":3000"); err != nil {
		fmt.Println("Unable to start on port 3000:", err.Error())
	}
}

func main() {
	Stores.ConnectMySQL()
	Stores.ConnectRedis()
	go AccountProcessor.ServeAccountDetails()
	go AccountProcessor.DatabaseAutoVacuum()

	MainApp := fiber.New(fiber.Config{
		AppName:          "BhariyaAuth",
		ProxyHeader:      fiber.HeaderXForwardedFor,
		ReadBufferSize:   4 * 1024,
		WriteBufferSize:  4 * 1024,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		BodyLimit:        10 * 1024,
		TrustProxyConfig: fiber.TrustProxyConfig{Loopback: true},
		JSONEncoder:      json.Marshal,
		JSONDecoder:      json.Unmarshal,
	})

	MainApp.Use(Middlewares.ProfilingMiddleware())

	AuthApp := MainApp.Group("/auth-backend")

	AccountRouters.AttachRoutes(AuthApp)
	PasswordResetRouters.AttachRoutes(AuthApp)
	StatusRouters.AttachRoutes(AuthApp)
	RegisterRouters.AttachRoutes(AuthApp)
	LoginRouters.AttachRoutes(AuthApp)
	SSORouters.AttachRoutes(AuthApp)
	SessionRouters.AttachRoutes(AuthApp)
	MFARouters.AttachRoutes(AuthApp)

	ReceiveCLIFlags(MainApp)
}
