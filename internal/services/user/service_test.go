package user

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	userMock "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
		mockRepo      func(t *testing.T) *userMock.UserRepository
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
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
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
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
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
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
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
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(validDomainUser, nil)
				mockRepo.On("GetRolesById", mock.Anything, "123").Return(nil, errors.New("repository failed"))
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
			mockRepo: func(t *testing.T) *userMock.UserRepository {
				mockRepo := userMock.NewUserRepository(t)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(validDomainUser, nil)
				mockRepo.On("GetRolesById", mock.Anything, "123").Return([]string{"user"}, nil)
				return mockRepo
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mocks.UserRepository{}
			if tt.mockRepo != nil {
				mockRepo = tt.mockRepo(t)
			}

			service := Service(nil, mockRepo, 24*time.Hour)
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
