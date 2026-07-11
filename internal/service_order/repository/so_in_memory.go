package repository

import (
	"context"
	"slices"
	"sort"
	"sync"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/domain"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/domain"
	"github.com/shopspring/decimal"
)

// statusPriority mirrors the SQL CASE priority used in so_postgres.go:
// IN_PROGRESS(1) → AWAITING_APPROVAL(2) → IN_DIAGNOSIS(3) → RECEIVED(4) → NEW(5) → other(6)
var statusPriority = map[domain.SERVICE_ORDER_STATUS]int{
	domain.SERVICE_ORDER_STATUS_IN_PROGRESS:       1,
	domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL: 2,
	domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS:      3,
	domain.SERVICE_ORDER_STATUS_RECEIVED:          4,
	domain.SERVICE_ORDER_STATUS_NEW:               5,
}

func statusPriorityOf(s domain.SERVICE_ORDER_STATUS) int {
	if p, ok := statusPriority[s]; ok {
		return p
	}
	return 6
}

// terminalStatusSet matches the postgres terminalStatuses slice.
var terminalStatusSet = map[domain.SERVICE_ORDER_STATUS]bool{
	domain.SERVICE_ORDER_STATUS_DELIVERED: true,
	domain.SERVICE_ORDER_STATUS_COMPLETED: true,
	domain.SERVICE_ORDER_STATUS_REJECTED:  true,
	domain.SERVICE_ORDER_STATUS_CANCELLED: true,
}

type soMemoryRepo struct {
	mu              sync.RWMutex
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
	r.mu.Lock()
	defer r.mu.Unlock()
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

// UpdatePricing mutates only the total amount, leaving the status untouched so
// it cannot overwrite a concurrent lifecycle transition.
func (r *soMemoryRepo) UpdatePricing(_ context.Context, serviceOrderID string, totalAmount decimal.Decimal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	so, ok := r.data[serviceOrderID]
	if !ok {
		return domain.ErrServiceOrderNotFound
	}
	so.TotalAmount = totalAmount
	r.data[serviceOrderID] = so
	return nil
}

func (r *soMemoryRepo) ExistsByID(_ context.Context, id string) (bool, domain.SERVICE_ORDER_STATUS, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[id]
	return ok, v.Status, nil
}

func (r *soMemoryRepo) FindByID(_ context.Context, id string) (domain.ServiceOrder, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	so, ok := r.data[id]
	if !ok {
		return so, domain.ErrServiceOrderNotFound
	}
	return so, nil
}

func (r *soMemoryRepo) Count(_ context.Context, params *domain.ServiceOrderFilterParams) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var total int64
	for _, so := range r.data {
		if params.Status == "" {
			if terminalStatusSet[so.Status] {
				continue
			}
		} else if so.Status.String() != params.Status {
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	filtered := make([]domain.ServiceOrder, 0)
	for _, so := range r.data {
		if params.Status == "" {
			if terminalStatusSet[so.Status] {
				continue
			}
		} else if so.Status.String() != params.Status {
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

	if params.SortBy == "status" {
		sort.Slice(filtered, func(i, j int) bool {
			pi := statusPriorityOf(filtered[i].Status)
			pj := statusPriorityOf(filtered[j].Status)
			if pi != pj {
				return pi < pj
			}
			return filtered[i].ID < filtered[j].ID
		})
	} else {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].ID < filtered[j].ID
		})
	}

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
	r.mu.RLock()
	defer r.mu.RUnlock()
	works := r.worksByOrder[serviceOrderID]
	if works == nil {
		return []workDomain.Work{}, nil
	}
	out := make([]workDomain.Work, len(works))
	copy(out, works)
	return out, nil
}

func (r *soMemoryRepo) AddWorkLink(_ context.Context, serviceOrderID, workID string, unitPrice decimal.Decimal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.mu.Lock()
	defer r.mu.Unlock()
	list := r.worksByOrder[serviceOrderID]
	idx := slices.IndexFunc(list, func(w workDomain.Work) bool { return w.ID == workID })
	if idx < 0 {
		return domain.ErrServiceOrderWorkNotFound
	}
	r.worksByOrder[serviceOrderID] = slices.Delete(list, idx, idx+1)
	return nil
}

func (r *soMemoryRepo) ListSuppliesByServiceOrderID(_ context.Context, serviceOrderID string) ([]supplyDomain.Supply, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	supplies := r.suppliesByOrder[serviceOrderID]
	if supplies == nil {
		return []supplyDomain.Supply{}, nil
	}
	out := make([]supplyDomain.Supply, len(supplies))
	copy(out, supplies)
	return out, nil
}

func (r *soMemoryRepo) AddSupplyLink(_ context.Context, serviceOrderID, supplyID string, amount int, unitPrice decimal.Decimal) error {
	r.mu.Lock()
	defer r.mu.Unlock()
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
	r.mu.Lock()
	defer r.mu.Unlock()
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
