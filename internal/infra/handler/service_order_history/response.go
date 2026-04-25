package service_order_history

import (
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type workStatusEntryDTO struct {
	PreviousStatus string    `json:"previousStatus"`
	NewStatus      string    `json:"newStatus"`
	CreatedAt      time.Time `json:"createdAt"`
}

type workTransitionGroupDTO struct {
	WorkID string               `json:"workId"`
	Status []workStatusEntryDTO `json:"status"`
}

type serviceOrderHistoryResponseDTO struct {
	PreviousStatus  string                   `json:"previousStatus"`
	NewStatus       string                   `json:"newStatus"`
	CreatedAt       time.Time                `json:"createdAt"`
	WorkTransitions []workTransitionGroupDTO `json:"workTransitions,omitempty"`
}

func mapResponseDTOFromDomain(so *domain.ServiceOrderHistoryItem) serviceOrderHistoryResponseDTO {
	dto := serviceOrderHistoryResponseDTO{
		PreviousStatus: so.PreviousStatus.String(),
		NewStatus:      so.NewStatus.String(),
		CreatedAt:      so.CreatedAt,
	}

	if len(so.WorkTransitions) > 0 {
		groups := make([]workTransitionGroupDTO, len(so.WorkTransitions))
		for i, g := range so.WorkTransitions {
			statuses := make([]workStatusEntryDTO, len(g.Status))
			for j, s := range g.Status {
				statuses[j] = workStatusEntryDTO{
					PreviousStatus: s.PreviousStatus.String(),
					NewStatus:      s.NewStatus.String(),
					CreatedAt:      s.CreatedAt,
				}
			}
			groups[i] = workTransitionGroupDTO{
				WorkID: g.WorkID,
				Status: statuses,
			}
		}
		dto.WorkTransitions = groups
	}

	return dto
}

func mapResponseDTOFromDomainList(soHistories []domain.ServiceOrderHistoryItem) []serviceOrderHistoryResponseDTO {
	items := make([]serviceOrderHistoryResponseDTO, len(soHistories))
	for i, h := range soHistories {
		history := h
		items[i] = mapResponseDTOFromDomain(&history)
	}

	return items
}
