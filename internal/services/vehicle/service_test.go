package vehicle

import (
	"context"
	"errors"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	uowMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow/mocks"
)

func TestService_Create_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "123",
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
		Save(ctx, vehicle).
		Return(nil)

	service := NewService(exec, repo)

	err := service.Create(ctx, vehicle)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Create_ValidationError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "INVALID", // inválida
		Year:         2020,
	}

	service := NewService(exec, repo)

	err := service.Create(ctx, vehicle)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if !errors.Is(err, domain.ErrVehicleInvalidPlate) {
		t.Fatalf("expected error to wrap ErrVehicleInvalidPlate, got %v", err)
	}

	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestService_Create_UowError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "123",
	}

	expectedErr := errors.New("uow error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	service := NewService(exec, repo)

	err := service.Create(ctx, vehicle)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to be %v, got %v", expectedErr, err)
	}

	repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
}

func TestService_Create_RepositoryError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "123",
	}

	expectedErr := errors.New("db error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	repo.
		EXPECT().
		Save(ctx, vehicle).
		Return(expectedErr)

	service := NewService(exec, repo)

	err := service.Create(ctx, vehicle)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestService_FindByLicensePlate_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	plate := "ABC1D23"

	expectedVehicle := &domain.Vehicle{
		ID:           "1",
		LicensePlate: plate,
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "123",
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
		Find(ctx, plate).
		Return(expectedVehicle, nil)

	service := NewService(exec, repo)

	result, err := service.FindByLicensePlate(ctx, plate)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result != expectedVehicle {
		t.Fatalf("expected vehicle %v, got %v", expectedVehicle, result)
	}
}

func TestService_FindByLicensePlate_InvalidPlate(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	invalidPlate := "INVALID"

	service := NewService(exec, repo)

	result, err := service.FindByLicensePlate(ctx, invalidPlate)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, domain.ErrVehicleInvalidPlate) {
		t.Fatalf("expected error to wrap ErrVehicleInvalidPlate, got %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}

	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Find", mock.Anything, mock.Anything)
}

func TestService_FindByLicensePlate_RepositoryError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	plate := "ABC1D23"
	expectedErr := errors.New("db error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	repo.
		EXPECT().
		Find(ctx, plate).
		Return(nil, expectedErr)

	service := NewService(exec, repo)

	result, err := service.FindByLicensePlate(ctx, plate)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to be %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestService_FindByLicensePlate_UowError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	plate := "ABC1D23"
	expectedErr := errors.New("uow error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	service := NewService(exec, repo)

	result, err := service.FindByLicensePlate(ctx, plate)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to be %v, got %v", expectedErr, err)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}

	repo.AssertNotCalled(t, "Find", mock.Anything, mock.Anything)
}

func TestService_Update_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "123",
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
		Update(ctx, vehicle).
		Return(nil)

	service := NewService(exec, repo)

	err := service.Update(ctx, vehicle)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Update_ValidationError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "INVALID", // inválida
		Year:         2020,
	}

	service := NewService(exec, repo)

	err := service.Update(ctx, vehicle)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	if !errors.Is(err, domain.ErrVehicleInvalidPlate) {
		t.Fatalf("expected error to wrap ErrVehicleInvalidPlate, got %v", err)
	}

	exec.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestService_Update_UowError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "123",
	}

	expectedErr := errors.New("uow error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	service := NewService(exec, repo)

	err := service.Update(ctx, vehicle)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to be %v, got %v", expectedErr, err)
	}

	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestService_Update_RepositoryError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	vehicle := &domain.Vehicle{
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "123",
	}

	expectedErr := errors.New("db error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	repo.
		EXPECT().
		Update(ctx, vehicle).
		Return(expectedErr)

	service := NewService(exec, repo)

	err := service.Update(ctx, vehicle)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
}

func TestService_Delete_Success(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	id := "veh-1"

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
		Delete(ctx, id).
		Return(nil)

	service := NewService(exec, repo)

	err := service.Delete(ctx, id)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestService_Delete_RepositoryError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	id := "veh-1"
	expectedErr := errors.New("db error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		RunAndReturn(func(_ context.Context, steps ...uow.Step) error {
			if len(steps) != 1 {
				t.Fatalf("expected 1 step, got %d", len(steps))
			}
			return steps[0](ctx)
		})

	repo.
		EXPECT().
		Delete(ctx, id).
		Return(expectedErr)

	service := NewService(exec, repo)

	err := service.Delete(ctx, id)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to be %v, got %v", expectedErr, err)
	}
}

func TestService_Delete_UowError(t *testing.T) {
	ctx := context.Background()

	exec := uowMocks.NewExecutor(t)
	repo := domainMocks.NewVehicleRepository(t)

	id := "veh-1"
	expectedErr := errors.New("uow error")

	exec.
		EXPECT().
		Execute(ctx, mock.Anything).
		Return(expectedErr)

	service := NewService(exec, repo)

	err := service.Delete(ctx, id)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to be %v, got %v", expectedErr, err)
	}

	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}
