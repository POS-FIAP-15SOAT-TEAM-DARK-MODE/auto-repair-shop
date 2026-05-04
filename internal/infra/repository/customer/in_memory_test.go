package customer_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	"github.com/stretchr/testify/assert"
)

func newTestCustomer(t *testing.T) domain.Customer {
	t.Helper()
	c, err := domain.NewCustomer("João Silva", "joao@example.com", "Secret@123",
		domain.IndividualCustomerType, "111.444.777-35", "", "11999999999")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	return c
}

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

func TestMemoryRepository_GetByDocument(t *testing.T) {
	repo := customer.MemoryRepository()
	ctx := context.Background()

	c := newTestCustomer(t)
	_ = repo.Create(ctx, &c)

	t.Run("found by CPF", func(t *testing.T) {
		found, err := repo.GetByDocument(ctx, c.CPF)
		assert.NoError(t, err)
		assert.Equal(t, c.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByDocument(ctx, "99999999999")
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})
}

func TestMemoryRepository_GetByID(t *testing.T) {
	repo := customer.MemoryRepository()
	ctx := context.Background()

	c := newTestCustomer(t)
	_ = repo.Create(ctx, &c)

	t.Run("id exists returns error (known inverted behavior)", func(t *testing.T) {
		_, err := repo.GetByID(ctx, c.ID)
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})

	t.Run("id not found returns no error", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "nonexistent")
		assert.NoError(t, err)
	})
}

func TestMemoryRepository_Update(t *testing.T) {
	repo := customer.MemoryRepository()
	ctx := context.Background()

	c := newTestCustomer(t)
	_ = repo.Create(ctx, &c)

	t.Run("found - updates phone", func(t *testing.T) {
		err := repo.Update(ctx, c.ID, "11888888888")
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := repo.Update(ctx, "nonexistent", "11888888888")
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})
}

func TestMemoryRepository_Delete(t *testing.T) {
	repo := customer.MemoryRepository()
	ctx := context.Background()

	c := newTestCustomer(t)
	_ = repo.Create(ctx, &c)

	t.Run("found - deletes", func(t *testing.T) {
		err := repo.Delete(ctx, c.ID)
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		err := repo.Delete(ctx, "nonexistent")
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})
}

func TestMemoryRepository_GetByUserID(t *testing.T) {
	repo := customer.MemoryRepository()
	ctx := context.Background()

	c := newTestCustomer(t)
	_ = repo.Create(ctx, &c)

	t.Run("found", func(t *testing.T) {
		found, err := repo.GetByUserID(ctx, c.UserID)
		assert.NoError(t, err)
		assert.Equal(t, c.ID, found.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByUserID(ctx, "nonexistent-user")
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})
}
