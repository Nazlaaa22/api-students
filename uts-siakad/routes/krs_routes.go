package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"uts-siakad/handlers"
	"uts-siakad/middleware"
)

func SetupKRSRoutes(app *fiber.App, db *pgxpool.Pool) {
	app.Get("/krs", handlers.GetAllKRS(db))

	api := app.Group("/api/v1")

	api.Post(
		"/enrollments",
		middleware.AuthRequired(),
		middleware.RequireRole("mahasiswa"),
		handlers.CreateEnrollment(db),
	)

	api.Delete(
		"/enrollments/:id",
		middleware.AuthRequired(),
		handlers.DeleteEnrollment(db),
	)
}
