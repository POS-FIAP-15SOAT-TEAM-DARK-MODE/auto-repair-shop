package service_order_history

import (
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order_history/repository"
)

func NewHTTPController(_ *sql.DB) interfaces.ServiceOrderHistoryHTTPController {
	repo := repository.NewPostgres()
	svc := NewService(repo)
	return NewController(svc)
}
