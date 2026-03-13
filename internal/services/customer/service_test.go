package customer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/customer/dto"
	svc "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/customer"
)

// executes fn with the provided repos — simulates a real transaction
func runFn(userRepo domain.UserRepository, customerRepo domain.CustomerRepository) func(context.Context, domain.TxFunc) error {
	return func(_ context.Context, fn domain.TxFunc) error {
		return fn(userRepo, customerRepo)
	}
}

func validIndividualRequest() dto.CreateCustomerRequest {
	return dto.CreateCustomerRequest{
		Name:     "João Silva",
		Email:    "joao@example.com",
		Password: "secret123",
		Type:     domain.IndividualCustomerType,
		Document: "11144477735",
		Phone:    "11999999999",
	}
}

func validCompanyRequest() dto.CreateCustomerRequest {
	return dto.CreateCustomerRequest{
		Name:        "Empresa SA",
		Email:       "empresa@example.com",
		Password:    "secret123",
		Type:        domain.CompanyCustomerType,
		Document:    "11222333000181",
		CompanyName: "Empresa SA",
		Phone:       "11999999999",
	}
}

func TestService_Create_Individual_Success(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.AnythingOfType("*domain.Customer")).Return(nil)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(runFn(userRepo, customerRepo))

	customer, err := svc.Service(transactor).Create(validIndividualRequest())

	assert.NoError(t, err)
	assert.NotEmpty(t, customer.ID)
	assert.Equal(t, "11144477735", customer.CPF)
	assert.Equal(t, domain.IndividualCustomerType, customer.Type)
	assert.NotNil(t, customer.User)
	assert.Equal(t, "joao@example.com", customer.User.Email)
}

func TestService_Create_Company_Success(t *testing.T) {
	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.AnythingOfType("*domain.Customer")).Return(nil)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(runFn(userRepo, customerRepo))

	customer, err := svc.Service(transactor).Create(validCompanyRequest())

	assert.NoError(t, err)
	assert.Equal(t, "11222333000181", customer.CNPJ)
	assert.Equal(t, "Empresa SA", customer.CompanyName)
}

func TestService_Create_InvalidCPF_ReturnsBadRequest(t *testing.T) {
	transactor := domainmocks.NewTransactor(t) // no expectations — must not be called

	req := validIndividualRequest()
	req.Document = "11111111111" // all-same

	_, err := svc.Service(transactor).Create(req)

	var badReq domain.BadRequestError
	assert.ErrorAs(t, err, &badReq)
}

func TestService_Create_InvalidCNPJ_ReturnsBadRequest(t *testing.T) {
	transactor := domainmocks.NewTransactor(t)

	req := validCompanyRequest()
	req.Document = "11222333000182" // wrong check digit

	_, err := svc.Service(transactor).Create(req)

	var badReq domain.BadRequestError
	assert.ErrorAs(t, err, &badReq)
}

func TestService_Create_MissingCompanyName_ReturnsBadRequest(t *testing.T) {
	transactor := domainmocks.NewTransactor(t)

	req := validCompanyRequest()
	req.CompanyName = ""

	_, err := svc.Service(transactor).Create(req)

	var badReq domain.BadRequestError
	assert.ErrorAs(t, err, &badReq)
}

func TestService_Create_DuplicateEmail_ReturnsConflict(t *testing.T) {
	conflictErr := domain.ConflictError{Message: "email already in use"}

	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t) // Create must NOT be called
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.AnythingOfType("*domain.User")).Return(conflictErr)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(runFn(userRepo, customerRepo))

	_, err := svc.Service(transactor).Create(validIndividualRequest())

	var conflictError domain.ConflictError
	assert.ErrorAs(t, err, &conflictError)
	assert.Equal(t, "email already in use", conflictError.Message)
}

func TestService_Create_DuplicateCPF_ReturnsConflict(t *testing.T) {
	conflictErr := domain.ConflictError{Message: "CPF already registered"}

	userRepo := domainmocks.NewUserRepository(t)
	customerRepo := domainmocks.NewCustomerRepository(t)
	transactor := domainmocks.NewTransactor(t)

	userRepo.EXPECT().Create(mock.AnythingOfType("*domain.User")).Return(nil)
	customerRepo.EXPECT().Create(mock.AnythingOfType("*domain.Customer")).Return(conflictErr)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).
		RunAndReturn(runFn(userRepo, customerRepo))

	_, err := svc.Service(transactor).Create(validIndividualRequest())

	var conflictError domain.ConflictError
	assert.ErrorAs(t, err, &conflictError)
}

func TestService_Create_TransactionFailure_ReturnsError(t *testing.T) {
	transactor := domainmocks.NewTransactor(t)
	transactor.EXPECT().WithTransaction(mock.Anything, mock.Anything).
		Return(errors.New("db connection lost"))

	_, err := svc.Service(transactor).Create(validIndividualRequest())

	assert.Error(t, err)
}

func TestService_Create_ValidationRunsBeforeTransaction(t *testing.T) {
	transactor := domainmocks.NewTransactor(t) // no expectations — must not be called

	req := validIndividualRequest()
	req.Document = "00000000000"

	_, err := svc.Service(transactor).Create(req)

	var badReq domain.BadRequestError
	assert.ErrorAs(t, err, &badReq)
}
