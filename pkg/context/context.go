package context

import (
	"context"
	"github.com/gofiber/fiber/v2"
)

const RequestIDKey = "request_id"

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	requestID, ok := ctx.Value(RequestIDKey).(string)
	if !ok || requestID == "" {
		return "unknown"
	}
	return requestID
}

func FromFiberCtx(c *fiber.Ctx) context.Context {
	ctx := context.Background()

	requestID, ok := c.Locals("X-Request-ID").(string)
	if !ok || requestID == "" {
		requestID = c.Get("X-Request-ID")

		if requestID == "" {
			requestID = "unknown"
		}
	}

	return WithRequestID(ctx, requestID)
}
