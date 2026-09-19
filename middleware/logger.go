package middleware

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequestLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		requestID := uuid.New().String()
		c.Set("X-Request-ID", requestID)

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		// Ambil user_id dan role dari JWT middleware.
		userID := c.Locals("user_id")
		role := c.Locals("role")

		logData := map[string]interface{}{
			"request_id": requestID,
			"user_id":    userID,
			"role":       role,
			"method":     c.Method(),
			"path":       c.OriginalURL(),
			"status":     status,
			"duration":   duration.String(),
		}

		data, marshalErr := json.Marshal(logData)

		if marshalErr == nil {
			log.Println(string(data))
		}

		return err
	}
}
