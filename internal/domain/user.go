package domain

import (
	"context"
	"errors"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       string
	Name     string
	Email    string
	Password string
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserService --with-expecter
type UserService interface {
	Create(ctx context.Context, req *User) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserRepository --with-expecter
type UserRepository interface {
	Create(ctx context.Context, c *User) error
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
	var err []error
	if strings.TrimSpace(u.Name) == "" {
		err = append(err, ErrEmptyUserName)
	}

	if strings.TrimSpace(u.Email) == "" {
		err = append(err, ErrEmptyUserEmail)
	}

	if strings.TrimSpace(u.Password) == "" {
		err = append(err, ErrEmptyUserPassword)
	}

	// TODO: Should have a minimun length for Password and Name? If true, add the errors.
	// TODO: Should validate e-mail format and add errors.

	if len(err) > 0 {
		return errors.Join(err...)
	}

	return nil
}

var hashingCost int

func init() {
	hashingCost = env.GetInt("BCRYPT_COST", bcrypt.DefaultCost)
}

func (u *User) HashPassword(ctx context.Context) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), hashingCost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return ErrUserPasswordTooLong
		}

		if errors.Is(err, bcrypt.ErrHashTooShort) {
			return ErrUserPasswordTooShort
		}

		logger.Of(ctx).Warn("HashPassword failed: bcrypt error", zap.String("email", u.Email), zap.Error(err))
		return err
	}
	u.Password = string(hash)
	return nil
}
