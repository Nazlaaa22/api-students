package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/service"
)

func SetupStudentRoutes(app *fiber.App, handler *service.Handler) {
	api := app.Group("/api/v1")

	api.Get("/students", handler.GetStudents)
	api.Get("/students/:id", handler.GetStudent)
	api.Post("/students", handler.CreateStudent)
	api.Put("/students/:id", handler.UpdateStudent)
	api.Patch("/students/:id", handler.PatchStudent)
	api.Delete("/students/:id", handler.DeleteStudent)
}
