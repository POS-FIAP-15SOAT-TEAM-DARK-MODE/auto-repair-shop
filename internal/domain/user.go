package domain

import (
	"errors"

	"github.com/google/uuid"
)

type User struct {
	Id       string
	Name     string
	Email    string
	Password string
}

type UserRepository interface {
	CreateUser(c *User) (*User, error)
	FindByEmail(email string) (*User, error)
}

func NewUser(name, email, password string) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email and password hash cannot be empty")
	}

	return &User{
		Id:       uuid.New().String(),
		Name:     name,
		Email:    email,
		Password: password,
	}, nil
}
