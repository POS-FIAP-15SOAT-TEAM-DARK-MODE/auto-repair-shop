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

	LoggedUser struct {
		User
		SessionToken     string
		SessionExpiresIn int
	}
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserService --with-expecter
type UserService interface {
	Create(ctx context.Context, req *User) error
	Login(ctx context.Context, loggedUser *LoggedUser) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=UserRepository --with-expecter
type UserRepository interface {
	Create(ctx context.Context, c *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetRolesByUserId(ctx context.Context, id string) ([]Role, error)
	AssignRole(ctx context.Context, userID, roleName string) error
	Update(ctx context.Context, id, name, email string) error
	Delete(ctx context.Context, id string) error
}

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
func (u *User) IsPasswordEmpty() bool {
	return u.Password == ""
}
func (u *User) IsEmailEmpty() bool {
	return u.Email == ""
}
func (u *User) IsValidPassword() error {
	if u.IsPasswordEmpty() {
		return ErrEmptyUserPassword
	}

	if len(u.Password) > passordMaxLength {
		return ErrUserPasswordTooLong
	}
	if len(u.Password) < passwordMinLength {
		return ErrUserPasswordTooShort
	}
	if !passwordHasUpper.MatchString(u.Password) || !passwordHasSpecial.MatchString(u.Password) {
		return ErrInvalidUserPassword
	}

	return nil
}
func (u *User) IsValidEmail() error {
	email := strings.TrimSpace(u.Email)
	if u.IsEmailEmpty() {
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
func (u *LoggedUser) Validate() error {
	var errs []error
	if u.IsEmailEmpty() {
		errs = append(errs, ErrEmptyUserEmail)
	}
	if u.IsPasswordEmpty() {
		errs = append(errs, ErrEmptyUserPassword)
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

		logger.Of(ctx).Warn("HashPassword failed: bcrypt error", zap.String("email", u.Email), zap.Error(err))
		return err
	}
	u.Password = string(hash)
	return nil
}
