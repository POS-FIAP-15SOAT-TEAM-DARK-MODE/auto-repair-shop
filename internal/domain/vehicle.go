package domain

import "context"

type (
	Vehicle struct {
		Id           string `json:"id"`
		LicensePlate string `json:"license_plate"`
		Brand        string `json:"brand"`
		Model        string `json:"model"`
		Year         string `json:"year"`
		CostumerId   string `json:"costumer_id"`
	}

	VehicleService interface {
		Create(ctx context.Context) error
	}

	VehicleRepository interface {
		Create(ctx context.Context, vehicle *Vehicle) error
	}
)
