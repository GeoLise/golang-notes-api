package models

import (
	"errors"
	"time"
)

type Session struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrSessionExpired = errors.New("Session expired")

func (s Session) ChechExpired() error {
	if s.ExpiresAt.Before(time.Now()) {
		return ErrSessionExpired
	}

	return nil
}

func (s Session) Validate() error {
	if s.TokenHash == "" {
		return errors.New("Token hash is required")
	}

	if s.ExpiresAt.IsZero() {
		return errors.New("Expires at is required")
	}

	return nil
}
