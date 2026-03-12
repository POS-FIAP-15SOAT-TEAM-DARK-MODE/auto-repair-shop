package customer

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/jmoiron/sqlx"
)

type (
	repository struct {
		db *sqlx.DB
	}
)

func Repository(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Create(customer *domain.Customer) error {
	return nil
}
