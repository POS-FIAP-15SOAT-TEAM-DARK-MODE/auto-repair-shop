package db

import (
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/lib/pq"
)

// PostgreSQL error codes
// https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation     pq.ErrorCode = "23505"
	pgCheckViolation      pq.ErrorCode = "23514"
	pgNotNullViolation    pq.ErrorCode = "23502"
	pgForeignKeyViolation pq.ErrorCode = "23503"
)

// constraintMessages maps PostgreSQL unique constraint names to human-readable conflict messages.
var constraintMessages = map[string]string{
	"user_email_key":    "email already in use",
	"customer_cpf_key":  "CPF already registered",
	"customer_cnpj_key": "CNPJ already registered",
}

// Error maps PostgreSQL errors to domain errors.
// Any unmapped error is returned as-is so the transactor can still rollback.
func Error(err error) error {
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		return err
	}

	switch pgErr.Code {
	case pgUniqueViolation:
		if msg, found := constraintMessages[pgErr.Constraint]; found {
			return domain.ConflictError{Message: msg}
		}
		return domain.ConflictError{Message: "resource already exists"}

	case pgCheckViolation:
		return domain.UnprocessableEntityError{Message: "data violates business constraints"}

	case pgNotNullViolation:
		return domain.UnprocessableEntityError{Message: "required field is missing"}

	case pgForeignKeyViolation:
		return domain.UnprocessableEntityError{Message: "referenced resource does not exist"}

	default:
		return err
	}
}
