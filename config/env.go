package config

import (
	"os"

	"github.com/joho/godotenv"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		// Menggunakan environment variable sistem jika .env tidak ditemukan
		return
	}
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
