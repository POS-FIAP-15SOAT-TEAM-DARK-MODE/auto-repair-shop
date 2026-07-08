package repository_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSOMemory_UpdatePricing_UpdatesTotalAndPreservesStatus(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewSOMemory()

	so := &domain.ServiceOrder{
		ID:          "so-1",
		Status:      domain.SERVICE_ORDER_STATUS_RECEIVED,
		TotalAmount: decimal.NewFromInt(100),
	}
	require.NoError(t, repo.Save(ctx, so))

	newTotal := decimal.NewFromInt(250)
	require.NoError(t, repo.UpdatePricing(ctx, "so-1", newTotal))

	got, err := repo.FindByID(ctx, "so-1")
	require.NoError(t, err)
	assert.True(t, got.TotalAmount.Equal(newTotal), "total should be updated")
	assert.Equal(t, domain.SERVICE_ORDER_STATUS_RECEIVED, got.Status, "status must be left untouched")
}

func TestSOMemory_UpdatePricing_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewSOMemory()

	err := repo.UpdatePricing(ctx, "missing", decimal.NewFromInt(10))
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}
