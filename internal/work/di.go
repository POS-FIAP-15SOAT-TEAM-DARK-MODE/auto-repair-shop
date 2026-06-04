package work

import (
	"database/sql"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/repository"
)

func NewHTTPController(db *sql.DB) interfaces.WorkHTTPController {
	unitOfWork := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres()
	svc := NewService(unitOfWork, repo)
	return NewController(svc)
}
