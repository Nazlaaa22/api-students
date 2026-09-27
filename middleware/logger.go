package middleware

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"api-students/helper"

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

		if err != nil {
			var appErr *helper.AppError
			var fiberErr *fiber.Error

			if errors.As(err, &appErr) {
				status = appErr.Status
			} else if errors.As(err, &fiberErr) {
				status = fiberErr.Code
			} else {
				status = fiber.StatusInternalServerError
			}
		}

		var appErr *helper.AppError
		var fiberErr *fiber.Error

		if errors.As(err, &appErr) {
			status = appErr.Status
		} else if errors.As(err, &fiberErr) {
			status = fiberErr.Code
		} else if err != nil {
			status = fiber.StatusInternalServerError
		}

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
