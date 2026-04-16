package supply

import (
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/json"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type createSupplyRequestDTO struct {
	Name          string          `json:"name" binding:"required"`
	Description   string          `json:"description" binding:"required"`
	UnitPrice     decimal.Decimal `json:"unitPrice" binding:"required"`
	StockQuantity int             `json:"stockQuantity" binding:"required"`
}
type UpdateSupplyRequest struct {
	Name          *string          `json:"name"`
	Description   *string          `json:"description"`
	UnitPrice     *decimal.Decimal `json:"unitPrice"`
	StockQuantity *int             `json:"stockQuantity"`
}

func mapBodyToRequestDTO(body *gin.Context) (*createSupplyRequestDTO, error) {
	req := new(createSupplyRequestDTO)

	if err := body.ShouldBindJSON(req); err != nil {
		return nil, err
	}

	return req, nil
}

func mapBodyToUpdateSupplyRequest(body *gin.Context) (*UpdateSupplyRequest, error) {
	var req UpdateSupplyRequest
	if err := body.ShouldBindJSON(&req); err != nil {
		return nil, json.CheckJsonError(err)
	}
	return &req, nil
}

func mapUpdateSupplyRequestDTOToDomain(id string, req *UpdateSupplyRequest) *domain.SupplyUpdate {
	return domain.UpdateSupply(id, req.Name, req.Description, req.UnitPrice, req.StockQuantity)
}

func (c *createSupplyRequestDTO) MapToDomain() *domain.Supply {
	supply := domain.NewSupply(c.Name, c.Description, c.UnitPrice, c.StockQuantity, 0)
	return supply
}

func mapListParamsToDomain(c *gin.Context) *domain.ListSupplyParams {
	page := int64(1)
	pageSize := int64(10)
	var status string

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

	status = c.Query("status")
	if err := domain.ValidWorkStatusStringValue(status); status != "" && err != nil {
		status = ""
	}

	return &domain.ListSupplyParams{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
	}
}
