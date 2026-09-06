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
	c *fiber.Ctx,
	status int,
	message string,
) error {
	return c.Status(status).JSON(fiber.Map{
		"data":    nil,
		"message": message,
		"success": false,
	})
}
