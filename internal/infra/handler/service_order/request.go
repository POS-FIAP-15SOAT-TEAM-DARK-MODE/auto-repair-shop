package service_order

import (
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

type (
	initServiceOrderDTO struct {
		ClientID  string `json:"client"`
		VehicleID string `json:"vehicle"`
	}

	getServiceOrderHistoryDTO struct {
		ID string `json:"id"`
	}
)

func mapBodyToRequestDTO(body *gin.Context) (*initServiceOrderDTO, error) {
	req := new(initServiceOrderDTO)
	if err := body.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	return req, nil
}

func (r *getServiceOrderHistoryDTO) validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return domain.ErrServiceOrderIDRequired
	}
	return nil
}
