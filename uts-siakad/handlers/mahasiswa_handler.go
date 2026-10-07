package handlers

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func GetMahasiswa(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		page, err := strconv.Atoi(c.Query("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}

		perPage, err := strconv.Atoi(c.Query("per_page", "10"))
		if err != nil || perPage < 1 {
			perPage = 10
		}
		if perPage > 50 {
			perPage = 50
		}

		var conditions []string
		var args []interface{}

		addFilter := func(condition string, value interface{}) {
			args = append(args, value)
			conditions = append(
				conditions,
				fmt.Sprintf(condition, len(args)),
			)
		}

		prodi := strings.TrimSpace(c.Query("prodi"))
		if prodi != "" {
			addFilter("program_studi ILIKE $%d", "%"+prodi+"%")
		}

		angkatan := strings.TrimSpace(c.Query("angkatan"))
		if angkatan != "" {
			year, err := strconv.Atoi(angkatan)
			if err != nil || year < 1900 || year > 9999 {
				return c.Status(422).JSON(fiber.Map{
					"success": false,
					"message": "Parameter angkatan tidak valid",
				})
			}
			addFilter("angkatan = $%d", year)
		}

		search := strings.TrimSpace(c.Query("search"))
		if search != "" {
			args = append(args, "%"+search+"%")
			n := len(args)
			conditions = append(
				conditions,
				fmt.Sprintf("(nim ILIKE $%d OR nama ILIKE $%d)", n, n),
			)
		}

		where := " WHERE deleted_at IS NULL"

		if len(conditions) > 0 {
			where += " AND " + strings.Join(conditions, " AND ")
		}

		sortParam := c.Query("sort", "nama")
		orderBy := "nama ASC, id ASC"

		switch sortParam {
		case "nama":
			orderBy = "nama ASC, id ASC"
		case "-ipk_terakhir":
			orderBy = "ipk_terakhir DESC NULLS LAST, id ASC"
		default:
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Parameter sort tidak valid",
				"errors": fiber.Map{
					"sort": []string{
						"Gunakan nama atau -ipk_terakhir",
					},
				},
			})
		}

		var total int64
		countQuery := "SELECT COUNT(*) FROM mahasiswa" + where
		if err := db.QueryRow(
			c.Context(), countQuery, args...,
		).Scan(&total); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal menghitung data mahasiswa",
			})
		}

		offset := (page - 1) * perPage
		queryArgs := append([]interface{}{}, args...)
		queryArgs = append(queryArgs, perPage, offset)

		limitPos := len(queryArgs) - 1
		offsetPos := len(queryArgs)

		query := fmt.Sprintf(`
			SELECT id, nim, nama, program_studi,
			       angkatan, ipk_terakhir
			FROM mahasiswa
			%s
			ORDER BY %s
			LIMIT $%d OFFSET $%d
		`, where, orderBy, limitPos, offsetPos)

		rows, err := db.Query(c.Context(), query, queryArgs...)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil data mahasiswa",
			})
		}
		defer rows.Close()

		data := make([]fiber.Map, 0)

		for rows.Next() {
			var (
				id       int64
				nim      string
				nama     string
				prodiDB  string
				angkatan *int32
				ipk      *float64
			)

			if err := rows.Scan(
				&id, &nim, &nama, &prodiDB,
				&angkatan, &ipk,
			); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"message": "Gagal membaca data mahasiswa",
				})
			}

			data = append(data, fiber.Map{
				"id":           id,
				"nim":          nim,
				"nama":         nama,
				"prodi":        prodiDB,
				"angkatan":     angkatan,
				"ipk_terakhir": ipk,
			})
		}

		if err := rows.Err(); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membaca data mahasiswa",
			})
		}

		lastPage := int64(0)
		if total > 0 {
			lastPage = (total + int64(perPage) - 1) /
				int64(perPage)
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Data mahasiswa berhasil diambil",
			"data":    data,
			"meta": fiber.Map{
				"current_page": page,
				"per_page":     perPage,
				"total":        total,
				"last_page":    lastPage,
			},
		})
	}
}

func CreateMahasiswa(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req struct {
			NIM         string   `json:"nim"`
			Nama        string   `json:"nama"`
			Email       string   `json:"email"`
			Prodi       string   `json:"prodi"`
			Angkatan    int      `json:"angkatan"`
			IPKTerakhir *float64 `json:"ipk_terakhir"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Validasi gagal",
			})
		}

		req.NIM = strings.TrimSpace(req.NIM)
		req.Nama = strings.TrimSpace(req.Nama)
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		req.Prodi = strings.TrimSpace(req.Prodi)

		fieldErrors := fiber.Map{}

		if !regexp.MustCompile(`^[0-9]{12}$`).MatchString(req.NIM) {
			fieldErrors["nim"] = []string{
				"NIM wajib terdiri dari 12 digit",
			}
		}

		if req.Nama == "" {
			fieldErrors["nama"] = []string{
				"Nama wajib diisi",
			}
		}

		email, emailErr := mail.ParseAddress(req.Email)
		if emailErr != nil || email.Address != req.Email {
			fieldErrors["email"] = []string{
				"Format email tidak valid",
			}
		}

		if req.Prodi == "" {
			fieldErrors["prodi"] = []string{
				"Program studi wajib diisi",
			}
		}

		currentYear := time.Now().Year()
		if req.Angkatan < 1000 ||
			req.Angkatan > currentYear {
			fieldErrors["angkatan"] = []string{
				"Angkatan wajib 4 digit dan tidak boleh melebihi tahun berjalan",
			}
		}

		if req.IPKTerakhir != nil &&
			(*req.IPKTerakhir < 0 || *req.IPKTerakhir > 4) {
			fieldErrors["ipk_terakhir"] = []string{
				"IPK harus berada antara 0 dan 4",
			}
		}

		if len(fieldErrors) > 0 {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Validasi gagal",
				"errors":  fieldErrors,
			})
		}

		// Password awal mahasiswa adalah NIM yang di-hash.
		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(req.NIM),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal memproses akun mahasiswa",
			})
		}

		tx, err := db.Begin(c.Context())
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal memulai transaksi",
			})
		}
		defer tx.Rollback(c.Context())

		// Buat data mahasiswa terlebih dahulu.
		var mahasiswaID int64
		err = tx.QueryRow(c.Context(), `
			INSERT INTO mahasiswa
				(nim, nama, email, program_studi, angkatan, ipk_terakhir)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id
		`,
			req.NIM,
			req.Nama,
			req.Email,
			req.Prodi,
			req.Angkatan,
			req.IPKTerakhir,
		).Scan(&mahasiswaID)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return c.Status(422).JSON(fiber.Map{
					"success": false,
					"message": "Validasi gagal",
					"errors": fiber.Map{
						"nim": []string{
							"NIM atau email sudah terdaftar",
						},
					},
				})
			}

			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membuat data mahasiswa",
			})
		}

		// Buat akun login yang terhubung ke mahasiswa.
		_, err = tx.Exec(c.Context(), `
			INSERT INTO users
				(nama, email, password_hash, role, mahasiswa_id)
			VALUES ($1, $2, $3, 'mahasiswa', $4)
		`,
			req.Nama,
			req.Email,
			string(passwordHash),
			mahasiswaID,
		)

		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return c.Status(422).JSON(fiber.Map{
					"success": false,
					"message": "Validasi gagal",
					"errors": fiber.Map{
						"email": []string{
							"Email sudah terdaftar",
						},
					},
				})
			}

			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membuat akun mahasiswa",
			})
		}

		if err := tx.Commit(c.Context()); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal menyimpan data mahasiswa",
			})
		}

		return c.Status(201).JSON(fiber.Map{
			"success": true,
			"message": "Mahasiswa berhasil ditambahkan",
			"data": fiber.Map{
				"id":           mahasiswaID,
				"nim":          req.NIM,
				"nama":         req.Nama,
				"email":        req.Email,
				"prodi":        req.Prodi,
				"angkatan":     req.Angkatan,
				"ipk_terakhir": req.IPKTerakhir,
			},
		})
	}
}

func GetMahasiswaByID(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "ID mahasiswa tidak valid",
			})
		}

		// Pastikan pengguna sudah terautentikasi.
		userID, ok := c.Locals("userID").(int64)
		role, roleOK := c.Locals("role").(string)

		if !ok || !roleOK {
			return c.Status(401).JSON(fiber.Map{
				"success": false,
				"message": "Autentikasi diperlukan",
			})
		}

		// Mahasiswa hanya boleh melihat datanya sendiri.
		if role == "mahasiswa" {
			var mahasiswaID int64

			err := db.QueryRow(
				c.Context(),
				`SELECT mahasiswa_id FROM users WHERE id = $1`,
				userID,
			).Scan(&mahasiswaID)

			if err != nil || mahasiswaID != id {
				return c.Status(403).JSON(fiber.Map{
					"success": false,
					"message": "Anda hanya dapat melihat data mahasiswa milik sendiri",
				})
			}
		} else if role != "admin" {
			return c.Status(403).JSON(fiber.Map{
				"success": false,
				"message": "Anda tidak memiliki akses",
			})
		}

		// Ambil detail mahasiswa.
		var (
			nim         string
			nama        string
			email       string
			prodi       string
			angkatan    int
			ipkTerakhir float64
		)

		err = db.QueryRow(c.Context(), `
			SELECT nim, nama, email, program_studi,
				angkatan, ipk_terakhir
			FROM mahasiswa
			WHERE id = $1
			AND deleted_at IS NULL
		`, id).Scan(
			&nim,
			&nama,
			&email,
			&prodi,
			&angkatan,
			&ipkTerakhir,
		)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.Status(404).JSON(fiber.Map{
					"success": false,
					"message": "Mahasiswa tidak ditemukan",
				})
			}

			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil detail mahasiswa",
			})
		}

		// Ambil mata kuliah mahasiswa dari KRS.
		rows, err := db.Query(c.Context(), `
			SELECT
				mk.id,
				mk.kode_mk,
				mk.nama_mk,
				mk.sks,
				k.semester
			FROM krs k
			JOIN mata_kuliah mk ON mk.id = k.mata_kuliah_id
			WHERE k.mahasiswa_id = $1
			ORDER BY k.semester, mk.kode_mk
		`, id)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil data KRS mahasiswa",
			})
		}
		defer rows.Close()

		mataKuliah := make([]fiber.Map, 0)
		totalSKS := 0

		for rows.Next() {
			var (
				mkID     int64
				kodeMK   string
				namaMK   string
				sks      int
				semester string
			)

			if err := rows.Scan(
				&mkID,
				&kodeMK,
				&namaMK,
				&sks,
				&semester,
			); err != nil {
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"message": "Gagal membaca data KRS",
				})
			}

			mataKuliah = append(mataKuliah, fiber.Map{
				"id":       mkID,
				"kode_mk":  kodeMK,
				"nama_mk":  namaMK,
				"sks":      sks,
				"semester": semester,
			})

			totalSKS += sks
		}

		// Menentukan batas SKS berdasarkan IPK terakhir
		batasSKS := 18

		if ipkTerakhir >= 3.00 {
			batasSKS = 24
		} else if ipkTerakhir >= 2.50 {
			batasSKS = 21
		}

		// Menghitung sisa SKS
		sisaSKS := batasSKS - totalSKS
		if sisaSKS < 0 {
			sisaSKS = 0
		}

		if err := rows.Err(); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membaca data KRS",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Detail mahasiswa berhasil diambil",
			"data": fiber.Map{
				"id":           id,
				"nim":          nim,
				"nama":         nama,
				"email":        email,
				"prodi":        prodi,
				"angkatan":     angkatan,
				"ipk_terakhir": ipkTerakhir,
				"mata_kuliah":  mataKuliah,
				"total_sks":    totalSKS,
				"batas_sks":    batasSKS,
				"sisa_sks":     sisaSKS,
			},
		})
	}
}

func UpdateMahasiswa(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Ambil ID mahasiswa dari URL
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Mahasiswa tidak ditemukan",
			})
		}

		// Request body
		// NIM tidak dimasukkan karena NIM tidak boleh diubah
		var req struct {
			Nama        string   `json:"nama"`
			Prodi       string   `json:"prodi"`
			Angkatan    int      `json:"angkatan"`
			IPKTerkahir *float64 `json:"ipk_terakhir"`
		}

		// Parse JSON
		if err := c.BodyParser(&req); err != nil {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Format request tidak valid",
			})
		}

		// Bersihkan input
		req.Nama = strings.TrimSpace(req.Nama)
		req.Prodi = strings.TrimSpace(req.Prodi)

		// ==========================================
		// VALIDASI
		// ==========================================

		errors := make(map[string][]string)

		// Nama wajib
		if req.Nama == "" {
			errors["nama"] = []string{
				"Nama wajib diisi",
			}
		}

		// Prodi wajib
		if req.Prodi == "" {
			errors["prodi"] = []string{
				"Program studi wajib diisi",
			}
		}

		// Angkatan wajib 4 digit
		currentYear := time.Now().Year()

		if req.Angkatan < 1000 || req.Angkatan > 9999 {
			errors["angkatan"] = []string{
				"Angkatan harus berupa 4 digit",
			}
		} else if req.Angkatan > currentYear {
			errors["angkatan"] = []string{
				"Angkatan tidak boleh melebihi tahun berjalan",
			}
		}

		// IPK opsional, tetapi jika dikirim harus 0-4
		if req.IPKTerkahir != nil {
			if *req.IPKTerkahir < 0 || *req.IPKTerkahir > 4 {
				errors["ipk_terakhir"] = []string{
					"IPK harus berada antara 0 dan 4",
				}
			}
		}

		// Jika validasi gagal
		if len(errors) > 0 {
			return c.Status(422).JSON(fiber.Map{
				"success": false,
				"message": "Validasi gagal",
				"errors":  errors,
			})
		}

		// ==========================================
		// UPDATE DATABASE
		// ==========================================

		var (
			nim         string
			nama        string
			email       string
			prodi       string
			angkatan    int
			ipkTerakhir *float64
		)

		err = db.QueryRow(
			c.Context(),
			`UPDATE mahasiswa
			 SET nama = $1,
			     program_studi = $2,
			     angkatan = $3,
			     ipk_terakhir = $4
			 WHERE id = $5
  				AND deleted_at IS NULL
			 RETURNING
			     nim,
			     nama,
			     email,
			     program_studi,
			     angkatan,
			     ipk_terakhir`,
			req.Nama,
			req.Prodi,
			req.Angkatan,
			req.IPKTerkahir,
			id,
		).Scan(
			&nim,
			&nama,
			&email,
			&prodi,
			&angkatan,
			&ipkTerakhir,
		)

		// ID mahasiswa tidak ditemukan
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Mahasiswa tidak ditemukan",
			})
		}

		// ==========================================
		// RESPONSE
		// ==========================================

		return c.Status(200).JSON(fiber.Map{
			"success": true,
			"message": "Data mahasiswa berhasil diperbarui",
			"data": fiber.Map{
				"id":           id,
				"nim":          nim,
				"nama":         nama,
				"email":        email,
				"prodi":        prodi,
				"angkatan":     angkatan,
				"ipk_terakhir": ipkTerakhir,
			},
		})
	}
}

func DeleteMahasiswa(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// Ambil ID dari URL
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Mahasiswa tidak ditemukan",
			})
		}

		// Soft delete mahasiswa
		result, err := db.Exec(
			c.Context(),
			`UPDATE mahasiswa
			 SET deleted_at = NOW()
			 WHERE id = $1
			   AND deleted_at IS NULL`,
			id,
		)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal menghapus mahasiswa",
			})
		}

		// Tidak ditemukan / sudah pernah dihapus
		if result.RowsAffected() == 0 {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"message": "Mahasiswa tidak ditemukan",
			})
		}

		// Berhasil soft delete
		return c.SendStatus(204)
	}
}
