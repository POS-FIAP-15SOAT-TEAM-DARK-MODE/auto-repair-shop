package vehicle

import (
	"context"
	"slices"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memoryRepo struct {
	data   map[string]domain.Vehicle
	plates map[string]string // normalized plate → vehicle ID
}

func MemoryRepository() domain.VehicleRepository {
	return &memoryRepo{
		data:   make(map[string]domain.Vehicle),
		plates: make(map[string]string),
	}
}

func (r *memoryRepo) Save(_ context.Context, v *domain.Vehicle) error {
	norm := domain.NormalizeLicensePlate(v.LicensePlate)
	if existingID, exists := r.plates[norm]; exists && existingID != v.ID {
		return domain.ErrDataConflict
	}
	r.data[v.ID] = *v
	r.plates[norm] = v.ID
	return nil
}

func (r *memoryRepo) Find(_ context.Context, params domain.FindVehicleParams) (*domain.Vehicle, error) {
	if params.ID != "" {
		v, ok := r.data[params.ID]
		if !ok {
			return nil, domain.ErrVehicleNotFound
		}
		return &v, nil
	}
	if params.LicensePlate != "" {
		norm := domain.NormalizeLicensePlate(params.LicensePlate)
		id, ok := r.plates[norm]
		if !ok {
			return nil, domain.ErrVehicleNotFound
		}
		v := r.data[id]
		return &v, nil
	}
	return nil, domain.ErrVehicleNotFound
}

func (r *memoryRepo) Update(_ context.Context, v *domain.Vehicle) error {
	old, ok := r.data[v.ID]
	if !ok {
		return domain.ErrVehicleNotFound
	}
	delete(r.plates, domain.NormalizeLicensePlate(old.LicensePlate))
	r.data[v.ID] = *v
	r.plates[domain.NormalizeLicensePlate(v.LicensePlate)] = v.ID
	return nil
}

func (r *memoryRepo) Delete(_ context.Context, id string) error {
	v, ok := r.data[id]
	if ok {
		delete(r.plates, domain.NormalizeLicensePlate(v.LicensePlate))
		delete(r.data, id)
	}
	return nil
}

func (r *memoryRepo) Search(_ context.Context, params *domain.SearchVehicleParams) ([]domain.Vehicle, error) {
	var filtered []domain.Vehicle
	for _, v := range r.data {
		if params.CustomerId != "" && v.CustomerId != params.CustomerId {
			continue
		}
		filtered = append(filtered, v)
	}
	slices.SortFunc(filtered, func(a, b domain.Vehicle) int {
		if a.ID < b.ID {
			return -1
		}
		if a.ID > b.ID {
			return 1
		}
		return 0
	})
	start := int(params.Offset)
	if start >= len(filtered) {
		return []domain.Vehicle{}, nil
	}
	end := start + int(params.Limit)
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], nil
}

func (r *memoryRepo) Count(_ context.Context, params *domain.SearchVehicleParams) (int64, error) {
	var total int64
	for _, v := range r.data {
		if params.CustomerId != "" && v.CustomerId != params.CustomerId {
			continue
		}
		total++
	}
	return total, nil
}
