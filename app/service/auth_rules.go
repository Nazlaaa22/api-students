package service

import (
	"errors"
	"strings"
	"unicode"

	"api-students/app/model"
)

const minPasswordLength = 8

// checkPasswordStrength memeriksa kekuatan password.
// Function ini murni: tidak bergantung pada fiber.Ctx,
// database, atau komponen lain.
func checkPasswordStrength(password string) string {
	if len(password) < minPasswordLength {
		return "minimal 8 karakter"
	}

	var hasLetter, hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	// Daftar pendek password yang umum/lemah.
	weak := map[string]bool{
		"password1":   true,
		"12345678":    true,
		"qwerty123":   true,
		"admin123":    true,
		"password123": true,
	}

	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}

	return ""
}

func ValidateRegister(req model.RegisterRequest) error {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" {
		return errors.New("username wajib diisi")
	}

	if req.Email == "" {
		return errors.New("email wajib diisi")
	}

	if err := checkPasswordStrength(req.Password); err != "" {
		return errors.New(err)
	}

	return nil
}

func ValidateLogin(req model.LoginRequest) error {
	req.Username = strings.TrimSpace(req.Username)

	if req.Username == "" {
		return errors.New("username wajib diisi")
	}

	if req.Password == "" {
		return errors.New("password wajib diisi")
	}

	// Kekuatan password tidak diperiksa saat login.
	// Password lama mungkin dibuat sebelum aturan password diperbarui.
	return nil
}
