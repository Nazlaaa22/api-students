package main

import "github.com/gofiber/fiber/v2"

func sendSuccess(
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

func sendError(
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
