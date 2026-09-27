package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Repository struct {
	db    *pgxpool.Pool
	cache *redis.Client
}

func NewRepository(db *pgxpool.Pool, cache *redis.Client) *Repository {
	return &Repository{db: db, cache: cache}
}

func (r *Repository) Create(ctx context.Context, u User) (User, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (name, age)
		VALUES ($1, $2)
		RETURNING id, created_at`,
		u.Name, u.Age,
	).Scan(&u.Id, &u.CreatedAt)

	return u, err
}

func (r *Repository) GetAll(ctx context.Context) ([]User, error) {
	rows, err := r.db.Query(ctx, `SELECT * FROM users;`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.Id, &u.Name, &u.Age, &u.CreatedAt); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, rows.Err()
}

func (r *Repository) GetById(ctx context.Context, id int) (User, error) {

	key := fmt.Sprintf("user:%d", id)

	if data, err := r.cache.Get(ctx, key).Bytes(); err == nil {
		var u User
		if err := json.Unmarshal(data, &u); err == nil {
			return u, nil
		}
	}

	var u User

	err := r.db.QueryRow(ctx,
		`SELECT id, name, age, created_at FROM users WHERE id = $1`,
		id,
	).Scan(&u.Id, &u.Name, &u.Age, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrorNotFound
	}

	if err != nil {
		return User{}, err
	}

	if data, err := json.Marshal(u); err == nil {
		r.cache.Set(ctx, key, data, time.Second*60*60*24)
	}

	return u, err
}

func (r *Repository) Update(ctx context.Context, id int, u User) (User, error) {

	key := fmt.Sprintf("user:%d", id)

	err := r.db.QueryRow(ctx, `
		UPDATE users SET name = $1, age = $2 WHERE id = $3
		RETURNING id, name, age, created_at;
	`, u.Name, u.Age, id).Scan(&u.Id, &u.Name, &u.Age, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrorNotFound
	}

	if err != nil {
		return User{}, err
	}

	r.cache.Del(ctx, key)

	return u, err
}

func (r *Repository) Delete(ctx context.Context, id int) (User, error) {
	var u User
	err := r.db.QueryRow(ctx, `
		DELETE FROM users WHERE id = $1 RETURNING id, name, age, created_at;
		`, id).Scan(&u.Id, &u.Name, &u.Age, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrorNotFound
	}

	if err != nil {
		return User{}, err
	}

	return u, err
}
