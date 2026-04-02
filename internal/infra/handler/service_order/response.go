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
