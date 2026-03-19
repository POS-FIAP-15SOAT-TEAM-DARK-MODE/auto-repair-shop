package domain

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string
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

var (
	passwordHasUpper   = regexp.MustCompile(`[A-Z]`)
	passwordHasSpecial = regexp.MustCompile(`[^a-zA-Z0-9]`)
	nameIsValid        = regexp.MustCompile(`^[a-zA-ZÀ-ÿ\s]+$`)
)

func NewUser(name, email, password string) *User {
	return &User{
		ID:       uuid.New().String(),
		Name:     name,
		Email:    email,
		Password: password,
	}
}

func CreateUserToDomain(name, email, password string) (*User, error) {
	u := NewUser(name, email, password)
	if err := u.Validate(); err != nil {
		return nil, err
	}
	if err := u.HashPassword(context.Background()); err != nil {
		return nil, err
	}
	return u, nil
}

func IsValidName(u *User) error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrEmptyUserName
	}
	if utf8.RuneCountInString(u.Name) < 3 {
		return ErrInvalidUserName
	}

	if !nameIsValid.MatchString(u.Name) {
		return ErrInvalidUserName
	}

	return nil
}
func IsValidPassword(u *User) error {
	if u.Password == "" {
		return ErrEmptyUserPassword
	}

	if len(u.Password) > 72 {
		return ErrUserPasswordTooLong
	}
	if len(u.Password) < 8 {
		return ErrUserPasswordTooShort
	}
	if !passwordHasUpper.MatchString(u.Password) || !passwordHasSpecial.MatchString(u.Password) {
		return ErrInvalidPassword
	}

	return nil
}
func IsValidEmail(u *User) error {
	email := strings.TrimSpace(u.Email)
	if email == "" {
		return ErrEmptyUserEmail
	}

	if strings.Contains(email, " ") {
		return ErrInvalidUserEmail
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ErrInvalidUserEmail
	}

	localPart := parts[0]
	domainPart := parts[1]

	if localPart == "" || domainPart == "" {
		return ErrInvalidUserEmail
	}

	if !strings.Contains(domainPart, ".") {
		return ErrInvalidUserEmail
	}

	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return ErrInvalidUserEmail
	}

	return nil
}
func (u *User) Validate() error {
	var errs []error
	if err := IsValidName(u); err != nil {
		errs = append(errs, err)
	}
	if err := IsValidEmail(u); err != nil {
		errs = append(errs, err)
	}
	if err := IsValidPassword(u); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
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
