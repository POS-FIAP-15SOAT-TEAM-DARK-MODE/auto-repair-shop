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

func NewUser(name, email, password string) *User {
	return &User{
		Id:       uuid.New().String(),
		Name:     name,
		Email:    email,
		Password: password,
	}
}

func (u *User) Validate() error {
	if strings.TrimSpace(u.Name) == "" || strings.TrimSpace(u.Email) == "" || strings.TrimSpace(u.Password) == "" {
		return errors.New("name, email and password hash cannot be empty")
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
