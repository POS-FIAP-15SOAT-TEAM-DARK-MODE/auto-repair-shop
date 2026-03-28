package customer

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type service struct {
	uow          uow.Executor
	userRepo     domain.UserRepository
	customerRepo domain.CustomerRepository
}

func Service(uow uow.Executor, userRepo domain.UserRepository, customerRepo domain.CustomerRepository) *service {
	return &service{
		uow:          uow,
		userRepo:     userRepo,
		customerRepo: customerRepo,
	}
}

func (s *service) Create(ctx context.Context, customer domain.Customer) error {
	return s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.userRepo.Create(txCtx, customer.User)
		},
		func(txCtx context.Context) error {
			return s.customerRepo.Create(txCtx, &customer)
		},
	)
}

func (s *service) GetByID(ctx context.Context, id string) (domain.Customer, error) {
	return s.customerRepo.GetByID(ctx, id)
}

func (s *service) GetByDocument(ctx context.Context, rawDocument string) (domain.Customer, error) {
	_, document, err := domain.ParseDocument(rawDocument)
	if err != nil {
		return domain.Customer{}, err
	}
	return s.customerRepo.GetByDocument(ctx, document)
}
