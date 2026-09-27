package middleware

import (
	"errors"

	"api-students/app/model"
	"api-students/helper"

	"github.com/gofiber/fiber/v2"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	// Ambil request ID dari response header yang dibuat middleware logger.
	requestID := c.GetRespHeader("X-Request-ID")

	// Jika error berasal dari aplikasi.
	var appErr *helper.AppError
	if errors.As(err, &appErr) {
		return c.Status(appErr.Status).JSON(model.ErrorResponse{
			Success:   false,
			Code:      appErr.Code,
			Message:   appErr.Message,
			Fields:    appErr.Fields,
			RequestID: requestID,
		})
	}

	// Default untuk error yang tidak dikenal.
	status := fiber.StatusInternalServerError
	code := helper.CodeInternal
	message := "terjadi kesalahan pada server"

	// Tangani error bawaan Fiber.
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		status = fiberErr.Code

		switch status {
		case fiber.StatusBadRequest:
			code = helper.CodeBadRequest
			message = fiberErr.Message

		case fiber.StatusUnauthorized:
			code = helper.CodeUnauthorized
			message = fiberErr.Message

		case fiber.StatusForbidden:
			code = helper.CodeForbidden
			message = fiberErr.Message

		case fiber.StatusNotFound:
			code = helper.CodeNotFound
			message = fiberErr.Message

		case fiber.StatusConflict:
			code = helper.CodeConflict
			message = fiberErr.Message

		case fiber.StatusUnsupportedMediaType:
			code = helper.CodeUnsupportedMedia
			message = fiberErr.Message

		case fiber.StatusNotAcceptable:
			code = helper.CodeNotAcceptable
			message = fiberErr.Message

		case fiber.StatusTooManyRequests:
			code = helper.CodeTooManyRequests
			message = fiberErr.Message

		default:
			if status < 500 {
				code = helper.CodeBadRequest
				message = fiberErr.Message
			}
		}
	}

	return c.Status(status).JSON(model.ErrorResponse{
		Success:   false,
		Code:      code,
		Message:   message,
		RequestID: requestID,
	})
}
