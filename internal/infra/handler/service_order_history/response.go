package service_order_history

import (
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type serviceOrderHistoryResponseDTO struct {
	ID             string    `json:"id"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	CreatedAt      time.Time `json:"created_at"`
}

func mapResponseDTOFromDomainList(so []domain.ServiceOrderHistory) []serviceOrderHistoryResponseDTO {
	soHistory := make([]serviceOrderHistoryResponseDTO, len(so))
	for i, s := range so {
		soHistory[i] = mapResponseDTOFromDomain(s)
	}
	return soHistory
}

func mapResponseDTOFromDomain(so domain.ServiceOrderHistory) serviceOrderHistoryResponseDTO {
	return serviceOrderHistoryResponseDTO{
		ID:             so.ID,
		PreviousStatus: so.PreviousStatus.String(),
		NewStatus:      so.NewStatus.String(),
		CreatedAt:      so.CreatedAt,
	}
}
