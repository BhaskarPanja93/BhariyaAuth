package main

import (
	Middlewares "BhariyaAuth/middlewares"
	AccountProcessor "BhariyaAuth/processors/account"
	AccountRouters "BhariyaAuth/routers/access"
	MFARouters "BhariyaAuth/routers/mfa"
	PasswordResetRouters "BhariyaAuth/routers/passwordreset"
	SessionRouters "BhariyaAuth/routers/sessions"
	SignInRouters "BhariyaAuth/routers/signin"
	SignUpRouters "BhariyaAuth/routers/signup"
	SSORouters "BhariyaAuth/routers/sso"
	StatusRouters "BhariyaAuth/routers/status"
	Stores "BhariyaAuth/stores"
	"flag"
	"fmt"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

// ReceiveCLIFlags parses CLI flags and determines how the server should bind.
//
// Overview:
// - Supports optional UNIX socket binding via `-bind` flag.
// - Falls back to TCP port 3000 if not provided.
//
// Flow:
//
//	parse flags → check socket → start server accordingly
//
// Parameters:
// - MainApp: initialized Fiber application instance.
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

// StartOnSocket starts the server on a UNIX domain socket.
//
// Overview:
// - Uses fiber's Unix socket listener.
// - Sets file permissions to 0760.
//
// Behavior:
// - Falls back to TCP port if socket binding fails.
//
// Parameters:
// - unixSocket: filesystem path to socket file.
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

// StartOnPort starts the server on TCP port 3000.
//
// Overview:
// - Default fallback when no socket is provided.
// - Logs startup attempt and failure if any.
func StartOnPort(MainApp *fiber.App) {
	fmt.Println("Attempt listen on port 3000")
	if err := MainApp.Listen(":3000"); err != nil {
		fmt.Println("Unable to start on port 3000:", err.Error())
	}
}

// main is the entry point of the application.
//
// Startup Sequence:
//  1. Initialize database (PostgreSQL).
//  2. Initialize Redis.
//  3. Start background workers.
//  4. Configure Fiber app.
//  5. Register routes and middleware.
//  6. Start server (socket or port).
//
// Concurrency:
// - Background workers run as goroutines.
// - HTTP server runs in main thread.
func main() {

	Stores.ConnectSQL()   // blocking until DB is available
	Stores.ConnectRedis() // blocking until Redis is available

	go AccountProcessor.ServeAccountDetails() // async worker
	go AccountProcessor.DatabaseAutoVacuum()  // periodic DB cleanup

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

		BodyLimit: 10 * 1024, // max request body size (10 KB)

		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	APIGroup := MainApp.Group("/auth/api")

	// Global middleware for API routes
	APIGroup.Use(Middlewares.ProfilingMiddleware())

	StatusRouters.AttachRoutes(APIGroup)
	AccountRouters.AttachRoutes(APIGroup)
	PasswordResetRouters.AttachRoutes(APIGroup)
	SignUpRouters.AttachRoutes(APIGroup)
	SignInRouters.AttachRoutes(APIGroup)
	SSORouters.AttachRoutes(APIGroup)
	SessionRouters.AttachRoutes(APIGroup)
	MFARouters.AttachRoutes(APIGroup)

	// WSGroup := MainApp.Group("/auth/ws")
	// ChatRouters.AttachRoutes(WSGroup)

	ReceiveCLIFlags(MainApp)
}
