package storage

import (
	"context"
	"errors"
	"notes-api/internal/auth"
	"notes-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sessions struct {
	db *pgxpool.Pool
}

func NewSessions(db *pgxpool.Pool) *Sessions {
	return &Sessions{db: db}
}

func (r *Sessions) Create(ctx context.Context, s models.Session) (models.Session, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`, s.UserId, s.TokenHash, s.ExpiresAt).Scan(&s.Id, &s.CreatedAt)

	return s, err
}

func (r *Sessions) GetByToken(ctx context.Context, token string) (models.User, error) {

	var u models.User
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.telegram_id, u.username, u.name, u.created_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > now()
	`, auth.HashToken(token)).Scan(&u.Id, &u.TelegramID, &u.Username, &u.Name, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrSessionExpired
	}

	if err != nil {
		return models.User{}, err
	}

	return u, err
}
