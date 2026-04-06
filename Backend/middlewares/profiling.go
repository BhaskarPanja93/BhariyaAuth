package middlewares

import (
	StringProcessor "BhariyaAuth/processors/string"
	"fmt"
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
// Context (Locals):
// - "request-start": time.Time → used across handlers for consistent timing.
// - "request-id": string → used for logging/tracing.
//
// Constants:
// - MaxResponseTimeAllowed: threshold for identifying slow requests.
//
// Returns:
// - Proceeds with next middleware/handler and enriches response with profiling metadata.
func ProfilingMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) error {

		// Capture request start time in UTC for consistency across systems
		received := time.Now().UTC()

		// Mark that request has reached API layer (useful for debugging proxies/CDNs)
		ctx.Set("X-Reached-API", "yes")

		// Generate a short, safe request identifier for tracing
		requestID := StringProcessor.SafeString(3)

		// Attach request metadata to response headers (early visibility)
		ctx.Set("X-Request-Start", fmt.Sprintf("%v", received))
		ctx.Set("X-Request-ID", requestID)

		// Store metadata in request context for downstream handlers
		ctx.Locals("request-id", requestID)
		ctx.Locals("request-start", time.Now().UTC())

		// Execute next middleware/handler in chain
		err := ctx.Next()

		// Capture request end time
		end := time.Now().UTC()

		// Calculate total duration
		dur := end.Sub(received)

		// Detect slow requests exceeding configured threshold
		if dur > MaxResponseTimeAllowed {
			// Placeholder for logging/alerting mechanism
		}

		// Attach response timing metadata
		ctx.Set("X-Response-End", fmt.Sprintf("%v", end))
		ctx.Set("X-Response-Time", fmt.Sprintf("%v", dur))
		return err
	}
}
