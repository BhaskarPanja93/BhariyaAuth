package main

import (
	Middlewares "BhariyaAuth/middlewares"
	AccountProcessor "BhariyaAuth/processors/account"
	Logs "BhariyaAuth/processors/logs"
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
	"time"

	"github.com/bytedance/sonic"
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

	Logs.RootLogger.Add(Logs.Info, "main", "", "Unix socket path received: "+*unixSocket)
	if *unixSocket == "" {
		Logs.RootLogger.Add(Logs.Warn, "main", "", "Unix socket path missing. Falling back to TCP")
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

// StartOnPort starts the server on TCP port 3000.
//
// Overview:
// - Default fallback when no socket is provided.
// - Logs startup attempt and failure if any.
func StartOnPort(MainApp *fiber.App) {
	Logs.RootLogger.Add(Logs.Intent, "main", "", "Attempting run on port 3000")
	if err := MainApp.Listen(":3000"); err != nil {
		Logs.RootLogger.Add(Logs.Error, "main", "", "TCP app start failed with error: "+err.Error())
		Logs.RootLogger.Add(Logs.Error, "main", "", "No Other Methods remaining")
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
// - HTTP server runs in the main thread.
func main() {
	Logs.RootLogger.Add(Logs.Info, "main", "", "Server startup")
	Stores.ConnectSQL()   // blocking until DB is available
	Stores.ConnectRedis() // blocking until Redis is available

	go AccountProcessor.ServeAccountDetails() // async worker
	go AccountProcessor.DatabaseAutoVacuum()  // periodic DB cleanup
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

		BodyLimit: 10 * 1024, // max request body size (10 KB)

		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})

	APIGroup := MainApp.Group("/auth/api")

	// Global middleware for API routes
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

	// WSGroup := MainApp.Group("/auth/ws")
	// ChatRouters.AttachRoutes(WSGroup)

	Logs.RootLogger.Add(Logs.Intent, "main", "", "Reading app run parameters")
	ReceiveCLIFlags(MainApp)

	Logs.RootLogger.Add(Logs.Info, "main", "", "Shutting down server")
}
