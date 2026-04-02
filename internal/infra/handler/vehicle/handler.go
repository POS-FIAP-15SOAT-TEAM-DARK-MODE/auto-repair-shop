package vehicle

import (
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	plateParam = "plate"
	idParam    = "id"
)

type handler struct {
	service domain.VehicleService
}

func HttpHandler(service domain.VehicleService) *handler {
	return &handler{
		service,
	}
}

func (h *handler) Create(c *gin.Context) {
	ctx := c.Request.Context()

	reqDTO, err := validateAndCreateDTO(c)
	if err != nil {
		status, resp := web.Error(err)
		logger.Of(ctx).Debug("Vehicle validation error",
			zap.String("operation", "create_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		c.JSON(status, resp)
		return
	}

	vehicleDomain := reqDTO.Domain()
	if err = h.service.Create(ctx, vehicleDomain); err != nil {
		status, resp := web.Error(err)
		logger.Of(ctx).Debug("Vehicle creation error",
			zap.String("operation", "create_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusCreated, domainToResponseDto(vehicleDomain))
}

func (h *handler) FindByLicensePlate(c *gin.Context) {
	ctx := c.Request.Context()
	plate := domain.NormalizeLicensePlate(c.Query(plateParam))

	vehicle, err := h.service.FindByLicensePlate(ctx, plate)
	if err != nil {
		status, resp := web.Error(err)
		logger.Of(ctx).Debug("Search vehicle by license plate error",
			zap.String("operation", "find_by_license_plate"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, domainToResponseDto(vehicle))
}

func (h *handler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	reqDTO, err := validateAndCreateDTO(c)
	if err != nil {
		status, resp := web.Error(err)
		logger.Of(ctx).Debug("Vehicle validation error",
			zap.String("operation", "update_vehicle"),
			zap.Error(err),
			zap.String("entity", "vehicle"),
		)
		c.JSON(status, resp)
		return
	}

	id := c.Param(idParam)
	if id == "" {
		status, resp := web.Error(domain.ErrInvalidVehicleId)
		logger.Of(ctx).Debug("Vehicle validation error",
			zap.String("operation", "update_vehicle"),
			zap.Error(domain.ErrInvalidVehicleId),
			zap.String("entity", "vehicle"),
		)
		c.JSON(status, resp)
		return
	}

	vehicleDomain := reqDTO.Domain()
	vehicleDomain.ID = id
	if err = h.service.Update(ctx, vehicleDomain); err != nil {
		status, resp := web.Error(err)
		logger.Of(ctx).Debug("Vehicle update error")
		c.JSON(status, resp)
		return
	}

	c.JSON(http.StatusOK, domainToResponseDto(vehicleDomain))
}
