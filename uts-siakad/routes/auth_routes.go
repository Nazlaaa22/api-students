package routes

import (
	"uts-siakad/handlers"
	"uts-siakad/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupAuthRoutes(app *fiber.App, db *pgxpool.Pool) {
	app.Post("/api/v1/auth/login", handlers.Login(db))

	app.Get(
		"/api/v1/auth/me",
		middleware.AuthRequired(),
		handlers.Me(db),
	)
}
