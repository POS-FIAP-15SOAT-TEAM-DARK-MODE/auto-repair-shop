package work

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db"
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

func (r *pg_repo) Count(ctx context.Context, params *domain.SearchWorkParams) (int64, error) {
	tx, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return 0, err
	}

	queryBuilder := db.QueryBuilder(countQuery)

	if params.Status != "" {
		queryBuilder.Add("status = ", domain.StringToWorkStatus(params.Status).Bool())
	}

	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", countQuery), zap.Any("params", args))
	var total int64
	if err = tx.QueryRowContext(ctx, query, args...).Scan(&total); err != nil {
		return 0, pgPkg.Error(ctx, err)
	}

	return total, nil
}

func (r *pg_repo) Search(ctx context.Context, params *domain.SearchWorkParams) ([]domain.Work, error) {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return nil, err
	}

	queryBuilder := db.QueryBuilder(searchQuery).
		OrderBy("s.created_at", db.ASC).
		AddPagination(params.Limit, params.Offset)

	if params.Status != "" {
		queryBuilder.Add("status = ", domain.StringToWorkStatus(params.Status).Bool())
	}

	query, args := queryBuilder.Build()

	logger.Of(ctx).Debug("Executing query", zap.String("query", query), zap.Any("params", args))
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	works := make([]domain.Work, 0, params.Limit)
	for rows.Next() {
		if err = rows.Err(); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		var work domain.Work
		if err = rows.Scan(
			&work.ID,
			&work.Name,
			&work.Description,
			&work.Price,
			&work.Status,
		); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}

		works = append(works, work)
	}

	return works, nil
}
