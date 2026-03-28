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
