package handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetMataKuliah(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {

		// ==========================================
		// QUERY PARAMETER
		// ==========================================

		semester := strings.TrimSpace(c.Query("semester"))
		search := strings.TrimSpace(c.Query("search"))
		available := c.Query("available")

		// ==========================================
		// QUERY DATA MATA KULIAH
		// ==========================================

		query := `
			SELECT
				mk.id,
				mk.kode_mk,
				mk.nama_mk,
				mk.sks,
				mk.semester,
				mk.kuota,
				COUNT(k.mahasiswa_id) AS terisi
			FROM mata_kuliah mk
			LEFT JOIN krs k
				ON k.mata_kuliah_id = mk.id
				AND k.semester = mk.semester
			WHERE 1=1
		`

		args := []interface{}{}
		arg := 1

		// Filter semester
		if semester != "" {
			query += ` AND mk.semester = $` + strconv.Itoa(arg)
			args = append(args, semester)
			arg++
		}

		// Search kode atau nama mata kuliah
		if search != "" {
			query += ` AND (
				mk.kode_mk ILIKE $` + strconv.Itoa(arg) + `
				OR mk.nama_mk ILIKE $` + strconv.Itoa(arg) + `
			)`

			args = append(args, "%"+search+"%")
			arg++
		}

		query += `
			GROUP BY
				mk.id,
				mk.kode_mk,
				mk.nama_mk,
				mk.sks,
				mk.semester,
				mk.kuota
		`

		// available=true → hanya yang belum penuh
		if available == "true" {
			query += ` HAVING COUNT(k.mahasiswa_id) < mk.kuota`
		}

		query += `
			ORDER BY mk.kode_mk ASC
		`

		rows, err := db.Query(
			c.Context(),
			query,
			args...,
		)

		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal mengambil daftar mata kuliah",
			})
		}

		defer rows.Close()

		// ==========================================
		// RESPONSE DATA
		// ==========================================

		data := make([]fiber.Map, 0)

		for rows.Next() {

			var (
				id         int64
				kodeMK     string
				namaMK     string
				sks        int
				semesterMK string
				kuota      int
				terisi     int64
			)

			if err := rows.Scan(
				&id,
				&kodeMK,
				&namaMK,
				&sks,
				&semesterMK,
				&kuota,
				&terisi,
			); err != nil {

				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"message": "Gagal membaca data mata kuliah",
				})
			}

			sisaKuota := kuota - int(terisi)

			if sisaKuota < 0 {
				sisaKuota = 0
			}

			data = append(data, fiber.Map{
				"id":         id,
				"kode_mk":    kodeMK,
				"nama_mk":    namaMK,
				"sks":        sks,
				"semester":   semesterMK,
				"kuota":      kuota,
				"terisi":     terisi,
				"sisa_kuota": sisaKuota,
			})
		}

		if err := rows.Err(); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Gagal membaca data mata kuliah",
			})
		}

		return c.JSON(fiber.Map{
			"success": true,
			"message": "Daftar mata kuliah berhasil diambil",
			"data":    data,
		})
	}
}
