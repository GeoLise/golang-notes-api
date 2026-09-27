package models

import (
	"errors"
	"time"
)

type User struct {
	Id         int       `json:"id"`
	TelegramID int64     `json:"telegram_id"`
	Username   string    `json:"username"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}

var ErrNotFound = errors.New("User not found")

func (u User) Validate() error {
	if u.Name == "" {
		return errors.New("Name is required")
	}

	return nil
}
