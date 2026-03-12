package domain

import "time"

type (
	Vehicle struct {
		Id           string    `json:"id"`
		LicensePlate string    `json:"license_plate"`
		Brand        string    `json:"brand"`
		Model        string    `json:"model"`
		Year         string    `json:"year"`
		CostumerId   string    `json:"costumer_id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
	}

	VehicleService interface {
		Create() (string, error)
		Update() error
		GetOnce() (Vehicle, error)
		GetByOwner() ([]Vehicle, error)
	}
)
