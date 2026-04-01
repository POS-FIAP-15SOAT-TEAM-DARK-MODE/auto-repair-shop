package supply

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type supplyRequestDTO struct {
	Name          string          `json:"name" binding:"required"`
	Description   string          `json:"description" binding:"required"`
	UnitPrice     decimal.Decimal `json:"unit_price" binding:"required"`
	StockQuantity int             `json:"stock_quantity" binding:"required"`
	Version       int             `json:"version" binding:"required"`
}

func mapBodyToCreateSupplyRequest(body *gin.Context) (*supplyRequestDTO, error) {
	var req supplyRequestDTO
	if err := body.ShouldBindJSON(&req); err != nil {
		return nil, json.CheckJsonError(err)
	}
	return &req, nil
}

func mapSupplyRequestDTOToDomain(req *supplyRequestDTO) *domain.Supply {
	return domain.NewSupply(req.Name, req.Description, req.UnitPrice, req.StockQuantity, req.Version)
}
