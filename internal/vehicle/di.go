package vehicle

import (
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/repository"
)

func NewHTTPController(db *sql.DB) interfaces.VehicleHTTPController {
	unitOfWork := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres()
	svc := NewService(unitOfWork, repo, id.NewIDGenerator())
	return NewController(svc)
}
