package main

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func sendSuccess(c *fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

func sendError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

func parseID(id string) (int, error) {
	return strconv.Atoi(id)
}
