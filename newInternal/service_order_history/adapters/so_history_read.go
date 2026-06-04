package adapters

import (
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/domain"
)

type WorkStatusEntryResponse struct {
	PreviousStatus string    `json:"previousStatus"`
	NewStatus      string    `json:"newStatus"`
	CreatedAt      time.Time `json:"createdAt"`
}

type WorkTransitionGroupResponse struct {
	WorkID string                    `json:"workId"`
	Status []WorkStatusEntryResponse `json:"status"`
}

type SOHistoryResponse struct {
	PreviousStatus  string                        `json:"previousStatus"`
	NewStatus       string                        `json:"newStatus"`
	CreatedAt       time.Time                     `json:"createdAt"`
	WorkTransitions []WorkTransitionGroupResponse `json:"workTransitions,omitempty"`
}

func MapToSOHistoryResponse(item *domain.ServiceOrderHistoryItem) SOHistoryResponse {
	resp := SOHistoryResponse{
		PreviousStatus: item.PreviousStatus.String(),
		NewStatus:      item.NewStatus.String(),
		CreatedAt:      item.CreatedAt,
	}

	if len(item.WorkTransitions) > 0 {
		groups := make([]WorkTransitionGroupResponse, len(item.WorkTransitions))
		for i, g := range item.WorkTransitions {
			statuses := make([]WorkStatusEntryResponse, len(g.Status))
			for j, s := range g.Status {
				statuses[j] = WorkStatusEntryResponse{
					PreviousStatus: s.PreviousStatus.String(),
					NewStatus:      s.NewStatus.String(),
					CreatedAt:      s.CreatedAt,
				}
			}
			groups[i] = WorkTransitionGroupResponse{
				WorkID: g.WorkID,
				Status: statuses,
			}
		}
		resp.WorkTransitions = groups
	}

	return resp
}

func MapToSOHistoryResponseList(items []domain.ServiceOrderHistoryItem) []SOHistoryResponse {
	result := make([]SOHistoryResponse, len(items))
	for i, item := range items {
		it := item
		result[i] = MapToSOHistoryResponse(&it)
	}
	return result
}
