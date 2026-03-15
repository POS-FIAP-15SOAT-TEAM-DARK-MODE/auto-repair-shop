package work

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
)

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	work, err := domain.NewWork(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain work: %v", err)
	}

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			if err := steps[0](ctx); err != nil {
				t.Fatalf("step returned error: %v", err)
			}
			return nil
		})

	repo.
		EXPECT().
		Save(ctx, work).
		Return(nil)

	appService := Service(exec, repo)

	err = appService.Create(ctx, work)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Create_ValidationError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	work, err := domain.NewWork(
		"",
		"Valid description",
		"10.00",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain work: %v", err)
	}

	appService := Service(exec, repo)

	err = appService.Create(ctx, work)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if !errors.Is(err, domain.ErrEmptyWorkName) {
		t.Fatalf("expected error to wrap ErrEmptyWorkName, got %v", err)
	}

	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestService_Create_UnitOfWorkError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	work, err := domain.NewWork(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain work: %v", err)
	}

	expectedErr := errors.New("uow execute failed")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	appService := Service(exec, repo)

	err = appService.Create(ctx, work)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func TestService_Create_RepositoryErrorViaExecutor(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	work, err := domain.NewWork(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain work: %v", err)
	}

	expectedRepoErr := errors.New("db error")

	repo.
		EXPECT().
		Save(ctx, work).
		Return(expectedRepoErr)

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	appService := Service(exec, repo)

	err = appService.Create(ctx, work)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedRepoErr) {
		t.Fatalf("expected error to wrap repo error %v, got %v", expectedRepoErr, err)
	}
}

func TestService_SaveRepositoryStep_RepositoryError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	work, err := domain.NewWork(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain work: %v", err)
	}

	appSvcConcrete := &service{
		uow:  exec,
		repo: repo,
	}

	expectedRepoErr := errors.New("db error")

	repo.
		EXPECT().
		Save(ctx, work).
		Return(expectedRepoErr)

	step := appSvcConcrete.saveRepositoryStep(work)

	err = step(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedRepoErr) {
		t.Fatalf("expected error to wrap repo error %v, got %v", expectedRepoErr, err)
	}
}
