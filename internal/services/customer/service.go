package customer

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

type service struct {
	transactor   domain.Transactor
	userRepo     domain.UserRepository
	customerRepo domain.CustomerRepository
}

func Service(transactor domain.Transactor, userRepo domain.UserRepository, customerRepo domain.CustomerRepository) *service {
	return &service{
		transactor:   transactor,
		userRepo:     userRepo,
		customerRepo: customerRepo,
	}
}

func (s *service) Create(ctx context.Context, customer domain.Customer) error {
	return s.transactor.WithTransaction(ctx, func(ctx context.Context) error {
		if err := s.userRepo.Create(ctx, customer.User); err != nil {
			return err
		}
		return s.customerRepo.Create(ctx, &customer)
	})
}
