package service_order

import (
	"context"
	"slices"
	"sort"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/domain"
	"github.com/shopspring/decimal"
)

type memory_repo struct {
	data            map[string]domain.ServiceOrder
	worksByOrder    map[string][]domain.Work
	suppliesByOrder map[string][]supplyDomain.Supply
	onStatusChange  func(soID string, prev *domain.SERVICE_ORDER_STATUS, newStatus domain.SERVICE_ORDER_STATUS)
}

func MemoryRepository() domain.ServiceOrderRepository {
	return &memory_repo{
		data:            make(map[string]domain.ServiceOrder),
		worksByOrder:    make(map[string][]domain.Work),
		suppliesByOrder: make(map[string][]supplyDomain.Supply),
	}
}

// MemoryRepositoryWithHistory creates an in-memory repo that notifies onStatusChange on every
// status transition so that history can be recorded alongside the state change.
func MemoryRepositoryWithHistory(onStatusChange func(soID string, prev *domain.SERVICE_ORDER_STATUS, newStatus domain.SERVICE_ORDER_STATUS)) domain.ServiceOrderRepository {
	return &memory_repo{
		data:            make(map[string]domain.ServiceOrder),
		worksByOrder:    make(map[string][]domain.Work),
		suppliesByOrder: make(map[string][]supplyDomain.Supply),
		onStatusChange:  onStatusChange,
	}
}

func (r *memory_repo) Save(_ context.Context, so *domain.ServiceOrder) error {
	if r.onStatusChange != nil {
		old, exists := r.data[so.ID]
		if !exists {
			r.onStatusChange(so.ID, nil, so.Status)
		} else if old.Status != so.Status {
			prev := old.Status
			r.onStatusChange(so.ID, &prev, so.Status)
		}
	}
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
	s := supplyDomain.Supply{
		ID:            supplyID,
		UnitPrice:     unitPrice,
		StockQuantity: amount,
		Name:          "",
		Description:   "",
	}
	r.suppliesByOrder[serviceOrderID] = append(r.suppliesByOrder[serviceOrderID], s)
	return nil
}

func (r *memory_repo) ListSuppliesByServiceOrderID(_ context.Context, serviceOrderID string) ([]supplyDomain.Supply, error) {
	supplies := r.suppliesByOrder[serviceOrderID]
	if supplies == nil {
		return []supplyDomain.Supply{}, nil
	}
	out := make([]supplyDomain.Supply, len(supplies))
	copy(out, supplies)
	return out, nil
}

func (r *memory_repo) RemoveSupplyLink(_ context.Context, serviceOrderID string, supplyID string) (int, error) {
	list := r.suppliesByOrder[serviceOrderID]
	idx := slices.IndexFunc(list, func(w supplyDomain.Supply) bool { return w.ID == supplyID })
	if idx < 0 {
		return 0, domain.ErrServiceOrderSupplyNotFound
	}
	qty := list[idx].StockQuantity
	r.suppliesByOrder[serviceOrderID] = slices.Delete(list, idx, idx+1)
	return qty, nil
}

func (r *memory_repo) FindByID(_ context.Context, id string) (domain.ServiceOrder, error) {
	so, ok := r.data[id]
	if !ok {
		return so, domain.ErrServiceOrderNotFound
	}
	return so, nil
}

func (r *memory_repo) Count(_ context.Context, params *domain.ServiceOrderFilterParams) (int64, error) {
	var total int64
	for _, so := range r.data {
		if params.Status != "" && so.Status.String() != params.Status {
			continue
		}
		if params.CustomerID != "" && (so.Customer == nil || so.Customer.ID != params.CustomerID) {
			continue
		}
		if params.VehicleID != "" && (so.Vehicle == nil || so.Vehicle.ID != params.VehicleID) {
			continue
		}
		total++
	}

	return total, nil
}

func (r *memory_repo) Search(_ context.Context, params *domain.ServiceOrderFilterParams) ([]domain.ServiceOrder, error) {
	items := make([]domain.ServiceOrder, 0, params.Limit)
	filtered := make([]domain.ServiceOrder, 0)
	for _, so := range r.data {
		if params.Status != "" && so.Status.String() != params.Status {
			continue
		}
		if params.CustomerID != "" && (so.Customer == nil || so.Customer.ID != params.CustomerID) {
			continue
		}
		if params.VehicleID != "" && (so.Vehicle == nil || so.Vehicle.ID != params.VehicleID) {
			continue
		}
		filtered = append(filtered, so)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ID < filtered[j].ID
	})

	start := int(params.Offset)
	if start >= len(filtered) {
		return items, nil
	}

	for i := start; i < len(filtered) && len(items) < int(params.Limit); i++ {
		items = append(items, filtered[i])
	}

	return items, nil
}

func (r *memory_repo) AverageExecutionTimeInHours(_ context.Context, _ []string) ([]domain.WorkExecutionTime, error) {
	return []domain.WorkExecutionTime{}, nil
}
