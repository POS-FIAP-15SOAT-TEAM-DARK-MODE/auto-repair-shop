package supply_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
	svc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/supply"
)

// runAllSteps executes all UoW steps with the provided context — simulates a real transaction.
func runAllSteps() func(context.Context, ...uow.Step) error {
	return func(ctx context.Context, steps ...uow.Step) error {
		for _, step := range steps {
			if err := step(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

func validSupply() *domain.Supply {
	return &domain.Supply{
		ID:            uuid.New().String(),
		Name:          "Brake Pad",
		Description:   "High-performance brake pads for daily driving",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
		Version:       1,
	}
}

// ─── Create ───────────────────────────────────────────────────────────────────

func TestService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     *domain.Supply
		mockSetup func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository)
		wantErr   error
	}{
		{
			name:  "success",
			input: validSupply(),
			mockSetup: func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository) {
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).Return(nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
		},
		{
			name:  "validation_error",
			input: &domain.Supply{}, // deliberately empty / invalid
			mockSetup: func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository) {
				// neither the UoW nor the repository should be called
			},
			wantErr: domain.ErrInvalidSupplyName,
		},
		{
			name:  "repository_conflict_returns_error",
			input: validSupply(),
			mockSetup: func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository) {
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Supply")).Return(domain.ErrDataConflict)
				executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
			wantErr: domain.ErrDataConflict,
		},
		{
			name:  "transaction_failure_returns_error",
			input: validSupply(),
			mockSetup: func(executor *uowmocks.Executor, _ *domainmocks.SupplyRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := uowmocks.NewExecutor(t)
			repo := domainmocks.NewSupplyRepository(t)

			tt.mockSetup(executor, repo)

			err := svc.Service(executor, repo).Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestService_List(t *testing.T) {
	tests := []struct {
		name      string
		mockSetup func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository)
		wantLen   int
		wantErr   error
	}{
		{
			name: "success_returns_supplies",
			mockSetup: func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository) {
				repo.EXPECT().
					List(mock.Anything).
					Return([]*domain.Supply{validSupply(), validSupply()}, nil)
				executor.EXPECT().
					Execute(mock.Anything, mock.Anything).
					RunAndReturn(runAllSteps())
			},
			wantLen: 2,
		},
		{
			name: "success_returns_empty_list",
			mockSetup: func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository) {
				repo.EXPECT().
					List(mock.Anything).
					Return([]*domain.Supply{}, nil)
				executor.EXPECT().
					Execute(mock.Anything, mock.Anything).
					RunAndReturn(runAllSteps())
			},
			wantLen: 0,
		},
		{
			name: "repository_error_returns_error",
			mockSetup: func(executor *uowmocks.Executor, repo *domainmocks.SupplyRepository) {
				repo.EXPECT().
					List(mock.Anything).
					Return([]*domain.Supply{}, assert.AnError)
				executor.EXPECT().
					Execute(mock.Anything, mock.Anything).
					RunAndReturn(runAllSteps())
			},
			wantErr: assert.AnError,
		},
		{
			name: "transaction_failure_returns_error",
			mockSetup: func(executor *uowmocks.Executor, _ *domainmocks.SupplyRepository) {
				executor.EXPECT().
					Execute(mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := uowmocks.NewExecutor(t)
			repo := domainmocks.NewSupplyRepository(t)

			tt.mockSetup(executor, repo)

			result, err := svc.Service(executor, repo).List(context.Background())

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}
		})
	}
}
