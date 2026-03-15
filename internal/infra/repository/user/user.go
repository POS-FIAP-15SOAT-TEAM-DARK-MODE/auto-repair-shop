package repository

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/jmoiron/sqlx"
)

type SqlxUserRepository struct {
	db *sqlx.DB
}

func NewSqlxUserRepository(db *sqlx.DB) *SqlxUserRepository {
	return &SqlxUserRepository{db: db}
}

func (u *SqlxUserRepository) GetByEmail(email string) (*domain.User, error) {
	return nil, nil
}

func (u *SqlxUserRepository) GetRolesById(id string) ([]string, error) {
	return nil, nil
}

func (u *SqlxUserRepository) Create(c *domain.User) (*domain.User, error) {
	// TODO: add logging
	_, err := u.db.Exec(CreateUser, c.Id, c.Name, c.Email, c.Password)
	if err != nil {
		return nil, postgres.Error(err)
	}
	return c, nil
}
