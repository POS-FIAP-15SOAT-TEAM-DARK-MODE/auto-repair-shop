package customer_test

import (
	"context"
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
			Password: "Senha@123",
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
			Password: "Senha@123",
		},
	}
}

// --- Create ---

func TestService_Create(t *testing.T) {
	tests := []struct {
		name      string
		input     domain.Customer
		mockSetup func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository)
		wantErr   error
	}{
		{
			name:  "success_individual",
			input: validIndividualCustomer(),
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
				userRepo.EXPECT().AssignRole(mock.Anything, "uuid-individual", domain.CUSTOMER).Return(nil)
				customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
		},
		{
			name:  "success_company",
			input: validCompanyCustomer(),
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
				userRepo.EXPECT().AssignRole(mock.Anything, "uuid-company", domain.CUSTOMER).Return(nil)
				customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
		},
		{
			name:  "duplicate_email_returns_conflict",
			input: validIndividualCustomer(),
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, _ *domainmocks.CustomerRepository) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(domain.ErrDataConflict)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
			wantErr: domain.ErrDataConflict,
		},
		{
			name:  "duplicate_cpf_returns_conflict",
			input: validIndividualCustomer(),
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				userRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)
				userRepo.EXPECT().AssignRole(mock.Anything, "uuid-individual", domain.CUSTOMER).Return(nil)
				customerRepo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.Customer")).Return(domain.ErrDataConflict)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
			wantErr: domain.ErrDataConflict,
		},
		{
			name:  "transaction_failure_returns_error",
			input: validIndividualCustomer(),
			mockSetup: func(executor *uowmocks.Executor, _ *domainmocks.UserRepository, _ *domainmocks.CustomerRepository) {
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := uowmocks.NewExecutor(t)
			userRepo := domainmocks.NewUserRepository(t)
			customerRepo := domainmocks.NewCustomerRepository(t)

			tt.mockSetup(executor, userRepo, customerRepo)

			err := svc.Service(executor, userRepo, customerRepo).Create(context.Background(), tt.input)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- GetByID ---

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		mockSetup    func(*domainmocks.CustomerRepository)
		wantCustomer domain.Customer
		wantErr      error
	}{
		{
			name: "success",
			id:   "uuid-individual",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
			},
			wantCustomer: validIndividualCustomer(),
		},
		{
			name: "not_found",
			id:   "unknown-id",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByID(mock.Anything, "unknown-id").Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			wantErr: domain.ErrCustomerNotFound,
		},
		{
			name: "repository_error",
			id:   "any-id",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByID(mock.Anything, "any-id").Return(domain.Customer{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerRepo := domainmocks.NewCustomerRepository(t)
			tt.mockSetup(customerRepo)

			result, err := svc.Service(nil, nil, customerRepo).GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCustomer.ID, result.ID)
			}
		})
	}
}

// --- Update ---

func TestService_Update(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name         string
		id           string
		nameArg      *string
		emailArg     *string
		phoneArg     *string
		mockSetup    func(*uowmocks.Executor, *domainmocks.UserRepository, *domainmocks.CustomerRepository)
		wantCustomer func() domain.Customer
		wantErr      error
	}{
		{
			name:     "success_update_all_fields",
			id:       "uuid-individual",
			nameArg:  strPtr("Novo Nome"),
			emailArg: strPtr("novo@example.com"),
			phoneArg: strPtr("11888888888"),
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
				userRepo.EXPECT().Update(mock.Anything, "uuid-individual", "Novo Nome", "novo@example.com").Return(nil)
				customerRepo.EXPECT().Update(mock.Anything, "uuid-individual", "11888888888").Return(nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
			wantCustomer: func() domain.Customer {
				c := validIndividualCustomer()
				c.User.Name = "Novo Nome"
				c.User.Email = "novo@example.com"
				c.Phone = "11888888888"
				return c
			},
		},
		{
			name:     "success_update_partial_only_phone",
			id:       "uuid-individual",
			nameArg:  nil,
			emailArg: nil,
			phoneArg: strPtr("11777777777"),
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				c := validIndividualCustomer()
				customerRepo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(c, nil)
				// name and email keep original values
				userRepo.EXPECT().Update(mock.Anything, "uuid-individual", c.User.Name, c.User.Email).Return(nil)
				customerRepo.EXPECT().Update(mock.Anything, "uuid-individual", "11777777777").Return(nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
			wantCustomer: func() domain.Customer {
				c := validIndividualCustomer()
				c.Phone = "11777777777"
				return c
			},
		},
		{
			name:    "customer_not_found",
			id:      "unknown-id",
			nameArg: strPtr("Novo Nome"),
			mockSetup: func(_ *uowmocks.Executor, _ *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "unknown-id").Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			wantErr: domain.ErrCustomerNotFound,
		},
		{
			name:     "duplicate_email_returns_conflict",
			id:       "uuid-individual",
			emailArg: strPtr("duplicate@example.com"),
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything, mock.Anything, "duplicate@example.com").Return(domain.ErrDataConflict)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
			wantErr: domain.ErrDataConflict,
		},
		{
			name:    "transaction_failure",
			id:      "uuid-individual",
			nameArg: strPtr("Novo Nome"),
			mockSetup: func(executor *uowmocks.Executor, _ *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := uowmocks.NewExecutor(t)
			userRepo := domainmocks.NewUserRepository(t)
			customerRepo := domainmocks.NewCustomerRepository(t)

			tt.mockSetup(executor, userRepo, customerRepo)

			result, err := svc.Service(executor, userRepo, customerRepo).Update(
				context.Background(), tt.id, tt.nameArg, tt.emailArg, tt.phoneArg,
			)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				want := tt.wantCustomer()
				assert.Equal(t, want.User.Name, result.User.Name)
				assert.Equal(t, want.User.Email, result.User.Email)
				assert.Equal(t, want.Phone, result.Phone)
			}
		})
	}
}

// --- Delete ---

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		mockSetup func(*uowmocks.Executor, *domainmocks.UserRepository, *domainmocks.CustomerRepository)
		wantErr   error
	}{
		{
			name: "success",
			id:   "uuid-individual",
			mockSetup: func(executor *uowmocks.Executor, userRepo *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
				customerRepo.EXPECT().Delete(mock.Anything, "uuid-individual").Return(nil)
				userRepo.EXPECT().Delete(mock.Anything, "uuid-individual").Return(nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
		},
		{
			name: "customer_not_found",
			id:   "unknown-id",
			mockSetup: func(_ *uowmocks.Executor, _ *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "unknown-id").Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			wantErr: domain.ErrCustomerNotFound,
		},
		{
			name: "has_service_orders_returns_conflict",
			id:   "uuid-individual",
			mockSetup: func(executor *uowmocks.Executor, _ *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
				customerRepo.EXPECT().Delete(mock.Anything, "uuid-individual").Return(domain.ErrCustomerHasServiceOrders)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).RunAndReturn(runAllSteps())
			},
			wantErr: domain.ErrCustomerHasServiceOrders,
		},
		{
			name: "transaction_failure",
			id:   "uuid-individual",
			mockSetup: func(executor *uowmocks.Executor, _ *domainmocks.UserRepository, customerRepo *domainmocks.CustomerRepository) {
				customerRepo.EXPECT().GetByID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
				executor.EXPECT().Execute(mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := uowmocks.NewExecutor(t)
			userRepo := domainmocks.NewUserRepository(t)
			customerRepo := domainmocks.NewCustomerRepository(t)

			tt.mockSetup(executor, userRepo, customerRepo)

			err := svc.Service(executor, userRepo, customerRepo).Delete(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// --- GetByDocument ---

func TestService_GetByDocument(t *testing.T) {
	tests := []struct {
		name         string
		rawDocument  string
		mockSetup    func(*domainmocks.CustomerRepository)
		wantCustomer domain.Customer
		wantErr      error
	}{
		{
			name:        "success_cpf_with_formatting",
			rawDocument: "111.444.777-35",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				// Service sanitizes "111.444.777-35" → "11144477735" before calling repo
				repo.EXPECT().GetByDocument(mock.Anything, "11144477735").Return(validIndividualCustomer(), nil)
			},
			wantCustomer: validIndividualCustomer(),
		},
		{
			name:        "success_cnpj_with_formatting",
			rawDocument: "11.222.333/0001-81",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				// Service sanitizes "11.222.333/0001-81" → "11222333000181" before calling repo
				repo.EXPECT().GetByDocument(mock.Anything, "11222333000181").Return(validCompanyCustomer(), nil)
			},
			wantCustomer: validCompanyCustomer(),
		},
		{
			name:        "invalid_document_format",
			rawDocument: "123",
			mockSetup:   func(_ *domainmocks.CustomerRepository) {},
			wantErr:     domain.ErrInvalidDocumentFormat,
		},
		{
			name:        "not_found",
			rawDocument: "111.444.777-35",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByDocument(mock.Anything, "11144477735").Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			wantErr: domain.ErrCustomerNotFound,
		},
		{
			name:        "repository_error",
			rawDocument: "111.444.777-35",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByDocument(mock.Anything, mock.Anything).Return(domain.Customer{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerRepo := domainmocks.NewCustomerRepository(t)
			tt.mockSetup(customerRepo)

			result, err := svc.Service(nil, nil, customerRepo).GetByDocument(context.Background(), tt.rawDocument)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCustomer.ID, result.ID)
			}
		})
	}
}

// --- GetByUserID ---

func TestService_GetByUserID(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		mockSetup    func(*domainmocks.CustomerRepository)
		wantCustomer domain.Customer
		wantErr      error
	}{
		{
			name: "success",
			id:   "uuid-individual",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, "uuid-individual").Return(validIndividualCustomer(), nil)
			},
			wantCustomer: validIndividualCustomer(),
		},
		{
			name: "not_found",
			id:   "unknown-id",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, "unknown-id").Return(domain.Customer{}, domain.ErrCustomerNotFound)
			},
			wantErr: domain.ErrCustomerNotFound,
		},
		{
			name: "repository_error",
			id:   "any-id",
			mockSetup: func(repo *domainmocks.CustomerRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, "any-id").Return(domain.Customer{}, assert.AnError)
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customerRepo := domainmocks.NewCustomerRepository(t)
			tt.mockSetup(customerRepo)

			result, err := svc.Service(nil, nil, customerRepo).GetByUserID(context.Background(), tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantCustomer.ID, result.ID)
			}
		})
	}
}
