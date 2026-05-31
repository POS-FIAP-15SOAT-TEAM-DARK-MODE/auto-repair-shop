package service_order_history

import (
	"context"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

// MemoryRepository returns a no-op implementation used by unit tests.
func MemoryRepository() domain.ServiceOrderHistoryRepository {
	return &memory_repo{data: nil}
}

type memory_repo struct {
	data map[string]domain.ServiceOrder
}

func (r *memory_repo) Search(_ context.Context, _ *domain.SearchServiceOrderHistoryParams) ([]domain.ServiceOrderHistoryItem, error) {
	return []domain.ServiceOrderHistoryItem{}, nil
}

func (r *memory_repo) SearchWorkTransitionsByServiceOrderID(_ context.Context, _ string) ([]domain.WorkTransitionGroup, error) {
	return []domain.WorkTransitionGroup{}, nil
}

func (r *memory_repo) InsertWorkHistory(_ context.Context, _, _ string, _ domain.SERVICE_ORDER_STATUS) error {
	return nil
}

// MemoryStore is a real in-memory history store used by integration tests.
// The service-order repo calls Record whenever a status transition occurs,
// and Search reads those entries back through the ServiceOrderHistoryRepository interface.
type MemoryStore struct {
	entries map[string][]domain.ServiceOrderHistoryItem
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{entries: make(map[string][]domain.ServiceOrderHistoryItem)}
}

// Record is called by the in-memory service-order repo when a status changes.
func (s *MemoryStore) Record(soID string, prev *domain.SERVICE_ORDER_STATUS, newStatus domain.SERVICE_ORDER_STATUS) {
	item := domain.ServiceOrderHistoryItem{NewStatus: newStatus, CreatedAt: time.Now()}
	if prev != nil {
		item.PreviousStatus = *prev
	}
	s.entries[soID] = append(s.entries[soID], item)
}

func (s *MemoryStore) Search(_ context.Context, params *domain.SearchServiceOrderHistoryParams) ([]domain.ServiceOrderHistoryItem, error) {
	return s.entries[params.ID], nil
}

func (s *MemoryStore) SearchWorkTransitionsByServiceOrderID(_ context.Context, _ string) ([]domain.WorkTransitionGroup, error) {
	return []domain.WorkTransitionGroup{}, nil
}

func (s *MemoryStore) InsertWorkHistory(_ context.Context, _, _ string, _ domain.SERVICE_ORDER_STATUS) error {
	return nil
}
