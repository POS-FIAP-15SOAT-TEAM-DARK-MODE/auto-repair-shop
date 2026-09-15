package interfaces

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
)

// ServiceOrderStatusDurationReader is consumed by the Prometheus metrics
// collector, kept separate from ServiceOrderHistoryRepository so the
// business-facing interface (and its generated mocks) are untouched.
type ServiceOrderStatusDurationReader interface {
	AverageDurationByStatusInHours(ctx context.Context) ([]domain.StatusDuration, error)
}
