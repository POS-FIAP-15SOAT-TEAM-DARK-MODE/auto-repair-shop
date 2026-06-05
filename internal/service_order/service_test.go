package service_order_test

import (
	"context"
	"errors"
	"testing"

	customerDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/customer/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	serviceOrder "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/adapters"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/interfaces"
	soRepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/service_order/repository"
	supplyDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/domain"
	supplyMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/supply/interfaces/mocks"
	vehicleDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/domain"
	vehicleMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/vehicle/interfaces/mocks"
	workDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/domain"
	workMocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/work/interfaces/mocks"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ─── helpers ──────────────────────────────────────────────────────────────────

// inMemoryUoW is a no-op unit-of-work that executes steps inline.
type inMemoryUoW struct{}

func (u *inMemoryUoW) Execute(ctx context.Context, steps ...uow.Step) error {
	for _, s := range steps {
		if err := s(ctx); err != nil {
			return err
		}
	}
	return nil
}

// stubCustomerFinder is a minimal CustomerFinder that always returns a preset customer.
type stubCustomerFinder struct {
	customer customerDomain.Customer
	err      error
}

func (f *stubCustomerFinder) GetByID(_ context.Context, _ string) (customerDomain.Customer, error) {
	return f.customer, f.err
}

func (f *stubCustomerFinder) GetByUserID(_ context.Context, _ string) (customerDomain.Customer, error) {
	return f.customer, f.err
}

// makeCustomer builds a minimal customer for tests.
func makeCustomer(id string) customerDomain.Customer {
	return customerDomain.Customer{ID: id}
}

// makeVehicle builds a minimal vehicle for tests.
func makeVehicle(id string) vehicleDomain.Vehicle {
	return vehicleDomain.Vehicle{
		ID:           id,
		LicensePlate: "TEST0001",
		Brand:        "Toyota",
		Model:        "Corolla",
		Year:         2021,
	}
}

// makeSupply builds a supply with the given stock quantity.
func makeSupply(id string, stock int) supplyDomain.Supply {
	p, _ := decimal.NewFromString("50.00")
	return supplyDomain.Supply{
		ID:            id,
		Name:          "Test Supply",
		Description:   "Test supply for unit tests",
		UnitPrice:     p,
		StockQuantity: stock,
		Version:       1,
	}
}

// makeWork builds a work domain object with the given price string.
func makeWork(id, priceStr string) workDomain.Work {
	p, _ := decimal.NewFromString(priceStr)
	return workDomain.Work{
		ID:          id,
		Name:        "Test Work",
		Description: "Test work for unit tests",
		Price:       p,
		Status:      workDomain.ACTIVE,
	}
}

// buildSvc is a convenience helper that wires up a service with the given mocks.
func buildSvc(
	workRepo *workMocks.WorkRepository,
	supplySvc *supplyMocks.SupplyService,
	vehicleSvc *vehicleMocks.VehicleService,
	customerFinder serviceOrder.CustomerFinder,
) interfaces.ServiceOrderService {
	memRepo := soRepo.NewSOMemory()
	wsoHistoryRepo := soRepo.NewWSOHistoryMemory()
	return serviceOrder.NewService(
		&inMemoryUoW{},
		memRepo,
		workRepo,
		supplySvc,
		wsoHistoryRepo,
		customerFinder,
		vehicleSvc,
	)
}

// ─── Create — no works / supplies (regression) ───────────────────────────────

func TestSOService_Create_NoWorksNoSupplies(t *testing.T) {
	ctx := context.Background()
	custID := "cust-1"
	vehID := "veh-1"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)
	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})
	resp, err := svc.Create(ctx, adapters.CreateSORequest{CustomerID: custID, VehicleID: vehID})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "NEW", resp.Status)
}

// ─── Create — customer not found ─────────────────────────────────────────────

func TestSOService_Create_CustomerNotFound(t *testing.T) {
	ctx := context.Background()

	svc := buildSvc(
		workMocks.NewWorkRepository(t),
		supplyMocks.NewSupplyService(t),
		vehicleMocks.NewVehicleService(t),
		&stubCustomerFinder{err: customerDomain.ErrCustomerNotFound},
	)

	_, err := svc.Create(ctx, adapters.CreateSORequest{CustomerID: "unknown", VehicleID: "veh-x"})
	assert.ErrorIs(t, err, customerDomain.ErrCustomerNotFound)
}

// ─── Create — with works: success ────────────────────────────────────────────

func TestSOService_Create_WithWorks_Success(t *testing.T) {
	ctx := context.Background()
	custID := "cust-works"
	vehID := "veh-works"
	workID := "work-abc"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)

	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)
	workRepo.EXPECT().FindByID(mock.Anything, workID).Return(makeWork(workID, "100.00"), nil)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	resp, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		WorkIDs:    []string{workID},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "NEW", resp.Status)

	// Verify the work is actually linked.
	worksResp, err := svc.GetWorks(ctx, resp.ID)
	require.NoError(t, err)
	require.Len(t, worksResp.Items, 1)
	assert.Equal(t, workID, worksResp.Items[0].ID)
}

// ─── Create — empty work ID → error ──────────────────────────────────────────

func TestSOService_Create_EmptyWorkID_ReturnsError(t *testing.T) {
	ctx := context.Background()
	custID := "cust-empty-work"
	vehID := "veh-empty-work"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)
	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)
	// FindByID must NOT be called when the work ID is empty.

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	_, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		WorkIDs:    []string{""},
	})

	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

// ─── Create — work repo error propagated ─────────────────────────────────────

func TestSOService_Create_WorkRepoError(t *testing.T) {
	ctx := context.Background()
	custID := "cust-work-err"
	vehID := "veh-work-err"
	workID := "work-err"
	repoErr := errors.New("work repo unavailable")

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)

	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)
	workRepo.EXPECT().FindByID(mock.Anything, workID).Return(workDomain.Work{}, repoErr)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	_, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		WorkIDs:    []string{workID},
	})

	assert.ErrorIs(t, err, repoErr)
}

// ─── Create — with supplies, sufficient stock ─────────────────────────────────

func TestSOService_Create_WithSupplies_Success(t *testing.T) {
	ctx := context.Background()
	custID := "cust-sup"
	vehID := "veh-sup"
	supplyID := "supply-123"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)

	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)
	supplySvc.EXPECT().FindById(mock.Anything, supplyID).Return(makeSupply(supplyID, 10), nil)
	supplySvc.EXPECT().DecrementStockQuantity(mock.Anything, supplyID, 2).Return(nil)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	resp, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		Supplies:   []adapters.AddSupplyItem{{ID: supplyID, Amount: 2}},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)
	assert.Equal(t, "NEW", resp.Status)

	// Verify the supply is linked.
	suppliesResp, err := svc.GetSupplies(ctx, resp.ID)
	require.NoError(t, err)
	require.Len(t, suppliesResp.Items, 1)
	assert.Equal(t, supplyID, suppliesResp.Items[0].ID)
}

// ─── Create — supply amount exceeds stock → ErrSupplyOutOfStock ──────────────

func TestSOService_Create_SupplyExceedsStock_ReturnsError(t *testing.T) {
	ctx := context.Background()
	custID := "cust-oos"
	vehID := "veh-oos"
	supplyID := "supply-oos"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)

	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)
	supplySvc.EXPECT().FindById(mock.Anything, supplyID).Return(makeSupply(supplyID, 1), nil)
	// DecrementStockQuantity must NOT be called.

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	_, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		Supplies:   []adapters.AddSupplyItem{{ID: supplyID, Amount: 5}},
	})

	assert.ErrorIs(t, err, domain.ErrSupplyOutOfStock)
}

// ─── Create — empty supply ID → ErrInvalidSupplyID ───────────────────────────

func TestSOService_Create_EmptySupplyID_ReturnsError(t *testing.T) {
	ctx := context.Background()
	custID := "cust-empty-sup"
	vehID := "veh-empty-sup"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)
	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	_, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		Supplies:   []adapters.AddSupplyItem{{ID: "", Amount: 1}},
	})

	assert.ErrorIs(t, err, domain.ErrInvalidSupplyID)
}

// ─── Create — supply amount ≤ 0 → ErrInvalidSupplyAmount ─────────────────────

func TestSOService_Create_InvalidSupplyAmount_ReturnsError(t *testing.T) {
	ctx := context.Background()
	custID := "cust-invalid-amount"
	vehID := "veh-invalid-amount"
	supplyID := "supply-invalid-amount"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)
	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	_, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		Supplies:   []adapters.AddSupplyItem{{ID: supplyID, Amount: 0}},
	})

	assert.ErrorIs(t, err, domain.ErrInvalidSupplyAmount)
}

// ─── Create — supply FindById error propagated ───────────────────────────────

func TestSOService_Create_SupplyFindByIdError(t *testing.T) {
	ctx := context.Background()
	custID := "cust-sup-err"
	vehID := "veh-sup-err"
	supplyID := "supply-find-err"
	svcErr := errors.New("supply service unavailable")

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)

	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)
	supplySvc.EXPECT().FindById(mock.Anything, supplyID).Return(supplyDomain.Supply{}, svcErr)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	_, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		Supplies:   []adapters.AddSupplyItem{{ID: supplyID, Amount: 1}},
	})

	assert.ErrorIs(t, err, svcErr)
}

// ─── Create — works and supplies together ────────────────────────────────────

func TestSOService_Create_WithWorksAndSupplies_Success(t *testing.T) {
	ctx := context.Background()
	custID := "cust-both"
	vehID := "veh-both"
	workID := "work-both"
	supplyID := "supply-both"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)

	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil)
	workRepo.EXPECT().FindByID(mock.Anything, workID).Return(makeWork(workID, "200.00"), nil)
	supplySvc.EXPECT().FindById(mock.Anything, supplyID).Return(makeSupply(supplyID, 5), nil)
	supplySvc.EXPECT().DecrementStockQuantity(mock.Anything, supplyID, 1).Return(nil)

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	resp, err := svc.Create(ctx, adapters.CreateSORequest{
		CustomerID: custID,
		VehicleID:  vehID,
		WorkIDs:    []string{workID},
		Supplies:   []adapters.AddSupplyItem{{ID: supplyID, Amount: 1}},
	})

	require.NoError(t, err)
	assert.NotEmpty(t, resp.ID)

	works, err := svc.GetWorks(ctx, resp.ID)
	require.NoError(t, err)
	assert.Len(t, works.Items, 1)

	supplies, err := svc.GetSupplies(ctx, resp.ID)
	require.NoError(t, err)
	assert.Len(t, supplies.Items, 1)
}

// ─── Lifecycle helpers ────────────────────────────────────────────────────────

// svcFixture holds a shared service + pre-created SO for lifecycle tests.
type svcFixture struct {
	svc    interfaces.ServiceOrderService
	soID   string
	custID string
	workID string
}

// newSvcFixture creates an SO at NEW status, ready for lifecycle tests.
func newSvcFixture(t *testing.T) svcFixture {
	t.Helper()
	ctx := context.Background()
	custID := "fixture-cust"
	vehID := "fixture-veh"
	workID := "fixture-work"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)

	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil).Maybe()
	workRepo.EXPECT().FindByID(mock.Anything, workID).Return(makeWork(workID, "150.00"), nil).Maybe()

	svc := buildSvc(workRepo, supplySvc, vehicleSvc, &stubCustomerFinder{customer: makeCustomer(custID)})

	resp, err := svc.Create(ctx, adapters.CreateSORequest{CustomerID: custID, VehicleID: vehID})
	require.NoError(t, err)

	// Pre-add a work so the SO has proper pricing later.
	err = svc.AddWorks(ctx, resp.ID, []string{workID})
	require.NoError(t, err)

	return svcFixture{svc: svc, soID: resp.ID, custID: custID, workID: workID}
}

// ─── Receive ─────────────────────────────────────────────────────────────────

func TestSOService_Receive_Success(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)

	err := fix.svc.Receive(ctx, fix.soID)
	require.NoError(t, err)

	status, err := fix.svc.GetStatus(ctx, fix.soID)
	require.NoError(t, err)
	assert.Equal(t, "RECEIVED", status.Status)
}

func TestSOService_Receive_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.Receive(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_Receive_NotFound_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.Receive(context.Background(), "nonexistent-id")
	assert.Error(t, err)
}

func TestSOService_Receive_AlreadyReceived_ReturnsError(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)
	require.NoError(t, fix.svc.Receive(ctx, fix.soID))

	// Second receive must fail.
	err := fix.svc.Receive(ctx, fix.soID)
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotNew)
}

// ─── SendToDiagnosis ──────────────────────────────────────────────────────────

func TestSOService_SendToDiagnosis_Success(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)
	require.NoError(t, fix.svc.Receive(ctx, fix.soID))

	err := fix.svc.SendToDiagnosis(ctx, fix.soID)
	require.NoError(t, err)

	status, _ := fix.svc.GetStatus(ctx, fix.soID)
	assert.Equal(t, "IN_DIAGNOSIS", status.Status)
}

func TestSOService_SendToDiagnosis_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.SendToDiagnosis(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

// ─── AddWorks ─────────────────────────────────────────────────────────────────

func TestSOService_AddWorks_EmptySOID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.AddWorks(context.Background(), "", []string{"w1"})
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_AddWorks_EmptyWorksList_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.AddWorks(context.Background(), "some-so-id", []string{})
	assert.ErrorIs(t, err, domain.ErrEmptyServicesList)
}

// ─── RemoveWork ───────────────────────────────────────────────────────────────

func TestSOService_RemoveWork_EmptySOID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.RemoveWork(context.Background(), "", "w1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_RemoveWork_EmptyWorkID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.RemoveWork(context.Background(), "so-id", "")
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

// ─── AddSupplies ──────────────────────────────────────────────────────────────

func TestSOService_AddSupplies_EmptySOID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.AddSupplies(context.Background(), "", []adapters.AddSupplyItem{{ID: "s1", Amount: 1}})
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_AddSupplies_EmptyList_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.AddSupplies(context.Background(), "so-id", []adapters.AddSupplyItem{})
	assert.ErrorIs(t, err, domain.ErrEmptyServicesList)
}

// ─── RemoveSupply ─────────────────────────────────────────────────────────────

func TestSOService_RemoveSupply_EmptySOID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.RemoveSupply(context.Background(), "", "s1")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_RemoveSupply_EmptySupplyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.RemoveSupply(context.Background(), "so-id", "")
	assert.ErrorIs(t, err, domain.ErrInvalidSupplyID)
}

// ─── Cancel ───────────────────────────────────────────────────────────────────

func TestSOService_Cancel_FromNew_Success(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)

	err := fix.svc.Cancel(ctx, fix.soID)
	require.NoError(t, err)

	status, _ := fix.svc.GetStatus(ctx, fix.soID)
	assert.Equal(t, "CANCELLED", status.Status)
}

func TestSOService_Cancel_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.Cancel(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_Cancel_AlreadyCancelled_ReturnsError(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)
	require.NoError(t, fix.svc.Cancel(ctx, fix.soID))

	err := fix.svc.Cancel(ctx, fix.soID)
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotCancelable)
}

// ─── GetStatus ────────────────────────────────────────────────────────────────

func TestSOService_GetStatus_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	_, err := svc.GetStatus(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_GetStatus_NotFound_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	_, err := svc.GetStatus(context.Background(), "nonexistent")
	assert.Error(t, err)
}

// ─── GetWorks / GetSupplies — input validation ────────────────────────────────

func TestSOService_GetWorks_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	_, err := svc.GetWorks(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_GetSupplies_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	_, err := svc.GetSupplies(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

// ─── GetFullByID ──────────────────────────────────────────────────────────────

func TestSOService_GetFullByID_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	_, err := svc.GetFullByID(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_GetFullByID_Success(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)

	detail, err := fix.svc.GetFullByID(ctx, fix.soID)
	require.NoError(t, err)
	assert.Equal(t, fix.soID, detail.ID)
	assert.Equal(t, "NEW", detail.Status)
}

// ─── List ─────────────────────────────────────────────────────────────────────

func TestSOService_List_ReturnsPaginatedResults(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)

	result, err := fix.svc.List(ctx, adapters.SOFilterParams{Page: 1, PageSize: 10})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, result.TotalItems, int64(1))
}

func TestSOService_List_FilterByStatus(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)

	result, err := fix.svc.List(ctx, adapters.SOFilterParams{Page: 1, PageSize: 10, Status: "NEW"})
	require.NoError(t, err)
	for _, item := range result.Items {
		assert.Equal(t, "NEW", item.Status)
	}
	assert.GreaterOrEqual(t, result.TotalItems, int64(1))
}

// ─── NextWork / CancelWork — input validation ─────────────────────────────────

func TestSOService_NextWork_EmptySOID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.NextWork(context.Background(), "", "work-id")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_NextWork_EmptyWorkID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.NextWork(context.Background(), "so-id", "")
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

func TestSOService_CancelWork_EmptySOID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.CancelWork(context.Background(), "", "work-id")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_CancelWork_EmptyWorkID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.CancelWork(context.Background(), "so-id", "")
	assert.ErrorIs(t, err, domain.ErrInvalidWorkId)
}

// ─── Full lifecycle — NEW → RECEIVED → IN_DIAGNOSIS → AWAITING_APPROVAL → IN_PROGRESS → COMPLETED → DELIVERED ─────

func TestSOService_FullLifecycle(t *testing.T) {
	ctx := context.Background()
	custID := "lifecycle-cust"
	vehID := "lifecycle-veh"
	workID := "lifecycle-work"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)
	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil).Maybe()
	workRepo.EXPECT().FindByID(mock.Anything, workID).Return(makeWork(workID, "300.00"), nil).Maybe()

	finder := &stubCustomerFinder{customer: makeCustomer(custID)}
	svc := buildSvc(workRepo, supplySvc, vehicleSvc, finder)

	// 1. Create SO.
	resp, err := svc.Create(ctx, adapters.CreateSORequest{CustomerID: custID, VehicleID: vehID})
	require.NoError(t, err)
	soID := resp.ID

	// 2. Add work while NEW.
	require.NoError(t, svc.AddWorks(ctx, soID, []string{workID}))

	// 3. Receive.
	require.NoError(t, svc.Receive(ctx, soID))
	st, _ := svc.GetStatus(ctx, soID)
	assert.Equal(t, "RECEIVED", st.Status)

	// 4. Start diagnosis.
	require.NoError(t, svc.SendToDiagnosis(ctx, soID))
	st, _ = svc.GetStatus(ctx, soID)
	assert.Equal(t, "IN_DIAGNOSIS", st.Status)

	// 5. Send for customer approval.
	require.NoError(t, svc.SendToCustomerApproval(ctx, soID))
	st, _ = svc.GetStatus(ctx, soID)
	assert.Equal(t, "AWAITING_APPROVAL", st.Status)

	// 6. Customer accepts.
	require.NoError(t, svc.Accept(ctx, soID, custID))
	st, _ = svc.GetStatus(ctx, soID)
	assert.Equal(t, "IN_PROGRESS", st.Status)

	// 7. Finish.
	require.NoError(t, svc.Finish(ctx, soID))
	st, _ = svc.GetStatus(ctx, soID)
	assert.Equal(t, "COMPLETED", st.Status)

	// 8. Deliver.
	require.NoError(t, svc.Deliver(ctx, soID))
	st, _ = svc.GetStatus(ctx, soID)
	assert.Equal(t, "DELIVERED", st.Status)

	// Terminal status must NOT appear in default list.
	listResult, err := svc.List(ctx, adapters.SOFilterParams{Page: 1, PageSize: 100})
	require.NoError(t, err)
	for _, item := range listResult.Items {
		assert.NotEqual(t, soID, item.ID, "delivered SO should not appear in default list")
	}

	// But it DOES appear when filtering by DELIVERED.
	listResult, err = svc.List(ctx, adapters.SOFilterParams{Page: 1, PageSize: 100, Status: "DELIVERED"})
	require.NoError(t, err)
	found := false
	for _, item := range listResult.Items {
		if item.ID == soID {
			found = true
		}
	}
	assert.True(t, found, "delivered SO must appear in status=DELIVERED filter")
}

// ─── Rejection lifecycle ──────────────────────────────────────────────────────

func TestSOService_Reject_Success(t *testing.T) {
	ctx := context.Background()
	custID := "reject-cust"
	vehID := "reject-veh"
	workID := "reject-work"

	workRepo := workMocks.NewWorkRepository(t)
	supplySvc := supplyMocks.NewSupplyService(t)
	vehicleSvc := vehicleMocks.NewVehicleService(t)
	vehicleSvc.EXPECT().FindById(mock.Anything, vehID).Return(makeVehicle(vehID), nil).Maybe()
	workRepo.EXPECT().FindByID(mock.Anything, workID).Return(makeWork(workID, "100.00"), nil).Maybe()

	finder := &stubCustomerFinder{customer: makeCustomer(custID)}
	svc := buildSvc(workRepo, supplySvc, vehicleSvc, finder)

	resp, err := svc.Create(ctx, adapters.CreateSORequest{CustomerID: custID, VehicleID: vehID})
	require.NoError(t, err)
	soID := resp.ID

	require.NoError(t, svc.AddWorks(ctx, soID, []string{workID}))
	require.NoError(t, svc.Receive(ctx, soID))
	require.NoError(t, svc.SendToDiagnosis(ctx, soID))
	require.NoError(t, svc.SendToCustomerApproval(ctx, soID))

	// Reject the SO.
	err = svc.Reject(ctx, soID, custID)
	require.NoError(t, err)

	st, _ := svc.GetStatus(ctx, soID)
	assert.Equal(t, "REJECTED", st.Status)
}

// ─── Finish / Deliver — wrong state ──────────────────────────────────────────

func TestSOService_Finish_NotInProgress_ReturnsError(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)

	err := fix.svc.Finish(ctx, fix.soID)
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotInProgress)
}

func TestSOService_Deliver_NotCompleted_ReturnsError(t *testing.T) {
	ctx := context.Background()
	fix := newSvcFixture(t)

	err := fix.svc.Deliver(ctx, fix.soID)
	assert.ErrorIs(t, err, domain.ErrServiceOrderNotCompleted)
}

func TestSOService_Finish_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.Finish(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

func TestSOService_Deliver_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.Deliver(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

// ─── SendToCustomerApproval — empty ID ───────────────────────────────────────

func TestSOService_SendToCustomerApproval_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.SendToCustomerApproval(context.Background(), "")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

// ─── Accept — empty ID ────────────────────────────────────────────────────────

func TestSOService_Accept_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.Accept(context.Background(), "", "user-id")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

// ─── Reject — empty ID ────────────────────────────────────────────────────────

func TestSOService_Reject_EmptyID_ReturnsError(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	err := svc.Reject(context.Background(), "", "user-id")
	assert.ErrorIs(t, err, domain.ErrInvalidServiceOrderId)
}

// ─── GetAverageExecutionTime ──────────────────────────────────────────────────

func TestSOService_GetAverageExecutionTime_Empty(t *testing.T) {
	svc := buildSvc(workMocks.NewWorkRepository(t), supplyMocks.NewSupplyService(t), vehicleMocks.NewVehicleService(t), &stubCustomerFinder{})
	result, err := svc.GetAverageExecutionTime(context.Background(), []string{"w1"})
	require.NoError(t, err)
	assert.Empty(t, result)
}
