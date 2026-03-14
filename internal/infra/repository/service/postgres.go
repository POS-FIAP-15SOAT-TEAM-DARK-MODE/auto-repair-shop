package service

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
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

func (r *pg_repo) Count(ctx context.Context, params *domain.SearchServiceParams) (int64, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	queryBuilder := db.QueryBuilder(countQuery)

	if params.Status != "" {
		queryBuilder.Add("status = ", domain.StringToServiceStatus(params.Status).Bool())
	}

	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", countQuery), zap.Any("params", args))
	var total int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		// TODO: add a property error handling
		return 0, err
	}

	return total, nil
}

func (r *pg_repo) Search(ctx context.Context, params *domain.SearchServiceParams) ([]domain.Service, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return nil, err
	}

	queryBuilder := db.QueryBuilder(searchQuery).
		OrderBy("s.created_at", db.ASC).
		AddPagination(params.Limit, params.Offset)

	if params.Status != "" {
		queryBuilder.Add("status = ", domain.StringToServiceStatus(params.Status).Bool())
	}

	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", args))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		// TODO: add a property error handling
		return nil, err
	}

	services := make([]domain.Service, 0, params.Limit)
	for rows.Next() {
		if err = rows.Err(); err != nil {
			// TODO: add a property error handling
			return nil, err
		}

		var service domain.Service
		if err = rows.Scan(
			&service.ID,
			&service.Name,
			&service.Description,
			&service.Price,
			&service.Status,
		); err != nil {
			// TODO: add a property error handling
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}
