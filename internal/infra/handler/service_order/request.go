package service_order

import "github.com/gin-gonic/gin"

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
