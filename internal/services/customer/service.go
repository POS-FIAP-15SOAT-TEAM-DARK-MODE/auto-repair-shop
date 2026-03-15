package customer

import (
	"context"

	"go.uber.org/zap"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
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
	err := s.transactor.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := s.userRepo.Create(txCtx, customer.User); err != nil {
			return err
		}
		return s.customerRepo.Create(txCtx, &customer)
	})

	if err != nil {
		return err
	}

	logger.Info("customer created",
		zap.String("operation", "create_customer"),
		zap.String("entity_id", customer.ID),
	)
	return nil
}
