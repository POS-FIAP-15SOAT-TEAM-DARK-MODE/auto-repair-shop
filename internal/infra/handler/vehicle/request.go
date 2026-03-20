package vehicle

import (
	"errors"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/gin-gonic/gin"
	"strings"
)

type vehicleRequestDto struct {
	LicensePlate string `json:"license_plate"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	Year         int    `json:"year"`
	CustomerId   string `json:"customer_id"`
}

func (v *vehicleRequestDto) Domain() *domain.Vehicle {
	return domain.NewVehicle(v.LicensePlate, v.Brand, v.Model, v.CustomerId, v.Year)
}

func validateAndCreateDTO(c *gin.Context) (*vehicleRequestDto, error) {
	reqDTO, err := reqBodyVehicleToDTO(c)
	if err != nil {
		return nil, err
	}

	if err = reqDTO.validate(); err != nil {
		return nil, err
	}

	return reqDTO, nil
}

func reqBodyVehicleToDTO(c *gin.Context) (*vehicleRequestDto, error) {
	var reqVehicle vehicleRequestDto
	if err := c.ShouldBindJSON(&reqVehicle); err != nil {
		return &reqVehicle, json.CheckJsonError(err)
	}
	return &reqVehicle, nil
}

func (v *vehicleRequestDto) validate() error {
	var errs []error

	if strings.TrimSpace(v.Brand) == "" {
		errs = append(errs, domain.ErrRequiredVehicleBrand)
	}
	if strings.TrimSpace(v.Model) == "" {
		errs = append(errs, domain.ErrRequiredVehicleModel)
	}
	if strings.TrimSpace(v.CustomerId) == "" {
		errs = append(errs, domain.ErrVehicleNoCustomerAssociated)
	}

	return errors.Join(errs...)
}
