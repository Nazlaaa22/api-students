package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"api-students/app/model"
	"api-students/app/repository"
	"api-students/helper"
)

const refreshTokenBytes = 32

type AuthService struct {
	users  repository.UserRepository
	tokens repository.TokenRepository
}

func NewAuthService(
	users repository.UserRepository,
	tokens repository.TokenRepository,
) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

// Register menangani pendaftaran user baru.
func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if err := ValidateRegister(req); err != nil {
		return helper.SendError(
			c,
			fiber.StatusUnprocessableEntity,
			err.Error(),
		)
	}

	// Password harus di-hash sebelum disimpan ke database.
	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.SendError(
			c,
			fiber.StatusInternalServerError,
			"gagal memproses password",
		)
	}

	// Role ditentukan oleh server, bukan dari request client.
	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     "user",
		IsActive: true,
	}

	created, err := s.users.Create(context.Background(), user)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return helper.SendError(
				c,
				fiber.StatusConflict,
				"username atau email sudah digunakan",
			)
		}

		return helper.SendError(
			c,
			fiber.StatusInternalServerError,
			"gagal mendaftarkan user",
		)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusCreated,
		"pendaftaran berhasil",
		created,
	)
}

// Login memeriksa username dan password lalu membuat token.
func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	if err := ValidateLogin(req); err != nil {
		return helper.SendError(
			c,
			fiber.StatusUnprocessableEntity,
			err.Error(),
		)
	}

	username := strings.TrimSpace(req.Username)

	user, err := s.users.FindByUsername(
		context.Background(),
		username,
	)

	if err != nil {
		// Tetap melakukan bcrypt agar waktu respons
		// mirip dengan username yang benar tetapi password salah.
		helper.VerifyDummyPassword(req.Password)

		return helper.SendError(
			c,
			fiber.StatusUnauthorized,
			"username atau password salah",
		)
	}

	if !helper.VerifyPassword(user.Password, req.Password) {
		return helper.SendError(
			c,
			fiber.StatusUnauthorized,
			"username atau password salah",
		)
	}

	if !user.IsActive {
		return helper.SendError(
			c,
			fiber.StatusForbidden,
			"akun dinonaktifkan",
		)
	}

	tokenPair, err := s.issueTokenPair(user)
	if err != nil {
		return helper.SendError(
			c,
			fiber.StatusInternalServerError,
			"gagal membuat token",
		)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"login berhasil",
		tokenPair,
	)
}

// Refresh menggunakan refresh token untuk membuat pasangan token baru.
func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)

	if req.RefreshToken == "" {
		return helper.SendError(
			c,
			fiber.StatusBadRequest,
			"refresh_token wajib diisi",
		)
	}

	tokenHash := helper.SHA256Hex(req.RefreshToken)

	storedToken, err := s.tokens.FindByHash(
		context.Background(),
		tokenHash,
	)

	if err != nil {
		return helper.SendError(
			c,
			fiber.StatusUnauthorized,
			"refresh token tidak valid atau sudah kedaluwarsa",
		)
	}

	user, err := s.users.FindByID(
		context.Background(),
		storedToken.UserID,
	)

	if err != nil || !user.IsActive {
		return helper.SendError(
			c,
			fiber.StatusUnauthorized,
			"akun tidak dapat digunakan",
		)
	}

	// Rotasi: token lama langsung dicabut.
	if err := s.tokens.Revoke(
		context.Background(),
		storedToken.ID,
	); err != nil {
		return helper.SendError(
			c,
			fiber.StatusInternalServerError,
			"gagal memperbarui token",
		)
	}

	tokenPair, err := s.issueTokenPair(user)
	if err != nil {
		return helper.SendError(
			c,
			fiber.StatusInternalServerError,
			"gagal membuat token",
		)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"token berhasil diperbarui",
		tokenPair,
	)
}

// Logout mencabut refresh token.
func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.SendError(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	req.RefreshToken = strings.TrimSpace(req.RefreshToken)

	if req.RefreshToken != "" {
		tokenHash := helper.SHA256Hex(req.RefreshToken)

		// Logout tetap dianggap berhasil meskipun token
		// sudah tidak aktif.
		if storedToken, err := s.tokens.FindByHash(
			context.Background(),
			tokenHash,
		); err == nil {
			_ = s.tokens.Revoke(
				context.Background(),
				storedToken.ID,
			)
		}
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"logout berhasil",
		nil,
	)
}

// Me mengambil profil user berdasarkan identitas dari JWT.
func (s *AuthService) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int)

	if !ok {
		return helper.SendError(
			c,
			fiber.StatusUnauthorized,
			"belum terautentikasi",
		)
	}

	user, err := s.users.FindByID(
		context.Background(),
		userID,
	)

	if err != nil {
		return helper.SendError(
			c,
			fiber.StatusUnauthorized,
			"user tidak ditemukan",
		)
	}

	return helper.SendSuccess(
		c,
		fiber.StatusOK,
		"profil berhasil diambil",
		user,
	)
}

// issueTokenPair membuat access token dan refresh token.
func (s *AuthService) issueTokenPair(
	user model.User,
) (model.TokenPair, error) {

	accessToken, err := helper.GenerateAccessToken(
		user.ID,
		user.Username,
		user.Role,
	)
	if err != nil {
		return model.TokenPair{}, err
	}

	refreshToken, err := helper.RandomToken(refreshTokenBytes)
	if err != nil {
		return model.TokenPair{}, err
	}

	_, err = s.tokens.Create(
		context.Background(),
		model.RefreshToken{
			UserID:    user.ID,
			TokenHash: helper.SHA256Hex(refreshToken),
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		},
	)

	if err != nil {
		return model.TokenPair{}, err
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60,
	}, nil
}
