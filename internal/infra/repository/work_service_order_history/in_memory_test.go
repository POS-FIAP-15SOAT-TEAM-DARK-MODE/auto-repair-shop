package work_service_order_history_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work_service_order_history"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepo_Insert(t *testing.T) {
	repo := work_service_order_history.MemoryRepository()
	ctx := context.Background()

	from := domain.WORK_SERVICE_ORDER_STATUS("AWAITING")
	err := repo.Insert(ctx, "so-1", "w-1", &from, domain.WORK_SERVICE_ORDER_STATUS("IN_PROGRESS"))

	assert.NoError(t, err)
}

func TestMemoryRepo_Search(t *testing.T) {
	repo := work_service_order_history.MemoryRepository()
	ctx := context.Background()

	results, err := repo.Search(ctx, domain.SearchWorkSOHistoryParams{ServiceOrderID: "so-1"})

	assert.NoError(t, err)
	assert.Nil(t, results)
}
