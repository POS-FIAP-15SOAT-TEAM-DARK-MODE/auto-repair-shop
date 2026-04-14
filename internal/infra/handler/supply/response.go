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

type paginatorResponseDTO struct {
	Items      []supplyResponseDTO `json:"items"`
	TotalItems int64               `json:"total_items"`
	TotalPages int64               `json:"total_pages"`
	PageSize   int64               `json:"page_size"`
	Page       int64               `json:"page"`
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

func domainListToResponseDTO(list *domain.PaginatorResponse[domain.Supply]) paginatorResponseDTO {
	items := make([]supplyResponseDTO, len(list.Items))
	for i, item := range list.Items {
		items[i] = mapSupplyToResponseDTO(&item)
	}
	return paginatorResponseDTO{
		Items:      items,
		TotalItems: list.TotalItems,
		TotalPages: list.TotalPages,
		PageSize:   list.PageSize,
		Page:       list.Page,
	}
}
