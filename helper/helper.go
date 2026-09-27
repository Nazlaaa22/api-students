package helper

import "github.com/gofiber/fiber/v2"

func SendSuccess(
	c *fiber.Ctx,
	status int,
	message string,
	data interface{},
) error {
	return c.Status(status).JSON(fiber.Map{
		"data":    data,
		"message": message,
		"success": true,
	})
}

func SendError(
	_ *fiber.Ctx,
	status int,
	message string,
) error {
	code := CodeInternal

	switch status {
	case fiber.StatusBadRequest:
		code = CodeBadRequest
	case fiber.StatusUnauthorized:
		code = CodeUnauthorized
	case fiber.StatusForbidden:
		code = CodeForbidden
	case fiber.StatusNotFound:
		code = CodeNotFound
	case fiber.StatusConflict:
		code = CodeConflict
	case fiber.StatusUnprocessableEntity:
		code = CodeValidation
	case fiber.StatusUnsupportedMediaType:
		code = CodeUnsupportedMedia
	case fiber.StatusNotAcceptable:
		code = CodeNotAcceptable
	case fiber.StatusTooManyRequests:
		code = CodeTooManyRequests
	}

	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}
