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
			return s.userRepo.AssignRole(txCtx, customer.User.ID, domain.CUSTOMER)
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

func (s *service) Update(ctx context.Context, id string, name, email, phone *string) (domain.Customer, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Customer{}, err
	}

	updatedName := customer.User.Name
	if name != nil {
		updatedName = *name
	}
	updatedEmail := customer.User.Email
	if email != nil {
		updatedEmail = *email
	}
	updatedPhone := customer.Phone
	if phone != nil {
		updatedPhone = *phone
	}

	if err := s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.userRepo.Update(txCtx, customer.UserID, updatedName, updatedEmail)
		},
		func(txCtx context.Context) error { return s.customerRepo.Update(txCtx, id, updatedPhone) },
	); err != nil {
		return domain.Customer{}, err
	}

	customer.User.Name = updatedName
	customer.User.Email = updatedEmail
	customer.Phone = updatedPhone
	return customer, nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.uow.Execute(ctx,
		func(txCtx context.Context) error { return s.customerRepo.Delete(txCtx, id) },
		func(txCtx context.Context) error { return s.userRepo.Delete(txCtx, customer.UserID) },
	)
}

func (s *service) GetByUserID(ctx context.Context, id string) (domain.Customer, error) {
	return s.customerRepo.GetByUserID(ctx, id)
}
