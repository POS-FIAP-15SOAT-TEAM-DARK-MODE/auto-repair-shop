package service_order

import (
	"context"
	"slices"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/shopspring/decimal"
)

type memory_repo struct {
	data map[string]domain.ServiceOrder
	// serviceOrderID -> ordered list of works linked to that order (price = snapshot at link time)
	worksByOrder map[string][]domain.Work
}

func MemoryRepository() domain.ServiceOrderRepository {
	return &memory_repo{
		data:         make(map[string]domain.ServiceOrder),
		worksByOrder: make(map[string][]domain.Work),
	}
}

func (r *memory_repo) Save(_ context.Context, so *domain.ServiceOrder) error {
	r.data[so.ID] = *so
	return nil
}

func (r *memory_repo) ExistsByID(_ context.Context, id string) (bool, error) {
	_, ok := r.data[id]
	return ok, nil
}

func (r *memory_repo) ListWorksByServiceOrderID(_ context.Context, serviceOrderID string) ([]domain.Work, error) {
	works := r.worksByOrder[serviceOrderID]
	if works == nil {
		return []domain.Work{}, nil
	}
	out := make([]domain.Work, len(works))
	copy(out, works)
	return out, nil
}

func (r *memory_repo) AddWorkLink(_ context.Context, serviceOrderID, workID string, unitPrice decimal.Decimal) error {
	for _, w := range r.worksByOrder[serviceOrderID] {
		if w.ID == workID {
			return nil
		}
	}
	w := domain.Work{
		ID:          workID,
		Price:       unitPrice,
		Status:      domain.ACTIVE,
		Name:        "",
		Description: "",
	}
	r.worksByOrder[serviceOrderID] = append(r.worksByOrder[serviceOrderID], w)
	return nil
}

func (r *memory_repo) RemoveWorkLink(_ context.Context, serviceOrderID, workID string) error {
	list := r.worksByOrder[serviceOrderID]
	idx := slices.IndexFunc(list, func(w domain.Work) bool { return w.ID == workID })
	if idx < 0 {
		return domain.ErrServiceOrderWorkNotFound
	}
	r.worksByOrder[serviceOrderID] = slices.Delete(list, idx, idx+1)
	return nil
}
