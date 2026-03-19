package work

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type createWorkResDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
}

func mapResponseDTOFromDomain(work *domain.Work) createWorkResDTO {
	price, _ := work.Price.Float64()
	return createWorkResDTO{
		ID:          work.ID,
		Name:        work.Name,
		Description: work.Description,
		Price:       price,
		Status:      work.Status.String(),
	}
}
