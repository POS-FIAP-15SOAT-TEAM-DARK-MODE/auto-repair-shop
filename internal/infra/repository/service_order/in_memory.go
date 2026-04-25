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
	worksByOrder    map[string][]domain.Work
	suppliesByOrder map[string][]domain.Supply
}

func MemoryRepository() domain.ServiceOrderRepository {
	return &memory_repo{
		data:            make(map[string]domain.ServiceOrder),
		worksByOrder:    make(map[string][]domain.Work),
		suppliesByOrder: make(map[string][]domain.Supply),
	}
}

func (r *memory_repo) Save(_ context.Context, so *domain.ServiceOrder) error {
	r.data[so.ID] = *so
	return nil
}

func (r *memory_repo) ExistsByID(_ context.Context, id string) (bool, domain.SERVICE_ORDER_STATUS, error) {
	v, ok := r.data[id]
	return ok, v.Status, nil
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

func (r *memory_repo) AddSupplyLink(_ context.Context, serviceOrderID string, supplyID string, amount int, unitPrice decimal.Decimal) error {
	for _, w := range r.worksByOrder[serviceOrderID] {
		if w.ID == supplyID {
			return nil
		}
	}
	s := domain.Supply{
		ID:            supplyID,
		UnitPrice:     unitPrice,
		StockQuantity: amount,
		Name:          "",
		Description:   "",
	}
	r.suppliesByOrder[serviceOrderID] = append(r.suppliesByOrder[serviceOrderID], s)
	return nil
}

func (r *memory_repo) ListSuppliesByServiceOrderID(_ context.Context, serviceOrderID string) ([]domain.Supply, error) {
	supplies := r.suppliesByOrder[serviceOrderID]
	if supplies == nil {
		return []domain.Supply{}, nil
	}
	out := make([]domain.Supply, len(supplies))
	copy(out, supplies)
	return out, nil
}

func (r *memory_repo) RemoveSupplyLink(_ context.Context, serviceOrderID string, supplyID string) (int, error) {
	list := r.suppliesByOrder[serviceOrderID]
	idx := slices.IndexFunc(list, func(w domain.Supply) bool { return w.ID == supplyID })
	if idx < 0 {
		return 0, domain.ErrServiceOrderSupplyNotFound
	}
	qty := list[idx].StockQuantity
	r.suppliesByOrder[serviceOrderID] = slices.Delete(list, idx, idx+1)
	return qty, nil
}

func (r *memory_repo) FindByID(ctx context.Context, id string) (domain.ServiceOrder, error) {
	so, ok := r.data[id]
	if !ok {
		return so, domain.ErrServiceOrderNotFound
	}

	return so, nil
}
