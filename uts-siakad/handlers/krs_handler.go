package handlers

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type KRS struct {
	ID           int64 `json:"id"`
	MahasiswaID  int64 `json:"mahasiswa_id"`
	MataKuliahID int64 `json:"mata_kuliah_id"`
}

func GetAllKRS(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query(
			c.Context(),
			`SELECT id, mahasiswa_id, mata_kuliah_id
			 FROM krs
			 ORDER BY id`,
		)
		if err != nil {
			log.Println("Query KRS:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"success": false,
					"message": "Gagal mengambil data KRS",
				},
			)
		}
		defer rows.Close()

		data := make([]KRS, 0)

		for rows.Next() {
			var k KRS

			if err := rows.Scan(
				&k.ID,
				&k.MahasiswaID,
				&k.MataKuliahID,
			); err != nil {
				log.Println("Scan KRS:", err)
				return c.Status(fiber.StatusInternalServerError).JSON(
					fiber.Map{
						"success": false,
						"message": "Gagal membaca data KRS",
					},
				)
			}

			data = append(data, k)
		}

		if err := rows.Err(); err != nil {
			log.Println("Rows KRS:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"success": false,
					"message": "Terjadi kesalahan saat membaca data",
				},
			)
		}

		return c.JSON(fiber.Map{
			"success": true,
			"data":    data,
		})
	}
}

func CreateEnrollment(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ==========================================
		// CEK ROLE
		// ==========================================

		role, ok := c.Locals("role").(string)

		if !ok || role != "mahasiswa" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Hanya mahasiswa yang dapat mengambil mata kuliah",
			})
		}

		// ==========================================
		// AMBIL USER ID DARI TOKEN
		// ==========================================

		userID, ok := c.Locals("userID").(int64)

		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token tidak valid",
			})
		}

		// ==========================================
		// REQUEST BODY
		// ==========================================

		var req struct {
			CourseID      int64  `json:"course_id"`
			TahunAkademik string `json:"tahun_akademik"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "Format request tidak valid",
			})
		}

		// ==========================================
		// VALIDASI COURSE ID
		// ==========================================

		if req.CourseID <= 0 {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "course_id wajib diisi",
			})
		}

		// ==========================================
		// VALIDASI TAHUN AKADEMIK
		// Contoh:
		// 2026/2027 Ganjil
		// 2026/2027 Genap
		// ==========================================

		req.TahunAkademik = strings.TrimSpace(req.TahunAkademik)

		pattern := regexp.MustCompile(
			`^\d{4}/\d{4} (Ganjil|Genap)$`,
		)

		if !pattern.MatchString(req.TahunAkademik) {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "Format tahun_akademik harus seperti 2026/2027 Ganjil",
			})
		}

		// ==========================================
		// MULAI TRANSACTION
		// ==========================================

		tx, err := db.Begin(c.Context())

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal memulai transaction",
			})
		}

		defer tx.Rollback(c.Context())

		// ==========================================
		// CARI MAHASISWA DARI USER LOGIN
		// ==========================================

		var mahasiswaID int64
		var ipkTerakhir float64

		err = tx.QueryRow(
			c.Context(),
			`
			SELECT m.id, COALESCE(m.ipk_terakhir, 0)
			FROM users u
			JOIN mahasiswa m ON m.id = u.mahasiswa_id
			WHERE u.id = $1
			  AND u.role = 'mahasiswa'
			  AND m.deleted_at IS NULL
			`,
			userID,
		).Scan(
			&mahasiswaID,
			&ipkTerakhir,
		)

		if err != nil {

			if err == pgx.ErrNoRows {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"message": "Data mahasiswa tidak ditemukan",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil data mahasiswa",
			})
		}

		// ==========================================
		// LOCK COURSE
		// ==========================================

		var (
			courseID int64
			kodeMK   string
			namaMK   string
			sks      int
			kuota    int
		)

		err = tx.QueryRow(
			c.Context(),
			`
			SELECT
				id,
				kode_mk,
				nama_mk,
				sks,
				kuota
			FROM mata_kuliah
			WHERE id = $1
			FOR UPDATE
			`,
			req.CourseID,
		).Scan(
			&courseID,
			&kodeMK,
			&namaMK,
			&sks,
			&kuota,
		)

		if err != nil {

			if err == pgx.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"message": "Mata kuliah tidak ditemukan",
				})
			}

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil data mata kuliah",
			})
		}

		// ==========================================
		// CEK DUPLIKASI
		// ==========================================

		var duplicate bool

		err = tx.QueryRow(
			c.Context(),
			`
			SELECT EXISTS (
				SELECT 1
				FROM krs
				WHERE mahasiswa_id = $1
				  AND mata_kuliah_id = $2
				  AND semester = $3
			)
			`,
			mahasiswaID,
			courseID,
			req.TahunAkademik,
		).Scan(&duplicate)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengecek duplikasi KRS",
			})
		}

		if duplicate {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"message": "Mata kuliah sudah pernah diambil pada tahun akademik tersebut",
			})
		}

		// ==========================================
		// HITUNG JUMLAH MAHASISWA DI COURSE
		// ==========================================

		var terisi int

		err = tx.QueryRow(
			c.Context(),
			`
			SELECT COUNT(*)
			FROM krs
			WHERE mata_kuliah_id = $1
			  AND semester = $2
			`,
			courseID,
			req.TahunAkademik,
		).Scan(&terisi)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengecek kuota mata kuliah",
			})
		}

		// ==========================================
		// CEK KUOTA
		// ==========================================

		if terisi >= kuota {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "Kuota mata kuliah sudah penuh",
			})
		}

		// ==========================================
		// TENTUKAN BATAS SKS
		// ==========================================

		batasSKS := 18

		if ipkTerakhir >= 3.00 {
			batasSKS = 24
		} else if ipkTerakhir >= 2.50 {
			batasSKS = 21
		}

		// ==========================================
		// HITUNG SKS YANG SUDAH DIAMBIL
		// ==========================================

		var totalSKS int

		err = tx.QueryRow(
			c.Context(),
			`
			SELECT COALESCE(SUM(mk.sks), 0)
			FROM krs k
			JOIN mata_kuliah mk
				ON mk.id = k.mata_kuliah_id
			WHERE k.mahasiswa_id = $1
			  AND k.semester = $2
			`,
			mahasiswaID,
			req.TahunAkademik,
		).Scan(&totalSKS)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal menghitung total SKS",
			})
		}

		// ==========================================
		// CEK BATAS SKS
		// ==========================================

		sisaSKS := batasSKS - totalSKS

		if sks > sisaSKS {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{
				"success": false,
				"message": "Total SKS melebihi batas",
				"errors": fiber.Map{
					"sisa_sks":        sisaSKS,
					"batas_sks":       batasSKS,
					"sks_mata_kuliah": sks,
				},
			})
		}

		// ==========================================
		// INSERT KRS
		// ==========================================

		_, err = tx.Exec(
			c.Context(),
			`
			INSERT INTO krs (
				mahasiswa_id,
				mata_kuliah_id,
				semester
			)
			VALUES ($1, $2, $3)
			`,
			mahasiswaID,
			courseID,
			req.TahunAkademik,
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil mata kuliah",
			})
		}

		// ==========================================
		// COMMIT
		// ==========================================

		if err := tx.Commit(c.Context()); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal menyimpan KRS",
			})
		}

		// ==========================================
		// RESPONSE
		// ==========================================

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"message": "Mata kuliah berhasil diambil",
			"data": fiber.Map{
				"course_id":      courseID,
				"kode_mk":        kodeMK,
				"nama_mk":        namaMK,
				"sks":            sks,
				"tahun_akademik": req.TahunAkademik,
				"total_sks":      totalSKS + sks,
				"batas_sks":      batasSKS,
				"sisa_sks":       batasSKS - (totalSKS + sks),
				"terisi":         terisi + 1,
				"sisa_kuota":     kuota - (terisi + 1),
			},
		})
	}
}

func DeleteEnrollment(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ==========================================
		// AMBIL ID ENROLLMENT / KRS DARI URL
		// ==========================================

		enrollmentID, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || enrollmentID <= 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Enrollment tidak ditemukan",
			})
		}

		// ==========================================
		// CEK ROLE
		// ==========================================

		role, ok := c.Locals("role").(string)
		if !ok || role != "mahasiswa" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Hanya mahasiswa yang dapat membatalkan mata kuliah",
			})
		}

		// ==========================================
		// AMBIL USER ID DARI TOKEN
		// ==========================================

		userID, ok := c.Locals("userID").(int64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"message": "Token tidak valid",
			})
		}

		// ==========================================
		// CARI MAHASISWA DARI USER LOGIN
		// users.mahasiswa_id -> mahasiswa.id
		// ==========================================

		var mahasiswaID int64

		err = db.QueryRow(
			c.Context(),
			`
			SELECT mahasiswa_id
			FROM users
			WHERE id = $1
			  AND role = 'mahasiswa'
			  AND mahasiswa_id IS NOT NULL
			`,
			userID,
		).Scan(&mahasiswaID)

		if err != nil {
			if err == pgx.ErrNoRows {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"success": false,
					"message": "Data mahasiswa tidak ditemukan",
				})
			}

			log.Println("Query mahasiswa:", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil data mahasiswa",
			})
		}

		// ==========================================
		// CARI PEMILIK KRS
		// ==========================================

		var ownerMahasiswaID int64

		err = db.QueryRow(
			c.Context(),
			`
			SELECT mahasiswa_id
			FROM krs
			WHERE id = $1
			`,
			enrollmentID,
		).Scan(&ownerMahasiswaID)

		if err != nil {

			// KRS tidak ditemukan
			if err == pgx.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"success": false,
					"message": "Enrollment tidak ditemukan",
				})
			}

			log.Println("Query enrollment:", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengecek enrollment",
			})
		}

		// ==========================================
		// CEK KEPEMILIKAN
		// ==========================================

		if ownerMahasiswaID != mahasiswaID {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"message": "Kamu tidak memiliki akses ke enrollment ini",
			})
		}

		// ==========================================
		// HAPUS KRS
		// ==========================================

		result, err := db.Exec(
			c.Context(),
			`
			DELETE FROM krs
			WHERE id = $1
			  AND mahasiswa_id = $2
			`,
			enrollmentID,
			mahasiswaID,
		)

		if err != nil {
			log.Println("Delete KRS:", err)

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membatalkan mata kuliah",
			})
		}

		// ==========================================
		// PASTIKAN DATA TERHAPUS
		// ==========================================

		if result.RowsAffected() == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Enrollment tidak ditemukan",
			})
		}

		// ==========================================
		// BERHASIL
		// ==========================================

		return c.SendStatus(fiber.StatusNoContent)
	}
}
