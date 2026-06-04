package service_order_history

import (
	"context"
	"net/http"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/interfaces"
	"go.uber.org/zap"
)

type controller struct {
	svc interfaces.ServiceOrderHistoryService
}

func NewController(svc interfaces.ServiceOrderHistoryService) interfaces.ServiceOrderHistoryHTTPController {
	return &controller{svc: svc}
}

func (c *controller) GetHistoryByID(ctx context.Context, req *http.Request) ([]adapters.SOHistoryResponse, error) {
	id, _ := ctx.Value("id").(string)
	params := &domain.SearchParams{ID: id}

	if err := params.Validate(); err != nil {
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Failed to validate service order history params",
			zap.String("operation", "get_service_order_history"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		return nil, err
	}

	items, err := c.svc.GetHistoryByID(ctx, params)
	if err != nil {
		logger.Of(ctx).Error(err)
		logger.Of(ctx).Debug("Service order history retrieval failed",
			zap.String("operation", "get_service_order_history"),
			zap.Error(err),
			zap.String("entity", "service_order"),
		)
		return nil, err
	}

	return adapters.MapToSOHistoryResponseList(items), nil
}
