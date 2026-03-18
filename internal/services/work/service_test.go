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
	"github.com/shopspring/decimal"
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

func TestService_List_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	params := &domain.ListWorkParams{
		Page:     1,
		PageSize: 10,
		Status:   "",
	}

	expectedItems := []domain.Work{
		{ID: "1", Name: "Oil Change", Description: "Complete oil change", Price: decimal.NewFromInt(50), Status: domain.ACTIVE},
	}
	var expectedTotal int64 = 1

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return(expectedTotal, nil)

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return(expectedItems, nil)

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	appService := Service(exec, repo)

	got, err := appService.List(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got == nil {
		t.Fatalf("expected non-nil response")
	}

	if got.TotalItems != expectedTotal {
		t.Fatalf("expected total %d, got %d", expectedTotal, got.TotalItems)
	}

	if len(got.Items) != len(expectedItems) {
		t.Fatalf("expected %d items, got %d", len(expectedItems), len(got.Items))
	}

	if got.Page != params.Page {
		t.Fatalf("expected page %d, got %d", params.Page, got.Page)
	}

	if got.PageSize != params.PageSize {
		t.Fatalf("expected page size %d, got %d", params.PageSize, got.PageSize)
	}

	if got.TotalPages != 1 {
		t.Fatalf("expected 1 total page, got %d", got.TotalPages)
	}
}

func TestService_List_UoWError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	params := &domain.ListWorkParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("uow execute failed")

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	appService := Service(exec, repo)

	got, err := appService.List(ctx, params)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil response on error")
	}
}

func TestService_List_CountError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	params := &domain.ListWorkParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("count db error")

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return(int64(0), expectedErr).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return([]domain.Work{}, nil).Maybe()

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	appService := Service(exec, repo)

	got, err := appService.List(ctx, params)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil response on error")
	}
}

func TestService_List_SearchError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	params := &domain.ListWorkParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("search db error")

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return(int64(0), nil).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return(nil, expectedErr).Maybe()

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	appService := Service(exec, repo)

	got, err := appService.List(ctx, params)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil response on error")
	}
}

func TestService_List_EmptyResult(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewWorkRepository(t)

	params := &domain.ListWorkParams{Page: 1, PageSize: 10}

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return(int64(0), nil)

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchWorkParams")).
		Return([]domain.Work{}, nil)

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	appService := Service(exec, repo)

	got, err := appService.List(ctx, params)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got == nil {
		t.Fatalf("expected non-nil response")
	}

	if got.TotalItems != 0 {
		t.Fatalf("expected total 0, got %d", got.TotalItems)
	}

	if len(got.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(got.Items))
	}

	if got.TotalPages != 0 {
		t.Fatalf("expected 0 total pages, got %d", got.TotalPages)
	}
}
