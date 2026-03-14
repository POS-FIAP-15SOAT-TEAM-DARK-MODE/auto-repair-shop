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

func TestNewService(t *testing.T) {
	t.Run("valid inputs create service with uuid", func(t *testing.T) {
		svc, err := domain.NewService("Oil Change", "Complete oil change", "49.99", domain.ACTIVE)
		assert.NoError(t, err)
		assert.NotNil(t, svc)
		assert.NotEmpty(t, svc.ID)
		assert.Equal(t, "Oil Change", svc.Name)
		assert.Equal(t, "Complete oil change", svc.Description)
		assert.Equal(t, domain.ACTIVE, svc.Status)
		assert.True(t, svc.Price.IsPositive())
	})

	t.Run("invalid price returns error", func(t *testing.T) {
		svc, err := domain.NewService("Oil Change", "Complete oil change", "not-a-number", domain.ACTIVE)
		assert.Error(t, err)
		assert.Nil(t, svc)
	})
}

func TestListServiceParams_SearchServiceParams(t *testing.T) {
	t.Run("maps limit offset and status correctly", func(t *testing.T) {
		params := domain.ListServiceParams{Page: 3, PageSize: 20, Status: "ACTIVE"}
		sp := params.SearchServiceParams()
		assert.Equal(t, int64(20), sp.Limit)
		assert.Equal(t, int64(40), sp.Offset)
		assert.Equal(t, "ACTIVE", sp.Status)
	})

	t.Run("page 1 produces zero offset", func(t *testing.T) {
		params := domain.ListServiceParams{Page: 1, PageSize: 10, Status: ""}
		sp := params.SearchServiceParams()
		assert.Equal(t, int64(10), sp.Limit)
		assert.Equal(t, int64(0), sp.Offset)
		assert.Equal(t, "", sp.Status)
	})
}
