package service_order_history_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	sohrepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order_history"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository(t *testing.T) {
	repo := sohrepo.MemoryRepository()
	ctx := context.Background()
	params := &domain.SearchServiceOrderHistoryParams{ID: "so-1"}

	t.Run("Search returns empty slice without error", func(t *testing.T) {
		histories, err := repo.Search(ctx, params)
		assert.NoError(t, err)
		assert.Empty(t, histories)
	})

	t.Run("Count returns zero without error", func(t *testing.T) {
		total, err := repo.Count(ctx, params)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
	})

	t.Run("WorkTimelineByServiceOrderID returns empty slice without error", func(t *testing.T) {
		timelines, err := repo.WorkTimelineByServiceOrderID(ctx, "so-1")
		assert.NoError(t, err)
		assert.Empty(t, timelines)
	})
}
