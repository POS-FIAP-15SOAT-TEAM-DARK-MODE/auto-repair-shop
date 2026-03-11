package customer

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type (
	Handler struct {
		service domain.CustomerService
	}
)

func NewHandler(service domain.CustomerService) *Handler {
	return &Handler{
		service: service,
	}
}
