package service

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.Service
}

func MemoryRepository() domain.ServiceRepository {
	return &memory_repo{
		data: make(map[string]domain.Service),
	}
}

func (r *memory_repo) Save(_ context.Context, svc *domain.Service) error {
	r.data[svc.ID] = *svc
	return nil
}

func (r *memory_repo) Count(_ context.Context, params *domain.SearchServiceParams) (int64, error) {
	if params.Status == "" {
		return int64(len(r.data)), nil
	}

	status := domain.StringToServiceStatus(params.Status)
	var total int64
	for _, v := range r.data {
		if v.Status == status {
			total++
		}
	}
	return total, nil
}

func (r *memory_repo) Search(_ context.Context, params *domain.SearchServiceParams) ([]domain.Service, error) {
	services := make([]domain.Service, 0, params.Limit)

	var filtered []domain.Service
	for _, v := range r.data {
		if params.Status != "" && v.Status != domain.StringToServiceStatus(params.Status) {
			continue
		}
		filtered = append(filtered, v)
	}

	start := int(params.Offset)
	if start >= len(filtered) {
		return services, nil
	}

	for i := start; i < len(filtered) && len(services) < int(params.Limit); i++ {
		services = append(services, filtered[i])
	}

	return services, nil
}
