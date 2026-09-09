package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"api-students/app/model"
)

type TokenRepository interface {
	Create(ctx context.Context, token model.RefreshToken) (model.RefreshToken, error)
	FindByHash(ctx context.Context, tokenHash string) (model.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
}

type tokenPostgresRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) TokenRepository {
	return &tokenPostgresRepository{
		pool: pool,
	}
}

func (r *tokenPostgresRepository) Create(
	ctx context.Context,
	token model.RefreshToken,
) (model.RefreshToken, error) {
	var created model.RefreshToken

	err := r.pool.QueryRow(
		ctx,
		`INSERT INTO refresh_tokens
			(user_id, token_hash, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, token_hash, expires_at, revoked_at, created_at`,
		token.UserID,
		token.TokenHash,
		token.ExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.TokenHash,
		&created.ExpiresAt,
		&created.RevokedAt,
		&created.CreatedAt,
	)

	if err != nil {
		return model.RefreshToken{}, fmt.Errorf("membuat refresh token: %w", err)
	}

	return created, nil
}

func (r *tokenPostgresRepository) FindByHash(
	ctx context.Context,
	tokenHash string,
) (model.RefreshToken, error) {
	var token model.RefreshToken

	err := r.pool.QueryRow(
		ctx,
		`SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		 FROM refresh_tokens
		 WHERE token_hash = $1
		   AND revoked_at IS NULL
		   AND expires_at > NOW()`,
		tokenHash,
	).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.RevokedAt,
		&token.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RefreshToken{}, ErrNotFound
		}

		return model.RefreshToken{}, fmt.Errorf("mencari refresh token: %w", err)
	}

	return token, nil
}

func (r *tokenPostgresRepository) Revoke(
	ctx context.Context,
	id int64,
) error {
	result, err := r.pool.Exec(
		ctx,
		`UPDATE refresh_tokens
		 SET revoked_at = NOW()
		 WHERE id = $1
		   AND revoked_at IS NULL`,
		id,
	)
	if err != nil {
		return fmt.Errorf("mencabut refresh token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
