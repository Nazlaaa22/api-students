package main

import (
	"log"

	"uts-siakad/seeders"

	"github.com/gofiber/fiber/v2"

	"uts-siakad/routes"
)

func main() {
	db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := seeders.SeedAuthUsers(db); err != nil {
		log.Fatal("Gagal membuat akun pengujian: ", err)
	}

	app := fiber.New()

	// Endpoint 1: halaman utama API
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"success": true,
			"message": "API SIAKAD Mini berhasil dijalankan",
		})
	})

	// Endpoint 2: memeriksa koneksi database
	app.Get("/db-check", func(c *fiber.Ctx) error {
		var dbName string

		err := db.QueryRow(
			c.Context(),
			"SELECT current_database()",
		).Scan(&dbName)

		if err != nil {
			log.Println("Database check:", err)

			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"success": false,
					"message": "Gagal memeriksa database",
				},
			)
		}

		return c.JSON(fiber.Map{
			"success":  true,
			"message":  "Koneksi PostgreSQL berhasil",
			"database": dbName,
		})
	})

	// Daftarkan endpoint mahasiswa
	routes.SetupMahasiswaRoutes(app, db)

	// Daftarkan endpoint mata kuliah
	routes.SetupMataKuliahRoutes(app, db)

	// Daftarkan endpoint dosen
	routes.SetupDosenRoutes(app, db)

	// Daftarkan endpoint KRS
	routes.SetupKRSRoutes(app, db)

	// Endpoint autentikasi
	routes.SetupAuthRoutes(app, db)

	// Menjalankan server
	log.Println("Server berjalan pada http://127.0.0.1:3000")
	log.Fatal(app.Listen(":3000"))
}
