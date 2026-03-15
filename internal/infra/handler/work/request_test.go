package work

import (
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestCreateWorkReqDTO_Validate_InvalidPriceSpecialForms(t *testing.T) {
	tests := []struct {
		name        string
		price       string
		expectedErr error
	}{
		{
			name:        "minus only",
			price:       "-",
			expectedErr: domain.ErrInvalidWorkPriceValue,
		},
		{
			name:        "dot only",
			price:       ".",
			expectedErr: domain.ErrInvalidWorkPriceValue,
		},
		{
			name:        "minus dot",
			price:       "-.",
			expectedErr: domain.ErrInvalidWorkPriceValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := &createWorkReqDTO{
				Name:        "Service",
				Description: "Valid description",
				Price:       tt.price,
				Status:      domain.ActiveString,
			}

			err := dto.Validate()
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error to wrap %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCreateWorkReqDTO_Validate_NormalizesPrice(t *testing.T) {
	dto := &createWorkReqDTO{
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
