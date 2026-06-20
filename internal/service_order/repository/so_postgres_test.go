package repository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	uowPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/repository"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestSOPostgres_UpdatePricing(t *testing.T) {
	updatePattern := `UPDATE service_order\s+SET total_amount = \$2, updated_at = NOW\(\)\s+WHERE id = \$1`

	total := decimal.NewFromInt(250)

	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		useTx     bool
		wantErr   error
	}{
		{
			name: "update success",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(updatePattern).
					WithArgs("so-1", total).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			useTx: true,
		},
		{
			name: "no rows affected returns not found",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(updatePattern).
					WithArgs("so-1", total).
					WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectRollback()
			},
			useTx:   true,
			wantErr: domain.ErrServiceOrderNotFound,
		},
		{
			name: "exec error",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(updatePattern).
					WithArgs("so-1", total).
					WillReturnError(errors.New("exec boom"))
				mock.ExpectRollback()
			},
			useTx:   true,
			wantErr: errors.New("exec boom"),
		},
		{
			name:    "missing transaction",
			useTx:   false,
			wantErr: uowPkg.ErrMissingPostgresTransaction,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewSOPostgres()

			if !tt.useTx {
				err := repo.UpdatePricing(context.Background(), "so-1", total)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			db, mock, err := sqlmock.New()
			assert.NoError(t, err)
			defer func() { _ = db.Close() }()

			uow := uowPkg.NewTransactionalUoW(db)
			if tt.setupMock != nil {
				tt.setupMock(mock)
			}

			err = uow.Execute(context.Background(), func(ctx context.Context) error {
				return repo.UpdatePricing(ctx, "so-1", total)
			})

			if tt.wantErr != nil {
				assert.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrServiceOrderNotFound) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
