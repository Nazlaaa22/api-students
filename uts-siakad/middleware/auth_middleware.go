package middleware

import (
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Token Bearer wajib disertakan",
			})
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Konfigurasi autentikasi belum tersedia",
			})
		}

		token, err := jwt.Parse(
			parts[1],
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
		)

		if err != nil || token == nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Token tidak valid atau kedaluwarsa",
			})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Klaim token tidak valid",
			})
		}

		sub, ok := claims["sub"].(string)
		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "ID pengguna pada token tidak valid",
			})
		}

		userID, err := strconv.ParseInt(sub, 10, 64)
		if err != nil || userID <= 0 {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "ID pengguna pada token tidak valid",
			})
		}

		role, ok := claims["role"].(string)
		if !ok || role == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Role pengguna pada token tidak valid",
			})
		}

		c.Locals("userID", userID)
		c.Locals("role", role)

		return c.Next()
	}
}

func RequireRole(allowedRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Autentikasi diperlukan",
			})
		}

		if role != allowedRole {
			return c.Status(403).JSON(fiber.Map{
				"success": false,
				"message": "Kamu tidak memiliki akses ke endpoint ini",
			})
		}

		return c.Next()
	}
}
