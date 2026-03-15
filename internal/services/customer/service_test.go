package customer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	svc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
)

// executes fn with the provided context — simulates a real transaction
func runFn() func(context.Context, domain.TxFunc) error {
	return func(ctx context.Context, fn domain.TxFunc) error {
		return fn(ctx)
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
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(nil)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).RunAndReturn(runFn())

	input := validIndividualCustomer()
	err := svc.Service(transactor, userRepo, customerRepo).Create(context.Background(), input)

	assert.NoError(t, err)
}

func TestService_Create_Company_Success(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(nil)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).RunAndReturn(runFn())

	err := svc.Service(transactor, userRepo, customerRepo).Create(context.Background(), validCompanyCustomer())

	assert.NoError(t, err)
}

func TestService_Create_DuplicateEmail_ReturnsConflict(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t) // Create must NOT be called
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(domain.ErrDataConflict)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).RunAndReturn(runFn())

	err := svc.Service(transactor, userRepo, customerRepo).Create(context.Background(), validIndividualCustomer())

	assert.ErrorIs(t, err, domain.ErrDataConflict)
}

func TestService_Create_DuplicateCPF_ReturnsConflict(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(domain.ErrDataConflict)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).RunAndReturn(runFn())

	err := svc.Service(transactor, userRepo, customerRepo).Create(context.Background(), validIndividualCustomer())

	assert.ErrorIs(t, err, domain.ErrDataConflict)
}

func TestService_Create_TransactionFailure_ReturnsError(t *testing.T) {
	transactor := domainmocks.NewTransactor(t)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).
		Return(errors.New("db connection lost"))

	err := svc.Service(transactor, nil, nil).Create(context.Background(), validIndividualCustomer())

	assert.Error(t, err)
}
