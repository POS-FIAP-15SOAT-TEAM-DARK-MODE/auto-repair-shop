package domain

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string
	Name     string
	Email    string
	Password string
}

type UserService interface {
	Create(ctx context.Context, user *User) (*User, error)
}

type UserRepository interface {
	Create(ctx context.Context, u *User) error
}

//go:generate mockery --name=UserService --with-expecter
//go:generate mockery --name=UserRepository --with-expecter

var (
	ErrEmptyName         = errors.New("name cannot be empty")
	ErrEmptyEmail        = errors.New("email cannot be empty")
	ErrEmptyPassword     = errors.New("password cannot be empty")
	ErrPasswordDontMatch = errors.New("passwords do not match")
	ErrPasswordTooLong   = errors.New("password too long")
	ErrPasswordTooShort  = errors.New("password too short")
)

func NewUser(name, email, password string) *User {
	return &User{
		ID:       uuid.New().String(),
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
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return ErrPasswordTooLong
		}
		if errors.Is(err, bcrypt.ErrHashTooShort) {
			return ErrPasswordTooShort
		}
		return err
	}
	u.Password = string(hash)
	return nil
}
