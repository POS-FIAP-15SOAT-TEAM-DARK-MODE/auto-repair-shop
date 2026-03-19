package domain

import (
	"context"
	"github.com/google/uuid"
)

type (
	Vehicle struct {
		Id           string
		LicensePlate string
		Brand        string
		Model        string
		Year         string
		CostumerId   string
	}

	VehicleService interface {
		Create(ctx context.Context) error
	}

	VehicleRepository interface {
		Create(ctx context.Context, vehicle *Vehicle) error
	}
)

func NewVehicle(plate, brand, model, year, costumerId string) *Vehicle {
	return &Vehicle{
		Id:           uuid.New().String(),
		LicensePlate: plate,
		Brand:        brand,
		Model:        model,
		Year:         year,
		CostumerId:   costumerId,
	}
}
