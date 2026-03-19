package customer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
	svc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
)

// runAllSteps executes all UoW steps with the provided context — simulates a real transaction.
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

func validIndividualCustomer() domain.Customer {
	return domain.Customer{
		ID:     "uuid-individual",
		UserID: "uuid-individual",
		Type:   domain.IndividualCustomerType,
		CPF:    "11144477735",
		Phone:  "11999999999",
		User: &domain.User{
			ID:       "uuid-individual",
			Name:     "João Silva",
			Email:    "joao@example.com",
			Password: "secret123",
		},
	}
}

func validCompanyCustomer() domain.Customer {
	return domain.Customer{
		ID:          "uuid-company",
		UserID:      "uuid-company",
		Type:        domain.CompanyCustomerType,
		CNPJ:        "11222333000181",
		CompanyName: "Empresa SA",
		Phone:       "11999999999",
		User: &domain.User{
			ID:       "uuid-company",
			Name:     "Empresa SA",
			Email:    "empresa@example.com",
			Password: "secret123",
		},
	}
}

func TestService_Create_Individual_Success(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	executor := uowmocks.NewExecutor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(nil)
	executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())

	input := validIndividualCustomer()
	err := svc.Service(executor, userRepo, customerRepo).Create(context.Background(), input)

	assert.NoError(t, err)
}

func TestService_Create_Company_Success(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	executor := uowmocks.NewExecutor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(nil)
	executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())

	err := svc.Service(executor, userRepo, customerRepo).Create(context.Background(), validCompanyCustomer())

	assert.NoError(t, err)
}

func TestService_Create_DuplicateEmail_ReturnsConflict(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t) // Create must NOT be called
	executor := uowmocks.NewExecutor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(domain.ErrDataConflict)
	executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())

	err := svc.Service(executor, userRepo, customerRepo).Create(context.Background(), validIndividualCustomer())

	assert.ErrorIs(t, err, domain.ErrDataConflict)
}

func TestService_Create_DuplicateCPF_ReturnsConflict(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	executor := uowmocks.NewExecutor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(domain.ErrDataConflict)
	executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())

	err := svc.Service(executor, userRepo, customerRepo).Create(context.Background(), validIndividualCustomer())

	assert.ErrorIs(t, err, domain.ErrDataConflict)
}

func TestService_Create_TransactionFailure_ReturnsError(t *testing.T) {
	executor := uowmocks.NewExecutor(t)
	executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("db connection lost"))

	err := svc.Service(executor, nil, nil).Create(context.Background(), validIndividualCustomer())

	assert.Error(t, err)
}
