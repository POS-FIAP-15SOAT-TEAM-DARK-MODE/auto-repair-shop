package domain

import (
	"time"

	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
)

type ServiceOrderHistoryItem struct {
	PreviousStatus  soDomain.SERVICE_ORDER_STATUS
	NewStatus       soDomain.SERVICE_ORDER_STATUS
	CreatedAt       time.Time
	WorkTransitions []WorkTransitionGroup
}

type WorkTransitionGroup struct {
	WorkID string
	Status []WorkStatusItem
}

type WorkStatusItem struct {
	PreviousStatus soDomain.SERVICE_ORDER_STATUS
	NewStatus      soDomain.SERVICE_ORDER_STATUS
	CreatedAt      time.Time
}

type SearchParams struct {
	ID string
}

// StatusDuration is the average time a service order takes to reach Status
// from whatever status preceded it.
type StatusDuration struct {
	Status       soDomain.SERVICE_ORDER_STATUS
	AverageHours float64
}

func (p *SearchParams) Validate() error {
	if p.ID == "" {
		return ErrServiceOrderIDRequired
	}
	return nil
}
