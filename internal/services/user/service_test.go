package user

import (
	"errors"
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	userMock "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/stretchr/testify/assert"
)

func TestService_Login(t *testing.T) {
	validDomainUser := &domain.User{
		Id:       "123",
		Name:     "John Doe",
		Email:    "test@example.com",
		Password: "$2a$10$2R8fcJ033oiMS2cRCO3L4Ojn0lGsivjOVjid3bfi2bJ.j8bvDRGC6",
	}

	tests := []struct {
		name          string
		user          *domain.User
		mockRepo      func(t *testing.T) *userMock.UserRepository
		expectedError error
	}{
		{
			name: "Should fail gracefully when get by email fails",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "password",
			},
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, errors.New("repository failed"))
				return mockRepo
			},
			expectedError: errors.New("repository failed"),
		},
		{
			name: "Should fail gracefully when get by email returns nil",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "password",
			},
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", "test@example.com").Return(nil, nil)
				return mockRepo
			},
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name: "Should fail gracefully when password does not match",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "invalid_password",
			},
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", "test@example.com").Return(validDomainUser, nil)
				return mockRepo
			},
			expectedError: domain.ErrInvalidCredentials,
		},
		{
			name: "Should fail gracefully when get roles by id fails",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "password",
			},
			mockRepo: func(t *testing.T) *mocks.UserRepository {
				mockRepo := mocks.NewUserRepository(t)
				mockRepo.On("GetByEmail", "test@example.com").Return(validDomainUser, nil)
				mockRepo.On("GetRolesById", "123").Return(nil, errors.New("repository failed"))
				return mockRepo
			},
			expectedError: errors.New("repository failed"),
		},
		{
			name: "Should return the token successfully",
			user: &domain.User{
				Email:    "test@example.com",
				Password: "password",
			},
			mockRepo: func(t *testing.T) *userMock.UserRepository {
				mockRepo := userMock.NewUserRepository(t)
				mockRepo.On("GetByEmail", "test@example.com").Return(validDomainUser, nil)
				mockRepo.On("GetRolesById", "123").Return([]string{"user"}, nil)
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

			service := Service(mockRepo, 24*time.Hour, "super-secret-key")
			_, err := service.Login(tt.user)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			}
		})
	}
}
