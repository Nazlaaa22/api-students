package middleware

import (
	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return helper.SendError(
				c,
				fiber.StatusUnauthorized,
				"user belum terautentikasi",
			)
		}

		permissions := helper.PermissionsForRole(role)

		if !permissions.Has(permission) {
			return helper.SendError(
				c,
				fiber.StatusForbidden,
				"tidak memiliki permission yang diperlukan",
			)
		}

		return c.Next()
	}
}
