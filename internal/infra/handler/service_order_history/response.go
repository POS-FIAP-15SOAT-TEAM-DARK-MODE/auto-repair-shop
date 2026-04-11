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

type paginatorResponseDTO struct {
	Items      []serviceOrderHistoryResponseDTO `json:"items"`
	TotalItems int64                            `json:"total_items"`
	TotalPages int64                            `json:"total_pages"`
	PageSize   int64                            `json:"page_size"`
	Page       int64                            `json:"page"`
}

func mapResponseDTOFromDomainList(soHistories *domain.PaginatorResponse[domain.ServiceOrderHistory]) paginatorResponseDTO {
	items := make([]serviceOrderHistoryResponseDTO, len(soHistories.Items))
	for i, h := range soHistories.Items {
		history := h
		items[i] = mapResponseDTOFromDomain(&history)
	}

	return paginatorResponseDTO{
		Items:      items,
		TotalItems: soHistories.TotalItems,
		TotalPages: soHistories.TotalPages,
		PageSize:   soHistories.PageSize,
		Page:       soHistories.Page,
	}
}
