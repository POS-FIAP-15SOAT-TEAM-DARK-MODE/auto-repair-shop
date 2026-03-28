package user_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_Create(t *testing.T) {
	repo := user.MemoryRepository()
	ctx := context.Background()

	u := domain.NewUser("Test User", "test@example.com", "Secret@123")
	err := repo.Create(ctx, u)

	assert.NoError(t, err)
}
