package config

import (
	"api-students/middleware"

	"github.com/gofiber/fiber/v2"
)

func NewApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})
}
