package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"notes-api/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Users struct {
	db    *pgxpool.Pool
	cache *redis.Client
}

func NewUsers(db *pgxpool.Pool, cache *redis.Client) *Users {
	return &Users{db: db, cache: cache}
}

func cacheKey(name string, id int) string {
	return fmt.Sprintf("%s:%d", name, id)
}

func (r *Users) Create(ctx context.Context, u models.User) (models.User, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name)
		VALUES ($1, $2)
		RETURNING id, created_at`,
		u.Name,
	).Scan(&u.Id, &u.CreatedAt)

	return u, err
}

func (r *Users) GetAll(ctx context.Context) ([]models.User, error) {
	rows, err := r.db.Query(ctx, `SELECT * FROM users;`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []models.User{}
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.Id, &u.Name, &u.CreatedAt); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *Users) GetById(ctx context.Context, id int) (models.User, error) {

	key := cacheKey("user", id)

	if data, err := r.cache.Get(ctx, key).Bytes(); err == nil {
		var u models.User
		if err := json.Unmarshal(data, &u); err == nil {
			return u, nil
		}
	}

	var u models.User

	err := r.db.QueryRow(ctx,
		`SELECT id, name, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.Id, &u.Name, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrNotFound
	}

	if err != nil {
		return models.User{}, err
	}

	if data, err := json.Marshal(u); err == nil {
		r.cache.Set(ctx, key, data, time.Second*60*60*24)
	}

	return u, err
}

func (r *Users) Update(ctx context.Context, id int, u models.User) (models.User, error) {

	err := r.db.QueryRow(ctx, `
		UPDATE users SET name = $1, WHERE id = $3
		RETURNING id, name, age, created_at;
	`, u.Name, id).Scan(&u.Id, &u.Name, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrNotFound
	}

	if err != nil {
		return models.User{}, err
	}

	r.cache.Del(ctx, cacheKey("user", u.Id))

	return u, err
}

func (r *Users) Delete(ctx context.Context, id int) (models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
		DELETE FROM users WHERE id = $1 RETURNING id, name, age, created_at;
		`, id).Scan(&u.Id, &u.Name, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, models.ErrNotFound
	}

	if err != nil {
		return models.User{}, err
	}

	return u, err
}

func (r *Users) UpsertByTelegram(ctx context.Context, telegramID int64, username, name string) (models.User, error) {
	var u models.User

	err := r.db.QueryRow(ctx,
		`INSERT INTO users (telegram_id, username, name)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (telegram_id)
		 DO UPDATE SET username = EXCLUDED.username, name = EXCLUDED.name
		 RETURNING id, telegram_id, username, name, created_at`,
		telegramID, username, name,
	).Scan(&u.Id, &u.TelegramID, &u.Username, &u.Name, &u.CreatedAt)
	if err != nil {
		return models.User{}, err
	}

	r.cache.Del(ctx, cacheKey("user", u.Id))
	return u, nil
}
