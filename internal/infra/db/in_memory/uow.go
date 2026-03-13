package inmemory

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func NewInMemoryUoW() domain.Executor {
	return &domain.UnitOfWork{
		OnStart: func(ctx context.Context) (context.Context, error) {
			return ctx, nil
		},
		OnSuccess: func(context.Context) error {
			return nil
		},
		OnFailure: func(_ context.Context, cause error) error {
			return cause
		},
	}
}
