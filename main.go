package main

import (
	Middlewares "BhariyaAuth/middlewares"
	AccountProcessor "BhariyaAuth/processors/account"
	AccountRouters "BhariyaAuth/routers/account"
	LoginRouters "BhariyaAuth/routers/login"
	PasswordResetRouters "BhariyaAuth/routers/passwordreset"
	RegisterRouters "BhariyaAuth/routers/register"
	SessionRouters "BhariyaAuth/routers/sessions"
	SSORouters "BhariyaAuth/routers/sso"
	StatusRouters "BhariyaAuth/routers/status"
	Stores "BhariyaAuth/stores"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func ReceiveCLIFlags(MainApp *fiber.App) {
	unixSocket := flag.String("bind", "", "Unix socket path (optional)")
	flag.Parse()
	if *unixSocket == "" {
		fmt.Println("No unix socket path provided. Fallback to port 3000.")
		StartOnPort(MainApp)
	} else {
		StartOnSocket(MainApp, unixSocket)
	}
}

func DeleteResidualSocket(unixSocket *string) bool {
	if _, err := os.Stat(*unixSocket); err == nil {
		fmt.Println("Unix socket path exists, trying to delete.")
		err = os.Remove(*unixSocket)
		if err != nil {
			fmt.Println("Failed to delete:", err.Error())
			return false
		}
	}
	return true
}

func StartOnSocket(MainApp *fiber.App, unixSocket *string) {
	if DeleteResidualSocket(unixSocket) {
		listener, err := net.Listen("unix", *unixSocket)
		if err != nil {
			fmt.Println("Unable to utilise unix socket")
			StartOnPort(MainApp)
		} else {
			if chmodErr := os.Chmod(*unixSocket, 0760); chmodErr != nil {
				fmt.Println("Warning: failed to set unix socket permissions:", chmodErr.Error())
				StartOnPort(MainApp)
			} else {
				fmt.Println("Listener announced on unix socket:", *unixSocket)
				fmt.Println("Unix socket permission set to 760")
				if err = MainApp.Listener(listener); err != nil {
					fmt.Println("Unable to start server on unix socket:", err.Error())
					StartOnPort(MainApp)
				}
			}
		}
	} else {
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

	MainApp := fiber.New(fiber.Config{
		AppName:          "BhariyaAuth",
		ProxyHeader:      fiber.HeaderXForwardedFor,
		ReadBufferSize:   8 * 1024,
		WriteBufferSize:  8 * 1024,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		BodyLimit:        10 * 1024,
		TrustProxyConfig: fiber.TrustProxyConfig{Loopback: true},
	})

	MainApp.Use(Middlewares.ProfilingMiddleware())
	MainApp.Use(recover.New())

	AuthApp := MainApp.Group("/auth")

	AccountRouters.AttachRoutes(AuthApp)
	PasswordResetRouters.AttachRoutes(AuthApp)
	StatusRouters.AttachRoutes(AuthApp)
	RegisterRouters.AttachRoutes(AuthApp)
	LoginRouters.AttachRoutes(AuthApp)
	SSORouters.AttachRoutes(AuthApp)
	SessionRouters.AttachRouters(AuthApp)

	ReceiveCLIFlags(MainApp)
}
