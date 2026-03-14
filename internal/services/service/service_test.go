package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
)

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	svcDomain, err := domain.NewService(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	uow.
		EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		RunAndReturn(func(_ context.Context, steps ...domain.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			// Simulate UnitOfWork executing the step and calling the repository.
			if err := steps[0](ctx); err != nil {
				t.Fatalf("step returned error: %v", err)
			}
			return nil
		})

	repo.
		EXPECT().
		Save(ctx, svcDomain).
		Return(nil)

	appService := Service(uow, repo)

	got, err := appService.Create(ctx, svcDomain)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got != svcDomain {
		t.Fatalf("expected returned service pointer to match input")
	}
}

func TestService_Create_ValidationError(t *testing.T) {
	ctx := context.Background()

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	// Invalid: empty name triggers domain.ErrEmptyName in Validate.
	svcDomain, err := domain.NewService(
		"",
		"Valid description",
		"10.00",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	appService := Service(uow, repo)

	got, err := appService.Create(ctx, svcDomain)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if got != nil {
		t.Fatalf("expected returned service to be nil on validation error")
	}

	if !errors.Is(err, domain.ErrEmptyName) {
		t.Fatalf("expected error to wrap ErrEmptyName, got %v", err)
	}

	// When validation fails, UoW and repository must not be called.
	uow.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestService_Create_UnitOfWorkError(t *testing.T) {
	ctx := context.Background()

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	svcDomain, err := domain.NewService(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	expectedErr := errors.New("uow execute failed")

	uow.
		EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		Return(expectedErr)

	appService := Service(uow, repo)

	got, err := appService.Create(ctx, svcDomain)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil result on error")
	}
}

func TestService_Create_RepositoryErrorViaExecutor(t *testing.T) {
	ctx := context.Background()

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	svcDomain, err := domain.NewService(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	expectedRepoErr := errors.New("db error")

	repo.
		EXPECT().
		Save(ctx, svcDomain).
		Return(expectedRepoErr)

	uow.
		EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		RunAndReturn(func(_ context.Context, steps ...domain.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			// Execute the step; it should call repo.Save and return a wrapped error.
			return steps[0](ctx)
		})

	appService := Service(uow, repo)

	got, err := appService.Create(ctx, svcDomain)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedRepoErr) {
		t.Fatalf("expected error to wrap repo error %v, got %v", expectedRepoErr, err)
	}

	if got != nil {
		t.Fatalf("expected nil result on error")
	}
}

func TestService_SaveRepositoryStep_RepositoryError(t *testing.T) {
	ctx := context.Background()

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	svcDomain, err := domain.NewService(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	appSvcConcrete := &service{
		uow:  uow,
		repo: repo,
	}

	expectedRepoErr := errors.New("db error")

	repo.
		EXPECT().
		Save(ctx, svcDomain).
		Return(expectedRepoErr)

	step := appSvcConcrete.saveRepositoryStep(svcDomain)

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

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	params := &domain.ListServiceParams{
		Page:     1,
		PageSize: 10,
		Status:   "",
	}

	expectedItems := []domain.Service{
		{ID: "1", Name: "Oil Change", Description: "Complete oil change", Status: domain.ACTIVE},
	}
	var expectedTotal int64 = 1

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return(expectedTotal, nil)

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return(expectedItems, nil)

	uow.EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		RunAndReturn(func(_ context.Context, steps ...domain.Step) error {
			return steps[0](ctx)
		})

	appService := Service(uow, repo)

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

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	params := &domain.ListServiceParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("uow execute failed")

	uow.EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		Return(expectedErr)

	appService := Service(uow, repo)

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

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	params := &domain.ListServiceParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("count db error")

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return(int64(0), expectedErr).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return([]domain.Service{}, nil).Maybe()

	uow.EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		RunAndReturn(func(_ context.Context, steps ...domain.Step) error {
			return steps[0](ctx)
		})

	appService := Service(uow, repo)

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

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	params := &domain.ListServiceParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("search db error")

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return(int64(0), nil).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return(nil, expectedErr).Maybe()

	uow.EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		RunAndReturn(func(_ context.Context, steps ...domain.Step) error {
			return steps[0](ctx)
		})

	appService := Service(uow, repo)

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

	uow := domainmocks.NewExecutor(t)
	repo := domainmocks.NewServiceRepository(t)

	params := &domain.ListServiceParams{Page: 1, PageSize: 10}

	repo.EXPECT().
		Count(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return(int64(0), nil)

	repo.EXPECT().
		Search(mock.Anything, mock.AnythingOfType("*domain.SearchServiceParams")).
		Return([]domain.Service{}, nil)

	uow.EXPECT().
		Execute(ctx, mock.AnythingOfType("domain.Step")).
		RunAndReturn(func(_ context.Context, steps ...domain.Step) error {
			return steps[0](ctx)
		})

	appService := Service(uow, repo)

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
