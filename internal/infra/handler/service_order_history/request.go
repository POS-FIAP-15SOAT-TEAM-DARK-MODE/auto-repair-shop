package service_order_history

import (
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
)

func mapListParamsToDomain(c *gin.Context) domain.SearchServiceOrderHistoryParams {
	return domain.SearchServiceOrderHistoryParams{ID: c.Param("id")}
}
