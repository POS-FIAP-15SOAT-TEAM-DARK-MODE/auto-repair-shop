package service_order

import (
	"context"
	"database/sql"

	customerAdapters "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/adapters"
	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/domain"
	customerInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/interfaces"
	customerRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/customer/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/id"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/uow"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/interfaces"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/service_order/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply"
	supplyRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/supply/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle"
	vehicleRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/repository"
	workInterfaces "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work/interfaces"
	workRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/work/repository"
)

func NewHTTPController(db *sql.DB) interfaces.ServiceOrderHTTPController {
	unitOfWork := uow.NewTransactionalUoW(db)
	idGen := id.NewIDGenerator()

	soRepo := repository.NewSOPostgres()
	wsoHistoryRepo := repository.NewWSOHistoryPostgres()

	custRepo := customerRepo.NewPostgres()
	customerFinder := newCustomerFinder(custRepo)

	vehicleRepository := vehicleRepo.NewPostgres()
	vehicleSvc := vehicle.NewService(unitOfWork, vehicleRepository, idGen)

	supplyRepository := supplyRepo.NewPostgres()
	supplySvc := supply.NewService(unitOfWork, supplyRepository, idGen)

	var wRepo workInterfaces.WorkRepository = workRepo.NewPostgres()

	svc := NewService(unitOfWork, soRepo, wRepo, supplySvc, wsoHistoryRepo, customerFinder, vehicleSvc)
	return NewController(svc)
}

// customerFinderAdapter bridges the SO module's CustomerFinder interface
// with the customer repository, returning domain objects the SO service needs.
type customerFinderAdapter struct {
	repo customerInterfaces.CustomerRepository
}

func newCustomerFinder(repo customerInterfaces.CustomerRepository) CustomerFinder {
	return &customerFinderAdapter{repo: repo}
}

func (f *customerFinderAdapter) GetByID(ctx context.Context, id string) (customerDomain.Customer, error) {
	items, err := f.repo.List(ctx, customerAdapters.ListCustomerParams{ID: id})
	if err != nil {
		return customerDomain.Customer{}, err
	}
	if len(items) == 0 {
		return customerDomain.Customer{}, customerDomain.ErrCustomerNotFound
	}
	return items[0], nil
}

func (f *customerFinderAdapter) GetByUserID(ctx context.Context, userID string) (customerDomain.Customer, error) {
	items, err := f.repo.List(ctx, customerAdapters.ListCustomerParams{UserID: userID})
	if err != nil {
		return customerDomain.Customer{}, err
	}
	if len(items) == 0 {
		return customerDomain.Customer{}, customerDomain.ErrCustomerNotFound
	}
	return items[0], nil
}
