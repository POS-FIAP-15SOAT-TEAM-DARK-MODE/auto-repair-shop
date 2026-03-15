package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type (
	User struct {
		Id       string
		Name     string
		Email    string
		Password string
	}

	LoginResponse struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
)

//go:generate mockery --name=UserService --output=mocks --outpkg=mocks
type UserService interface {
	Create(req *User) (*User, error)
	Login(user *User) (*LoginResponse, error)
}

//go:generate mockery --name=UserRepository --output=mocks --outpkg=mocks
type UserRepository interface {
	Create(c *User) (*User, error)
	GetByEmail(email string) (*User, error)
	GetRolesById(id string) ([]string, error)
}

var (
	ErrEmptyName          = errors.New("name cannot be empty")
	ErrEmptyEmail         = errors.New("email cannot be empty")
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrPasswordDontMatch  = errors.New("passwords do not match")
	ErrPasswordTooLong    = errors.New("password too long")
	ErrPasswordTooShort   = errors.New("password too short")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
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

	// TODO: Should have a minimun length for Password and Name? If true, add the errors.
	// TODO: Should validate e-mail format and add errors.

	return nil
}

func (u *User) HashPassword() error {
	// TODO: Use .env variable and add logging
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
