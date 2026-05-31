package service_order

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_ReturnsServiceOrderWithDefaults(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, nil, vehicleId, nil)
	s := Service(executor, mockRepo, workRepo, supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerMock, vehicleMock)
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
	s := Service(executor, mockRepo, workRepo, supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
}

func TestService_Create_VerifyVehicleID(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, nil, vehicleId, domain.ErrVehicleNotFound)
	s := Service(executor, mockRepo, workRepo, supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrVehicleNotFound)
}

func TestService_Create_InsertSO_RepositoryFailure(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, domain.ErrInfraConflict, c.ID, nil, vehicleId, nil)
	s := Service(executor, mockRepo, workRepo, supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerMock, vehicleMock)
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 5}})
	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
}

func TestAddSupplies_ServiceOrderNotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveSupply(context.Background(), "so-1", "sup-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderSupplyNotFound)
}

func TestRemoveSupply_ServiceOrderNotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestFinish_ServiceOrderNotInProgress(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.NoError(t, err)
}

func TestFinish_SaveFailure(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_PROGRESS}, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(assert.AnError)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.ErrorIs(t, err, assert.AnError)
}

// --- SendToDiagnosis tests ---

func TestSendToDiagnosis_EmptyServiceOrderID(t *testing.T) {
	s := newBasicService(t)
	err := s.SendToDiagnosis(context.Background(), "   ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSendToDiagnosis_ServiceOrderNotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToDiagnosis(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestSendToDiagnosis_ServiceOrderNotInReceived(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToDiagnosis(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotInReceived)
}

func TestSendToDiagnosis_SaveFailure(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_RECEIVED}, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(domain.ErrInfraConflict)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToDiagnosis(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

func TestSendToDiagnosis_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_RECEIVED}, nil)
	repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(so *domain.ServiceOrder) bool {
		return so.Status == domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS
	})).Return(nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToDiagnosis(context.Background(), "so-1")
	assert.NoError(t, err)
}

func newBasicService(t *testing.T) *svc {
	t.Helper()
	return Service(
		uowmocks.NewExecutor(t),
		mocks.NewServiceOrderRepository(t),
		mocks.NewWorkRepository(t),
		mocks.NewSupplyRepository(t),
		mocks.NewServiceOrderHistoryRepository(t),
		mocks.NewWorkServiceOrderHistoryRepository(t),
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

// --- AverageExecutionTime tests ---

func TestAverageExecutionTime_ReturnsAllWorks(t *testing.T) {
	expected := []domain.WorkExecutionTime{
		{WorkID: "w-1", WorkName: "Troca de óleo", AverageHours: 1.5},
		{WorkID: "w-2", WorkName: "Revisão de freios", AverageHours: 3.0},
	}
	repo := mocks.NewServiceOrderRepository(t)
	repo.EXPECT().AverageExecutionTimeInHours(mock.Anything, []string{}).Return(expected, nil)

	s := newBasicServiceWith(t, repo)
	result, err := s.AverageExecutionTime(context.Background(), []string{})
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestAverageExecutionTime_FilterByWorkIDs(t *testing.T) {
	expected := []domain.WorkExecutionTime{
		{WorkID: "w-1", WorkName: "Troca de óleo", AverageHours: 1.5},
	}
	repo := mocks.NewServiceOrderRepository(t)
	repo.EXPECT().AverageExecutionTimeInHours(mock.Anything, []string{"w-1"}).Return(expected, nil)

	s := newBasicServiceWith(t, repo)
	result, err := s.AverageExecutionTime(context.Background(), []string{"w-1"})
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestAverageExecutionTime_ReturnsEmptyWhenNoCompletedWorks(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	repo.EXPECT().AverageExecutionTimeInHours(mock.Anything, []string{}).Return([]domain.WorkExecutionTime{}, nil)

	s := newBasicServiceWith(t, repo)
	result, err := s.AverageExecutionTime(context.Background(), []string{})
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestAverageExecutionTime_PropagatesRepositoryError(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	repo.EXPECT().AverageExecutionTimeInHours(mock.Anything, []string{}).Return(nil, domain.ErrInfraConflict)

	s := newBasicServiceWith(t, repo)
	_, err := s.AverageExecutionTime(context.Background(), []string{})
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

// --- Cancel tests ---

func TestCancel_CancelableStatus_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	so := domain.ServiceOrder{
		ID:       "so-1",
		Status:   domain.SERVICE_ORDER_STATUS_NEW,
		Customer: &domain.Customer{ID: "cust-1"},
	}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil)

	s := newBasicServiceWith(t, repo)
	s.uow = executor
	err := s.Cancel(context.Background(), "so-1")
	assert.NoError(t, err)
}

func TestCancel_NonCancelableStatus_ReturnsError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	so := domain.ServiceOrder{
		ID:       "so-1",
		Status:   domain.SERVICE_ORDER_STATUS_DELIVERED,
		Customer: &domain.Customer{ID: "cust-1"},
	}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)

	s := newBasicServiceWith(t, repo)
	s.uow = executor
	err := s.Cancel(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotCancelable)
}

func TestCancel_EmptyID_ReturnsError(t *testing.T) {
	s := newBasicService(t)
	err := s.Cancel(context.Background(), "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func newBasicServiceWith(t *testing.T, repo domain.ServiceOrderRepository) *svc {
	t.Helper()
	return Service(
		uowmocks.NewExecutor(t),
		repo,
		mocks.NewWorkRepository(t),
		mocks.NewSupplyRepository(t),
		mocks.NewServiceOrderHistoryRepository(t),
		mocks.NewWorkServiceOrderHistoryRepository(t),
		mocks.NewCustomerService(t),
		mocks.NewVehicleService(t),
	)
}

func TestService_GetStatus_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	serviceOrderID := "123"
	expectedStatus := domain.SERVICE_ORDER_STATUS("NEW")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	repo.
		EXPECT().
		ExistsByID(ctx, serviceOrderID).
		Return(true, expectedStatus, nil)

	service := Service(exec, repo, nil, nil, nil, nil, nil, nil)

	result, err := service.GetStatus(ctx, serviceOrderID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expectedStatus {
		t.Fatalf("expected status %v, got %v", expectedStatus, result)
	}
}

func TestService_GetStatus_InvalidID(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	service := Service(exec, repo, nil, nil, nil, nil, nil, nil)

	result, err := service.GetStatus(ctx, "   ")

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrInvalidServiceOrderId) {
		t.Fatalf("expected ErrInvalidServiceOrderId, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}

	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "ExistsByID", mock.Anything, mock.Anything)
}

func TestService_GetStatus_RepositoryError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	serviceOrderID := "123"
	expectedErr := errors.New("repo error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	repo.
		EXPECT().
		ExistsByID(ctx, serviceOrderID).
		Return(false, "", expectedErr)

	service := Service(exec, repo, nil, nil, nil, nil, nil, nil)

	result, err := service.GetStatus(ctx, serviceOrderID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestService_GetStatus_NotFound(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	serviceOrderID := "123"

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	repo.
		EXPECT().
		ExistsByID(ctx, serviceOrderID).
		Return(false, "", nil)

	service := Service(exec, repo, nil, nil, nil, nil, nil, nil)

	result, err := service.GetStatus(ctx, serviceOrderID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrServiceOrderNotFound) {
		t.Fatalf("expected ErrServiceOrderNotFound, got %v", err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}
}

func TestService_GetStatus_UowError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	serviceOrderID := "123"
	expectedErr := errors.New("uow error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	service := Service(exec, repo, nil, nil, nil, nil, nil, nil)

	result, err := service.GetStatus(ctx, serviceOrderID)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if result != "" {
		t.Fatalf("expected empty result, got %v", result)
	}

	repo.AssertNotCalled(t, "ExistsByID", mock.Anything, mock.Anything)
}

func newWorkSOHistoryServiceWith(t *testing.T, workSOHistoryRepo *mocks.WorkServiceOrderHistoryRepository) *svc {
	t.Helper()
	executor := uowmocks.NewExecutor(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps()).Maybe()
	return Service(
		executor,
		mocks.NewServiceOrderRepository(t),
		mocks.NewWorkRepository(t),
		mocks.NewSupplyRepository(t),
		mocks.NewServiceOrderHistoryRepository(t),
		workSOHistoryRepo,
		mocks.NewCustomerService(t),
		mocks.NewVehicleService(t),
	)
}

// --- NextWork tests ---

func TestNextWork_EmptyServiceOrderID(t *testing.T) {
	s := newBasicService(t)
	err := s.NextWork(context.Background(), "  ", "work-1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestNextWork_EmptyWorkID(t *testing.T) {
	s := newBasicService(t)
	err := s.NextWork(context.Background(), "so-1", "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

func TestNextWork_NoHistory_ReturnsNotFound(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return(nil, nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.NextWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrWorkServiceOrderNotFound)
}

func TestNextWork_FromAwaitingStart_InsertsInProgress(t *testing.T) {
	prev := domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: prev}}, nil)
	workSOHistoryRepo.EXPECT().Insert(mock.Anything, "so-1", "work-1", &prev, domain.WORK_SERVICE_ORDER_STATUS_IN_PROGRESS).Return(nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.NextWork(context.Background(), "so-1", "work-1")
	assert.NoError(t, err)
}

func TestNextWork_FromInProgress_InsertsCompleted(t *testing.T) {
	prev := domain.WORK_SERVICE_ORDER_STATUS_IN_PROGRESS
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: prev}}, nil)
	workSOHistoryRepo.EXPECT().Insert(mock.Anything, "so-1", "work-1", &prev, domain.WORK_SERVICE_ORDER_STATUS_COMPLETED).Return(nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.NextWork(context.Background(), "so-1", "work-1")
	assert.NoError(t, err)
}

func TestNextWork_AlreadyCompleted_ReturnsError(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: domain.WORK_SERVICE_ORDER_STATUS_COMPLETED}}, nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.NextWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrWorkServiceOrderAlreadyCompleted)
}

func TestNextWork_AlreadyCancelled_ReturnsError(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: domain.WORK_SERVICE_ORDER_STATUS_CANCELLED}}, nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.NextWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrWorkServiceOrderAlreadyCancelled)
}

func TestNextWork_SearchError_Propagates(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return(nil, domain.ErrInfraConflict)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.NextWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

// --- CancelWork tests ---

func TestCancelWork_EmptyServiceOrderID(t *testing.T) {
	s := newBasicService(t)
	err := s.CancelWork(context.Background(), "  ", "work-1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestCancelWork_EmptyWorkID(t *testing.T) {
	s := newBasicService(t)
	err := s.CancelWork(context.Background(), "so-1", "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

func TestCancelWork_NoHistory_ReturnsNotFound(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return(nil, nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.CancelWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrWorkServiceOrderNotFound)
}

func TestCancelWork_FromAwaitingStart_Success(t *testing.T) {
	prev := domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: prev}}, nil)
	workSOHistoryRepo.EXPECT().Insert(mock.Anything, "so-1", "work-1", &prev, domain.WORK_SERVICE_ORDER_STATUS_CANCELLED).Return(nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.CancelWork(context.Background(), "so-1", "work-1")
	assert.NoError(t, err)
}

func TestCancelWork_FromInProgress_Success(t *testing.T) {
	prev := domain.WORK_SERVICE_ORDER_STATUS_IN_PROGRESS
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: prev}}, nil)
	workSOHistoryRepo.EXPECT().Insert(mock.Anything, "so-1", "work-1", &prev, domain.WORK_SERVICE_ORDER_STATUS_CANCELLED).Return(nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.CancelWork(context.Background(), "so-1", "work-1")
	assert.NoError(t, err)
}

func TestCancelWork_AlreadyCompleted_ReturnsError(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: domain.WORK_SERVICE_ORDER_STATUS_COMPLETED}}, nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.CancelWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrWorkServiceOrderAlreadyCompleted)
}

func TestCancelWork_AlreadyCancelled_ReturnsError(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return([]domain.WorkServiceOrderHistory{{Status: domain.WORK_SERVICE_ORDER_STATUS_CANCELLED}}, nil)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.CancelWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrWorkServiceOrderAlreadyCancelled)
}

func TestCancelWork_SearchError_Propagates(t *testing.T) {
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	workSOHistoryRepo.EXPECT().Search(mock.Anything, domain.SearchWorkSOHistoryParams{
		ServiceOrderID: "so-1", WorkID: "work-1",
	}).Return(nil, domain.ErrInfraConflict)

	s := newWorkSOHistoryServiceWith(t, workSOHistoryRepo)
	err := s.CancelWork(context.Background(), "so-1", "work-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

// --- List tests ---

func TestList_Success(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	params := &domain.ServiceOrderFilterParams{Page: 1, PageSize: 10}
	items := []domain.ServiceOrder{{ID: "so-1"}, {ID: "so-2"}}
	repo.EXPECT().Count(mock.Anything, params).Return(int64(2), nil)
	repo.EXPECT().Search(mock.Anything, params).Return(items, nil)

	s := newBasicServiceWith(t, repo)
	result, err := s.List(context.Background(), params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), result.TotalItems)
	assert.Len(t, result.Items, 2)
}

func TestList_RepoCountError(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	params := &domain.ServiceOrderFilterParams{Page: 1, PageSize: 10}
	repo.EXPECT().Count(mock.Anything, params).Return(int64(0), domain.ErrInfraConflict).Maybe()
	repo.EXPECT().Search(mock.Anything, params).Return(nil, nil).Maybe()

	s := newBasicServiceWith(t, repo)
	_, err := s.List(context.Background(), params)
	assert.Error(t, err)
}

func TestList_RepoSearchError(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	params := &domain.ServiceOrderFilterParams{Page: 1, PageSize: 10}
	repo.EXPECT().Count(mock.Anything, params).Return(int64(0), nil).Maybe()
	repo.EXPECT().Search(mock.Anything, params).Return(nil, domain.ErrInfraConflict).Maybe()

	s := newBasicServiceWith(t, repo)
	_, err := s.List(context.Background(), params)
	assert.Error(t, err)
}

// --- ListWorks tests ---

func TestListWorks_EmptyID(t *testing.T) {
	s := newBasicService(t)
	_, err := s.ListWorks(context.Background(), "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestListWorks_NotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	_, err := s.ListWorks(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestListWorks_ExistsByIDError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), domain.ErrInfraConflict)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	_, err := s.ListWorks(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

func TestListWorks_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	works := []domain.Work{{ID: "w-1"}, {ID: "w-2"}}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)
	repo.EXPECT().ListWorksByServiceOrderID(mock.Anything, "so-1").Return(works, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	got, err := s.ListWorks(context.Background(), "so-1")
	assert.NoError(t, err)
	assert.Equal(t, works, got)
}

// --- AddWorks tests ---

func TestAddWorks_EmptyID(t *testing.T) {
	s := newBasicService(t)
	err := s.AddWorks(context.Background(), "  ", []string{"w-1"})
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestAddWorks_EmptyList(t *testing.T) {
	s := newBasicService(t)
	err := s.AddWorks(context.Background(), "so-1", []string{})
	assert.ErrorIs(t, err, domain.ErrEmptyServicesList)
}

func TestAddWorks_SONotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddWorks(context.Background(), "so-1", []string{"w-1"})
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestAddWorks_SONotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_RECEIVED, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddWorks(context.Background(), "so-1", []string{"w-1"})
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotNew)
}

func TestAddWorks_BlankWorkID(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddWorks(context.Background(), "so-1", []string{"  "})
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

func TestAddWorks_WorkNotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	workRepo := mocks.NewWorkRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)
	workRepo.EXPECT().FindByID(mock.Anything, "w-1").Return(domain.Work{}, domain.ErrWorkNotFound)

	s := Service(executor, repo, workRepo, mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddWorks(context.Background(), "so-1", []string{"w-1"})
	assert.ErrorIs(t, err, domain.ErrWorkNotFound)
}

func TestAddWorks_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	workRepo := mocks.NewWorkRepository(t)
	workSOHistoryRepo := mocks.NewWorkServiceOrderHistoryRepository(t)
	w := domain.Work{ID: "w-1"}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)
	workRepo.EXPECT().FindByID(mock.Anything, "w-1").Return(w, nil)
	repo.EXPECT().AddWorkLink(mock.Anything, "so-1", "w-1", w.Price).Return(nil)
	workSOHistoryRepo.EXPECT().Insert(mock.Anything, "so-1", "w-1", mock.Anything, domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START).Return(nil)
	expectReviewOSPricing(executor, repo, "so-1")

	s := Service(executor, repo, workRepo, mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), workSOHistoryRepo, mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddWorks(context.Background(), "so-1", []string{"w-1"})
	assert.NoError(t, err)
}

// --- RemoveWork tests ---

func TestRemoveWork_EmptySOID(t *testing.T) {
	s := newBasicService(t)
	err := s.RemoveWork(context.Background(), "  ", "w-1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestRemoveWork_EmptyWorkID(t *testing.T) {
	s := newBasicService(t)
	err := s.RemoveWork(context.Background(), "so-1", "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

func TestRemoveWork_SONotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveWork(context.Background(), "so-1", "w-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestRemoveWork_SONotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_RECEIVED, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveWork(context.Background(), "so-1", "w-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotNew)
}

func TestRemoveWork_RemoveLinkError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)
	repo.EXPECT().RemoveWorkLink(mock.Anything, "so-1", "w-1").Return(domain.ErrServiceOrderWorkNotFound)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveWork(context.Background(), "so-1", "w-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderWorkNotFound)
}

func TestRemoveWork_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)
	repo.EXPECT().RemoveWorkLink(mock.Anything, "so-1", "w-1").Return(nil)
	expectReviewOSPricing(executor, repo, "so-1")

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveWork(context.Background(), "so-1", "w-1")
	assert.NoError(t, err)
}

// --- ListSupplies tests ---

func TestListSupplies_EmptyID(t *testing.T) {
	s := newBasicService(t)
	_, err := s.ListSupplies(context.Background(), "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestListSupplies_SONotFound(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), nil)

	s := newBasicServiceWith(t, repo)
	_, err := s.ListSupplies(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestListSupplies_ExistsByIDError(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), domain.ErrInfraConflict)

	s := newBasicServiceWith(t, repo)
	_, err := s.ListSupplies(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

func TestListSupplies_Success(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	supplies := []domain.Supply{{ID: "sup-1"}}
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, "so-1").Return(supplies, nil)

	s := newBasicServiceWith(t, repo)
	got, err := s.ListSupplies(context.Background(), "so-1")
	assert.NoError(t, err)
	assert.Equal(t, supplies, got)
}

// --- SendToCustomerApproval tests ---

func TestSendToCustomerApproval_EmptyID(t *testing.T) {
	s := newBasicService(t)
	err := s.SendToCustomerApproval(context.Background(), "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSendToCustomerApproval_FindError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToCustomerApproval(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestSendToCustomerApproval_WrongStatus(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToCustomerApproval(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotInDiagnosis)
}

func TestSendToCustomerApproval_SaveError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS}, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(domain.ErrInfraConflict)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToCustomerApproval(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

func TestSendToCustomerApproval_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	// uow.Execute called twice: once inside, once for reviewOSPricing
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps()).Times(2)
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS}, nil).Times(2)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(nil).Times(2)
	repo.EXPECT().ListWorksByServiceOrderID(mock.Anything, "so-1").Return([]domain.Work{}, nil)
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, "so-1").Return([]domain.Supply{}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.SendToCustomerApproval(context.Background(), "so-1")
	assert.NoError(t, err)
}

// --- Receive tests ---

func TestReceive_EmptyID(t *testing.T) {
	s := newBasicService(t)
	err := s.Receive(context.Background(), "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestReceive_FindError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Receive(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestReceive_WrongStatus(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_RECEIVED}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Receive(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotNew)
}

func TestReceive_SaveError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(domain.ErrInfraConflict)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Receive(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

func TestReceive_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)
	repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(so *domain.ServiceOrder) bool {
		return so.Status == domain.SERVICE_ORDER_STATUS_RECEIVED
	})).Return(nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Receive(context.Background(), "so-1")
	assert.NoError(t, err)
}

// --- Accept tests ---

func TestAccept_EmptyID(t *testing.T) {
	s := newBasicService(t)
	err := s.Accept(context.Background(), "  ", "user-1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestAccept_FindError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	customerSvc := mocks.NewCustomerService(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound).Maybe()
	customerSvc.EXPECT().GetByUserID(mock.Anything, "user-1").Return(domain.Customer{ID: "cust-1"}, nil).Maybe()

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerSvc, mocks.NewVehicleService(t))
	err := s.Accept(context.Background(), "so-1", "user-1")
	assert.Error(t, err)
}

func TestAccept_WrongStatus(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	customerSvc := mocks.NewCustomerService(t)
	cust := domain.Customer{ID: "cust-1"}
	so := domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW, Customer: &cust}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	customerSvc.EXPECT().GetByUserID(mock.Anything, "user-1").Return(cust, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerSvc, mocks.NewVehicleService(t))
	err := s.Accept(context.Background(), "so-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotAwaitingApproval)
}

func TestAccept_CustomerMismatch(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	customerSvc := mocks.NewCustomerService(t)
	soOwner := domain.Customer{ID: "cust-owner"}
	otherCust := domain.Customer{ID: "cust-other"}
	so := domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, Customer: &soOwner}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	customerSvc.EXPECT().GetByUserID(mock.Anything, "user-1").Return(otherCust, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerSvc, mocks.NewVehicleService(t))
	err := s.Accept(context.Background(), "so-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrInvalidCustomerProperty)
}

func TestAccept_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	customerSvc := mocks.NewCustomerService(t)
	cust := domain.Customer{ID: "cust-1"}
	so := domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, Customer: &cust}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	customerSvc.EXPECT().GetByUserID(mock.Anything, "user-1").Return(cust, nil)
	repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(s *domain.ServiceOrder) bool {
		return s.Status == domain.SERVICE_ORDER_STATUS_IN_PROGRESS
	})).Return(nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerSvc, mocks.NewVehicleService(t))
	err := s.Accept(context.Background(), "so-1", "user-1")
	assert.NoError(t, err)
}

// --- Reject tests ---

func TestReject_EmptyID(t *testing.T) {
	s := newBasicService(t)
	err := s.Reject(context.Background(), "  ", "user-1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestReject_WrongStatus(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	customerSvc := mocks.NewCustomerService(t)
	cust := domain.Customer{ID: "cust-1"}
	so := domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_PROGRESS, Customer: &cust}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	customerSvc.EXPECT().GetByUserID(mock.Anything, "user-1").Return(cust, nil)
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_PROGRESS, nil)
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, "so-1").Return([]domain.Supply{}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerSvc, mocks.NewVehicleService(t))
	err := s.Reject(context.Background(), "so-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotAwaitingApproval)
}

func TestReject_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	customerSvc := mocks.NewCustomerService(t)
	supplyRepo := mocks.NewSupplyRepository(t)
	cust := domain.Customer{ID: "cust-1"}
	sup := domain.Supply{ID: "sup-1", StockQuantity: 2}
	so := domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, Customer: &cust}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	customerSvc.EXPECT().GetByUserID(mock.Anything, "user-1").Return(cust, nil)
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, nil)
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, "so-1").Return([]domain.Supply{sup}, nil)
	supplyRepo.EXPECT().RestoreStock(mock.Anything, "sup-1", 2).Return(nil)
	repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(s *domain.ServiceOrder) bool {
		return s.Status == domain.SERVICE_ORDER_STATUS_REJECTED
	})).Return(nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerSvc, mocks.NewVehicleService(t))
	err := s.Reject(context.Background(), "so-1", "user-1")
	assert.NoError(t, err)
}

// --- Deliver tests ---

func TestDeliver_EmptyID(t *testing.T) {
	s := newBasicService(t)
	err := s.Deliver(context.Background(), "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestDeliver_FindError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Deliver(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestDeliver_WrongStatus(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_PROGRESS}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Deliver(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotCompleted)
}

func TestService_Deliver_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_COMPLETED}, nil)
	repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(so *domain.ServiceOrder) bool {
		return so.Status == domain.SERVICE_ORDER_STATUS_DELIVERED
	})).Return(nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.Deliver(context.Background(), "so-1")
	assert.NoError(t, err)
}

func TestService_ReviewOSPricing_Success(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	so := domain.ServiceOrder{ID: "so-1"}
	works := []domain.Work{{ID: "w-1", Price: decimal.NewFromInt(100)}}
	supplies := []domain.Supply{{ID: "sup-1", UnitPrice: decimal.NewFromInt(10), StockQuantity: 2}}

	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	repo.EXPECT().ListWorksByServiceOrderID(mock.Anything, "so-1").Return(works, nil)
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, "so-1").Return(supplies, nil)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().Save(mock.Anything, mock.MatchedBy(func(s *domain.ServiceOrder) bool {
		return s.TotalAmount.Equal(decimal.NewFromInt(120))
	})).Return(nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.reviewOSPricing(context.Background(), "so-1")
	assert.NoError(t, err)
}

// --- GetFullOSByID tests ---

func TestGetFullOSByID_EmptyID(t *testing.T) {
	s := newBasicService(t)
	_, err := s.GetFullOSByID(context.Background(), "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestGetFullOSByID_FindError(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound).Maybe()
	repo.EXPECT().ListWorksByServiceOrderID(mock.Anything, "so-1").Return([]domain.Work{}, nil).Maybe()
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, "so-1").Return([]domain.Supply{}, nil).Maybe()

	s := newBasicServiceWith(t, repo)
	_, err := s.GetFullOSByID(context.Background(), "so-1")
	assert.Error(t, err)
}

func TestGetFullOSByID_Success(t *testing.T) {
	repo := mocks.NewServiceOrderRepository(t)
	so := domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}
	works := []domain.Work{{ID: "w-1"}}
	supplies := []domain.Supply{{ID: "sup-1"}}
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	repo.EXPECT().ListWorksByServiceOrderID(mock.Anything, "so-1").Return(works, nil)
	repo.EXPECT().ListSuppliesByServiceOrderID(mock.Anything, "so-1").Return(supplies, nil)

	s := newBasicServiceWith(t, repo)
	got, err := s.GetFullOSByID(context.Background(), "so-1")
	assert.NoError(t, err)
	assert.Equal(t, so.ID, got.ID)
	assert.Equal(t, works, got.Works)
	assert.Equal(t, supplies, got.Supplies)
}

// --- Cancel additional branch coverage ---

func TestCancel_FindError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{}, domain.ErrServiceOrderNotFound)

	s := newBasicServiceWith(t, repo)
	s.uow = executor
	err := s.Cancel(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestCancel_NilCustomer_ReturnsNotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW, Customer: nil}, nil)

	s := newBasicServiceWith(t, repo)
	s.uow = executor
	err := s.Cancel(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestCancel_SaveError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	so := domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW, Customer: &domain.Customer{ID: "c-1"}}
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(so, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(domain.ErrInfraConflict)

	s := newBasicServiceWith(t, repo)
	s.uow = executor
	err := s.Cancel(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrInfraConflict)
}

// --- AddSupplies missing branch: SO not found and blank supply ID ---

func TestAddSupplies_SONotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 1}})
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestAddSupplies_BlankSupplyID(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "  ", Amount: 1}})
	assert.ErrorIs(t, err, domain.ErrInvalidSupplyID)
}

// --- RemoveSupply missing branches ---

func TestRemoveSupply_EmptySOID(t *testing.T) {
	s := newBasicService(t)
	err := s.RemoveSupply(context.Background(), "  ", "sup-1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestRemoveSupply_EmptySupplyID(t *testing.T) {
	s := newBasicService(t)
	err := s.RemoveSupply(context.Background(), "so-1", "  ")
	assert.ErrorIs(t, err, domain.ErrInvalidSupplyID)
}

func TestRemoveSupply_SONotFound(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(false, domain.SERVICE_ORDER_STATUS(""), nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks.NewVehicleService(t))
	err := s.RemoveSupply(context.Background(), "so-1", "sup-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}
