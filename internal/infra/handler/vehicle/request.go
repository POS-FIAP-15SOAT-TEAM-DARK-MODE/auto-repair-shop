package vehicle

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/gin-gonic/gin"
	"regexp"
	"strings"
)

var (
	platePattern = regexp.MustCompile(`^[A-Za-z]{3}[0-9][A-Za-z0-9][0-9]{2}$`)
)

type vehicleRequestDto struct {
	LicensePlate string `json:"license_plate"`
	Brand        string `json:"brand"`
	Model        string `json:"model"`
	Year         string `json:"year"`
	CostumerId   string `json:"costumer_id"`
}

func (v *vehicleRequestDto) Domain() *domain.Vehicle {
	return domain.NewVehicle(v.LicensePlate, v.Brand, v.Model, v.Year, v.CostumerId)
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
	plateTrim := strings.TrimSpace(v.LicensePlate)
	if plateTrim == "" {
		return domain.ErrPlateIsRequired
	}

	if strings.TrimSpace(v.CostumerId) == "" {
		return domain.ErrNoCustomerAssociated
	}

	if len(plateTrim) != 7 || !platePattern.MatchString(plateTrim) {
		return domain.ErrInvalidPlate
	}

	return nil
}
