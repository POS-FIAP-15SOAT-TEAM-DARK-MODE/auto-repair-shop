package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestError_Mapping(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			"UniqueViolation",
			&pq.Error{Code: "23505", Message: "duplicate key"},
			domain.ErrDataConflict,
		},
		{
			"CheckViolation",
			&pq.Error{Code: "23514", Message: "check constraint"},
			domain.ErrDataViolation,
		},
		{
			"ForeignKeyViolation",
			&pq.Error{Code: "23503", Message: "fk violation"},
			domain.ErrDataViolation,
		},
		{
			"NotNullViolation",
			&pq.Error{Code: "23502", Message: "not null violation"},
			domain.ErrDataViolation,
		},
		{
			"SerializationFailure",
			&pq.Error{Code: "40001", Message: "serialization failure"},
			domain.ErrInfraConflict,
		},
		{
			"DeadlockDetected",
			&pq.Error{Code: "40P01", Message: "deadlock detected"},
			domain.ErrInfraConflict,
		},
		{
			"NonPQError",
			errors.New("standard error"),
			errors.New("standard error"),
		},
		{
			"UnhandledPQError",
			&pq.Error{Code: "99999", Message: "unknown pq error"},
			&pq.Error{Code: "99999", Message: "unknown pq error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := postgres.Error(ctx, tt.err)
			if tt.name == "NonPQError" || tt.name == "UnhandledPQError" {
				assert.Equal(t, tt.expected.Error(), got.Error())
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
