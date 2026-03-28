package customer_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_Create(t *testing.T) {
	repo := customer.MemoryRepository()
	ctx := context.Background()

	c, _ := domain.NewCustomer(
		"João Silva",
		"joao@example.com",
		"Secret@123",
		domain.IndividualCustomerType,
		"111.444.777-35",
		"",
		"11999999999",
	)
	err := repo.Create(ctx, &c)

	assert.NoError(t, err)
}
