package service_order

import (
	"context"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_ReturnsServiceOrderWithDefaults(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, nil, vehicleId, nil)
	s := Service(executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock)
	ctx := context.Background()

	so, err := s.Create(ctx, customerId, vehicleId)
	assert.NoError(t, err)
	assert.Equal(t, domain.SERVICE_ORDER_STATUS_NEW, so.Status)
	assert.NotEmpty(t, so.ID)
	assert.Empty(t, so.Services)
	assert.Empty(t, so.Supplies)
	assert.Zero(t, so.TotalAmount.InexactFloat64())
}

func TestService_Create_VerifyCustomerID(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, domain.ErrCustomerNotFound, vehicleId, nil)
	s := Service(executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
}

func TestService_Create_VerifyVehicleID(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, nil, vehicleId, domain.ErrVehicleNotFound)
	s := Service(executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrVehicleNotFound)
}

func TestService_Create_InsertSO_RepositoryFailure(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, domain.ErrInfraConflict, c.ID, nil, vehicleId, nil)
	s := Service(executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock)
	ctx := context.Background()

	so, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
	assert.NotNil(t, so)
}

func setupCreateTestMock(t *testing.T,
	err error,
	customerID string,
	customerError error,
	vehicleID string,
	vehicleError error) (uow.Executor,
	domain.ServiceOrderRepository,
	domain.WorkRepository,
	domain.SupplyRepository,
	domain.CustomerService,
	domain.VehicleService) {
	executor := uowmocks.NewExecutor(t)
	mockRepo := mocks.NewServiceOrderRepository(t)
	mockWorkRepo := mocks.NewWorkRepository(t)
	mockSupplyRepo := mocks.NewSupplyRepository(t)
	mockCustomerSvc := mocks.NewCustomerService(t)
	mockVehicleSvc := mocks.NewVehicleService(t)

	mockCustomerSvc.EXPECT().GetByID(mock.Anything, customerID).Return(domain.Customer{}, customerError)

	if customerError == nil {
		mockVehicleSvc.EXPECT().FindByID(mock.Anything, vehicleID).Return(&domain.Vehicle{}, vehicleError)
		if vehicleError == nil {
			executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			mockRepo.EXPECT().Save(mock.Anything, mock.Anything).Return(err)
		}
	}

	return executor, mockRepo, mockWorkRepo, mockSupplyRepo, mockCustomerSvc, mockVehicleSvc
}

func runAllSteps() func(context.Context, ...uow.Step) error {
	return func(ctx context.Context, steps ...uow.Step) error {
		for _, step := range steps {
			if err := step(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// --- AddSupplies tests ---

func TestAddSupplies_EmptyServiceOrderID(t *testing.T) {
	s := newBasicService(t)
	err := s.AddSupplies(context.Background(), "  ", []domain.AddSupply{{ID: "sup-1", Amount: 1}})
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestAddSupplies_EmptyList(t *testing.T) {
	s := newBasicService(t)
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{})
	assert.ErrorIs(t, err, domain.ErrEmptyServicesList)
}

func TestAddSupplies_InvalidAmount(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	supplyRepo := mocks.NewSupplyRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 0}})
	assert.ErrorIs(t, err, domain.ErrInvalidSupplyAmount)
}

func TestAddSupplies_OutOfStock_ZeroStock(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	supplyRepo := mocks.NewSupplyRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)
	supplyRepo.EXPECT().FindById(mock.Anything, "sup-1").Return(domain.Supply{ID: "sup-1", StockQuantity: 0}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 1}})
	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
}

func TestAddSupplies_OutOfStock_InsufficientQuantity(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	supplyRepo := mocks.NewSupplyRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)
	supplyRepo.EXPECT().FindById(mock.Anything, "sup-1").Return(domain.Supply{ID: "sup-1", StockQuantity: 2}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 5}})
	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
}

func TestAddSupplies_ServiceOrderNotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 1}})
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotInDiagnosis)
}

func TestAddSupplies_Success_DecrementsStock(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	supplyRepo := mocks.NewSupplyRepository(t)

	supply := domain.Supply{ID: "sup-1", StockQuantity: 10}

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)
	supplyRepo.EXPECT().FindById(mock.Anything, "sup-1").Return(supply, nil)
	repo.EXPECT().AddSupplyLink(mock.Anything, "so-1", "sup-1", 3, supply.UnitPrice).Return(nil)
	supplyRepo.EXPECT().DecrementStock(mock.Anything, "sup-1", 3).Return(nil)
	expectReviewOSPricing(executor, repo, "so-1")

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 3}})
	assert.NoError(t, err)
}

func TestAddSupplies_DecrementStock_ConcurrentFailure(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	supplyRepo := mocks.NewSupplyRepository(t)

	supply := domain.Supply{ID: "sup-1", StockQuantity: 5}

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)
	supplyRepo.EXPECT().FindById(mock.Anything, "sup-1").Return(supply, nil)
	repo.EXPECT().AddSupplyLink(mock.Anything, "so-1", "sup-1", 5, supply.UnitPrice).Return(nil)
	// Simulates concurrent depletion: stock was taken by another transaction.
	supplyRepo.EXPECT().DecrementStock(mock.Anything, "sup-1", 5).Return(domain.ErrSupplyOutOfStock)

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 5}})
	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
}

// --- RemoveSupply tests ---

func TestRemoveSupply_Success_RestoresStock(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	supplyRepo := mocks.NewSupplyRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)
	repo.EXPECT().RemoveSupplyLink(mock.Anything, "so-1", "sup-1").Return(3, nil)
	supplyRepo.EXPECT().RestoreStock(mock.Anything, "sup-1", 3).Return(nil)
	expectReviewOSPricing(executor, repo, "so-1")

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveSupply(context.Background(), "so-1", "sup-1")
	assert.NoError(t, err)
}

func TestRemoveSupply_SupplyNotLinked(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	supplyRepo := mocks.NewSupplyRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)
	repo.EXPECT().RemoveSupplyLink(mock.Anything, "so-1", "sup-1").Return(0, domain.ErrServiceOrderSupplyNotFound)

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveSupply(context.Background(), "so-1", "sup-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderSupplyNotFound)
}

func TestRemoveSupply_ServiceOrderNotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveSupply(context.Background(), "so-1", "sup-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotInDiagnosis)
}

// --- Finish tests ---

func TestFinish_EmptyServiceOrderID(t *testing.T) {
	s := newBasicService(t)
	err := s.Finish(context.Background(), "   ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestFinish_RepositoryFindFailure(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestFinish_ServiceOrderNotInProgress(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotInProgress)
}

func TestFinish_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_PROGRESS}, nil)
	repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(so *domain.ServiceOrder) bool {
		return so.ID == "so-1" && so.Status == domain.SERVICE_ORDER_STATUS_COMPLETED
	})).Return(nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.NoError(t, err)
}

func TestFinish_SaveFailure(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_PROGRESS}, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(assert.AnError)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.ErrorIs(t, err, assert.AnError)
}

func newBasicService(t *testing.T) *svc {
	t.Helper()
	return Service(
		uowmocks.NewExecutor(t),
		mocks.NewServiceOrderRepository(t),
		mocks.NewWorkRepository(t),
		mocks.NewSupplyRepository(t),
		mocks.NewCustomerService(t),
		mocks.NewVehicleService(t),
	)
}

func expectReviewOSPricing(executor *uowmocks.Executor, repo *mocks.ServiceOrderRepository, serviceOrderID string) {
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps()).Maybe()
	repo.EXPECT().FindByID(mock.Anything, serviceOrderID).Return(domain.ServiceOrder{ID: serviceOrderID}, nil).Maybe()
	repo.EXPECT().ListWorksByServiceOrderID(mock.Anything, serviceOrderID).Return([]domain.Work{}, nil).Maybe()
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, serviceOrderID).Return([]domain.Supply{}, nil).Maybe()
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil).Maybe()
}
