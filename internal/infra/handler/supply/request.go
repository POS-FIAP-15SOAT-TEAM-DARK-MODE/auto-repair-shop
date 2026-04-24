package supply

import (
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type writeSupplyRequestDTO struct {
	Name          string          `json:"name" binding:"required"`
	Description   string          `json:"description" binding:"required"`
	UnitPrice     decimal.Decimal `json:"unitPrice" binding:"required"`
	StockQuantity int             `json:"stockQuantity" binding:"required"`
}

func mapBodyToRequestDTO(body *gin.Context) (*writeSupplyRequestDTO, error) {
	req := new(writeSupplyRequestDTO)

	if err := body.ShouldBindJSON(req); err != nil {
		return nil, json.CheckJsonError(err)
	}

	return req, nil
}

func (c *writeSupplyRequestDTO) MapToDomain() *domain.Supply {
	supply := domain.NewSupply(c.Name, c.Description, c.UnitPrice, c.StockQuantity, 0)
	return supply
}

func mapListParamsToDomain(c *gin.Context) *domain.ListSupplyParams {
	var page, pageSize int64 = 1, 10

	if p := c.Query("page"); p != "" {
		if v, err := strconv.ParseInt(p, 10, 64); err == nil && v > 0 {
			page = v
		}
	}

	if ps := c.Query("pageSize"); ps != "" {
		if v, err := strconv.ParseInt(ps, 10, 64); err == nil && v > 0 {
			pageSize = v
		}
	}

	return &domain.ListSupplyParams{
		Page:     page,
		PageSize: pageSize,
		Version:  c.Query("version"),
	}
}
