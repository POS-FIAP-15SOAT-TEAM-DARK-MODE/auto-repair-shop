package user

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type memory_repo struct {
	data map[string]domain.User
}

func MemoryRepository() domain.UserRepository {
	return &memory_repo{
		data: make(map[string]domain.User),
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

func (r *memory_repo) GetRolesByUserId(_ context.Context, _ string) ([]string, error) {
	return []string{}, nil
}

func (r *memory_repo) AssignRole(_ context.Context, _, _ string) error {
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
