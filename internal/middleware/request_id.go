package middleware

import (
	"ProjectGolang/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"time"
)

const RequestIDKey = "X-Request-ID"

func NewRequestIDMiddleware() fiber.Handler {
	utilsInstance := utils.New()

	return func(c *fiber.Ctx) error {
		requestID := c.Get(RequestIDKey)

		if requestID == "" {
			requestID, _ = utilsInstance.NewULIDFromTimestamp(time.Now())
		}

		c.Locals(RequestIDKey, requestID)
		c.Set(RequestIDKey, requestID)

		return c.Next()
	}
}
