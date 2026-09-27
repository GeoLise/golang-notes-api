package user

import (
	"errors"
	"time"
)

type User struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Age       int       `json:"age"`
	CreatedAt time.Time `json:"created_at"`
}

var ErrorNotFound = errors.New("User not found")

func (u User) Validate() error {
	if u.Name == "" {
		return errors.New("Name is required")
	}

	if u.Age <= 0 {
		return errors.New("Age must be greater than 0")
	}

	return nil
}
