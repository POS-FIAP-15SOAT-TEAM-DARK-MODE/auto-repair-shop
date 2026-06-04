package customer

import (
	"database/sql"
	"time"

	authRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/interfaces"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

func NewHTTPController(db *sql.DB) interfaces.CustomerHTTPController {
	unitOfWork := uow.NewTransactionalUoW(db)
	custRepo := customerRepo.NewPostgres()
	userRepo := authRepo.NewPostgres(time.Duration(0)) // session duration unused by customer service
	svc := NewService(unitOfWork, custRepo, userRepo)
	return NewController(svc)
}
