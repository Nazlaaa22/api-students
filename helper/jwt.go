package helper

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	jwtIssuer        = "praktikum-backend"
	jwtAccessTTL     = 15 * time.Minute
	minJWTSecretSize = 32
)

var (
	ErrExpiredToken = errors.New("token JWT sudah kedaluwarsa")
	ErrInvalidToken = errors.New("token JWT tidak valid")
)

type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func jwtSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")

	if len(secret) < minJWTSecretSize {
		return nil, errors.New("JWT_SECRET minimal 32 karakter")
	}

	return []byte(secret), nil
}

func GenerateAccessToken(
	userID int,
	username string,
	role string,
) (string, error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}

	now := time.Now()

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			Issuer:    jwtIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtAccessTTL)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(secret)
}

func ParseAccessToken(tokenString string) (*JWTClaims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Pastikan algoritma benar-benar HS256.
			if token.Method != jwt.SigningMethodHS256 {
				return nil, ErrInvalidToken
			}

			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuer(jwtIssuer),
	)

	if err != nil {
		// Bedakan token expired dengan token invalid.
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}

		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
