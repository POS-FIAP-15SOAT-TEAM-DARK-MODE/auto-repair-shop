package domain

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateLicensePlate(t *testing.T) {
	tests := []struct {
		name        string
		plate       string
		expectedErr error
	}{
		{
			name:        "valid mercosul plate",
			plate:       "ABC1D23",
			expectedErr: nil,
		},
		{
			name:        "invalid plate format",
			plate:       "AB12345",
			expectedErr: ErrVehicleInvalidPlate,
		},
		{
			name:        "invalid plate with special chars",
			plate:       "ABC-1234",
			expectedErr: ErrVehicleInvalidPlate,
		},
		{
			name:        "empty plate",
			plate:       "",
			expectedErr: ErrVehicleInvalidPlate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLicensePlate(tt.plate)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.True(t, errors.Is(err, tt.expectedErr),
				"expected error to wrap %v, got %v", tt.expectedErr, err)
		})
	}
}

func TestVehicle_Validate(t *testing.T) {
	tests := []struct {
		name        string
		vehicle     *Vehicle
		expectedErr error
	}{
		{
			name: "valid vehicle",
			vehicle: &Vehicle{
				LicensePlate: "ABC1D23",
				Year:         2020,
			},
			expectedErr: nil,
		},
		{
			name: "invalid plate",
			vehicle: &Vehicle{
				LicensePlate: "INVALID",
				Year:         2020,
			},
			expectedErr: ErrVehicleInvalidPlate,
		},
		{
			name: "invalid year",
			vehicle: &Vehicle{
				LicensePlate: "ABC1D23",
				Year:         1800,
			},
			expectedErr: ErrParamVehicleYear,
		},
		{
			name: "invalid plate and year (joined error)",
			vehicle: &Vehicle{
				LicensePlate: "XXX",
				Year:         1000,
			},
			expectedErr: ErrVehicleInvalidPlate,
		},
		{
			name: "plate normalized with spaces and hyphen",
			vehicle: &Vehicle{
				LicensePlate: " abc-1d23 ",
				Year:         2020,
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vehicle.Validate()

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				return
			}

			assert.True(t, errors.Is(err, tt.expectedErr),
				"expected error to wrap %v, got %v", tt.expectedErr, err)
		})
	}
}

func TestNewVehicle(t *testing.T) {
	tests := []struct {
		name          string
		plate         string
		brand         string
		model         string
		customerId    string
		year          int
		expectedPlate string
	}{
		{
			name:          "should uppercase plate",
			plate:         "abc1d23",
			brand:         "chevrolet",
			model:         "onix",
			customerId:    "123",
			year:          2020,
			expectedPlate: "ABC1D23",
		},
		{
			name:          "already uppercase plate",
			plate:         "ABC1D23",
			brand:         "fiat",
			model:         "uno",
			customerId:    "456",
			year:          2010,
			expectedPlate: "ABC1D23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewVehicle(tt.plate, tt.brand, tt.model, tt.customerId, tt.year)

			assert.NotEmpty(t, v.ID)
			assert.Equal(t, tt.expectedPlate, v.LicensePlate)
			assert.Equal(t, tt.brand, v.Brand)
			assert.Equal(t, tt.model, v.Model)
			assert.Equal(t, tt.year, v.Year)
			assert.Equal(t, tt.customerId, v.CustomerId)
		})
	}
}
