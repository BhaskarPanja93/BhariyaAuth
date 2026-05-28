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

func ProfilingMiddleware() fiber.Handler {
	return func(ctx fiber.Ctx) (err error) {

		defer func() {
			if r := recover(); r != nil {
				Logs.RootLogger.Add(Logs.Error, "middleware/profiling", "", fmt.Sprintf("Panic: %v\n%s", r, debug.Stack()))
				err = ctx.SendStatus(fiber.StatusInternalServerError)
			}
		}()

		path := ctx.Path()

		if strings.HasPrefix(path, "/auth/api/status") {
			return ctx.Next()
		}

		ctx.Set("X-Reached-API-At", time.Now().UTC().Format(time.StampNano))
		ctx.Set("X-URL-Path", path)

		requestID := RequestProcessor.SetRequestId(ctx)
		Logs.RootLogger.Add(Logs.Intent, "middleware/profiling", requestID, "Received request from "+ctx.IP()+" for path "+path)

		ctx.Set(fiber.HeaderXRequestID, requestID)

		processingStarted := RequestProcessor.SetRequestTime(ctx)
		err = ctx.Next()

		processingEnded := time.Now().UTC()

		processingDuration := processingEnded.Sub(processingStarted)

		if processingDuration > MaxResponseTimeAllowed {
			Logs.RootLogger.Add(Logs.Warn, "middleware/profiling", requestID, "Slow response: "+processingDuration.String())
		}

		ctx.Set("X-Rate-Limit-Weight", strconv.Itoa(int(RequestProcessor.GetRateLimitWeight(ctx))))
		ctx.Set("X-Left-API-At", processingEnded.Format(time.StampNano))
		ctx.Set("X-Time-Taken", processingDuration.String())
		return err
	}
}
