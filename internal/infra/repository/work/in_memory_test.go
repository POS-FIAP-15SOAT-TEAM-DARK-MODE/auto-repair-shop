package work_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_Save(t *testing.T) {
	repo := work.MemoryRepository()
	ctx := context.Background()

	w, _ := domain.NewWork("Oil Change", "Description", "100.00", domain.ACTIVE)
	err := repo.Save(ctx, w)

	assert.NoError(t, err)

	count, _ := repo.Count(ctx, &domain.SearchWorkParams{Limit: 10, Offset: 0})
	assert.Equal(t, int64(1), count)
}

func TestMemoryRepository_Count(t *testing.T) {
	repo := work.MemoryRepository()
	ctx := context.Background()

	w1, _ := domain.NewWork("Work 1", "Description 1", "10.00", domain.ACTIVE)
	w2, _ := domain.NewWork("Work 2", "Description 2", "20.00", domain.INACTIVE)
	_ = repo.Save(ctx, w1)
	_ = repo.Save(ctx, w2)

	// Total count
	count, _ := repo.Count(ctx, &domain.SearchWorkParams{})
	assert.Equal(t, int64(2), count)

	// Active count
	count, _ = repo.Count(ctx, &domain.SearchWorkParams{Status: "ACTIVE"})
	assert.Equal(t, int64(1), count)

	// Inactive count
	count, _ = repo.Count(ctx, &domain.SearchWorkParams{Status: "INACTIVE"})
	assert.Equal(t, int64(1), count)
}

func TestMemoryRepository_Search(t *testing.T) {
	repo := work.MemoryRepository()
	ctx := context.Background()

	w1, _ := domain.NewWork("Work 1", "Description 1", "10.00", domain.ACTIVE)
	w2, _ := domain.NewWork("Work 2", "Description 2", "20.00", domain.ACTIVE)
	_ = repo.Save(ctx, w1)
	_ = repo.Save(ctx, w2)

	// Search all with limit
	results, _ := repo.Search(ctx, &domain.SearchWorkParams{Limit: 1, Offset: 0})
	assert.Len(t, results, 1)

	// Search all with offset
	results, _ = repo.Search(ctx, &domain.SearchWorkParams{Limit: 10, Offset: 1})
	assert.Len(t, results, 1)

	// Search with status
	results, _ = repo.Search(ctx, &domain.SearchWorkParams{Limit: 10, Offset: 0, Status: "ACTIVE"})
	assert.Len(t, results, 2)

	// Search with empty results due to offset
	results, _ = repo.Search(ctx, &domain.SearchWorkParams{Limit: 10, Offset: 10})
	assert.Len(t, results, 0)
}

func TestMemoryRepository_Delete(t *testing.T) {
	repo := work.MemoryRepository()
	ctx := context.Background()

	w, _ := domain.NewWork("Work", "Description", "10.00", domain.ACTIVE)
	_ = repo.Save(ctx, w)

	err := repo.Delete(ctx, w.ID)
	assert.NoError(t, err)

	count, _ := repo.Count(ctx, &domain.SearchWorkParams{})
	assert.Equal(t, int64(0), count)
}
