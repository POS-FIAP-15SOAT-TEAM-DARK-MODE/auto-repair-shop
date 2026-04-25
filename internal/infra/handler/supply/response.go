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

type paginatorResponseDTO struct {
	Items      []supplyResponseDTO `json:"items"`
	TotalItems int64               `json:"totalItems"`
	TotalPages int64               `json:"totalPages"`
	PageSize   int64               `json:"pageSize"`
	Page       int64               `json:"page"`
}

func mapListResponseDTOFromDomain(supplies *domain.PaginatorResponse[domain.Supply]) paginatorResponseDTO {
	items := make([]supplyResponseDTO, len(supplies.Items))
	for i, s := range supplies.Items {
		supply := s
		items[i] = mapSupplyToResponseDTO(&supply)
	}

	return paginatorResponseDTO{
		Items:      items,
		TotalItems: supplies.TotalItems,
		TotalPages: supplies.TotalPages,
		PageSize:   supplies.PageSize,
		Page:       supplies.Page,
	}
}
