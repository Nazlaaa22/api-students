package middleware

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/helper"
)

const (
	maxLoginAttempts = 5
	loginWindow      = time.Minute
)

type loginAttempt struct {
	count       int
	windowStart time.Time
}

var (
	loginAttempts   = make(map[string]loginAttempt)
	loginAttemptsMu sync.Mutex
)

// LoginRateLimiter membatasi percobaan login yang gagal.
// Maksimal 5 kegagalan dalam 1 menit untuk kombinasi
// IP address dan username.
func LoginRateLimiter() fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := c.FormValue("username")

		// Karena request login menggunakan JSON,
		// ambil username dari body setelah BodyParser
		// dilakukan oleh handler. Untuk limiter, kita gunakan IP
		// sebagai identitas tambahan.
		key := c.IP() + ":" + username

		loginAttemptsMu.Lock()

		attempt, exists := loginAttempts[key]
		now := time.Now()

		if !exists || now.Sub(attempt.windowStart) >= loginWindow {
			attempt = loginAttempt{
				count:       0,
				windowStart: now,
			}
		}

		if attempt.count >= maxLoginAttempts {
			loginAttemptsMu.Unlock()

			c.Set("Retry-After", "60")

			return helper.SendError(
				c,
				fiber.StatusTooManyRequests,
				"terlalu banyak percobaan login, coba lagi setelah 60 detik",
			)
		}

		loginAttemptsMu.Unlock()

		err := c.Next()

		status := c.Response().StatusCode()

		loginAttemptsMu.Lock()
		defer loginAttemptsMu.Unlock()

		// Login berhasil → reset percobaan.
		if status == fiber.StatusOK {
			delete(loginAttempts, key)
			return err
		}

		// Login gagal karena username/password salah.
		if status == fiber.StatusUnauthorized {
			attempt.count++
			loginAttempts[key] = attempt
		}

		return err
	}
}
