package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"uts-siakad/handlers"
)

func SetupDosenRoutes(app *fiber.App, db *pgxpool.Pool) {
	app.Get("/dosen", handlers.GetAllDosen(db))
}
