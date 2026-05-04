package supply

import (
	"context"
	"slices"
	"sync"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memoryRepo struct {
	mu   sync.Mutex
	data map[string]domain.Supply
}

func MemoryRepository() domain.SupplyRepository {
	return &memoryRepo{data: make(map[string]domain.Supply)}
}

func (r *memoryRepo) Save(_ context.Context, s *domain.Supply) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[s.ID] = *s
	return nil
}

func (r *memoryRepo) FindById(_ context.Context, id string) (domain.Supply, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.data[id]
	if !ok {
		return domain.Supply{}, domain.ErrSupplyNotFound
	}
	return s, nil
}

func (r *memoryRepo) Search(_ context.Context, params *domain.ListSupplyParams) ([]domain.Supply, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	all := make([]domain.Supply, 0, len(r.data))
	for _, s := range r.data {
		all = append(all, s)
	}
	slices.SortFunc(all, func(a, b domain.Supply) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	offset := int(params.Offset())
	if offset >= len(all) {
		return []domain.Supply{}, nil
	}
	end := offset + int(params.PageSize)
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *memoryRepo) Count(_ context.Context, _ *domain.ListSupplyParams) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return int64(len(r.data)), nil
}

func (r *memoryRepo) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}

func (r *memoryRepo) DecrementStock(_ context.Context, id string, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.data[id]
	if !ok {
		return domain.ErrSupplyNotFound
	}
	if s.StockQuantity < amount {
		return domain.ErrSupplyOutOfStock
	}
	s.StockQuantity -= amount
	r.data[id] = s
	return nil
}

func (r *memoryRepo) RestoreStock(_ context.Context, id string, amount int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.data[id]
	if !ok {
		return domain.ErrSupplyNotFound
	}
	s.StockQuantity += amount
	r.data[id] = s
	return nil
}
