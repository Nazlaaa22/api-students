package route

import (
	"github.com/gofiber/fiber/v2"

	"api-students/app/service"
	"api-students/middleware"
)

func SetupStudentRoutes(app *fiber.App, handler *service.Handler) {
	api := app.Group("/api/v1")

	students := api.Group("/students", middleware.RequireAuth())

	// Permission dapat ditentukan tanpa membaca data student.
	students.Get(
		"/",
		middleware.RequirePermission("student:list"),
		handler.GetStudents,
	)

	students.Post(
		"/",
		middleware.RequirePermission("student:create"),
		handler.CreateStudent,
	)

	students.Delete(
		"/:id",
		middleware.RequirePermission("student:delete"),
		handler.DeleteStudent,
	)

	// Endpoint berikut membutuhkan pemeriksaan ownership
	// sehingga pengecekannya dilakukan di service.
	students.Get("/:id", handler.GetStudent)
	students.Put("/:id", handler.UpdateStudent)
	students.Patch("/:id", handler.PatchStudent)
}
