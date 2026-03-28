package service_order

import (
	"context"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
)

type svc struct {
	uow  uow.Executor
	repo domain.ServiceOrderRepository
}

func Service(uow uow.Executor, repo domain.ServiceOrderRepository) *svc {
	return &svc{uow, repo}
}

func (s *svc) Create(ctx context.Context, customerId string, vehicleId string) (domain.ServiceOrder, error) {
	// TODO: Verify customerID
	// TODO: Verify vehicleID
	so := domain.NewServiceOrder(nil, nil)
	err := s.uow.Execute(ctx, func(ctx context.Context) error {
		return s.repo.Save(ctx, so)
	})
	return *so, err
}
