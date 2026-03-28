package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_Create_Success(t *testing.T) {
	mockExecutor := uowmocks.NewExecutor(t)
	repo := domainmocks.NewUserRepository(t)
	svc := user.Service(mockExecutor, repo, time.Minute)

	mockExecutor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
		for _, step := range steps {
			if err := step(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	u := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Secret@123",
	}

	err := svc.Create(context.Background(), u)

	assert.NoError(t, err)
}

func TestService_Create_ValidationError(t *testing.T) {
	mockExecutor := uowmocks.NewExecutor(t)
	repo := domainmocks.NewUserRepository(t)
	svc := user.Service(mockExecutor, repo, time.Minute)

	u := &domain.User{
		Name:     "Ts", // too short
		Email:    "test@example.com",
		Password: "Secret@123",
	}

	err := svc.Create(context.Background(), u)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user validation failed")
}

func TestService_Create_TransactionError(t *testing.T) {
	mockExecutor := uowmocks.NewExecutor(t)
	repo := domainmocks.NewUserRepository(t)
	svc := user.Service(mockExecutor, repo, time.Minute)

	mockExecutor.EXPECT().Execute(mock.Anything, mock.Anything).Return(errors.New("db error"))

	u := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Secret@123",
	}

	err := svc.Create(context.Background(), u)

	assert.Error(t, err)
	assert.Equal(t, "db error", err.Error())
}

func TestService_Create_RepoError(t *testing.T) {
	mockExecutor := uowmocks.NewExecutor(t)
	repo := domainmocks.NewUserRepository(t)
	svc := user.Service(mockExecutor, repo, time.Minute)

	mockExecutor.EXPECT().Execute(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, steps ...uow.Step) error {
		for _, step := range steps {
			if err := step(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("repo error"))

	u := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Secret@123",
	}

	err := svc.Create(context.Background(), u)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository.Create failed")
}

func TestService_Create_HashPasswordError(t *testing.T) {
	mockExecutor := uowmocks.NewExecutor(t)
	repo := domainmocks.NewUserRepository(t)

	expectedErr := errors.New("hashing failed")
	// Using the exported test constructor to inject the mock hashPassword function
	svc := user.NewServiceForTest(mockExecutor, repo, time.Minute, func(u *domain.User, ctx context.Context) error {
		return expectedErr
	})

	u := &domain.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "Secret@123",
	}

	err := svc.Create(context.Background(), u)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user password hashing failed")
	assert.ErrorIs(t, err, expectedErr)
}

func TestService_Login(t *testing.T) {
	validDomainUser := &domain.User{
		ID:       "123",
		Name:     "John Doe",
		Email:    "test@example.com",
		Password: "$2a$10$2R8fcJ033oiMS2cRCO3L4Ojn0lGsivjOVjid3bfi2bJ.j8bvDRGC6",
	}

	tests := []struct {
		name          string
		loggedUser    *domain.LoggedUser
		mockRepo      func(t *testing.T) *domainmocks.UserRepository
		expectedError error
	}{
		{
			name: "Should fail gracefully when email is empty",
			loggedUser: &domain.LoggedUser{
				User: domain.User{
					Email:    "",
					Password: "password",
				},
			},
			expectedError: domain.ErrEmptyUserEmail,
		},
		{
			name: "Should fail gracefully when password is empty",
			loggedUser: &domain.LoggedUser{
				User: domain.User{
					Email:    "test@example.com",
					Password: "",
				},
			},
			expectedError: domain.ErrEmptyUserPassword,
		},
		{
			name: "Should fail gracefully when get by email fails",
			loggedUser: &domain.LoggedUser{
				User: domain.User{
					Email:    "test@example.com",
					Password: "password",
				},
			},
			mockRepo: func(t *testing.T) *domainmocks.UserRepository {
				mockRepo := domainmocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("repository failed"))
				return mockRepo
			},
			expectedError: errors.New("repository failed"),
		},
		{
			name: "Should fail gracefully when get by email returns nil",
			loggedUser: &domain.LoggedUser{
				User: domain.User{
					Email:    "test@example.com",
					Password: "password",
				},
			},
			mockRepo: func(t *testing.T) *domainmocks.UserRepository {
				mockRepo := domainmocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				return mockRepo
			},
			expectedError: domain.ErrInvalidUserCredentials,
		},
		{
			name: "Should fail gracefully when password does not match",
			loggedUser: &domain.LoggedUser{
				User: domain.User{
					Email:    "test@example.com",
					Password: "invalid_password",
				},
			},
			mockRepo: func(t *testing.T) *domainmocks.UserRepository {
				mockRepo := domainmocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(validDomainUser, nil)
				return mockRepo
			},
			expectedError: domain.ErrInvalidUserCredentials,
		},
		{
			name: "Should fail gracefully when get roles by id fails",
			loggedUser: &domain.LoggedUser{
				User: domain.User{
					Email:    "test@example.com",
					Password: "password",
				},
			},
			mockRepo: func(t *testing.T) *domainmocks.UserRepository {
				mockRepo := domainmocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(validDomainUser, nil)
				mockRepo.On("GetRolesByUserId", mock.Anything, "123").Return(nil, errors.New("repository failed"))
				return mockRepo
			},
			expectedError: errors.New("repository failed"),
		},
		{
			name: "Should return the token successfully",
			loggedUser: &domain.LoggedUser{
				User: domain.User{
					Email:    "test@example.com",
					Password: "password",
				},
			},
			mockRepo: func(t *testing.T) *domainmocks.UserRepository {
				mockRepo := domainmocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(validDomainUser, nil)
				mockRepo.On("GetRolesByUserId", mock.Anything, "123").Return([]domain.Role{domain.ADMIN}, nil)
				return mockRepo
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &domainmocks.UserRepository{}
			if tt.mockRepo != nil {
				mockRepo = tt.mockRepo(t)
			}

			service := user.Service(nil, mockRepo, 24*time.Hour)
			err := service.Login(context.Background(), tt.loggedUser)
			if tt.expectedError != nil {
				assert.Error(t, err)
				if err == tt.expectedError {
					return
				}
				if errors.Is(err, tt.expectedError) {
					return
				}
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.expectedError)
				}
				if strings.Contains(err.Error(), tt.expectedError.Error()) {
					return
				}
				t.Fatalf("expected error to match %v, got %v", tt.expectedError, err)
			}
		})
	}
}
