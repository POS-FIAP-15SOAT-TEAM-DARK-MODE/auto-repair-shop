package vehicle

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestVehicleRequestDto_Domain(t *testing.T) {
	tests := []struct {
		name  string
		input vehicleRequestDto
	}{
		{
			name: "valid mapping to domain",
			input: vehicleRequestDto{
				LicensePlate: "abc1d23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Domain()

			assert.NotEmpty(t, result.ID)
			assert.Equal(t, "ABC1D23", result.LicensePlate)
			assert.Equal(t, tt.input.Brand, result.Brand)
			assert.Equal(t, tt.input.Model, result.Model)
			assert.Equal(t, tt.input.Year, result.Year)
			assert.Equal(t, tt.input.CustomerId, result.CustomerId)
		})
	}
}

func TestVehicleRequestDto_Validate(t *testing.T) {
	tests := []struct {
		name        string
		input       vehicleRequestDto
		expectedErr error
	}{
		{
			name: "valid dto",
			input: vehicleRequestDto{
				Brand:      "chevrolet",
				Model:      "onix",
				CustomerId: "123",
			},
			expectedErr: nil,
		},
		{
			name: "empty brand",
			input: vehicleRequestDto{
				Brand:      "   ",
				Model:      "onix",
				CustomerId: "123",
			},
			expectedErr: domain.ErrRequiredVehicleBrand,
		},
		{
			name: "empty model",
			input: vehicleRequestDto{
				Brand:      "chevrolet",
				Model:      " ",
				CustomerId: "123",
			},
			expectedErr: domain.ErrRequiredVehicleModel,
		},
		{
			name: "empty customerId",
			input: vehicleRequestDto{
				Brand:      "chevrolet",
				Model:      "onix",
				CustomerId: " ",
			},
			expectedErr: domain.ErrVehicleNoCustomerAssociated,
		},
		{
			name: "multiple errors (joined)",
			input: vehicleRequestDto{
				Brand:      " ",
				Model:      "",
				CustomerId: "",
			},
			expectedErr: domain.ErrRequiredVehicleBrand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.validate()

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.True(t, errors.Is(err, tt.expectedErr),
				"expected error to wrap %v, got %v", tt.expectedErr, err)
		})
	}
}

func TestReqBodyVehicleToDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		body        string
		expectError bool
	}{
		{
			name: "valid json",
			body: `{
				"license_plate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customer_id": "123"
			}`,
			expectError: false,
		},
		{
			name:        "invalid json",
			body:        `{"license_plate":`,
			expectError: true,
		},
		{
			name:        "empty body",
			body:        ``,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			dto, err := reqBodyVehicleToDTO(c)

			if tt.expectError {
				assert.Error(t, err)
				assert.NotNil(t, dto)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, dto)
		})
	}
}

func TestValidateAndCreateDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		body        string
		expectedErr error
	}{
		{
			name: "valid request",
			body: `{
				"license_plate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customer_id": "123"
			}`,
			expectedErr: nil,
		},
		{
			name:        "invalid json",
			body:        `{"license_plate":`,
			expectedErr: errors.New("json error"),
		},
		{
			name: "validation error",
			body: `{
				"license_plate": "ABC1D23",
				"brand": "",
				"model": "",
				"year": 2020,
				"customer_id": ""
			}`,
			expectedErr: domain.ErrRequiredVehicleBrand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			dto, err := validateAndCreateDTO(c)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.NotNil(t, dto)
				return
			}

			assert.Error(t, err)

			if errors.Is(tt.expectedErr, domain.ErrRequiredVehicleBrand) {
				assert.True(t, errors.Is(err, tt.expectedErr))
			}
		})
	}
}
