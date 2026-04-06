package service_order

import (
	"github.com/gin-gonic/gin"
)

type (
	initServiceOrderDTO struct {
		ClientID  string `json:"client"`
		VehicleID string `json:"vehicle"`
	}
)

func mapBodyToRequestDTO(body *gin.Context) (*initServiceOrderDTO, error) {
	req := new(initServiceOrderDTO)
	if err := body.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	return req, nil
}
