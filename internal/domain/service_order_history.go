package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type (
	ServiceOrderHistory struct {
	ID             string
	ServiceOrderID string
	PreviousStatus SERVICE_ORDER_STATUS
	NewStatus      SERVICE_ORDER_STATUS
	CreatedAt      time.Time
}

	ListServiceOrderHistoryByIDParams struct {
		ID       string
		Page     int64
		PageSize int64
	}

	FindServiceOrderHistoryParams struct {
		ID       string
		Page     int64
		PageSize int64
	}
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryRepository --with-expecter
type (
	ServiceOrderHistoryService interface {
		GetHistoryByID(ctx context.Context, id string) ([]ServiceOrderHistory, error)
	}

	ServiceOrderHistoryRepository interface {
		Find(ctx context.Context, params FindServiceOrderHistoryParams) ([]ServiceOrderHistory, error)
	}
)

func (p *ListServiceOrderHistoryByIDParams) Validate() error {
	if p.ID == "" {
		return ErrServiceOrderIDRequired
	}
	return nil
}

func NewServiceOrderHistory(previousStatus, newStatus SERVICE_ORDER_STATUS, createdAt time.Time) *ServiceOrderHistory {
	return &ServiceOrderHistory{
		ID:             uuid.New().String(),
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		CreatedAt:      createdAt,
	}
}
