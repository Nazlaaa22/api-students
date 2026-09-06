package config

import (
	"io"
	"log"
	"os"
)

func SetupLogger() {
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Println("Gagal membuat folder logs:", err)
		return
	}

	file, err := os.OpenFile(
		"logs/app.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)

	if err != nil {
		log.Println("Gagal membuka file log:", err)
		return
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetFlags(log.Ldate | log.Ltime)
}
