package domain_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestValidationError_Error(t *testing.T) {
	msg := "test error message"
	err := domain.ValidationError{Message: msg}
	if err.Error() != msg {
		t.Errorf("expected %q, got %q", msg, err.Error())
	}
}
