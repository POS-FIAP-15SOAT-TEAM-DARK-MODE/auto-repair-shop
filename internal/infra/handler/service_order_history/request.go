package service_order

import (
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type getServiceOrderHistoryDTO struct {
	ID string `json:"id"`
}

func (r *getServiceOrderHistoryDTO) validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return domain.ErrServiceOrderIDRequired
	}
	return nil
}
