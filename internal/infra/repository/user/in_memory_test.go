package user_test

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository_FullFlow(t *testing.T) {
	repo := user.MemoryRepository()
	ctx := context.Background()

	u := domain.NewUser("Test User", "test@example.com", "Secret@123")
	u.ID = "u1"
	_ = repo.Create(ctx, u)

	t.Run("GetByEmail", func(t *testing.T) {
		found, err := repo.GetByEmail(ctx, "test@example.com")
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "u1", found.ID)

		notFound, err := repo.GetByEmail(ctx, "other@e.com")
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("Roles", func(t *testing.T) {
		_ = repo.AssignRole(ctx, "u1", domain.ADMIN)
		roles, err := repo.GetRolesByUserId(ctx, "u1")
		assert.NoError(t, err)
		assert.Len(t, roles, 1)
		assert.Equal(t, domain.ADMIN, roles[0])

		_ = repo.UpdateRole(ctx, "u1", domain.MECHANIC)
		roles, _ = repo.GetRolesByUserId(ctx, "u1")
		assert.Len(t, roles, 1)
		assert.Equal(t, domain.MECHANIC, roles[0])
	})

	t.Run("Update", func(t *testing.T) {
		err := repo.Update(ctx, "u1", "New Name", "new@e.com")
		assert.NoError(t, err)

		found, _ := repo.GetByEmail(ctx, "new@e.com")
		assert.Equal(t, "New Name", found.Name)

		err = repo.Update(ctx, "non-existent", "Name", "e@e.com")
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		err := repo.Delete(ctx, "u1")
		assert.NoError(t, err)

		found, _ := repo.GetByEmail(ctx, "new@e.com")
		assert.Nil(t, found)
	})
}
