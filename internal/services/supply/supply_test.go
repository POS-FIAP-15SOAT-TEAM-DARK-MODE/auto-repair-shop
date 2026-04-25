package supply

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	uowmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
	"github.com/shopspring/decimal"
)

// -------------------------
// Create
// -------------------------

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
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
		Save(ctx, supply).
		Return(nil)

	appService := Service(exec, repo)

	err := appService.Create(ctx, supply)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Create_ValidationError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{
		Name: "", // invalid
	}

	expectedErr := errors.New("validation error")
	exec.EXPECT().Execute(ctx, mock.Anything).Return(expectedErr)
	appService := Service(exec, repo)

	err := appService.Create(ctx, supply)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")

	assert.Error(t, errors.New("validation error"), err.Error())

}

func TestService_Create_UnitOfWorkError(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	expectedErr := errors.New("uow execute failed")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	appService := Service(exec, repo)

	err := appService.Create(ctx, supply)
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
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	expectedRepoErr := errors.New("db error")

	repo.
		EXPECT().
		Save(ctx, supply).
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

	err := appService.Create(ctx, supply)
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
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	appSvcConcrete := &service{
		uow:  exec,
		repo: repo,
	}

	expectedRepoErr := errors.New("db error")

	repo.
		EXPECT().
		Save(ctx, supply).
		Return(expectedRepoErr)

	step := appSvcConcrete.saveRepositoryStep(supply)

	err := step(ctx)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedRepoErr) {
		t.Fatalf("expected error to wrap repo error %v, got %v", expectedRepoErr, err)
	}
}

// -------------------------
// List
// -------------------------

func TestService_List_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	params := &domain.ListSupplyParams{
		Page:     1,
		PageSize: 10,
	}

	expectedItems := []domain.Supply{
		{ID: "1", Name: "Brake Pad", Description: "High performance brake pad", UnitPrice: decimal.NewFromFloat(49.99), StockQuantity: 10},
	}
	var expectedTotal int64 = 1

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(expectedTotal, nil)

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(expectedItems, nil)

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
	// Arrange
	ctx := context.Background()

	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	params := &domain.ListSupplyParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("uow execute failed")

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), expectedErr)

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(nil, nil)
	// action
	appService := Service(exec, repo)

	//assert
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
	repo := domainmocks.NewSupplyRepository(t)

	params := &domain.ListSupplyParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("count db error")

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), expectedErr).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return([]domain.Supply{}, nil).Maybe()

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
	repo := domainmocks.NewSupplyRepository(t)

	params := &domain.ListSupplyParams{Page: 1, PageSize: 10}
	expectedErr := errors.New("search db error")

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), nil).Maybe()

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return(nil, expectedErr).Maybe()

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
	repo := domainmocks.NewSupplyRepository(t)

	params := &domain.ListSupplyParams{Page: 1, PageSize: 10}

	repo.EXPECT().
		Count(mock.Anything, mock.Anything).
		Return(int64(0), nil)

	repo.EXPECT().
		Search(mock.Anything, mock.Anything).
		Return([]domain.Supply{}, nil)

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

// -------------------------
// Update
// -------------------------

func TestService_Update_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	repo.EXPECT().Save(ctx, supply).Return(nil)

	appService := Service(exec, repo)
	err := appService.Update(ctx, supply)

	assert.NoError(t, err)
}

func TestService_Update_ValidationError(t *testing.T) {
	ctx := context.Background()
	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{Name: ""} // invalid

	expectedErr := errors.New("validation error")
	exec.EXPECT().Execute(ctx, mock.Anything).Return(expectedErr)

	appService := Service(exec, repo)

	err := appService.Update(ctx, supply)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation error")
}

func TestService_Update_UnitOfWorkError(t *testing.T) {
	ctx := context.Background()
	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	supply := &domain.Supply{
		Name:          "Brake Pad",
		Description:   "High performance brake pad",
		UnitPrice:     decimal.NewFromFloat(49.99),
		StockQuantity: 10,
	}
	expectedErr := errors.New("uow execute failed")

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	appService := Service(exec, repo)
	err := appService.Update(ctx, supply)

	assert.ErrorIs(t, err, expectedErr)
}
func TestService_Delete_Success(t *testing.T) {
	ctx := context.Background()
	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	id := "supply-id"

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			return steps[0](ctx)
		})

	repo.EXPECT().Delete(ctx, id).Return(nil)

	appService := Service(exec, repo)
	err := appService.Delete(ctx, id)

	assert.NoError(t, err)
}
func TestService_Delete_Error(t *testing.T) {
	ctx := context.Background()
	exec := uowmocks.NewExecutor(t)
	repo := domainmocks.NewSupplyRepository(t)

	id := "supply-id"
	expectedErr := errors.New("delete failed")

	exec.EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	appService := Service(exec, repo)
	err := appService.Delete(ctx, id)

	assert.ErrorIs(t, err, expectedErr)
}
