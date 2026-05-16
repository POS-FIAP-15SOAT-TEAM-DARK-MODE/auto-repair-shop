package service_order

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
	domainV2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/interfaces"
	mocks2 "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/vehicle/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_ReturnsServiceOrderWithDefaults(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := "mock-id"

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
	vehicleId := "mock-id"

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, domain.ErrCustomerNotFound, vehicleId, nil)
	s := Service(executor, mockRepo, workRepo, supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
}

func TestService_Create_VerifyVehicleID(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := "mock-id"

	executor, mockRepo, workRepo, supplyRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, nil, vehicleId, domain.ErrVehicleNotFound)
	s := Service(executor, mockRepo, workRepo, supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrVehicleNotFound)
}

func TestService_Create_InsertSO_RepositoryFailure(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := "mock-id"

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
	_ string,
	vehicleError error) (uow.Executor,
	domain.ServiceOrderRepository,
	domain.WorkRepository,
	domain.SupplyRepository,
	domain.CustomerService,
	interfaces.VehicleService) {
	executor := uowmocks.NewExecutor(t)
	mockRepo := mocks.NewServiceOrderRepository(t)
	mockWorkRepo := mocks.NewWorkRepository(t)
	mockSupplyRepo := mocks.NewSupplyRepository(t)
	mockCustomerSvc := mocks.NewCustomerService(t)
	mockVehicleSvc := mocks2.NewVehicleService(t)

	mockCustomerSvc.EXPECT().GetByID(mock.Anything, customerID).Return(domain.Customer{}, customerError)

	if customerError == nil {
		mockVehicleSvc.EXPECT().FindById(mock.Anything, mock.AnythingOfType("string")).Return(domainV2.Vehicle{}, vehicleError)
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
	err := s.AddSupplies(context.Background(), "so-1", []domain.AddSupply{{ID: "sup-1", Amount: 5}})
	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
}

func TestAddSupplies_ServiceOrderNotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), supplyRepo, mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
	err := s.RemoveSupply(context.Background(), "so-1", "sup-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderSupplyNotFound)
}

func TestRemoveSupply_ServiceOrderNotNew(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().ExistsByID(mock.Anything, "so-1").Return(true, domain.SERVICE_ORDER_STATUS_NEW, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestFinish_ServiceOrderNotInProgress(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
	err := s.Finish(context.Background(), "so-1")
	assert.NoError(t, err)
}

func TestFinish_SaveFailure(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_IN_PROGRESS}, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(assert.AnError)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
	err := s.SendToDiagnosis(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
}

func TestSendToDiagnosis_ServiceOrderNotInReceived(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_NEW}, nil)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
	err := s.SendToDiagnosis(context.Background(), "so-1")
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotInReceived)
}

func TestSendToDiagnosis_SaveFailure(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	repo := mocks.NewServiceOrderRepository(t)

	executor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
	repo.EXPECT().FindByID(mock.Anything, "so-1").Return(domain.ServiceOrder{ID: "so-1", Status: domain.SERVICE_ORDER_STATUS_RECEIVED}, nil)
	repo.EXPECT().Save(mock.Anything, mock.Anything).Return(domain.ErrInfraConflict)

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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

	s := Service(executor, repo, mocks.NewWorkRepository(t), mocks.NewSupplyRepository(t), mocks.NewServiceOrderHistoryRepository(t), mocks.NewWorkServiceOrderHistoryRepository(t), mocks.NewCustomerService(t), mocks2.NewVehicleService(t))
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
		mocks2.NewVehicleService(t),
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
		mocks2.NewVehicleService(t),
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
		mocks2.NewVehicleService(t),
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
