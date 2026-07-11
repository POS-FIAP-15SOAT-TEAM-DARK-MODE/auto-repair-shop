package service_order

import (
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/domain"
)

func mapWorksToResponse(works []workDomain.Work) []adapters.WorkItemResponse {
	if len(works) == 0 {
		return nil
	}
	items := make([]adapters.WorkItemResponse, 0, len(works))
	for _, w := range works {
		items = append(items, adapters.WorkItemResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			Price:       w.Price.String(),
			Status:      w.Status.String(),
		})
	}
	return items
}

func mapSuppliesToResponse(supplies []supplyDomain.Supply) []adapters.SupplyItemResponse {
	if len(supplies) == 0 {
		return nil
	}
	items := make([]adapters.SupplyItemResponse, 0, len(supplies))
	for _, s := range supplies {
		items = append(items, adapters.SupplyItemResponse{
			ID:          s.ID,
			Name:        s.Name,
			Description: s.Description,
			Price:       s.UnitPrice.String(),
			Version:     s.Version,
			Amount:      s.StockQuantity,
		})
	}
	return items
}

func mapSOToListItem(so domain.ServiceOrder) adapters.SOListItemResponse {
	item := adapters.SOListItemResponse{
		ID:         so.ID,
		Status:     so.Status.String(),
		TotalValue: so.TotalAmount.InexactFloat64(),
	}
	if so.Customer != nil {
		item.Customer = mapCustomerResponse(so)
	}
	if so.Vehicle != nil {
		item.Vehicle = adapters.VehicleResponse{
			ID:           so.Vehicle.ID,
			BrandModel:   fmt.Sprintf("%s - %s", so.Vehicle.Brand, so.Vehicle.Model),
			Year:         so.Vehicle.Year,
			LicensePlate: so.Vehicle.LicensePlate,
		}
	}
	return item
}

func mapSOToDetailResponse(so domain.ServiceOrder, works []adapters.WorkItemResponse, supplies []adapters.SupplyItemResponse) adapters.SODetailResponse {
	resp := adapters.SODetailResponse{
		ID:         so.ID,
		Status:     so.Status.String(),
		TotalValue: so.TotalAmount.InexactFloat64(),
		Works:      works,
		Supplies:   supplies,
	}
	if so.Customer != nil {
		resp.Customer = mapCustomerResponse(so)
	}
	if so.Vehicle != nil {
		resp.Vehicle = adapters.VehicleResponse{
			ID:           so.Vehicle.ID,
			BrandModel:   fmt.Sprintf("%s - %s", so.Vehicle.Brand, so.Vehicle.Model),
			Year:         so.Vehicle.Year,
			LicensePlate: so.Vehicle.LicensePlate,
		}
	}
	return resp
}

func mapCustomerResponse(so domain.ServiceOrder) adapters.CustomerResponse {
	c := so.Customer
	resp := adapters.CustomerResponse{
		ID:          c.ID,
		Type:        c.Type.String(),
		CompanyName: c.CompanyName,
		Phone:       c.Phone,
	}
	if c.Type == "INDIVIDUAL" {
		resp.Document = c.CPF
	} else {
		resp.Document = c.CNPJ
	}
	if c.User != nil {
		resp.Name = c.User.Name
		resp.Email = c.User.Email
	}
	return resp
}
