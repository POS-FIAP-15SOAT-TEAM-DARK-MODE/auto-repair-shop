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
	stmt, err := tx.PrepareContext(ctx, upsertQuery)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := stmt.Close(); cerr != nil {
			logger.Of(ctx).Warn("failed to close statement", zap.Error(cerr))
		}
	}()

	if _, err = stmt.ExecContext(
		ctx,
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
