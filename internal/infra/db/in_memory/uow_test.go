package inmemory_test

import (
	"context"
	"errors"
	"testing"

	inmemory "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/in_memory"
	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryUoW(t *testing.T) {
	uow := inmemory.NewInMemoryUoW()
	assert.NotNil(t, uow)

	ctx := context.Background()

	// Test success
	err := uow.Execute(ctx, func(ctx context.Context) error {
		return nil
	})
	assert.NoError(t, err)

	// Test failure
	expectedErr := errors.New("boom")
	err = uow.Execute(ctx, func(ctx context.Context) error {
		return expectedErr
	})
	assert.ErrorIs(t, err, expectedErr)
}
