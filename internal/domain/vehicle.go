package domain

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"regexp"
	"strings"
)

const (
	FirstCarYear        = 1886
	NumberOfPlateDigits = 7
)

var (
	PlatePattern = regexp.MustCompile(`^[A-Za-z]{3}[0-9][A-Za-z0-9][0-9]{2}$`)
)

type (
	Vehicle struct {
		Id           string
		LicensePlate string
		Brand        string
		Model        string
		Year         int
		CustomerId   string
	}

	VehicleService interface {
		Create(ctx context.Context, vehicle *Vehicle) error
	}

	VehicleRepository interface {
		Save(ctx context.Context, vehicle *Vehicle) error
	}
)

func NewVehicle(plate, brand, model, customerId string, year int) *Vehicle {
	return &Vehicle{
		Id:           uuid.New().String(),
		LicensePlate: strings.ToUpper(plate),
		Brand:        brand,
		Model:        model,
		Year:         year,
		CustomerId:   customerId,
	}
}

func (v *Vehicle) Validate() error {
	var errs []error

	if strings.TrimSpace(v.Brand) == "" {
		errs = append(errs, ErrRequiredVehicleBrand)
	}
	if strings.TrimSpace(v.Model) == "" {
		errs = append(errs, ErrRequiredVehicleModel)
	}
	if strings.TrimSpace(v.CustomerId) == "" {
		errs = append(errs, ErrNoCustomerAssociated)
	}
	if len(v.LicensePlate) != NumberOfPlateDigits {
		errs = append(errs, ErrInvalidPlate)
	}

	return errors.Join(errs...)
}
