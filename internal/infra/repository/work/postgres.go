package work

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

type pg_repo struct{}

func Repository() domain.WorkRepository {
	return &pg_repo{}
}

func (r *pg_repo) Save(ctx context.Context, w *domain.Work) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", upsertQuery), zap.Any("params", w))
	if _, err = tx.ExecContext(
		ctx,
		upsertQuery,
		w.ID,
		w.Name,
		w.Description,
		w.Price,
		w.Status.Bool(),
	); err != nil {
		return pgPkg.Error(ctx, err)
	}

	return nil
}
