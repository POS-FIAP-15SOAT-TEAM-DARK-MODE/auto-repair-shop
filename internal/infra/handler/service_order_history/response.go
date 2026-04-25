package service_order_history

import (
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type workStatusEntryDTO struct {
	ID             string    `json:"id"`
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	CreatedAt      time.Time `json:"created_at"`
}

type workTransitionGroupDTO struct {
	WorkID string               `json:"work_id"`
	Status []workStatusEntryDTO `json:"status"`
}

type serviceOrderHistoryResponseDTO struct {
	ID              string                   `json:"id"`
	PreviousStatus  string                   `json:"previous_status"`
	NewStatus       string                   `json:"new_status"`
	CreatedAt       time.Time                `json:"created_at"`
	WorkTransitions []workTransitionGroupDTO `json:"work_transitions,omitempty"`
}

type listResponseDTO struct {
	Items []serviceOrderHistoryResponseDTO `json:"items"`
}

func mapResponseDTOFromDomain(so *domain.ServiceOrderHistory) serviceOrderHistoryResponseDTO {
	dto := serviceOrderHistoryResponseDTO{
		ID:             so.ID,
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
					ID:             s.ID,
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

func mapResponseDTOFromDomainList(soHistories []domain.ServiceOrderHistory) listResponseDTO {
	items := make([]serviceOrderHistoryResponseDTO, len(soHistories))
	for i, h := range soHistories {
		history := h
		items[i] = mapResponseDTOFromDomain(&history)
	}

	return listResponseDTO{Items: items}
}
