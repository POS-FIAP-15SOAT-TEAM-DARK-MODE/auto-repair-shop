package repository

import (
	"context"
	"time"

	soDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order_history/interfaces"
)

type inMemory struct{}

func NewInMemory() interfaces.ServiceOrderHistoryRepository {
	return &inMemory{}
}

func (r *inMemory) Search(_ context.Context, _ *domain.SearchParams) ([]domain.ServiceOrderHistoryItem, error) {
	return []domain.ServiceOrderHistoryItem{}, nil
}

func (r *inMemory) SearchWorkTransitionsByServiceOrderID(_ context.Context, _ string) ([]domain.WorkTransitionGroup, error) {
	return []domain.WorkTransitionGroup{}, nil
}

// MemoryStore is a real in-memory history store for integration tests.
// The service-order repo calls Record when a status transition occurs.
type MemoryStore struct {
	entries map[string][]domain.ServiceOrderHistoryItem
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{entries: make(map[string][]domain.ServiceOrderHistoryItem)}
}

func (s *MemoryStore) Record(soID string, prev *soDomain.SERVICE_ORDER_STATUS, newStatus soDomain.SERVICE_ORDER_STATUS) {
	item := domain.ServiceOrderHistoryItem{NewStatus: newStatus, CreatedAt: time.Now()}
	if prev != nil {
		item.PreviousStatus = *prev
	}
	s.entries[soID] = append(s.entries[soID], item)
}

func (s *MemoryStore) Search(_ context.Context, params *domain.SearchParams) ([]domain.ServiceOrderHistoryItem, error) {
	return s.entries[params.ID], nil
}

func (s *MemoryStore) SearchWorkTransitionsByServiceOrderID(_ context.Context, _ string) ([]domain.WorkTransitionGroup, error) {
	return []domain.WorkTransitionGroup{}, nil
}
