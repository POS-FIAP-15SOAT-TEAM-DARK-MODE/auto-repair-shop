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

func (r *memory_repo) GetByDocument(ctx context.Context, document string) (domain.Customer, error) {
	for _, c := range r.data {
		if c.CPF == document || c.CNPJ == document {
			return c, nil
		}
	}

	return domain.Customer{}, domain.ErrCustomerNotFound
}

func (r *memory_repo) GetByID(ctx context.Context, id string) (domain.Customer, error) {
	c, ok := r.data[id]
	if !ok {
		return c, nil
	}

	return domain.Customer{}, domain.ErrCustomerNotFound
}

func (r *memory_repo) Update(_ context.Context, id, phone string) error {
	c, ok := r.data[id]
	if !ok {
		return domain.ErrCustomerNotFound
	}
	c.Phone = phone
	r.data[id] = c
	return nil
}

func (r *memory_repo) Delete(_ context.Context, id string) error {
	if _, ok := r.data[id]; !ok {
		return domain.ErrCustomerNotFound
	}
	delete(r.data, id)
	return nil
}

func (r *memory_repo) GetByUserID(_ context.Context, id string) (domain.Customer, error) {
	for _, c := range r.data {
		if c.UserID == id {
			return c, nil
		}
	}

	return domain.Customer{}, domain.ErrCustomerNotFound
}
