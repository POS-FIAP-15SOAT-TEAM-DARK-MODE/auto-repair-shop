package postgres

import (
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/lib/pq"
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
func Error(err error) error {
	// TODO: add logging
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case pgUniqueViolation:
		// HTTP 409 Conflict
		return domain.ErrDataConflict

	case pgCheckViolation, pgForeignKeyViolation, pgNotNullViolation, pgExclusionViolation:
		// HTTP 422 Unprocessable Entity
		return domain.ErrDataViolation

	case pgSerializationFailure, pgDeadlockDetected:
		// HTTP 409 Conflict (transient, should retry)
		// This error means a serializable transaction couldn't complete due to concurrent updates.
		return domain.ErrInfraConflict

	default:
		return err
	}
}
