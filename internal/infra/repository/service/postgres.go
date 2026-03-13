package service

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

type pg_repo struct{}

func Repository() domain.ServiceRepository {
	return &pg_repo{}
}

func (r *pg_repo) Save(ctx context.Context, svc *domain.Service) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", upsertQuery), zap.Any("params", svc))
	if _, err = tx.ExecContext(
		ctx,
		upsertQuery,
		svc.ID,
		svc.Name,
		svc.Description,
		svc.Price,
		svc.Status.Bool(),
	); err != nil {
		return err
	}

	return nil
}
