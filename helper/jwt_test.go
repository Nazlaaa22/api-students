package helper

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestParseAccessTokenRejectsAlgNone(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodNone, JWTClaims{
		UserID:   1,
		Username: "nazlatest",
		Role:     "user",
	})

	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("gagal membuat token test: %v", err)
	}

	_, err = ParseAccessToken(tokenString)

	if err == nil {
		t.Fatal("seharusnya token dengan alg:none ditolak")
	}
}
