package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"api-students/app/repository"
	"api-students/database"
)

func main() {
	// Membaca konfigurasi dari file .env
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, menggunakan environment variable sistem")
	}

	// Membuat koneksi ke PostgreSQL
	db, err := database.NewPostgresPool()
	if err != nil {
		log.Fatal("Gagal terhubung ke PostgreSQL:", err)
	}

	defer db.Close()

	log.Println("Berhasil terhubung ke PostgreSQL")

	// Membuat repository
	studentRepo := repository.NewStudentRepository(db)

	// Membuat handler
	handler := NewHandler(studentRepo)

	// Membuat aplikasi Fiber
	app := fiber.New()

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Ping(c.Context()); err != nil {
			return sendError(
				c,
				503,
				"PostgreSQL tidak dapat dihubungi",
			)
		}

		return sendSuccess(
			c,
			200,
			"API dan PostgreSQL berjalan normal",
			fiber.Map{
				"database": "connected",
			},
		)
	})

	// API group
	api := app.Group("/api/v1")

	// Student endpoints
	api.Get("/students", handler.GetStudents)
	api.Get("/students/:id", handler.GetStudent)
	api.Post("/students", handler.CreateStudent)
	api.Put("/students/:id", handler.UpdateStudent)
	api.Patch("/students/:id", handler.PatchStudent)
	api.Delete("/students/:id", handler.DeleteStudent)

	// Menjalankan server
	log.Println("Server berjalan di http://127.0.0.1:3000")

	log.Fatal(app.Listen(":3000"))
}
