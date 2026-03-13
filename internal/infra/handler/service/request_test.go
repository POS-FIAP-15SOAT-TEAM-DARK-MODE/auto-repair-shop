package service

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestCreateServiceReqDTO_Validate_InvalidPriceSpecialForms(t *testing.T) {
	tests := []struct {
		name        string
		price       string
		expectedErr error
	}{
		{
			name:        "minus only",
			price:       "-",
			expectedErr: domain.ErrInvalidPriceValue,
		},
		{
			name:        "dot only",
			price:       ".",
			expectedErr: domain.ErrInvalidPriceValue,
		},
		{
			name:        "minus dot",
			price:       "-.",
			expectedErr: domain.ErrInvalidPriceValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := &createServiceReqDTO{
				Name:        "Service",
				Description: "Valid description",
				Price:       tt.price,
				Status:      domain.ActiveString,
			}

			err := dto.Validate()
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCreateServiceReqDTO_Validate_NormalizesPrice(t *testing.T) {
	dto := &createServiceReqDTO{
		Name:        "Service",
		Description: "Valid description",
		Price:       "1,234.50",
		Status:      domain.ActiveString,
	}

	if err := dto.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := dto.Price, "1234.50"; got != want {
		t.Fatalf("expected normalized price %q, got %q", want, got)
	}
}
