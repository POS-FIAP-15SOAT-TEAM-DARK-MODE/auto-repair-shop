package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SqlxUserRepository struct {
	db *sqlx.DB
}

func NewSqlxUserRepository(db *sqlx.DB) *SqlxUserRepository {
	return &SqlxUserRepository{db: db}
}

func (r *SqlxUserRepository) FindByEmail(email string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.Get(user, GetUserByEmail, email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return user, nil
}

func (u *SqlxUserRepository) CreateUser(c *domain.User) (*domain.User, error) {
	emailExists, err := u.FindByEmail(c.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if emailExists != nil {
		return nil, fmt.Errorf("user with email %s already exists", c.Email)
	}
	_, err = u.db.Exec(CreateUser, c.Id, c.Name, c.Email, c.Password)
	if err != nil {
		return nil, err
	}
	return c, nil

}
