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

func mapResponseDTOFromDomain(so *domain.ServiceOrderHistory) serviceOrderHistoryResponseDTO {
	return serviceOrderHistoryResponseDTO{
		ID:             so.ID,
		PreviousStatus: so.PreviousStatus.String(),
		NewStatus:      so.NewStatus.String(),
		CreatedAt:      so.CreatedAt,
	}
}

type workStatusHistoryEntryDTO struct {
	PreviousStatus string    `json:"previous_status"`
	NewStatus      string    `json:"new_status"`
	CreatedAt      time.Time `json:"created_at"`
}

type workStatusTimelineDTO struct {
	WorkID  string                      `json:"work_id"`
	History []workStatusHistoryEntryDTO `json:"history"`
}

func mapWorkTimelineDTOFromDomain(w *domain.WorkStatusTimeline) workStatusTimelineDTO {
	history := make([]workStatusHistoryEntryDTO, len(w.History))
	for i, entry := range w.History {
		history[i] = workStatusHistoryEntryDTO{
			PreviousStatus: entry.PreviousStatus.String(),
			NewStatus:      entry.NewStatus.String(),
			CreatedAt:      entry.CreatedAt,
		}
	}

	return workStatusTimelineDTO{
		WorkID:  w.WorkID,
		History: history,
	}
}

type paginatorResponseDTO struct {
	Items      []serviceOrderHistoryResponseDTO `json:"items"`
	Works      []workStatusTimelineDTO          `json:"works"`
	TotalItems int64                            `json:"total_items"`
	TotalPages int64                            `json:"total_pages"`
	PageSize   int64                            `json:"page_size"`
	Page       int64                            `json:"page"`
}

func mapResponseDTOFromDomainList(response *domain.ServiceOrderHistoryResponse) paginatorResponseDTO {
	items := make([]serviceOrderHistoryResponseDTO, len(response.Page.Items))
	for i, h := range response.Page.Items {
		history := h
		items[i] = mapResponseDTOFromDomain(&history)
	}

	works := make([]workStatusTimelineDTO, len(response.Works))
	for i, w := range response.Works {
		work := w
		works[i] = mapWorkTimelineDTOFromDomain(&work)
	}

	return paginatorResponseDTO{
		Items:      items,
		Works:      works,
		TotalItems: response.Page.TotalItems,
		TotalPages: response.Page.TotalPages,
		PageSize:   response.Page.PageSize,
		Page:       response.Page.Page,
	}
}
