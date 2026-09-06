package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/middleware"
	"api-students/route"
)

func main() {
	// Memuat environment variable dari file .env
	if err := godotenv.Load(); err != nil {
		log.Println("File .env tidak ditemukan, menggunakan environment variable sistem")
	}

	// Mengaktifkan logger ke terminal dan logs/app.log
	config.SetupLogger()

	// Membuat koneksi ke PostgreSQL
	db, err := database.NewPostgresPool()
	if err != nil {
		log.Fatal("Gagal terhubung ke PostgreSQL:", err)
	}
	defer db.Close()

	log.Println("Berhasil terhubung ke PostgreSQL")

	// Membuat aplikasi Fiber
	app := config.NewApp()

	// Middleware logger untuk mencatat setiap request
	app.Use(middleware.RequestLogger())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		if err := db.Ping(c.Context()); err != nil {
			return helper.SendError(
				c,
				503,
				"PostgreSQL tidak dapat dihubungi",
			)
		}

		return helper.SendSuccess(
			c,
			200,
			"API dan PostgreSQL berjalan normal",
			fiber.Map{
				"database": "connected",
			},
		)
	})

	// Membuat repository
	studentRepo := repository.NewStudentRepository(db)

	// Membuat handler/service
	handler := service.NewHandler(studentRepo)

	// Mendaftarkan seluruh route mahasiswa
	route.SetupStudentRoutes(app, handler)

	log.Println("Server berjalan di http://127.0.0.1:3000")

	// Menjalankan server
	log.Fatal(app.Listen(":3000"))
}
