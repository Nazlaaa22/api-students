package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"uts-siakad/handlers"
	"uts-siakad/middleware"
)

func SetupMahasiswaRoutes(app *fiber.App, db *pgxpool.Pool) {
	api := app.Group("/api/v1")

	api.Get(
		"/students",
		middleware.AuthRequired(),
		middleware.RequireRole("admin"),
		handlers.GetMahasiswa(db),
	)

	api.Post(
		"/students",
		middleware.AuthRequired(),
		middleware.RequireRole("admin"),
		handlers.CreateMahasiswa(db),
	)

	app.Get(
		"/api/v1/students/:id",
		middleware.AuthRequired(),
		handlers.GetMahasiswaByID(db),
	)

	api.Put(
		"/students/:id",
		middleware.AuthRequired(),
		middleware.RequireRole("admin"),
		handlers.UpdateMahasiswa(db),
	)

	api.Delete(
		"/students/:id",
		middleware.AuthRequired(),
		middleware.RequireRole("admin"),
		handlers.DeleteMahasiswa(db),
	)
}
