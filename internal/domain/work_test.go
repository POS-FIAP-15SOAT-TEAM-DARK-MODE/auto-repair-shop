package domain_test

import (
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestWork_Validate(t *testing.T) {
	tests := []struct {
		name        string
		work        *domain.Work
		expectedErr error
	}{
		{
			name: "valid work",
			work: &domain.Work{
				Name:        "Oil Change",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: nil,
		},
		{
			name: "empty name error",
			work: &domain.Work{
				Name:        "",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrEmptyWorkName,
		},
		{
			name: "short name error",
			work: &domain.Work{
				Name:        "ab",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrWorkNameShorterThenRequired,
		},
		{
			name: "empty description error",
			work: &domain.Work{
				Name:        "Oil Change",
				Description: "",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrEmptyWorkDescription,
		},
		{
			name: "short description error",
			work: &domain.Work{
				Name:        "Oil Change",
				Description: "Too short",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrWorkDescriptionShorterThenRequired,
		},
		{
			name: "zero price error",
			work: &domain.Work{
				Name:        "Oil Change",
				Description: "Complete oil change service",
				Price:       decimal.Zero,
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrWorkPriceLessThenOrEqualZero,
		},
		{
			name: "negative price error",
			work: &domain.Work{
				Name:        "Oil Change",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(-10),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrWorkPriceLessThenOrEqualZero,
		},
		{
			name: "valid work with whitespace trimming",
			work: &domain.Work{
				Name:        "  Oil Change  ",
				Description: "   Complete oil change service   ",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.work.Validate()
			if tt.expectedErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.True(t, errors.Is(err, tt.expectedErr), "expected error to wrap %v, got %v", tt.expectedErr, err)
		})
	}
}

func TestWorkStatus_String(t *testing.T) {
	active := domain.ACTIVE
	inactive := domain.INACTIVE

	assert.Equal(t, "ACTIVE", active.String())
	assert.Equal(t, "INACTIVE", inactive.String())
}

func TestWorkStatus_Bool(t *testing.T) {
	active := domain.ACTIVE
	inactive := domain.INACTIVE

	assert.True(t, active.Bool())
	assert.False(t, inactive.Bool())
}

func TestBoolToWorkStatus(t *testing.T) {
	active := domain.BoolToWorkStatus(true)
	inactive := domain.BoolToWorkStatus(false)

	assert.Equal(t, domain.ACTIVE, active)
	assert.Equal(t, domain.INACTIVE, inactive)
}

func TestStringToWorkStatus(t *testing.T) {
	active := domain.StringToWorkStatus("ACTIVE")
	inactive := domain.StringToWorkStatus("INACTIVE")
	defaultInactive := domain.StringToWorkStatus("INVALID")

	assert.Equal(t, domain.ACTIVE, active)
	assert.Equal(t, domain.INACTIVE, inactive)
	assert.Equal(t, domain.INACTIVE, defaultInactive)
}

func TestValidWorkStatusStringValue(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectedErr error
	}{
		{
			name:        "valid active status",
			value:       "ACTIVE",
			expectedErr: nil,
		},
		{
			name:        "valid inactive status",
			value:       "INACTIVE",
			expectedErr: nil,
		},
		{
			name:        "invalid status",
			value:       "INVALID",
			expectedErr: domain.ErrInvalidWorkStatusValue,
		},
		{
			name:        "lowercase active status",
			value:       "active",
			expectedErr: domain.ErrInvalidWorkStatusValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidWorkStatusStringValue(tt.value)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
