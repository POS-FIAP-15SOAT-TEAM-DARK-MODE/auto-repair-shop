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

// Construtor
func NewSqlxUserRepository(db *sqlx.DB) *SqlxUserRepository {
	return &SqlxUserRepository{db: db}
}

func (r *SqlxUserRepository) FindByEmail(email string) (*domain.User, error) {
	user := &domain.User{}
	query := `SELECT email FROM "user" WHERE email = $1`
	err := r.db.Get(user, query, email)
	if err != nil {
		return nil, err
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

	query := `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	_, err = u.db.Exec(query, c.Id, c.Name, c.Email, c.Password)
	if err != nil {
		return nil, err
	}
	return c, nil

}
