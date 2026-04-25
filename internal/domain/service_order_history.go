package domain

import (
	"context"
	"time"

	"github.com/oklog/ulid/v2"
)

type (
	ServiceOrderHistory struct {
		ID              string
		ServiceOrderID  string
		PreviousStatus  SERVICE_ORDER_STATUS
		NewStatus       SERVICE_ORDER_STATUS
		CreatedAt       time.Time
		WorkTransitions []WorkTransitionGroup
	}

	WorkTransitionGroup struct {
		WorkID string
		Status []WorkServiceOrderHistory
	}

	WorkServiceOrderHistory struct {
		ID             string
		WorkID         string
		ServiceOrderID string
		PreviousStatus SERVICE_ORDER_STATUS
		NewStatus      SERVICE_ORDER_STATUS
		CreatedAt      time.Time
	}

	SearchServiceOrderHistoryParams struct {
		ID string
	}
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=ServiceOrderHistoryRepository --with-expecter
type (
	ServiceOrderHistoryService interface {
		GetHistoryByID(ctx context.Context, params *SearchServiceOrderHistoryParams) ([]ServiceOrderHistory, error)
	}

	ServiceOrderHistoryRepository interface {
		Search(ctx context.Context, params *SearchServiceOrderHistoryParams) ([]ServiceOrderHistory, error)
		SearchWorkTransitions(ctx context.Context, serviceOrderID string) ([]WorkServiceOrderHistory, error)
	}
)

func (p *SearchServiceOrderHistoryParams) Validate() error {
	if p.ID == "" {
		return ErrServiceOrderIDRequired
	}
	return nil
}

func NewServiceOrderHistory(previousStatus, newStatus SERVICE_ORDER_STATUS, createdAt time.Time) *ServiceOrderHistory {
	return &ServiceOrderHistory{
		ID:             ulid.Make().String(),
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		CreatedAt:      createdAt,
	}
}
