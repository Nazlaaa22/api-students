package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dosen struct {
	ID    int64   `json:"id"`
	NIDN  string  `json:"nidn"`
	Nama  string  `json:"nama"`
	Email *string `json:"email"`
}

func GetAllDosen(db *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rows, err := db.Query(
			c.Context(),
			`SELECT id, nidn, nama, email
			 FROM dosen
			 ORDER BY id`,
		)
		if err != nil {
			log.Println("Query dosen:", err)
			return c.Status(fiber.StatusInternalServerError).JSON(
				fiber.Map{
					"success": false,
					"message": "Gagal mengambil data dosen",
				},
			)
		}
		defer rows.Close()

		data := make([]Dosen, 0)

		for rows.Next() {
			var d Dosen

			if err := rows.Scan(
				&d.ID,
				&d.NIDN,
				&d.Nama,
				&d.Email,
			); err != nil {
				log.Println("Scan dosen:", err)
				return c.Status(fiber.StatusInternalServerError).JSON(
					fiber.Map{
						"success": false,
						"message": "Gagal membaca data dosen",
					},
				)
			}

			data = append(data, d)
		}

		if err := rows.Err(); err != nil {
			log.Println("Rows dosen:", err)
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
