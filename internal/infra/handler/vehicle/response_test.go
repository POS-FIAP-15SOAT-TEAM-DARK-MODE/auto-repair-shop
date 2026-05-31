package vehicle

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestDomainToResponseDto(t *testing.T) {
	tests := []struct {
		name     string
		input    *domain.Vehicle
		expected vehicleResponseDTO
	}{
		{
			name: "valid vehicle",
			input: &domain.Vehicle{
				ID:           "1",
				LicensePlate: "ABC1D23",
				Brand:        "chevrolet",
				Model:        "onix",
				Year:         2020,
				CustomerId:   "123",
			},
			expected: vehicleResponseDTO{
				ID:           "1",
				LicensePlate: "ABC1D23",
				BrandModel:   "chevrolet - onix",
				Year:         2020,
				CustomerId:   "123",
			},
		},
		{
			name: "empty fields",
			input: &domain.Vehicle{
				ID:           "",
				LicensePlate: "",
				Brand:        "",
				Model:        "",
				Year:         0,
				CustomerId:   "",
			},
			expected: vehicleResponseDTO{
				ID:           "",
				LicensePlate: "",
				BrandModel:   " - ",
				Year:         0,
				CustomerId:   "",
			},
		},
		{
			name: "only brand filled",
			input: &domain.Vehicle{
				ID:           "2",
				LicensePlate: "XYZ1A23",
				Brand:        "fiat",
				Model:        "",
				Year:         2015,
				CustomerId:   "456",
			},
			expected: vehicleResponseDTO{
				ID:           "2",
				LicensePlate: "XYZ1A23",
				BrandModel:   "fiat - ",
				Year:         2015,
				CustomerId:   "456",
			},
		},
		{
			name: "only model filled",
			input: &domain.Vehicle{
				ID:           "3",
				LicensePlate: "AAA1B23",
				Brand:        "",
				Model:        "civic",
				Year:         2018,
				CustomerId:   "789",
			},
			expected: vehicleResponseDTO{
				ID:           "3",
				LicensePlate: "AAA1B23",
				BrandModel:   " - civic",
				Year:         2018,
				CustomerId:   "789",
			},
		},
		{
			name: "special characters in brand and model",
			input: &domain.Vehicle{
				ID:           "4",
				LicensePlate: "BBB1C23",
				Brand:        "bmw",
				Model:        "x5 m-sport",
				Year:         2022,
				CustomerId:   "999",
			},
			expected: vehicleResponseDTO{
				ID:           "4",
				LicensePlate: "BBB1C23",
				BrandModel:   "bmw - x5 m-sport",
				Year:         2022,
				CustomerId:   "999",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domainToResponseDto(tt.input)

			assert.Equal(t, tt.expected.ID, result.ID)
			assert.Equal(t, tt.expected.LicensePlate, result.LicensePlate)
			assert.Equal(t, tt.expected.BrandModel, result.BrandModel)
			assert.Equal(t, tt.expected.Year, result.Year)
			assert.Equal(t, tt.expected.CustomerId, result.CustomerId)
		})
	}
}

func TestDomainListToResponseDto(t *testing.T) {
	tests := []struct {
		name     string
		input    *domain.PaginatorResponse[domain.Vehicle]
		expected paginatorResponseDTO
	}{
		{
			name: "empty list",
			input: &domain.PaginatorResponse[domain.Vehicle]{
				Items:      []domain.Vehicle{},
				TotalItems: 0,
				TotalPages: 0,
				PageSize:   10,
				Page:       1,
			},
			expected: paginatorResponseDTO{
				Items:      []vehicleResponseDTO{},
				TotalItems: 0,
				TotalPages: 0,
				PageSize:   10,
				Page:       1,
			},
		},
		{
			name: "single item",
			input: &domain.PaginatorResponse[domain.Vehicle]{
				Items: []domain.Vehicle{
					{
						ID:           "veh-1",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "cust-1",
					},
				},
				TotalItems: 1,
				TotalPages: 1,
				PageSize:   10,
				Page:       1,
			},
			expected: paginatorResponseDTO{
				Items: []vehicleResponseDTO{
					{
						ID:           "veh-1",
						LicensePlate: "ABC1D23",
						BrandModel:   "chevrolet - onix",
						Year:         2020,
						CustomerId:   "cust-1",
					},
				},
				TotalItems: 1,
				TotalPages: 1,
				PageSize:   10,
				Page:       1,
			},
		},
		{
			name: "multiple items",
			input: &domain.PaginatorResponse[domain.Vehicle]{
				Items: []domain.Vehicle{
					{
						ID:           "veh-1",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "cust-1",
					},
					{
						ID:           "veh-2",
						LicensePlate: "XYZ9Z99",
						Brand:        "fiat",
						Model:        "uno",
						Year:         2010,
						CustomerId:   "cust-2",
					},
				},
				TotalItems: 2,
				TotalPages: 1,
				PageSize:   10,
				Page:       1,
			},
			expected: paginatorResponseDTO{
				Items: []vehicleResponseDTO{
					{
						ID:           "veh-1",
						LicensePlate: "ABC1D23",
						BrandModel:   "chevrolet - onix",
						Year:         2020,
						CustomerId:   "cust-1",
					},
					{
						ID:           "veh-2",
						LicensePlate: "XYZ9Z99",
						BrandModel:   "fiat - uno",
						Year:         2010,
						CustomerId:   "cust-2",
					},
				},
				TotalItems: 2,
				TotalPages: 1,
				PageSize:   10,
				Page:       1,
			},
		},
		{
			name: "preserves pagination metadata",
			input: &domain.PaginatorResponse[domain.Vehicle]{
				Items: []domain.Vehicle{
					{
						ID:           "veh-1",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "cust-1",
					},
				},
				TotalItems: 50,
				TotalPages: 5,
				PageSize:   10,
				Page:       3,
			},
			expected: paginatorResponseDTO{
				Items: []vehicleResponseDTO{
					{
						ID:           "veh-1",
						LicensePlate: "ABC1D23",
						BrandModel:   "chevrolet - onix",
						Year:         2020,
						CustomerId:   "cust-1",
					},
				},
				TotalItems: 50,
				TotalPages: 5,
				PageSize:   10,
				Page:       3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := domainListToResponseDto(tt.input)

			assert.Equal(t, tt.expected, result)
		})
	}
}
