package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       string
	Name     string
	Email    string
	Password string
}

type UserService interface {
	Create(req *User) (*User, error)
}

type UserRepository interface {
	Create(c *User) (*User, error)
}

var (
	ErrEmptyName     = errors.New("name cannot be empty")
	ErrEmptyEmail    = errors.New("email cannot be empty")
	ErrEmptyPassword = errors.New("password cannot be empty")
)

func NewUser(name, email, password string) *User {
	return &User{
		Id:       uuid.New().String(),
		Name:     name,
		Email:    email,
		Password: password,
	}
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrEmptyName
	}

	if strings.TrimSpace(u.Email) == "" {
		return ErrEmptyEmail
	}

	if strings.TrimSpace(u.Password) == "" {
		return ErrEmptyPassword
	}

	return nil

}

func (u *User) HashPassword() error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return nil
}
