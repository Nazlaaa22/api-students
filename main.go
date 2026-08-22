package main

import "github.com/gofiber/fiber/v2"

func main() {
	app := fiber.New()

	// API group
	api := app.Group("/api/v1")

	// Student endpoints
	api.Get("/students", getStudents)
	api.Get("/students/:id", getStudent)
	api.Post("/students", createStudent)
	api.Put("/students/:id", updateStudent)
	api.Patch("/students/:id", patchStudent)
	api.Delete("/students/:id", deleteStudent)

	// Jalankan server
	app.Listen(":3000")
}
