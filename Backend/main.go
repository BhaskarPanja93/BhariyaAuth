package main

import (
	Middlewares "BhariyaAuth/middlewares"
	AccountProcessor "BhariyaAuth/processors/account"
	AccountRouters "BhariyaAuth/routers/account"
	ChatRouters "BhariyaAuth/routers/chat"
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
	Stores.ConnectSQL()
	Stores.ConnectRedis()
	go AccountProcessor.ServeAccountDetails()
	go AccountProcessor.DatabaseAutoVacuum()

	MainApp := fiber.New(fiber.Config{
		AppName:          "BhariyaAuth",
		ProxyHeader:      fiber.HeaderXForwardedFor,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		BodyLimit:        10 * 1024,
		TrustProxy:       true,
		TrustProxyConfig: fiber.TrustProxyConfig{Loopback: true},
		JSONEncoder:      json.Marshal,
		JSONDecoder:      json.Unmarshal,
	})

	APIGroup := MainApp.Group("/auth/api")
	APIGroup.Use(Middlewares.ProfilingMiddleware())

	WSGroup := MainApp.Group("/auth/ws")

	StatusRouters.AttachRoutes(APIGroup)
	AccountRouters.AttachRoutes(APIGroup)
	PasswordResetRouters.AttachRoutes(APIGroup)
	RegisterRouters.AttachRoutes(APIGroup)
	LoginRouters.AttachRoutes(APIGroup)
	SSORouters.AttachRoutes(APIGroup)
	SessionRouters.AttachRoutes(APIGroup)
	MFARouters.AttachRoutes(APIGroup)
	ChatRouters.AttachRoutes(WSGroup)

	ReceiveCLIFlags(MainApp)
}
