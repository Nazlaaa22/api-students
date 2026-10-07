package seeders

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

func SeedAuthUsers(db *pgxpool.Pool) error {
	ctx := context.Background()

	// Cari satu mahasiswa untuk akun pengujian.
	var mahasiswaID int64
	err := db.QueryRow(ctx,
		"SELECT id FROM mahasiswa ORDER BY id LIMIT 1",
	).Scan(&mahasiswaID)

	if err != nil {
		return fmt.Errorf("belum ada data mahasiswa: %w", err)
	}

	// Cari satu dosen untuk akun pengujian.
	var dosenID int64
	err = db.QueryRow(ctx,
		"SELECT id FROM dosen ORDER BY id LIMIT 1",
	).Scan(&dosenID)

	if err != nil {
		return fmt.Errorf("belum ada data dosen: %w", err)
	}

	accounts := []struct {
		nama        string
		email       string
		password    string
		role        string
		mahasiswaID *int64
		dosenID     *int64
	}{
		{
			nama:     "Administrator",
			email:    "admin@siakad.test",
			password: "Admin123!",
			role:     "admin",
		},
		{
			nama:        "Mahasiswa Uji",
			email:       "mahasiswa@siakad.test",
			password:    "Mhs12345!",
			role:        "mahasiswa",
			mahasiswaID: &mahasiswaID,
		},
		{
			nama:     "Dosen Uji",
			email:    "dosen@siakad.test",
			password: "Dosen123!",
			role:     "dosen",
			dosenID:  &dosenID,
		},
	}

	for _, account := range accounts {
		hash, err := bcrypt.GenerateFromPassword(
			[]byte(account.password),
			bcrypt.DefaultCost,
		)
		if err != nil {
			return fmt.Errorf("gagal melakukan hash password: %w", err)
		}

		_, err = db.Exec(ctx, `
			INSERT INTO users
				(nama, email, password_hash, role, mahasiswa_id, dosen_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT DO NOTHING
		`,
			account.nama,
			account.email,
			string(hash),
			account.role,
			account.mahasiswaID,
			account.dosenID,
		)
		if err != nil {
			return fmt.Errorf(
				"gagal membuat akun %s: %w",
				account.email, err,
			)
		}
	}

	return nil
}
