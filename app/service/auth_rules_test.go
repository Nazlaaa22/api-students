package service

import "testing"

func TestCheckPasswordStrengthTooShort(t *testing.T) {
	got := checkPasswordStrength("abc123")

	if got == "" {
		t.Fatal("password yang terlalu pendek seharusnya ditolak")
	}
}

func TestCheckPasswordStrengthWithoutDigit(t *testing.T) {
	got := checkPasswordStrength("abcdefgh")

	if got == "" {
		t.Fatal("password tanpa angka seharusnya ditolak")
	}
}

func TestCheckPasswordStrengthWeakPassword(t *testing.T) {
	got := checkPasswordStrength("password1")

	if got == "" {
		t.Fatal("password umum seharusnya ditolak")
	}
}

func TestCheckPasswordStrengthValid(t *testing.T) {
	got := checkPasswordStrength("nazla1234")

	if got != "" {
		t.Fatalf("password yang valid seharusnya diterima, dapat error: %s", got)
	}
}
