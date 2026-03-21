package vehicle

import (
	"fmt"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type vehicleResponseDTO struct {
	ID           string `json:"id"`
	LicensePlate string `json:"license_plate"`
	BrandModel   string `json:"brand_model"`
	Year         int    `json:"year"`
	CustomerId   string `json:"customer_id"`
}

func domainToResponseDto(v *domain.Vehicle) vehicleResponseDTO {
	return vehicleResponseDTO{
		ID:           v.ID,
		BrandModel:   fmt.Sprintf("%s - %s", v.Brand, v.Model),
		Year:         v.Year,
		LicensePlate: v.LicensePlate,
		CustomerId:   v.CustomerId,
	}
}
