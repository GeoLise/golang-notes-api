package storage

import (
	"context"
	"encoding/json"
	"errors"
	"notes-api/internal/models"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Notes struct {
	db    *pgxpool.Pool
	cache *redis.Client
}

func NewNotes(db *pgxpool.Pool, cache *redis.Client) *Notes {
	return &Notes{db: db, cache: cache}
}

func (s *Notes) Create(ctx context.Context, n models.Note) (models.Note, error) {
	err := s.db.QueryRow(ctx, `
		INSERT INTO notes (user_id, name, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at;
	`, n.UserId, n.Name, n.Content).Scan(&n.Id, &n.CreatedAt)

	s.cache.Del(ctx, getCacheKey("notes", strconv.Itoa(n.UserId)))

	return n, err
}

func (s *Notes) Update(ctx context.Context, id int, n models.Note) (models.Note, error) {
	err := s.db.QueryRow(ctx, `
		UPDATE notes SET name = $1, content = $2, 
		WHERE id = $3 AND user_id = $4
		RETURNING id, name, content, created_at, user_id;
	`, n.Name, n.Content, id, n.UserId).Scan(&n.Id, &n.Name, &n.Content, &n.CreatedAt, &n.UserId)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Note{}, models.ErrNotFound
	}

	s.cache.Del(ctx, getCacheKey("notes", strconv.Itoa(n.UserId)))

	return n, nil
}

func (s *Notes) Get(ctx context.Context, userid int) ([]models.Note, error) {
	cacheKey := getCacheKey("notes", strconv.Itoa(userid))

	if data, err := s.cache.Get(ctx, cacheKey).Bytes(); err == nil {
		var n []models.Note
		if err := json.Unmarshal(data, &n); err == nil {
			return n, nil
		}
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, name, content, user_id, created_at FROM notes WHERE user_id = $1;
	`, userid)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	notes := []models.Note{}

	for rows.Next() {
		var n models.Note
		if err := rows.Scan(&n.Id, &n.Name, &n.Content, &n.UserId, &n.CreatedAt); err != nil {
			return nil, err
		}

		notes = append(notes, n)
	}

	if data, err := json.Marshal(notes); err == nil {
		s.cache.Set(ctx, cacheKey, data, DEFAULT_TTL)
	}

	return notes, nil

}
