package handlers

import (
	"log"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"uts-siakad/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// =====================================================
// RATE LIMITER LOGIN
// Maksimal 5 percobaan login gagal dalam 1 menit
// =====================================================

var (
	loginMu       sync.Mutex
	loginAttempts = make(map[string][]time.Time)
)

func isRateLimited(key string) bool {
	loginMu.Lock()
	defer loginMu.Unlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	attempts := loginAttempts[key]

	var recent []time.Time

	for _, t := range attempts {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}

	loginAttempts[key] = recent

	return len(recent) >= 5
}

func recordFailedLogin(key string) {
	loginMu.Lock()
	defer loginMu.Unlock()

	now := time.Now()
	windowStart := now.Add(-1 * time.Minute)

	attempts := loginAttempts[key]

	var recent []time.Time

	for _, t := range attempts {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}

	recent = append(recent, now)

	loginAttempts[key] = recent
}

func clearLoginAttempts(key string) {
	loginMu.Lock()
	defer loginMu.Unlock()

	delete(loginAttempts, key)
}

// =====================================================
// LOGIN
// POST /api/v1/auth/login
// =====================================================

func Login(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var req models.LoginRequest

		if err := c.BodyParser(&req); err != nil {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Format request tidak valid",
			})
		}

		req.Email = strings.TrimSpace(
			strings.ToLower(req.Email),
		)

		// -----------------------------
		// VALIDASI
		// -----------------------------

		if req.Email == "" {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Email wajib diisi",
			})
		}

		if _, err := mail.ParseAddress(req.Email); err != nil {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Format email tidak valid",
			})
		}

		if req.Password == "" {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Password wajib diisi",
			})
		}

		if len(req.Password) < 8 {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Password minimal 8 karakter",
			})
		}

		// -----------------------------
		// RATE LIMIT
		// -----------------------------

		rateLimitKey := c.IP() + "|" + req.Email

		if isRateLimited(rateLimitKey) {
			return c.Status(429).JSON(fiber.Map{
				"success": false,
				"message": "Terlalu banyak percobaan login. Silakan coba lagi nanti",
			})
		}

		// -----------------------------
		// CARI USER
		// -----------------------------

		var user models.User
		var passwordHash string

		err := db.QueryRow(
			c.Context(),
			`SELECT id, nama, email, password_hash, role
			 FROM users
			 WHERE LOWER(email) = $1`,
			req.Email,
		).Scan(
			&user.ID,
			&user.Nama,
			&user.Email,
			&passwordHash,
			&user.Role,
		)

		if err != nil {
			log.Println("Login query:", err)

			recordFailedLogin(rateLimitKey)

			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Email atau password salah",
			})
		}

		// -----------------------------
		// CEK PASSWORD
		// -----------------------------

		if err := bcrypt.CompareHashAndPassword(
			[]byte(passwordHash),
			[]byte(req.Password),
		); err != nil {

			recordFailedLogin(rateLimitKey)

			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Email atau password salah",
			})
		}

		if user.Role == "mahasiswa" {
			var deletedAt *time.Time

			err := db.QueryRow(
				c.Context(),
				`SELECT m.deleted_at
				FROM users u
				JOIN mahasiswa m ON m.id = u.mahasiswa_id
				WHERE u.id = $1
				AND u.role = 'mahasiswa'`,
				user.ID,
			).Scan(&deletedAt)

			if err != nil {
				log.Println("Cek status mahasiswa:", err)

				return c.Status(401).JSON(fiber.Map{
					"success": false,
					"message": "Akun mahasiswa tidak valid",
				})
			}

			if deletedAt != nil {
				return c.Status(403).JSON(fiber.Map{
					"success": false,
					"message": "Akun mahasiswa sudah dihapus",
				})
			}
		}

		clearLoginAttempts(rateLimitKey)

		// -----------------------------
		// JWT SECRET
		// -----------------------------

		secret := os.Getenv("JWT_SECRET")

		if secret == "" {
			log.Println("JWT_SECRET belum dikonfigurasi")

			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Konfigurasi autentikasi belum tersedia",
			})
		}

		// -----------------------------
		// JWT EXPIRED
		// -----------------------------

		hours, err := strconv.Atoi(
			os.Getenv("JWT_EXPIRES_HOURS"),
		)

		if err != nil || hours <= 0 {
			hours = 24
		}

		now := time.Now()

		expiresAt := now.Add(
			time.Duration(hours) * time.Hour,
		)

		claims := jwt.MapClaims{
			"sub":  strconv.FormatInt(user.ID, 10),
			"role": user.Role,
			"iat":  now.Unix(),
			"exp":  expiresAt.Unix(),
		}

		// -----------------------------
		// GENERATE TOKEN
		// -----------------------------

		token := jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			claims,
		)

		tokenString, err := token.SignedString(
			[]byte(secret),
		)

		if err != nil {
			log.Println("Generate token:", err)

			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membuat access token",
			})
		}

		// -----------------------------
		// RESPONSE
		// -----------------------------

		return c.JSON(fiber.Map{
			"success":      true,
			"message":      "Login berhasil",
			"access_token": tokenString,
			"token_type":   "Bearer",
			"expires_in":   int64(hours) * 3600,
			"user":         user,
		})
	}
}

// =====================================================
// ME
// GET /api/v1/auth/me
// =====================================================

func Me(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// -----------------------------
		// AMBIL USER ID DARI TOKEN
		// -----------------------------

		userID, ok := c.Locals("userID").(int64)

		if !ok {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Token tidak valid",
			})
		}

		// -----------------------------
		// AMBIL DATA USER
		// -----------------------------

		var user models.User

		err := db.QueryRow(
			c.Context(),
			`SELECT id, nama, email, role
			 FROM users
			 WHERE id = $1`,
			userID,
		).Scan(
			&user.ID,
			&user.Nama,
			&user.Email,
			&user.Role,
		)

		if err != nil {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Pengguna tidak ditemukan",
			})
		}

		// =================================================
		// JIKA ROLE MAHASISWA
		// Ambil data dari tabel mahasiswa
		// =================================================

		if user.Role == "mahasiswa" {

			var student struct {
				NIM      string `json:"nim"`
				Nama     string `json:"nama"`
				Prodi    string `json:"prodi"`
				Angkatan int    `json:"angkatan"`
			}

			err := db.QueryRow(
				c.Context(),
				`SELECT
					m.nim,
					m.nama,
					m.program_studi,
					m.angkatan
				 FROM users u
				 JOIN mahasiswa m
					ON m.id = u.mahasiswa_id
				 WHERE u.id = $1
				   AND u.role = 'mahasiswa'`,
				userID,
			).Scan(
				&student.NIM,
				&student.Nama,
				&student.Prodi,
				&student.Angkatan,
			)

			if err != nil {
				log.Println("Get student profile:", err)

				return c.Status(404).JSON(fiber.Map{
					"success": false,
					"message": "Data mahasiswa tidak ditemukan",
				})
			}

			// -----------------------------
			// RESPONSE MAHASISWA
			// -----------------------------

			return c.JSON(fiber.Map{
				"success": true,
				"message": "Profil pengguna berhasil diambil",
				"data": fiber.Map{
					"id":      user.ID,
					"nama":    user.Nama,
					"email":   user.Email,
					"role":    user.Role,
					"student": student,
				},
			})
		}

		// =================================================
		// ADMIN / DOSEN
		// =================================================

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Profil pengguna berhasil diambil",
			"data":    user,
		})
	}
}
