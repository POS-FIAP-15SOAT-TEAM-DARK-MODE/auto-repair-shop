package work

import "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"

type createWorkResDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Status      string `json:"status"`
}

func mapResponseDTOFromDomain(work *domain.Work) createWorkResDTO {
	return createWorkResDTO{
		ID:          work.ID,
		Name:        work.Name,
		Description: work.Description,
		Price:       work.Price.String(),
		Status:      work.Status.String(),
	}
}

type paginatorResponseDTO struct {
	Items      []createWorkResDTO
	TotalItems int64
	TotalPages int64
	PageSize   int64
	Page       int64
}

func mapListResponseDTOFromDomain(works *domain.PaginatorResponse[domain.Work]) paginatorResponseDTO {
	items := make([]createWorkResDTO, len(works.Items))
	for i, w := range works.Items {
		work := w
		items[i] = mapResponseDTOFromDomain(&work)
	}

	return paginatorResponseDTO{
		Items:      items,
		TotalItems: works.TotalItems,
		TotalPages: works.TotalPages,
		PageSize:   works.PageSize,
		Page:       works.Page,
	}
}
