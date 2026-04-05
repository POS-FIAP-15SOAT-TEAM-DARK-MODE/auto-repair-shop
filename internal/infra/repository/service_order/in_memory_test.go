package service_order_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	service_order "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order"
	"github.com/stretchr/testify/assert"
)

func buildServiceOrder() *domain.ServiceOrder {
	customer := &domain.Customer{ID: "customer-id"}
	vehicle := domain.NewVehicle("ABC-1234", "Toyota", "Corolla", customer.ID, 2020)
	return domain.NewServiceOrder(customer, vehicle)
}

func TestMemoryRepository_Save(t *testing.T) {
	repo := service_order.MemoryRepository()
	ctx := context.Background()

	so := buildServiceOrder()
	err := repo.Save(ctx, so)

	assert.NoError(t, err)
}

func TestMemoryRepository_SaveAndGetHistory(t *testing.T) {
	repo := service_order.MemoryRepository()

	// create a simple service order
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	v := domain.NewVehicle("", "", "", "", 2026)
	so := domain.NewServiceOrder(&c, v)

	// Save should not error
	err := repo.Save(context.Background(), so)
	assert.NoError(t, err)

	// Current in-memory implementation returns empty history slice
	hist, err := repo.GetHistoryByID(context.Background(), so.ID)
	assert.NoError(t, err)
	assert.Empty(t, hist)
}

func TestMemoryRepository_GetHistoryByID_Unknown(t *testing.T) {
	repo := service_order.MemoryRepository()

	hist, err := repo.GetHistoryByID(context.Background(), "non-existent")
	assert.NoError(t, err)
	assert.Empty(t, hist)
}
