package domain_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestService_Validate(t *testing.T) {
	tests := []struct {
		name        string
		service     *domain.Service
		expectedErr error
	}{
		{
			name: "valid service",
			service: &domain.Service{
				Name:        "Oil Change",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: nil,
		},
		{
			name: "empty name error",
			service: &domain.Service{
				Name:        "",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrEmptyName,
		},
		{
			name: "short name error",
			service: &domain.Service{
				Name:        "ab",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrNameShorterThenRequired,
		},
		{
			name: "empty description error",
			service: &domain.Service{
				Name:        "Oil Change",
				Description: "",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrEmptyDescription,
		},
		{
			name: "short description error",
			service: &domain.Service{
				Name:        "Oil Change",
				Description: "Too short",
				Price:       decimal.NewFromInt(50),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrDescriptionShorterThenRequired,
		},
		{
			name: "zero price error",
			service: &domain.Service{
				Name:        "Oil Change",
				Description: "Complete oil change service",
				Price:       decimal.Zero,
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrPriceLessThenOrEqualZero,
		},
		{
			name: "negative price error",
			service: &domain.Service{
				Name:        "Oil Change",
				Description: "Complete oil change service",
				Price:       decimal.NewFromInt(-10),
				Status:      domain.ACTIVE,
			},
			expectedErr: domain.ErrPriceLessThenOrEqualZero,
		},
		{
			name: "valid service with whitespace trimming",
			service: &domain.Service{
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
			err := tt.service.Validate()
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestServiceStatus_String(t *testing.T) {
	active := domain.ACTIVE
	inactive := domain.INACTIVE

	assert.Equal(t, "ACTIVE", active.String())
	assert.Equal(t, "INACTIVE", inactive.String())
}

func TestServiceStatus_Bool(t *testing.T) {
	active := domain.ACTIVE
	inactive := domain.INACTIVE

	assert.True(t, active.Bool())
	assert.False(t, inactive.Bool())
}

func TestBoolToServiceStatus(t *testing.T) {
	active := domain.BoolToServiceStatus(true)
	inactive := domain.BoolToServiceStatus(false)

	assert.Equal(t, domain.ACTIVE, active)
	assert.Equal(t, domain.INACTIVE, inactive)
}

func TestStringToServiceStatus(t *testing.T) {
	active := domain.StringToServiceStatus("ACTIVE")
	inactive := domain.StringToServiceStatus("INACTIVE")
	defaultInactive := domain.StringToServiceStatus("INVALID")

	assert.Equal(t, domain.ACTIVE, active)
	assert.Equal(t, domain.INACTIVE, inactive)
	assert.Equal(t, domain.INACTIVE, defaultInactive)
}

func TestValidServiceStatusStringValue(t *testing.T) {
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
			expectedErr: domain.ErrInvalidStatusValue,
		},
		{
			name:        "lowercase active status",
			value:       "active",
			expectedErr: domain.ErrInvalidStatusValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := domain.ValidServiceStatusStringValue(tt.value)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}
