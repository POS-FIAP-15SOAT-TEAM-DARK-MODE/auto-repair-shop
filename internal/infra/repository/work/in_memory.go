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

func (r *memory_repo) Count(_ context.Context, params *domain.SearchWorkParams) (int64, error) {
	if params.Status == "" {
		return int64(len(r.data)), nil
	}

	status := domain.StringToWorkStatus(params.Status)
	var total int64
	for _, v := range r.data {
		if v.Status == status {
			total++
		}
	}
	return total, nil
}

func (r *memory_repo) Search(_ context.Context, params *domain.SearchWorkParams) ([]domain.Work, error) {
	work := make([]domain.Work, 0, params.Limit)

	var filtered []domain.Work
	for _, v := range r.data {
		if params.Status != "" && v.Status != domain.StringToWorkStatus(params.Status) {
			continue
		}
		filtered = append(filtered, v)
	}

	start := int(params.Offset)
	if start >= len(filtered) {
		return work, nil
	}

	for i := start; i < len(filtered) && len(work) < int(params.Limit); i++ {
		work = append(work, filtered[i])
	}

	return work, nil
}
