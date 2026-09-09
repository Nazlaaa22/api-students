package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/service"
	"api-students/middleware"
)

func SetupStudentRoutes(app *fiber.App, handler *service.Handler) {
	api := app.Group("/api/v1")

	students := api.Group("/students", middleware.RequireAuth())

	students.Get("/", handler.GetStudents)
	students.Get("/:id", handler.GetStudent)
	students.Post("/", handler.CreateStudent)
	students.Put("/:id", handler.UpdateStudent)
	students.Patch("/:id", handler.PatchStudent)
	students.Delete("/:id", handler.DeleteStudent)
}
