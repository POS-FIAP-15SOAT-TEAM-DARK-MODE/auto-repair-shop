package user

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestLoginRequestDTOValidate(t *testing.T) {
	tests := []struct {
		name        string
		dto         loginRequestDTO
		expectedErr error
	}{
		{
			name:        "valid request",
			dto:         loginRequestDTO{Email: "test@example.com", Password: "password"},
			expectedErr: nil,
		},
		{
			name:        "empty email",
			dto:         loginRequestDTO{Email: "", Password: "password"},
			expectedErr: domain.ErrEmptyUserEmail,
		},
		{
			name:        "email with only spaces",
			dto:         loginRequestDTO{Email: "   ", Password: "password"},
			expectedErr: domain.ErrEmptyUserEmail,
		},
		{
			name:        "empty password",
			dto:         loginRequestDTO{Email: "test@example.com", Password: ""},
			expectedErr: domain.ErrEmptyUserPassword,
		},
		{
			name:        "password with only spaces",
			dto:         loginRequestDTO{Email: "test@example.com", Password: "   "},
			expectedErr: domain.ErrEmptyUserPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dto.Validate()
			if tt.expectedErr == nil && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.expectedErr != nil && err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
