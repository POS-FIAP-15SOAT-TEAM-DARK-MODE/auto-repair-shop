package vehicle

import (
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/repository"
)

func NewHTTPController(db *sql.DB) interfaces.VehicleHTTPController {
	unitOfWork := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres()
	svc := NewService(unitOfWork, repo, id.NewIDGenerator())
	return NewController(svc)
}
