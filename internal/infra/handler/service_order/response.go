package service_order

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type serviceOrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func mapResponseDTOFromDomain(so domain.ServiceOrder) serviceOrderResponse {
	return serviceOrderResponse{
		ID:     so.ID,
		Status: so.Status.String(),
	}
}

type workItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      string `json:"status"`
}

type supplyItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Version     int    `json:"version"`
	Amount      int    `json:"amount"`
}

type listResponse[T any] struct {
	Items      []T   `json:"items"`
	TotalItems int64 `json:"totalItems"`
	TotalPages int64 `json:"totalPages"`
	PageSize   int64 `json:"pageSize"`
	Page       int64 `json:"page"`
}

func mapWorksListToResponse(works []domain.Work) listResponse[workItemResponse] {
	items := make([]workItemResponse, 0, len(works))
	for _, w := range works {
		items = append(items, workItemResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			Price:       w.Price.String(),
			Status:      w.Status.String(),
		})
	}
	n := int64(len(items))
	return listResponse[workItemResponse]{
		Items:      items,
		TotalItems: n,
		TotalPages: 1,
		PageSize:   n,
		Page:       1,
	}
}

type averageExecutionTimeResponse struct {
	AverageExecutionTimeHours float64 `json:"average_execution_time_hours"`
}

func mapSuppliesListToResponse(supplies []domain.Supply) listResponse[supplyItemResponse] {
	items := make([]supplyItemResponse, 0, len(supplies))
	for _, w := range supplies {
		items = append(items, supplyItemResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			Price:       w.UnitPrice.String(),
			Version:     w.Version,
			Amount:      w.StockQuantity,
		})
	}
	n := int64(len(items))
	return listResponse[supplyItemResponse]{
		Items:      items,
		TotalItems: n,
		TotalPages: 1,
		PageSize:   n,
		Page:       1,
	}
}
