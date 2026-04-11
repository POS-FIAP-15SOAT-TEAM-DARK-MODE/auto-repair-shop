package service_order_history

import (
	"strconv"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

const (
	defaultPage     = int64(1)
	defaultPageSize = int64(10)
)

func mapListParamsToDomain(c *gin.Context) domain.ListServiceOrderHistoryByIDParams {
	page := defaultPage
	pageSize := defaultPageSize

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

	id := c.Param("id")

	return domain.ListServiceOrderHistoryByIDParams{Page: page, PageSize: pageSize, ID: id}
}
