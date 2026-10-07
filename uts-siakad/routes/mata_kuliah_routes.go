package routes

import (
	"uts-siakad/handlers"
	"uts-siakad/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupMataKuliahRoutes(app *fiber.App, db *pgxpool.Pool) {

	api := app.Group("/api/v1")

	api.Get(
		"/courses",
		middleware.AuthRequired(),
		handlers.GetMataKuliah(db),
	)
}
