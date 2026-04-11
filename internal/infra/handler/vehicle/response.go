package vehicle

import (
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type (
	vehicleResponseDTO struct {
		ID           string `json:"id"`
		LicensePlate string `json:"licensePlate"`
		BrandModel   string `json:"brandModel"`
		Year         int    `json:"year"`
		CustomerId   string `json:"customerId"`
	}

	paginatorResponseDTO struct {
		Items      []vehicleResponseDTO `json:"items"`
		TotalItems int64                `json:"totalItems"`
		TotalPages int64                `json:"totalPages"`
		PageSize   int64                `json:"pageSize"`
		Page       int64                `json:"page"`
	}
)

func domainToResponseDto(v *domain.Vehicle) vehicleResponseDTO {
	return vehicleResponseDTO{
		ID:           v.ID,
		BrandModel:   fmt.Sprintf("%s - %s", v.Brand, v.Model),
		Year:         v.Year,
		LicensePlate: v.LicensePlate,
		CustomerId:   v.CustomerId,
	}
}

func domainListToResponseDto(list *domain.PaginatorResponse[domain.Vehicle]) paginatorResponseDTO {
	vehicles := make([]vehicleResponseDTO, len(list.Items))
	for i, item := range list.Items {
		vehicles[i] = domainToResponseDto(&item)
	}

	return paginatorResponseDTO{
		Items:      vehicles,
		TotalItems: list.TotalItems,
		TotalPages: list.TotalPages,
		PageSize:   list.PageSize,
		Page:       list.Page,
	}
}
