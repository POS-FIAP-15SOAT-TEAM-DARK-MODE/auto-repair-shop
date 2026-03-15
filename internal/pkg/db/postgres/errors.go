package postgres

import (
	"context"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// PostgreSQL error codes
// See: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation      pq.ErrorCode = "23505"
	pgCheckViolation       pq.ErrorCode = "23514"
	pgNotNullViolation     pq.ErrorCode = "23502"
	pgForeignKeyViolation  pq.ErrorCode = "23503"
	pgSerializationFailure pq.ErrorCode = "40001"
	pgDeadlockDetected     pq.ErrorCode = "40P01"
	pgExclusionViolation   pq.ErrorCode = "23P01"
	// Add more codes if needed for further business rules.
)

// Error maps PostgreSQL-specific errors to domain errors.
// Any unmapped error is returned as-is so the transactor can still rollback.
// If the database changes (e.g. MySQL), this package must be replaced or adapted.
func Error(ctx context.Context, err error) error {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case pgUniqueViolation:
		logger.Of(ctx).Warn("PostgreSQL unique constraint violation",
			zap.String("code", string(pgErr.Code)),
			zap.String("constraint", pgErr.Constraint),
			zap.String("message", pgErr.Message),
		)
		return domain.ErrDataConflict

	case pgCheckViolation, pgForeignKeyViolation, pgNotNullViolation, pgExclusionViolation:
		logger.Of(ctx).Warn("PostgreSQL data violation",
			zap.String("code", string(pgErr.Code)),
			zap.String("constraint", pgErr.Constraint),
			zap.String("message", pgErr.Message),
		)
		return domain.ErrDataViolation

	case pgSerializationFailure, pgDeadlockDetected:
		logger.Of(ctx).Warn("PostgreSQL serialization/deadlock issue",
			zap.String("code", string(pgErr.Code)),
			zap.String("message", pgErr.Message),
		)
		return domain.ErrInfraConflict

	default:
		logger.Of(ctx).Warn("Unhandled PostgreSQL error",
			zap.String("code", string(pgErr.Code)),
			zap.String("message", pgErr.Message),
		)
		return err
	}
}
