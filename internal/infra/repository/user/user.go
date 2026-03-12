package repository

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/jmoiron/sqlx"
)

type SqlxUserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *SqlxUserRepository {
	return &SqlxUserRepository{db: db}
}

func (u *SqlxUserRepository) Create(c *domain.User) (*domain.User, error) {
	_, err := u.db.Exec(CreateUser, c.ID, c.Name, c.Email, c.Password)
	if err != nil {
		return nil, err
	}
	return c, nil
}
