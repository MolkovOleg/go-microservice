package models

import (
	"errors"
	"regexp"
	"sync/atomic"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("Name is required")
	}
	if u.Email == "" {
		return errors.New("Email is required")
	}
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("Invalid email format")
	}
	return nil
}

var nextID int64

func GenerateID() int {
	return int(atomic.AddInt64(&nextID, 1))
}