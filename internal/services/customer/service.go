package customer

import (
	"context"

	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
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
	err := s.uow.Execute(ctx,
		func(txCtx context.Context) error {
			return s.userRepo.Create(txCtx, customer.User)
		},
		func(txCtx context.Context) error {
			return s.customerRepo.Create(txCtx, &customer)
		},
	)

	if err != nil {
		return err
	}

	logger.Of(ctx).Info("customer created",
		zap.String("operation", "create_customer"),
		zap.String("entity_id", customer.ID),
	)
	return nil
}
