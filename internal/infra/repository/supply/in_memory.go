package supply

import (
	"context"
	"sort"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.Supply
}

func MemoryRepository() domain.SupplyRepository {
	return &memory_repo{
		data: make(map[string]domain.Supply),
	}
}

func (r *memory_repo) Save(_ context.Context, s *domain.Supply) error {
	r.data[s.ID] = *s
	return nil
}

func (r *memory_repo) FindById(_ context.Context, id string) (domain.Supply, error) {
	s, ok := r.data[id]
	if !ok {
		return domain.Supply{}, domain.ErrSupplyNotFound
	}
	return s, nil
}

func (r *memory_repo) Delete(_ context.Context, id string) error {
	delete(r.data, id)
	return nil
}

func (r *memory_repo) Count(_ context.Context, params *domain.ListSupplyParams) (int64, error) {
	return int64(len(r.data)), nil
}

func (r *memory_repo) Search(_ context.Context, params *domain.ListSupplyParams) ([]domain.Supply, error) {
	items := make([]domain.Supply, 0, len(r.data))
	for _, v := range r.data {
		items = append(items, v)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	start := int(params.Offset())
	if start >= len(items) {
		return []domain.Supply{}, nil
	}

	end := start + int(params.PageSize)
	if end > len(items) {
		end = len(items)
	}

	return items[start:end], nil
}

func (r *memory_repo) DecrementStock(_ context.Context, id string, amount int) error {
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

func (r *memory_repo) RestoreStock(_ context.Context, id string, amount int) error {
	s, ok := r.data[id]
	if !ok {
		return domain.ErrSupplyNotFound
	}
	s.StockQuantity += amount
	r.data[id] = s
	return nil
}
