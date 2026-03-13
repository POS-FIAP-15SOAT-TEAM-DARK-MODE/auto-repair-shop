package service

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type createServiceResDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
}

func mapResponseDTOFromDomain(service *domain.Service) createServiceResDTO {
	price, _ := service.Price.Float64()
	return createServiceResDTO{
		ID:          service.ID,
		Name:        service.Name,
		Description: service.Description,
		Price:       price,
		Status:      service.Status.String(),
	}
}
