package models

import (
	"errors"
	"time"
)

type Note struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrNoteValidationError = errors.New("")

func (n *Note) Validate() error {
	if n.Name == "" {
		return errors.New("Name is required")
	}
	if n.Content == "" {
		return errors.New("Content in reuired")
	}

	return nil
}
