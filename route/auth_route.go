package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/service"
	"api-students/middleware"
)

func SetupAuthRoutes(app *fiber.App, authService *service.AuthService) {
	api := app.Group("/api/v1/auth")

	// Endpoint authentication yang tidak membutuhkan access token
	api.Post("/register", authService.Register)
	api.Post("/login", middleware.LoginRateLimiter(), authService.Login)
	api.Post("/refresh", authService.Refresh)
	api.Post("/logout", authService.Logout)

	// Endpoint profil membutuhkan access token
	api.Get("/me", middleware.RequireAuth(), authService.Me)
}
