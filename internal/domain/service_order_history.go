package domain

import (
	"context"
	"time"
)

type (
	ServiceOrderHistoryItem struct {
		PreviousStatus  SERVICE_ORDER_STATUS
		NewStatus       SERVICE_ORDER_STATUS
		CreatedAt       time.Time
		WorkTransitions []WorkTransitionGroup
	}

	WorkTransitionGroup struct {
		WorkID string
		Status []WorkStatusItem
	}

	WorkStatusItem struct {
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
		GetHistoryByID(ctx context.Context, params *SearchServiceOrderHistoryParams) ([]ServiceOrderHistoryItem, error)
	}

	ServiceOrderHistoryRepository interface {
		Search(ctx context.Context, params *SearchServiceOrderHistoryParams) ([]ServiceOrderHistoryItem, error)
		SearchWorkTransitionsByServiceOrderID(ctx context.Context, serviceOrderID string) ([]WorkTransitionGroup, error)
	}
)

func (p *SearchServiceOrderHistoryParams) Validate() error {
	if p.ID == "" {
		return ErrServiceOrderIDRequired
	}
	return nil
}

func NewServiceOrderHistory(previousStatus, newStatus SERVICE_ORDER_STATUS, createdAt time.Time) *ServiceOrderHistoryItem {
	return &ServiceOrderHistoryItem{
		PreviousStatus: previousStatus,
		NewStatus:      newStatus,
		CreatedAt:      createdAt,
	}
}
