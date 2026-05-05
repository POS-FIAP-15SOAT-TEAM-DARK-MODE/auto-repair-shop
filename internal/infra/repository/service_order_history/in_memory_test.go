package service_order_history_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	sohrepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order_history"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_Search(t *testing.T) {
	repo := sohrepo.MemoryRepository()

	items, err := repo.Search(context.Background(), &domain.SearchServiceOrderHistoryParams{ID: "so-1"})

	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestMemoryRepository_SearchWorkTransitions(t *testing.T) {
	repo := sohrepo.MemoryRepository()

	groups, err := repo.SearchWorkTransitionsByServiceOrderID(context.Background(), "so-1")

	assert.NoError(t, err)
	assert.Empty(t, groups)
}

func TestMemoryRepository_InsertWorkHistory(t *testing.T) {
	repo := sohrepo.MemoryRepository()

	err := repo.InsertWorkHistory(context.Background(), "so-1", "w-1", domain.SERVICE_ORDER_STATUS_NEW)

	assert.NoError(t, err)
}
