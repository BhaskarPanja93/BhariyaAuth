package middlewares

import (
	Logs "BhariyaAuth/processors/logs"
	RequestProcessor "BhariyaAuth/processors/request"
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

const (
	MaxResponseTimeAllowed = 5 * time.Second
)

// ProfilingMiddleware attaches request-level metadata for tracing and performance monitoring.
//
// This middleware instruments each incoming request by:
//  1. Assigning a unique request ID.
//  2. Recording request start time.
//  3. Propagating metadata via headers and context (Locals).
//  4. Measuring total request duration after processing.
//  5. Attaching response timing metadata to headers.
//  6. Flagging slow requests for logging/analysis.
//
// Flow Summary:
//
//	generate metadata → attach headers → store context → process request → measure duration → attach response headers
//
// Observability Features:
// - Unique Request ID for tracing across services/logs.
// - Start/End timestamps for debugging latency.
// - Total execution time exposed to client.
// - Threshold-based slow request detection.
//
// Headers Added:
// - Reached-API: indicates request reached application layer.
// - Requested-Start: request start timestamp (UTC).
// - Requested-ID: unique request identifier.
// - Request-End: request end timestamp (UTC).
// - Time-Taken: total request duration.
//
// Constants:
// - MaxResponseTimeAllowed: threshold for identifying slow requests.
//
// Returns:
// - Proceeds with next middleware/handler and enriches response with profiling metadata.
func ProfilingMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) (err error) {

		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				Logs.RootLogger.Add(Logs.Error, "middleware/profiling", "", fmt.Sprintf("Panic: %v\n%s", r, debug.Stack()))
				// Send 500 response
				err = ctx.SendStatus(fiber.StatusInternalServerError)
			}
		}()

		path := ctx.Path()

		// Skip if request belongs to API status router
		if strings.HasPrefix(path, "/auth/api/status") {
			return ctx.Next()
		}

		// Request start time will be sent in the client's header
		ctx.Set("X-Reached-API-At", time.Now().UTC().Format(time.StampNano))
		ctx.Set("X-URL-Path", path)

		// Set a request identifier for tracing
		requestID := RequestProcessor.SetRequestId(ctx)
		Logs.RootLogger.Add(Logs.Intent, "middleware/profiling", requestID, "Received request from "+ctx.IP()+" for path "+path)

		// Attach request metadata to response headers (early visibility)
		ctx.Set(fiber.HeaderXRequestID, requestID)

		processingStarted := RequestProcessor.SetRequestTime(ctx)
		// Execute next middleware/handler in chain
		err = ctx.Next()

		// Capture request end time
		processingEnded := time.Now().UTC()

		// Calculate total duration
		processingDuration := processingEnded.Sub(processingStarted)

		// Detect slow requests exceeding configured threshold
		if processingDuration > MaxResponseTimeAllowed {
			Logs.RootLogger.Add(Logs.Warn, "middleware/profiling", requestID, "Slow response: "+processingDuration.String())
		}

		// Attach response timing metadata
		ctx.Set("X-Rate-Limit-Weight", strconv.Itoa(int(RequestProcessor.GetRateLimitWeight(ctx))))
		ctx.Set("X-Left-API-At", processingEnded.Format(time.StampNano))
		ctx.Set("X-Time-Taken", processingDuration.String())
		return err
	}
}
