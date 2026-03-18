package work

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.Work
}

func MemoryRepository() domain.WorkRepository {
	return &memory_repo{
		data: make(map[string]domain.Work),
	}
}

func (r *memory_repo) Save(_ context.Context, svc *domain.Work) error {
	r.data[svc.ID] = *svc
	return nil
}
