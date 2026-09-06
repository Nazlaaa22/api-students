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

		logData := map[string]interface{}{
			"request_id": requestID,
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
