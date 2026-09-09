package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")

		// Tidak ada Authorization header
		if authHeader == "" {
			c.Set("WWW-Authenticate", `Bearer`)
			return helper.SendError(
				c,
				fiber.StatusUnauthorized,
				"authorization header diperlukan",
			)
		}

		// Format harus: Bearer <token>
		parts := strings.Fields(authHeader)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Set("WWW-Authenticate", `Bearer`)
			return helper.SendError(
				c,
				fiber.StatusUnauthorized,
				"format authorization harus Bearer token",
			)
		}

		claims, err := helper.ParseAccessToken(parts[1])
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer`)

			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.SendError(
					c,
					fiber.StatusUnauthorized,
					"access token sudah kedaluwarsa",
				)
			}

			return helper.SendError(
				c,
				fiber.StatusUnauthorized,
				"access token tidak valid",
			)
		}

		// Menyimpan informasi user untuk handler berikutnya.
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
