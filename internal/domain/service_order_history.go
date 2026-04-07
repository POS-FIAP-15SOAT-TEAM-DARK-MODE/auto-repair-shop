package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ServiceOrderHistory struct {
	ID             string               `json:"id"`
	ServiceOrderID string               `json:"service_order_id"`
	PreviousStatus SERVICE_ORDER_STATUS `json:"previous_status"`
	NewStatus      SERVICE_ORDER_STATUS `json:"new_status"`
	CreatedAt      time.Time            `json:"created_at"`
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryRepository --with-expecter
type (
	ServiceOrderHistoryService interface {
		GetHistoryByID(ctx context.Context, id string) ([]ServiceOrderHistory, error)
	}

	ServiceOrderHistoryRepository interface {
		Find(ctx context.Context, params FindServiceOrderHistoryParams) ([]ServiceOrderHistory, error)
	}

	FindServiceOrderHistoryParams struct {
		ID string
	}
)

func NewServiceOrderHistory(previousStatus, newStatus SERVICE_ORDER_STATUS, createdAt time.Time) *ServiceOrderHistory {
	return &ServiceOrderHistory{
		ID:             uuid.New().String(),
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		CreatedAt:      createdAt,
	}
}
