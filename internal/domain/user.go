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

type (
	User struct {
		ID       string
		Name     string
		Email    string
		Password string
	}

	LoginResponse struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expires_in"`
	}
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserService --with-expecter
type UserService interface {
	Create(ctx context.Context, req *User) error
	Login(ctx context.Context, user *User) (*LoginResponse, error)
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserRepository --with-expecter
type UserRepository interface {
	Create(ctx context.Context, c *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetRolesById(ctx context.Context, id string) ([]string, error)
}

var (
	ErrEmptyEmail         = errors.New("email cannot be empty")
	ErrEmptyPassword      = errors.New("password cannot be empty")
	ErrPasswordTooLong    = errors.New("password too long")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

var (
	passwordMinLength  = 8
	passordMaxLength   = 72
	minNameLength      = 3
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

func (u *User) IsValidName() error {
	if strings.TrimSpace(u.Name) == "" {
		return ErrEmptyUserName
	}
	if utf8.RuneCountInString(u.Name) < minNameLength {
		return ErrInvalidUserName
	}

	if !nameIsValid.MatchString(u.Name) {
		return ErrInvalidUserName
	}

	return nil
}
func (u *User) IsValidPassword() error {
	if u.Password == "" {
		return ErrEmptyUserPassword
	}

	if len(u.Password) > passordMaxLength {
		return ErrUserPasswordTooLong
	}
	if len(u.Password) < passwordMinLength {
		return ErrUserPasswordTooShort
	}
	if !passwordHasUpper.MatchString(u.Password) || !passwordHasSpecial.MatchString(u.Password) {
		return ErrInvalidPassword
	}

	return nil
}
func (u *User) IsValidEmail() error {
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
	if err := u.IsValidName(); err != nil {
		errs = append(errs, err)
	}
	if err := u.IsValidEmail(); err != nil {
		errs = append(errs, err)
	}
	if err := u.IsValidPassword(); err != nil {
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
