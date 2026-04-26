package user

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data  map[string]domain.User
	roles map[string][]domain.Role
}

func MemoryRepository() domain.UserRepository {
	return &memory_repo{
		data:  make(map[string]domain.User),
		roles: make(map[string][]domain.Role),
	}
}

func (r *memory_repo) Create(_ context.Context, u *domain.User) error {
	r.data[u.ID] = *u
	return nil
}

func (r *memory_repo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range r.data {
		if u.Email == email {
			return &u, nil
		}
	}
	return nil, nil
}

func (r *memory_repo) GetRolesByUserId(_ context.Context, userID string) ([]domain.Role, error) {
	return r.roles[userID], nil
}

func (r *memory_repo) AssignRole(_ context.Context, userID string, role domain.Role) error {
	r.roles[userID] = append(r.roles[userID], role)
	return nil
}

func (r *memory_repo) UpdateRole(_ context.Context, userID string, role domain.Role) error {
	r.roles[userID] = []domain.Role{role}
	return nil
}

func (r *memory_repo) Update(_ context.Context, id, name, email string) error {
	u, ok := r.data[id]
	if !ok {
		return nil
	}
	u.Name = name
	u.Email = email
	r.data[id] = u
	return nil
}

func (r *memory_repo) Delete(_ context.Context, id string) error {
	delete(r.data, id)
	return nil
}
