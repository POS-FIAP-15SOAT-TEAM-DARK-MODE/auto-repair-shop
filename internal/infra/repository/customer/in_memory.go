package customer

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.Customer
}

func MemoryRepository() domain.CustomerRepository {
	return &memory_repo{
		data: make(map[string]domain.Customer),
	}
}

func (r *memory_repo) Create(_ context.Context, c *domain.Customer) error {
	r.data[c.ID] = *c
	return nil
}
