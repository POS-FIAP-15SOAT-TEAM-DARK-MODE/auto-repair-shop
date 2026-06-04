package repository

import (
	"context"
	"slices"
	"sort"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work/domain"
	"github.com/shopspring/decimal"
)

type soMemoryRepo struct {
	data            map[string]domain.ServiceOrder
	worksByOrder    map[string][]workDomain.Work
	suppliesByOrder map[string][]supplyDomain.Supply
	onStatusChange  func(soID string, prev *domain.SERVICE_ORDER_STATUS, newStatus domain.SERVICE_ORDER_STATUS)
}

func NewSOMemory() domain.ServiceOrderRepository {
	return &soMemoryRepo{
		data:            make(map[string]domain.ServiceOrder),
		worksByOrder:    make(map[string][]workDomain.Work),
		suppliesByOrder: make(map[string][]supplyDomain.Supply),
	}
}

func NewSOMemoryWithHistory(onStatusChange func(soID string, prev *domain.SERVICE_ORDER_STATUS, newStatus domain.SERVICE_ORDER_STATUS)) domain.ServiceOrderRepository {
	return &soMemoryRepo{
		data:            make(map[string]domain.ServiceOrder),
		worksByOrder:    make(map[string][]workDomain.Work),
		suppliesByOrder: make(map[string][]supplyDomain.Supply),
		onStatusChange:  onStatusChange,
	}
}

func (r *soMemoryRepo) Save(_ context.Context, so *domain.ServiceOrder) error {
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

func (r *soMemoryRepo) ExistsByID(_ context.Context, id string) (bool, domain.SERVICE_ORDER_STATUS, error) {
	v, ok := r.data[id]
	return ok, v.Status, nil
}

func (r *soMemoryRepo) FindByID(_ context.Context, id string) (domain.ServiceOrder, error) {
	so, ok := r.data[id]
	if !ok {
		return so, domain.ErrServiceOrderNotFound
	}
	return so, nil
}

func (r *soMemoryRepo) Count(_ context.Context, params *domain.ServiceOrderFilterParams) (int64, error) {
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

func (r *soMemoryRepo) Search(_ context.Context, params *domain.ServiceOrderFilterParams) ([]domain.ServiceOrder, error) {
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

	items := make([]domain.ServiceOrder, 0, params.Limit)
	start := int(params.Offset)
	if start >= len(filtered) {
		return items, nil
	}
	for i := start; i < len(filtered) && len(items) < int(params.Limit); i++ {
		items = append(items, filtered[i])
	}
	return items, nil
}

func (r *soMemoryRepo) ListWorksByServiceOrderID(_ context.Context, serviceOrderID string) ([]workDomain.Work, error) {
	works := r.worksByOrder[serviceOrderID]
	if works == nil {
		return []workDomain.Work{}, nil
	}
	out := make([]workDomain.Work, len(works))
	copy(out, works)
	return out, nil
}

func (r *soMemoryRepo) AddWorkLink(_ context.Context, serviceOrderID, workID string, unitPrice decimal.Decimal) error {
	for _, w := range r.worksByOrder[serviceOrderID] {
		if w.ID == workID {
			return nil
		}
	}
	r.worksByOrder[serviceOrderID] = append(r.worksByOrder[serviceOrderID], workDomain.Work{
		ID:    workID,
		Price: unitPrice,
	})
	return nil
}

func (r *soMemoryRepo) RemoveWorkLink(_ context.Context, serviceOrderID, workID string) error {
	list := r.worksByOrder[serviceOrderID]
	idx := slices.IndexFunc(list, func(w workDomain.Work) bool { return w.ID == workID })
	if idx < 0 {
		return domain.ErrServiceOrderWorkNotFound
	}
	r.worksByOrder[serviceOrderID] = slices.Delete(list, idx, idx+1)
	return nil
}

func (r *soMemoryRepo) ListSuppliesByServiceOrderID(_ context.Context, serviceOrderID string) ([]supplyDomain.Supply, error) {
	supplies := r.suppliesByOrder[serviceOrderID]
	if supplies == nil {
		return []supplyDomain.Supply{}, nil
	}
	out := make([]supplyDomain.Supply, len(supplies))
	copy(out, supplies)
	return out, nil
}

func (r *soMemoryRepo) AddSupplyLink(_ context.Context, serviceOrderID, supplyID string, amount int, unitPrice decimal.Decimal) error {
	for _, s := range r.suppliesByOrder[serviceOrderID] {
		if s.ID == supplyID {
			return nil
		}
	}
	r.suppliesByOrder[serviceOrderID] = append(r.suppliesByOrder[serviceOrderID], supplyDomain.Supply{
		ID:            supplyID,
		UnitPrice:     unitPrice,
		StockQuantity: amount,
	})
	return nil
}

func (r *soMemoryRepo) RemoveSupplyLink(_ context.Context, serviceOrderID, supplyID string) (int, error) {
	list := r.suppliesByOrder[serviceOrderID]
	idx := slices.IndexFunc(list, func(s supplyDomain.Supply) bool { return s.ID == supplyID })
	if idx < 0 {
		return 0, domain.ErrServiceOrderSupplyNotFound
	}
	qty := list[idx].StockQuantity
	r.suppliesByOrder[serviceOrderID] = slices.Delete(list, idx, idx+1)
	return qty, nil
}

func (r *soMemoryRepo) AverageExecutionTimeInHours(_ context.Context, _ []string) ([]domain.WorkExecutionTime, error) {
	return []domain.WorkExecutionTime{}, nil
}
