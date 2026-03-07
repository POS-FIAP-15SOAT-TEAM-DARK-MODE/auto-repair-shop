package ping_test

import (
	"testing"

	ping "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/ping"
)

func TestPingUseCaseExecute(t *testing.T) {
	service := ping.Service()
	result := service.Execute()

	if result.Message != "pong" {
		t.Errorf("expected 'pong', got '%s'", result.Message)
	}
}
