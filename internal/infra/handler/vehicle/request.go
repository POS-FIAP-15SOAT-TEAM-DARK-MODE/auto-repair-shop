package vehicle

import (
	"errors"
	"strconv"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/gin-gonic/gin"
)

const (
	customerIDParam = "customerId"
	pageParam       = "page"
	pageSizeParam   = "pageSize"
	defaultPageSize = 10
	defaultPage     = 1
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

func createListParams(ctx *gin.Context) (*domain.ListVehicleParams, error) {
	pg := int64(defaultPage)
	pgSize := int64(defaultPageSize)

	id := ctx.Param(customerIDParam)
	if id == "" {
		return nil, domain.ErrInvalidCustomerId
	}

	if page := ctx.Query(pageParam); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > defaultPage {
			pg = int64(p)
		}
	}
	if pageSize := ctx.Query(pageSizeParam); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > defaultPageSize {
			pgSize = int64(ps)
		}
	}

	return &domain.ListVehicleParams{
		CustomerID: id,
		Page:       pg,
		PageSize:   pgSize,
	}, nil
}
