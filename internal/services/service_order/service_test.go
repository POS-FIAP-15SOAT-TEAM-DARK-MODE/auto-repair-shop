package service_order

import (
	"context"
	"testing"
	"time"

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

	executor, mockRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, nil, vehicleId, nil)
	s := Service(executor, mockRepo, customerMock, vehicleMock)
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

	executor, mockRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, domain.ErrCustomerNotFound, vehicleId, nil)
	s := Service(executor, mockRepo, customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
}

func TestService_Create_VerifyVehicleID(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, customerMock, vehicleMock := setupCreateTestMock(t, nil, customerId, nil, vehicleId, domain.ErrVehicleNotFound)
	s := Service(executor, mockRepo, customerMock, vehicleMock)
	ctx := context.Background()

	_, err := s.Create(ctx, customerId, vehicleId)
	assert.ErrorIs(t, err, domain.ErrVehicleNotFound)
}

func TestService_Create_InsertSO_RepositoryFailure(t *testing.T) {
	c, _ := domain.NewCustomer("", "", "", domain.IndividualCustomerType, "", "", "")
	customerId := c.ID
	vehicleId := domain.NewVehicle("", "", "", "", 2026).ID

	executor, mockRepo, customerMock, vehicleMock := setupCreateTestMock(t, domain.ErrInfraConflict, c.ID, nil, vehicleId, nil)
	s := Service(executor, mockRepo, customerMock, vehicleMock)
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
	domain.CustomerService,
	domain.VehicleService) {
	executor := uowmocks.NewExecutor(t)
	mockRepo := mocks.NewServiceOrderRepository(t)
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

	return executor, mockRepo, mockCustomerSvc, mockVehicleSvc
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

func TestService_GetHistoryByID_Success(t *testing.T) {
	// table-driven tests for GetHistoryByID
	expSuccess := []domain.ServiceOrderHistory{
		{
			ID:             "h-1",
			PreviousStatus: domain.SERVICE_ORDER_STATUS_RECEIVED,
			NewStatus:      domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS,
			CreatedAt:      time.Date(2026, 4, 5, 12, 34, 56, 0, time.UTC),
		},
	}

	tests := []struct {
		name        string
		serviceID   string
		mockSetup   func(m *mocks.ServiceOrderRepository)
		expected    []domain.ServiceOrderHistory
		expectedErr error
	}{
		{
			name:      "success",
			serviceID: "so-1",
			mockSetup: func(m *mocks.ServiceOrderRepository) {
				m.EXPECT().GetHistoryByID(mock.Anything, "so-1").Return(expSuccess, nil)
			},
			expected:    expSuccess,
			expectedErr: nil,
		},
		{
			name:      "repository error",
			serviceID: "so-2",
			mockSetup: func(m *mocks.ServiceOrderRepository) {
				m.EXPECT().GetHistoryByID(mock.Anything, "so-2").Return(nil, domain.ErrInfraConflict)
			},
			expected:    nil,
			expectedErr: domain.ErrInfraConflict,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewServiceOrderRepository(t)
			tt.mockSetup(mockRepo)

			s := Service(nil, mockRepo, nil, nil)
			history, err := s.GetHistoryByID(context.Background(), tt.serviceID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, history)
			}
		})
	}
}
