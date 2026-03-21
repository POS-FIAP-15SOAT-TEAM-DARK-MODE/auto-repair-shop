package domain

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"regexp"
	"strings"
)

const (
	FirstCarYear = 1886
)

var (
	platePattern = regexp.MustCompile(`^[A-Za-z]{3}[0-9][A-Za-z0-9][0-9]{2}$`)
)

//go:generate go run github.com/vektra/mockery/v2@latest --name=VehicleService --with-expecter
//go:generate go run github.com/vektra/mockery/v2@latest --name=VehicleRepository --with-expecter
type (
	Vehicle struct {
		ID           string
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
		ID:           uuid.New().String(),
		LicensePlate: strings.ToUpper(plate),
		Brand:        brand,
		Model:        model,
		Year:         year,
		CustomerId:   customerId,
	}
}

func (v *Vehicle) Validate() error {
	var errs []error

	v.LicensePlate = strings.TrimSpace(strings.ReplaceAll(v.LicensePlate, "-", ""))
	if err := ValidateLicensePlate(v.LicensePlate); err != nil {
		errs = append(errs, err)
	}
	if v.Year < FirstCarYear {
		errs = append(errs, ErrParamVehicleYear)
	}

	return errors.Join(errs...)
}

func ValidateLicensePlate(plate string) error {
	if !platePattern.MatchString(plate) {
		return ErrVehicleInvalidPlate
	}
	return nil
}
