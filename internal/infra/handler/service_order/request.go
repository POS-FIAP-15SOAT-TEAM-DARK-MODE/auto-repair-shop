package service_order

import (
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

type initServiceOrderDTO struct {
	ClientID  string `json:"client"`
	VehicleID string `json:"vehicle"`
}

func mapBodyToRequestDTO(body *gin.Context) (*initServiceOrderDTO, error) {
	req := new(initServiceOrderDTO)
	if err := body.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	return req, nil
}

type addWorksToServiceOrderDTO struct {
	Services []string `json:"services"`
}

func mapAddWorksBody(c *gin.Context) (*addWorksToServiceOrderDTO, error) {
	req := new(addWorksToServiceOrderDTO)
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}
	return req, nil
}

type addSuppliesToServiceOrderDTO struct {
	Supplies []struct {
		ID     string `json:"id"`
		Amount int    `json:"amount"`
	} `json:"supplies"`
}

func mapAddSuppliesBody(c *gin.Context) ([]domain.AddSupply, error) {
	req := new(addSuppliesToServiceOrderDTO)
	if err := c.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	sups := make([]domain.AddSupply, 0, len(req.Supplies))
	for _, r := range req.Supplies {
		sups = append(sups, domain.AddSupply{
			ID:     r.ID,
			Amount: r.Amount,
		})
	}

	return sups, nil
}

func mapListServiceOrderParamsToDomain(c *gin.Context) *domain.ServiceOrderFilterParams {
	var page, pageSize int64 = 1, 10

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

	return &domain.ServiceOrderFilterParams{
		Page:       page,
		PageSize:   pageSize,
		Limit:      pageSize,
		Offset:     (page - 1) * pageSize,
		Status:     c.Query("status"),
		CustomerID: c.Query("customerId"),
		VehicleID:  c.Query("vehicleId"),
	}
}
