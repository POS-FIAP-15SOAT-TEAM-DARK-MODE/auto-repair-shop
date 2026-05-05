package supply_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/supply"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_DecrementStock(t *testing.T) {
	repo := supply.MemoryRepository()
	ctx := context.Background()

	s := domain.NewSupply("Oil", "Motor oil", decimal.NewFromInt(50), 10, 1)
	_ = repo.Save(ctx, s)

	err := repo.DecrementStock(ctx, s.ID, 3)
	assert.NoError(t, err)

	found, _ := repo.FindById(ctx, s.ID)
	assert.Equal(t, 7, found.StockQuantity)

	err = repo.DecrementStock(ctx, s.ID, 10)
	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
}

func TestMemoryRepository_RestoreStock(t *testing.T) {
	repo := supply.MemoryRepository()
	ctx := context.Background()

	s := domain.NewSupply("Oil", "Motor oil", decimal.NewFromInt(50), 10, 1)
	_ = repo.Save(ctx, s)

	err := repo.RestoreStock(ctx, s.ID, 5)
	assert.NoError(t, err)

	found, _ := repo.FindById(ctx, s.ID)
	assert.Equal(t, 15, found.StockQuantity)

	err = repo.RestoreStock(ctx, "non-existent", 5)
	assert.ErrorIs(t, err, domain.ErrSupplyNotFound)
}

func TestMemoryRepository_FindById_NotFound(t *testing.T) {
	repo := supply.MemoryRepository()
	_, err := repo.FindById(context.Background(), "none")
	assert.ErrorIs(t, err, domain.ErrSupplyNotFound)
}

func TestMemoryRepository_Delete(t *testing.T) {
	repo := supply.MemoryRepository()
	ctx := context.Background()
	s := domain.NewSupply("Oil", "Motor oil", decimal.NewFromInt(50), 10, 1)
	_ = repo.Save(ctx, s)

	err := repo.Delete(ctx, s.ID)
	assert.NoError(t, err)

	_, err = repo.FindById(ctx, s.ID)
	assert.ErrorIs(t, err, domain.ErrSupplyNotFound)
}

func TestMemoryRepository_SearchAndCount(t *testing.T) {
	repo := supply.MemoryRepository()
	ctx := context.Background()

	s1 := domain.NewSupply("A", "Desc 10 chars", decimal.NewFromInt(10), 10, 1)
	s1.ID = "1"
	_ = repo.Save(ctx, s1)

	s2 := domain.NewSupply("B", "Desc 10 chars", decimal.NewFromInt(20), 20, 1)
	s2.ID = "2"
	_ = repo.Save(ctx, s2)

	count, _ := repo.Count(ctx, nil)
	assert.Equal(t, int64(2), count)

	items, _ := repo.Search(ctx, &domain.ListSupplyParams{PageSize: 10, Page: 1})
	assert.Len(t, items, 2)
	assert.Equal(t, "1", items[0].ID)

	items, _ = repo.Search(ctx, &domain.ListSupplyParams{PageSize: 1, Page: 2})
	assert.Len(t, items, 1)
	assert.Equal(t, "2", items[0].ID)

	items, _ = repo.Search(ctx, &domain.ListSupplyParams{PageSize: 10, Page: 10})
	assert.Empty(t, items)
}

func TestMemoryRepository_DecrementStock_NotFound(t *testing.T) {
	repo := supply.MemoryRepository()
	err := repo.DecrementStock(context.Background(), "none", 1)
	assert.ErrorIs(t, err, domain.ErrSupplyNotFound)
}
