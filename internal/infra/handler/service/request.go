package service

import (
	"strconv"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

type createServiceReqDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      string `json:"status"`
}

func mapBodyToRequestDTO(body *gin.Context) (*createServiceReqDTO, error) {
	req := new(createServiceReqDTO)

	// TODO: replace by a performant alternative, build a custom serializer
	if err := body.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	return req, nil
}

func (c *createServiceReqDTO) Validate() error {
	if err := domain.ValidServiceStatusStringValue(c.Status); err != nil {
		return err
	}

	// Acceptable: optional minus, digits, at most one dot, any number of decimals (negative, unlimited decimals)
	// e.g., -123.45, 0.1, 44, -.50
	// Regex: ^-?\d*(\.\d*)?$   (but must not be just minus or just dot)
	if c.Price == "-" || c.Price == "." || c.Price == "-." {
		return domain.ErrInvalidPriceValue
	}

	// Remove thousands separators if present (like "1,000.23" -> "1000.23")
	// Only allow dot as decimal separator
	c.Price = strings.TrimSpace(strings.ReplaceAll(c.Price, ",", ""))

	return nil
}

func (c *createServiceReqDTO) MapToDomain() (*domain.Service, error) {
	return domain.NewService(c.Name, c.Description, c.Price, domain.StringToServiceStatus(c.Status))
}

func mapListParamsToDomain(c *gin.Context) *domain.ListServiceParams {
	page := int64(1)
	pageSize := int64(10)
	var status string

	if p := c.Query("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = int64(v)
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 {
			pageSize = int64(v)
		}
	}

	status = c.Query("status")
	if err := domain.ValidServiceStatusStringValue(status); status != "" && err != nil {
		status = ""
	}

	return &domain.ListServiceParams{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	}
}
