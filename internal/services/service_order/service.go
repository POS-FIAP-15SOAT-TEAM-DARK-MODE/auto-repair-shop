package service_order

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type svc struct {
	uow             uow.Executor
	repo            domain.ServiceOrderRepository
	customerService domain.CustomerService
}

func Service(uow uow.Executor, repo domain.ServiceOrderRepository, customerService domain.CustomerService) *svc {
	return &svc{uow, repo, customerService}
}

func (s *svc) Create(ctx context.Context, customerId string, vehicleId string) (domain.ServiceOrder, error) {
	customer, err := s.customerService.GetByID(ctx, customerId)
	if err != nil {
		return domain.ServiceOrder{}, err
	}

	// TODO: Verify vehicleID
	so := domain.NewServiceOrder(&customer, nil)
	if err = s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Save(ctx, so)
	}); err != nil {
		return domain.ServiceOrder{}, err
	}

	return *so, nil
}
