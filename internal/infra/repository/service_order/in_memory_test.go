package service_order_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	service_order "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func buildServiceOrder() *domain.ServiceOrder {
	customer := &domain.Customer{ID: "customer-id"}
	vehicle := &domain.Vehicle{ID: "vehicle-id"}
	return domain.NewServiceOrder(customer, vehicle)
}

func TestMemoryRepository_Save(t *testing.T) {
	repo := service_order.MemoryRepository()
	ctx := context.Background()

	so := buildServiceOrder()
	err := repo.Save(ctx, so)

	assert.NoError(t, err)

	found, err := repo.FindByID(ctx, so.ID)
	assert.NoError(t, err)
	assert.Equal(t, so.ID, found.ID)
}

func TestMemoryRepository_ExistsByID(t *testing.T) {
	repo := service_order.MemoryRepository()
	ctx := context.Background()
	so := buildServiceOrder()
	_ = repo.Save(ctx, so)

	ok, status, err := repo.ExistsByID(ctx, so.ID)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, so.Status, status)

	ok, _, err = repo.ExistsByID(ctx, "non-existent")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestMemoryRepository_WorkLinks(t *testing.T) {
	repo := service_order.MemoryRepository()
	ctx := context.Background()
	soID := "so-1"

	err := repo.AddWorkLink(ctx, soID, "w-1", decimal.NewFromInt(100))
	assert.NoError(t, err)

	works, err := repo.ListWorksByServiceOrderID(ctx, soID)
	assert.NoError(t, err)
	assert.Len(t, works, 1)
	assert.Equal(t, "w-1", works[0].ID)

	err = repo.RemoveWorkLink(ctx, soID, "w-1")
	assert.NoError(t, err)

	works, _ = repo.ListWorksByServiceOrderID(ctx, soID)
	assert.Empty(t, works)
}

func TestMemoryRepository_SupplyLinks(t *testing.T) {
	repo := service_order.MemoryRepository()
	ctx := context.Background()
	soID := "so-1"

	err := repo.AddSupplyLink(ctx, soID, "sup-1", 5, decimal.NewFromInt(10))
	assert.NoError(t, err)

	supplies, err := repo.ListSuppliesByServiceOrderID(ctx, soID)
	assert.NoError(t, err)
	assert.Len(t, supplies, 1)
	assert.Equal(t, "sup-1", supplies[0].ID)
	assert.Equal(t, 5, supplies[0].StockQuantity)

	qty, err := repo.RemoveSupplyLink(ctx, soID, "sup-1")
	assert.NoError(t, err)
	assert.Equal(t, 5, qty)
}

func TestMemoryRepository_SearchAndCount(t *testing.T) {
	repo := service_order.MemoryRepository()
	ctx := context.Background()

	so1 := buildServiceOrder()
	so1.ID = "a"
	so1.Status = domain.SERVICE_ORDER_STATUS_NEW
	so1.Customer = &domain.Customer{ID: "c1"}
	so1.Vehicle = &domain.Vehicle{ID: "v1"}
	_ = repo.Save(ctx, so1)

	so2 := buildServiceOrder()
	so2.ID = "b"
	so2.Status = domain.SERVICE_ORDER_STATUS_COMPLETED
	so2.Customer = &domain.Customer{ID: "c2"}
	so2.Vehicle = &domain.Vehicle{ID: "v2"}
	_ = repo.Save(ctx, so2)

	so3 := buildServiceOrder()
	so3.ID = "c"
	so3.Status = domain.SERVICE_ORDER_STATUS_COMPLETED
	so3.Customer = nil
	so3.Vehicle = nil
	_ = repo.Save(ctx, so3)

	params := &domain.ServiceOrderFilterParams{Status: "NEW"}
	count, _ := repo.Count(ctx, params)
	assert.Equal(t, int64(1), count)

	params = &domain.ServiceOrderFilterParams{CustomerID: "c2"}
	count, _ = repo.Count(ctx, params)
	assert.Equal(t, int64(1), count)

	params = &domain.ServiceOrderFilterParams{VehicleID: "v2"}
	count, _ = repo.Count(ctx, params)
	assert.Equal(t, int64(1), count)

	params = &domain.ServiceOrderFilterParams{CustomerID: "c1", Limit: 10}
	items, _ := repo.Search(ctx, params)
	assert.Len(t, items, 1)

	params = &domain.ServiceOrderFilterParams{VehicleID: "v1", Limit: 10}
	items, _ = repo.Search(ctx, params)
	assert.Len(t, items, 1)

	items, _ = repo.Search(ctx, &domain.ServiceOrderFilterParams{Limit: 10, Offset: 0})
	assert.Len(t, items, 3)

	items, _ = repo.Search(ctx, &domain.ServiceOrderFilterParams{Limit: 1, Offset: 1})
	assert.Len(t, items, 1)
	assert.Equal(t, "b", items[0].ID)

	items, _ = repo.Search(ctx, &domain.ServiceOrderFilterParams{Limit: 10, Offset: 10})
	assert.Empty(t, items)
}

func TestMemoryRepository_AverageExecutionTimeInHours(t *testing.T) {
	repo := service_order.MemoryRepository()
	res, err := repo.AverageExecutionTimeInHours(context.Background(), nil)
	assert.NoError(t, err)
	assert.Empty(t, res)
}
