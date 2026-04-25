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
