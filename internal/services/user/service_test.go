package user_test

import (
	"context"
	"errors"
	"testing"

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
	svc := user.Service(mockExecutor, repo)

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
	svc := user.Service(mockExecutor, repo)

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
	svc := user.Service(mockExecutor, repo)

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
	svc := user.Service(mockExecutor, repo)

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
	svc := user.NewServiceForTest(mockExecutor, repo, func(u *domain.User, ctx context.Context) error {
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
