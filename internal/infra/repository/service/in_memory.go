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
