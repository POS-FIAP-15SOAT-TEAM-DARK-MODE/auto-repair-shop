package supply

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/shopspring/decimal"
)

type supplyResponseDTO struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	UnitPrice     decimal.Decimal `json:"unit_price"`
	StockQuantity int             `json:"stock_quantity"`
	Version       int             `json:"version"`
}

func mapSupplyToResponseDTO(supply *domain.Supply) supplyResponseDTO {
	return supplyResponseDTO{
		ID:            supply.ID,
		Name:          supply.Name,
		Description:   supply.Description,
		UnitPrice:     supply.UnitPrice,
		StockQuantity: supply.StockQuantity,
		Version:       supply.Version,
	}
}
