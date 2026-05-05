package service_order

import (
	"fmt"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type (
	serviceOrderResponse struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	workItemResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       string `json:"price"`
		Status      string `json:"status"`
	}

	supplyItemResponse struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       string `json:"price"`
		Version     int    `json:"version"`
		Amount      int    `json:"amount"`
	}

	vehicleResponseDTO struct {
		ID           string `json:"id"`
		LicensePlate string `json:"licensePlate"`
		BrandModel   string `json:"brandModel"`
		Year         int    `json:"year"`
	}

	customerResponse struct {
		ID          string              `json:"id"`
		Name        string              `json:"name"`
		Email       string              `json:"email"`
		Type        domain.CustomerType `json:"type"`
		Document    string              `json:"document"`
		CompanyName string              `json:"companyName,omitempty"`
		Phone       string              `json:"phone"`
	}

	listResponse[T any] struct {
		Items      []T   `json:"items"`
		TotalItems int64 `json:"totalItems"`
		TotalPages int64 `json:"totalPages"`
		PageSize   int64 `json:"pageSize"`
		Page       int64 `json:"page"`
	}

	serviceOrderDetailResponse struct {
		ID         string               `json:"id"`
		Status     string               `json:"status"`
		TotalValue float64              `json:"totalValue"`
		Customer   customerResponse     `json:"customer"`
		Vehicle    vehicleResponseDTO   `json:"vehicle"`
		Works      []workItemResponse   `json:"works,omitempty"`
		Supplies   []supplyItemResponse `json:"supplies,omitempty"`
	}
)

func mapResponseDTOFromDomain(so domain.ServiceOrder) serviceOrderResponse {
	return serviceOrderResponse{
		ID:     so.ID,
		Status: so.Status.String(),
	}
}

func mapWorks(works []domain.Work) []workItemResponse {
	if len(works) == 0 {
		return nil
	}
	items := make([]workItemResponse, 0, len(works))
	for _, w := range works {
		items = append(items, workItemResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			Price:       w.Price.String(),
			Status:      w.Status.String(),
		})
	}
	return items
}

func mapWorksListToResponse(works []domain.Work) listResponse[workItemResponse] {
	items := mapWorks(works)
	n := int64(len(items))
	return listResponse[workItemResponse]{
		Items:      items,
		TotalItems: n,
		TotalPages: 1,
		PageSize:   n,
		Page:       1,
	}
}

type workExecutionTimeResponse struct {
	WorkID               string  `json:"workId"`
	WorkName             string  `json:"workName"`
	AverageExecutionTime float64 `json:"averageExecutionTimeHours"`
}

func mapSupplies(supplies []domain.Supply) []supplyItemResponse {
	if len(supplies) == 0 {
		return nil
	}
	items := make([]supplyItemResponse, 0, len(supplies))
	for _, w := range supplies {
		items = append(items, supplyItemResponse{
			ID:          w.ID,
			Name:        w.Name,
			Description: w.Description,
			Price:       w.UnitPrice.String(),
			Version:     w.Version,
			Amount:      w.StockQuantity,
		})
	}
	return items
}

func mapSuppliesListToResponse(supplies []domain.Supply) listResponse[supplyItemResponse] {
	items := mapSupplies(supplies)
	n := int64(len(items))
	return listResponse[supplyItemResponse]{
		Items:      items,
		TotalItems: n,
		TotalPages: 1,
		PageSize:   n,
		Page:       1,
	}
}

func mapCustomerResponse(c domain.Customer) customerResponse {
	var document string
	if c.Type == domain.IndividualCustomerType {
		document = c.CPF
	} else {
		document = c.CNPJ
	}

	return customerResponse{
		ID:          c.ID,
		Name:        c.User.Name,
		Email:       c.User.Email,
		Type:        c.Type,
		Document:    document,
		CompanyName: c.CompanyName,
		Phone:       c.Phone,
	}
}

func mapVehicleResponse(v domain.Vehicle) vehicleResponseDTO {
	return vehicleResponseDTO{
		ID:           v.ID,
		BrandModel:   fmt.Sprintf("%s - %s", v.Brand, v.Model),
		Year:         v.Year,
		LicensePlate: v.LicensePlate,
	}
}

func mapServiceOrderDetailResponse(order domain.FullServiceOrder) serviceOrderDetailResponse {
	return serviceOrderDetailResponse{
		ID:         order.ID,
		Status:     order.Status.String(),
		TotalValue: order.TotalAmount.InexactFloat64(),
		Customer:   mapCustomerResponse(*order.Customer),
		Vehicle:    mapVehicleResponse(*order.Vehicle),
		Works:      mapWorks(order.Works),
		Supplies:   mapSupplies(order.Supplies),
	}
}

type serviceOrderListItemResponse struct {
	ID         string             `json:"id"`
	Status     string             `json:"status"`
	TotalValue float64            `json:"totalValue"`
	Customer   customerResponse   `json:"customer"`
	Vehicle    vehicleResponseDTO `json:"vehicle"`
}

func mapServiceOrderListItemResponse(so domain.ServiceOrder) serviceOrderListItemResponse {
	var customer customerResponse
	var vehicle vehicleResponseDTO

	if so.Customer != nil {
		customer = mapCustomerResponse(*so.Customer)
	}
	if so.Vehicle != nil {
		vehicle = mapVehicleResponse(*so.Vehicle)
	}

	return serviceOrderListItemResponse{
		ID:         so.ID,
		Status:     so.Status.String(),
		TotalValue: so.TotalAmount.InexactFloat64(),
		Customer:   customer,
		Vehicle:    vehicle,
	}
}

func mapServiceOrderListResponse(page *domain.PaginatorResponse[domain.ServiceOrder]) listResponse[serviceOrderListItemResponse] {
	items := make([]serviceOrderListItemResponse, 0, len(page.Items))
	for _, so := range page.Items {
		items = append(items, mapServiceOrderListItemResponse(so))
	}

	return listResponse[serviceOrderListItemResponse]{
		Items:      items,
		TotalItems: page.TotalItems,
		TotalPages: page.TotalPages,
		PageSize:   page.PageSize,
		Page:       page.Page,
	}
}
