//go:build integration

package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Setup helpers ────────────────────────────────────────────────────────────

// soFixture holds the IDs needed to run a service order through its lifecycle.
type soFixture struct {
	CustomerID     string
	CustomerUserID string
	VehicleID      string
	WorkID         string
	SupplyID       string
	SupplyQty      int
}

// newSOFixture creates a minimal fixture for service-order tests.
// Each call generates unique customer/vehicle data automatically.
func newSOFixture(t *testing.T) soFixture {
	t.Helper()

	email := nextEmail()
	customer := createCustomer(t, "SO Test Customer", email, nextCPF(), nextPhone())
	customerUserID := getUserID(t, email)

	vehicle := createVehicle(t, nextPlate(), "Toyota", "Corolla", 2021, customer.ID)

	// Use the first seeded work and supply.
	works := listWorks(t)
	require.NotEmpty(t, works, "seed must provide at least one work")

	supplies := listSupplies(t)
	require.NotEmpty(t, supplies, "seed must provide at least one supply")

	return soFixture{
		CustomerID:     customer.ID,
		CustomerUserID: customerUserID,
		VehicleID:      vehicle.ID,
		WorkID:         works[0].ID,
		SupplyID:       supplies[0].ID,
		SupplyQty:      1,
	}
}

// ─── Creation tests ───────────────────────────────────────────────────────────

func TestCreateServiceOrder(t *testing.T) {
	fix := newSOFixture(t)

	resp := doRequest(t, http.MethodPost, "/v1/service-order", map[string]string{
		"client":  fix.CustomerID,
		"vehicle": fix.VehicleID,
	}, attendantToken(t))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	so := decodeJSON[serviceOrderResponse](t, resp)
	assert.NotEmpty(t, so.ID)
	assert.Equal(t, "NEW", so.Status)
}

func TestCreateServiceOrder_CustomerNotFound(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/service-order", map[string]string{
		"client":  "00000000-0000-0000-0000-000000000000",
		"vehicle": "00000000-0000-0000-0000-000000000001",
	}, attendantToken(t))

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateServiceOrder_Unauthorized_Mechanic(t *testing.T) {
	fix := newSOFixture(t)

	resp := doRequest(t, http.MethodPost, "/v1/service-order", map[string]string{
		"client":  fix.CustomerID,
		"vehicle": fix.VehicleID,
	}, mechanicToken(t))

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

// ─── Work / Supply management ─────────────────────────────────────────────────

func TestAddWorksToServiceOrder(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)

	addWorkToSO(t, so.ID, []string{fix.WorkID}, attendantToken(t))

	// Verify works appear in the service order.
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("/v1/service-order/%s/works", so.ID), nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var works struct {
		Items []workResponse `json:"items"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&works))
	resp.Body.Close()
	assert.Len(t, works.Items, 1)
	assert.Equal(t, fix.WorkID, works.Items[0].ID)
}

func TestRemoveWorkFromServiceOrder(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	addWorkToSO(t, so.ID, []string{fix.WorkID}, attendantToken(t))

	resp := doRequest(t, http.MethodDelete,
		fmt.Sprintf("/v1/service-order/%s/works/%s", so.ID, fix.WorkID), nil, attendantToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestAddSupplies_DecrementsStock(t *testing.T) {
	fix := newSOFixture(t)

	// Only applicable if the supply isn't already used in a conflicting test.
	// To avoid interference, use a dedicated supply created for this test.
	createSupplyResp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Rolamento dianteiro",
		"description":   "Rolamento dianteiro para testes de estoque e integração",
		"unitPrice":     "45.00",
		"stockQuantity": 10,
	}, mechanicToken(t))
	require.Equal(t, http.StatusCreated, createSupplyResp.StatusCode)
	supply := decodeJSON[supplyResponse](t, createSupplyResp)

	stockBefore := getSupplyStock(t, supply.ID)
	assert.Equal(t, 10, stockBefore)

	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	// AddSupplies requires IN_DIAGNOSIS status.
	receiveServiceOrder(t, so.ID)
	startDiagnosis(t, so.ID)
	addSupplyToSO(t, so.ID, supply.ID, 3, mechanicToken(t))

	stockAfter := getSupplyStock(t, supply.ID)
	assert.Equal(t, 7, stockAfter, "stock should decrease by the added amount")
}

func TestRemoveSupply_RestoresStock(t *testing.T) {
	fix := newSOFixture(t)

	createSupplyResp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Tensor de correia",
		"description":   "Tensor de correia dentada para teste de restauração de estoque",
		"unitPrice":     "85.00",
		"stockQuantity": 5,
	}, mechanicToken(t))
	require.Equal(t, http.StatusCreated, createSupplyResp.StatusCode)
	supply := decodeJSON[supplyResponse](t, createSupplyResp)

	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	// AddSupplies and RemoveSupply require IN_DIAGNOSIS status.
	receiveServiceOrder(t, so.ID)
	startDiagnosis(t, so.ID)
	addSupplyToSO(t, so.ID, supply.ID, 2, mechanicToken(t))
	assert.Equal(t, 3, getSupplyStock(t, supply.ID))

	// Remove supply → stock must be restored.
	resp := doRequest(t, http.MethodDelete,
		fmt.Sprintf("/v1/service-order/%s/supplies/%s", so.ID, supply.ID), nil, mechanicToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	assert.Equal(t, 5, getSupplyStock(t, supply.ID), "stock should be restored after removal")
}

func TestAddSupply_OutOfStock(t *testing.T) {
	fix := newSOFixture(t)

	createSupplyResp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Suprimento limitado",
		"description":   "Suprimento com apenas 1 unidade em estoque para o teste",
		"unitPrice":     "200.00",
		"stockQuantity": 1,
	}, mechanicToken(t))
	require.Equal(t, http.StatusCreated, createSupplyResp.StatusCode)
	supply := decodeJSON[supplyResponse](t, createSupplyResp)

	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	// AddSupplies requires IN_DIAGNOSIS status.
	receiveServiceOrder(t, so.ID)
	startDiagnosis(t, so.ID)

	// Request more than in stock.
	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/v1/service-order/%s/supplies", so.ID),
		map[string]any{
			"supplies": []map[string]any{{"id": supply.ID, "amount": 5}},
		}, mechanicToken(t))
	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	resp.Body.Close()
}

func TestCannotAddWorkToNonNewSO(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)

	// Advance to RECEIVED.
	receiveServiceOrder(t, so.ID)

	// Attempt to add work to a non-NEW order must fail.
	// ErrServiceOrderNotNew is mapped to 409 Conflict.
	resp := doRequest(t, http.MethodPost, fmt.Sprintf("/v1/service-order/%s/works", so.ID),
		map[string]any{"services": []string{fix.WorkID}}, attendantToken(t))
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()
}

// ─── Full happy-path lifecycle ────────────────────────────────────────────────

// TestServiceOrder_FullLifecycle walks a service order all the way from NEW to DELIVERED,
// verifying the status at each transition and checking that the history accumulates.
func TestServiceOrder_FullLifecycle(t *testing.T) {
	fix := newSOFixture(t)

	// 1. Create SO → NEW.
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	assert.Equal(t, "NEW", so.Status)

	// 2. Add work (ATTENDANT, requires NEW status).
	addWorkToSO(t, so.ID, []string{fix.WorkID}, attendantToken(t))

	// 3. Receive (NEW → RECEIVED).
	receiveServiceOrder(t, so.ID)
	assert.Equal(t, "RECEIVED", fetchSOStatus(t, so.ID))

	// 4. Start diagnosis (RECEIVED → IN_DIAGNOSIS).
	startDiagnosis(t, so.ID)
	assert.Equal(t, "IN_DIAGNOSIS", fetchSOStatus(t, so.ID))

	// 5. Send for customer approval (IN_DIAGNOSIS → AWAITING_APPROVAL).
	sendForApproval(t, so.ID)
	assert.Equal(t, "AWAITING_APPROVAL", fetchSOStatus(t, so.ID))

	// 6. Customer accepts (AWAITING_APPROVAL → IN_PROGRESS).
	acceptServiceOrder(t, so.ID, fix.CustomerUserID)
	assert.Equal(t, "IN_PROGRESS", fetchSOStatus(t, so.ID))

	// 7. Finish (IN_PROGRESS → COMPLETED).
	finishServiceOrder(t, so.ID)
	assert.Equal(t, "COMPLETED", fetchSOStatus(t, so.ID))

	// 8. Deliver (COMPLETED → DELIVERED).
	deliverServiceOrder(t, so.ID)
	assert.Equal(t, "DELIVERED", fetchSOStatus(t, so.ID))

	// 9. Verify history has all transitions.
	histResp := doRequest(t, http.MethodGet, fmt.Sprintf("/v1/service-order/%s/history", so.ID), nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, histResp.StatusCode)

	var hist []struct {
		NewStatus string `json:"newStatus"`
	}
	require.NoError(t, json.NewDecoder(histResp.Body).Decode(&hist))
	histResp.Body.Close()

	// At least 6 entries: NEW→RECEIVED→IN_DIAGNOSIS→AWAITING_APPROVAL→IN_PROGRESS→COMPLETED→DELIVERED
	assert.GreaterOrEqual(t, len(hist), 6)
}

// ─── Rejection flow ───────────────────────────────────────────────────────────

func TestServiceOrder_RejectionFlow(t *testing.T) {
	fix := newSOFixture(t)

	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	addWorkToSO(t, so.ID, []string{fix.WorkID}, attendantToken(t))
	receiveServiceOrder(t, so.ID)
	startDiagnosis(t, so.ID)
	sendForApproval(t, so.ID)
	assert.Equal(t, "AWAITING_APPROVAL", fetchSOStatus(t, so.ID))

	// Customer rejects.
	tok := customerToken(t, fix.CustomerUserID)
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("/v1/service-order/%s/reject", so.ID), nil, tok)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()

	assert.Equal(t, "REJECTED", fetchSOStatus(t, so.ID))
}

// ─── Cancellation flow ────────────────────────────────────────────────────────

func TestServiceOrder_CancelFromNew(t *testing.T) {
	fix := newSOFixture(t)

	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	assert.Equal(t, "NEW", so.Status)

	resp := doRequest(t, http.MethodPut, fmt.Sprintf("/v1/service-order/%s/cancel", so.ID), nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Verify cancelled status via the public status endpoint.
	assert.Equal(t, "CANCELLED", fetchSOStatus(t, so.ID))
}

func TestServiceOrder_CannotCancelDelivered(t *testing.T) {
	fix := newSOFixture(t)

	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	addWorkToSO(t, so.ID, []string{fix.WorkID}, attendantToken(t))
	receiveServiceOrder(t, so.ID)
	startDiagnosis(t, so.ID)
	sendForApproval(t, so.ID)
	acceptServiceOrder(t, so.ID, fix.CustomerUserID)
	finishServiceOrder(t, so.ID)
	deliverServiceOrder(t, so.ID)

	// Cannot cancel a DELIVERED order. ErrServiceOrderNotCancelable maps to 409 Conflict.
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("/v1/service-order/%s/cancel", so.ID), nil, attendantToken(t))
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	resp.Body.Close()
}

// ─── Budget calculation ───────────────────────────────────────────────────────

func TestServiceOrder_BudgetCalculated(t *testing.T) {
	fix := newSOFixture(t)

	// Create a work with a known price.
	workResp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Alinhamento de precisão",
		"description": "Serviço de alinhamento de precisão para cálculo de orçamento",
		"price":       "300.00",
		"status":      "ACTIVE",
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, workResp.StatusCode)
	work := decodeJSON[workResponse](t, workResp)

	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	addWorkToSO(t, so.ID, []string{work.ID}, attendantToken(t))

	// Get full SO detail to inspect total amount.
	detailResp := doRequest(t, http.MethodGet, fmt.Sprintf("/v1/service-order/%s", so.ID), nil, attendantToken(t))
	require.Equal(t, http.StatusOK, detailResp.StatusCode)

	var detail struct {
		TotalValue float64 `json:"totalValue"`
		Status     string  `json:"status"`
	}
	require.NoError(t, json.NewDecoder(detailResp.Body).Decode(&detail))
	detailResp.Body.Close()

	assert.InDelta(t, 300.0, detail.TotalValue, 0.01, "total amount should equal work price")
}

// ─── List service orders ──────────────────────────────────────────────────────

func TestListServiceOrders(t *testing.T) {
	fix := newSOFixture(t)
	createServiceOrder(t, fix.CustomerID, fix.VehicleID)

	resp := doRequest(t, http.MethodGet, "/v1/service-order", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var page struct {
		Items      []serviceOrderResponse `json:"items"`
		TotalItems int64                  `json:"totalItems"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
	resp.Body.Close()
	assert.GreaterOrEqual(t, page.TotalItems, int64(1))
}

func TestListServiceOrders_FilterByStatus(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	receiveServiceOrder(t, so.ID)

	resp := doRequest(t, http.MethodGet, "/v1/service-order?status=RECEIVED", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var page struct {
		Items []serviceOrderResponse `json:"items"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
	resp.Body.Close()

	for _, item := range page.Items {
		assert.Equal(t, "RECEIVED", item.Status)
	}
}

// ─── GetStatus (public) ───────────────────────────────────────────────────────

func TestGetSOStatus_NoAuthRequired(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)

	// No auth token.
	resp := doRequest(t, http.MethodGet, fmt.Sprintf("/v1/service-order/%s/status", so.ID), nil, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	resp.Body.Close()
	assert.Equal(t, "NEW", result.Status)
}

// ─── RBAC enforcement ─────────────────────────────────────────────────────────

func TestSOReceive_Unauthorized_Mechanic(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)

	resp := doRequest(t, http.MethodPut, fmt.Sprintf("/v1/service-order/%s/received", so.ID), nil, mechanicToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestSOAccept_Unauthorized_Attendant(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	addWorkToSO(t, so.ID, []string{fix.WorkID}, attendantToken(t))
	receiveServiceOrder(t, so.ID)
	startDiagnosis(t, so.ID)
	sendForApproval(t, so.ID)

	// Attendant cannot accept — only CUSTOMER (and ADMIN).
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("/v1/service-order/%s/accept", so.ID), nil, attendantToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestSOStartDiagnosis_Unauthorized_Attendant(t *testing.T) {
	fix := newSOFixture(t)
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	receiveServiceOrder(t, so.ID)

	// Attendant cannot start diagnosis — only MECHANIC (and ADMIN).
	resp := doRequest(t, http.MethodPut, fmt.Sprintf("/v1/service-order/%s/start-diagnosis", so.ID), nil, attendantToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

// ─── Feature 1: Works + supplies at SO creation ───────────────────────────────

// TestCreateSOWithWorksAndSupplies verifies that works and supplies provided at
// SO creation time are atomically attached and stock is decremented.
func TestCreateSOWithWorksAndSupplies(t *testing.T) {
	fix := newSOFixture(t)

	// Create a dedicated supply with known stock so we can check the decrement.
	createSupplyResp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Filtro de combustivel",
		"description":   "Filtro de combustivel para teste de criacao de OS",
		"unitPrice":     "55.00",
		"stockQuantity": 20,
	}, mechanicToken(t))
	require.Equal(t, http.StatusCreated, createSupplyResp.StatusCode)
	supply := decodeJSON[supplyResponse](t, createSupplyResp)

	// Create a work with a known price for budget assertion.
	createWorkResp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Troca de filtro completo",
		"description": "Troca de filtro de combustivel do veiculo",
		"price":       "150.00",
		"status":      "ACTIVE",
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, createWorkResp.StatusCode)
	work := decodeJSON[workResponse](t, createWorkResp)

	stockBefore := getSupplyStock(t, supply.ID)
	require.Equal(t, 20, stockBefore)

	// Create SO with works and supplies in the request body.
	resp := doRequest(t, http.MethodPost, "/v1/service-order", map[string]any{
		"client":  fix.CustomerID,
		"vehicle": fix.VehicleID,
		"works":   []string{work.ID},
		"supplies": []map[string]any{
			{"id": supply.ID, "amount": 3},
		},
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	so := decodeJSON[serviceOrderResponse](t, resp)
	assert.NotEmpty(t, so.ID)
	assert.Equal(t, "NEW", so.Status)

	// Assert work is attached.
	worksResp := doRequest(t, http.MethodGet, fmt.Sprintf("/v1/service-order/%s/works", so.ID), nil, attendantToken(t))
	require.Equal(t, http.StatusOK, worksResp.StatusCode)
	var worksResult struct {
		Items []workResponse `json:"items"`
	}
	require.NoError(t, json.NewDecoder(worksResp.Body).Decode(&worksResult))
	worksResp.Body.Close()
	require.Len(t, worksResult.Items, 1)
	assert.Equal(t, work.ID, worksResult.Items[0].ID)

	// Assert supply is attached.
	suppliesResp := doRequest(t, http.MethodGet, fmt.Sprintf("/v1/service-order/%s/supplies", so.ID), nil, attendantToken(t))
	require.Equal(t, http.StatusOK, suppliesResp.StatusCode)
	var suppliesResult struct {
		Items []struct {
			ID     string `json:"id"`
			Amount int    `json:"amount"`
		} `json:"items"`
	}
	require.NoError(t, json.NewDecoder(suppliesResp.Body).Decode(&suppliesResult))
	suppliesResp.Body.Close()
	require.Len(t, suppliesResult.Items, 1)
	assert.Equal(t, supply.ID, suppliesResult.Items[0].ID)
	assert.Equal(t, 3, suppliesResult.Items[0].Amount)

	// Assert stock was decremented.
	stockAfter := getSupplyStock(t, supply.ID)
	assert.Equal(t, 17, stockAfter, "stock should decrease by the amount added at creation")
}

// TestCreateSOWithSupply_InsufficientStock verifies that creating an SO with a
// supply whose requested amount exceeds stock_quantity returns a non-2xx error.
func TestCreateSOWithSupply_InsufficientStock(t *testing.T) {
	fix := newSOFixture(t)

	// Create a supply with just 1 unit.
	createSupplyResp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Suprimento escasso criacao OS",
		"description":   "Suprimento com apenas 1 unidade para teste de insuficiencia",
		"unitPrice":     "100.00",
		"stockQuantity": 1,
	}, mechanicToken(t))
	require.Equal(t, http.StatusCreated, createSupplyResp.StatusCode)
	supply := decodeJSON[supplyResponse](t, createSupplyResp)

	// Attempt to create SO requesting 5 units (more than available).
	resp := doRequest(t, http.MethodPost, "/v1/service-order", map[string]any{
		"client":  fix.CustomerID,
		"vehicle": fix.VehicleID,
		"supplies": []map[string]any{
			{"id": supply.ID, "amount": 5},
		},
	}, attendantToken(t))

	// Should be a non-2xx error (ErrSupplyOutOfStock).
	assert.True(t, resp.StatusCode >= 400, "expected error status, got %d", resp.StatusCode)
	resp.Body.Close()

	// Stock must remain unchanged.
	assert.Equal(t, 1, getSupplyStock(t, supply.ID), "stock must not change on failed creation")
}

// TestCreateSOWithEmptyWorkID verifies that an empty string in the works list
// at creation time returns an error and no SO is persisted.
func TestCreateSOWithEmptyWorkID(t *testing.T) {
	fix := newSOFixture(t)

	resp := doRequest(t, http.MethodPost, "/v1/service-order", map[string]any{
		"client":  fix.CustomerID,
		"vehicle": fix.VehicleID,
		"works":   []string{""},
	}, attendantToken(t))

	assert.True(t, resp.StatusCode >= 400, "expected error status, got %d", resp.StatusCode)
	resp.Body.Close()
}

// ─── Feature 2: SO list filtering improvements ────────────────────────────────

// TestListServiceOrders_ExcludesTerminalStatuses verifies that the default list
// (no status filter) does not include SOs in terminal statuses.
func TestListServiceOrders_ExcludesTerminalStatuses(t *testing.T) {
	fix := newSOFixture(t)

	// Create and fully deliver an SO so it reaches a terminal status (DELIVERED).
	so := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	addWorkToSO(t, so.ID, []string{fix.WorkID}, attendantToken(t))
	receiveServiceOrder(t, so.ID)
	startDiagnosis(t, so.ID)
	sendForApproval(t, so.ID)
	acceptServiceOrder(t, so.ID, fix.CustomerUserID)
	finishServiceOrder(t, so.ID)
	deliverServiceOrder(t, so.ID)
	assert.Equal(t, "DELIVERED", fetchSOStatus(t, so.ID))

	// Default list must NOT include the delivered SO.
	resp := doRequest(t, http.MethodGet, "/v1/service-order", nil, attendantToken(t))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var page struct {
		Items []serviceOrderResponse `json:"items"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
	resp.Body.Close()

	for _, item := range page.Items {
		assert.NotEqual(t, "DELIVERED", item.Status, "DELIVERED SO should not appear in default list")
		assert.NotEqual(t, "COMPLETED", item.Status, "COMPLETED SO should not appear in default list")
		assert.NotEqual(t, "REJECTED", item.Status, "REJECTED SO should not appear in default list")
		assert.NotEqual(t, "CANCELLED", item.Status, "CANCELLED SO should not appear in default list")
	}
}

// TestListServiceOrders_FilterByDelivered verifies that ?status=DELIVERED returns
// only DELIVERED SOs (terminal status filter works when explicitly requested).
func TestListServiceOrders_FilterByDelivered(t *testing.T) {
	fix := newSOFixture(t)

	// Advance one SO to DELIVERED.
	soDelivered := createServiceOrder(t, fix.CustomerID, fix.VehicleID)
	addWorkToSO(t, soDelivered.ID, []string{fix.WorkID}, attendantToken(t))
	receiveServiceOrder(t, soDelivered.ID)
	startDiagnosis(t, soDelivered.ID)
	sendForApproval(t, soDelivered.ID)
	acceptServiceOrder(t, soDelivered.ID, fix.CustomerUserID)
	finishServiceOrder(t, soDelivered.ID)
	deliverServiceOrder(t, soDelivered.ID)

	resp := doRequest(t, http.MethodGet, "/v1/service-order?status=DELIVERED", nil, attendantToken(t))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var page struct {
		Items []serviceOrderResponse `json:"items"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
	resp.Body.Close()

	require.NotEmpty(t, page.Items, "at least one DELIVERED SO should appear")
	for _, item := range page.Items {
		assert.Equal(t, "DELIVERED", item.Status)
	}

	found := false
	for _, item := range page.Items {
		if item.ID == soDelivered.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "the delivered SO should appear in the filtered list")
}

// TestListServiceOrders_SortByStatus verifies that ?sort_by=status returns SOs in
// workflow-priority order: IN_PROGRESS before AWAITING_APPROVAL before IN_DIAGNOSIS.
func TestListServiceOrders_SortByStatus(t *testing.T) {
	// We need at least three SOs in distinct active statuses to verify ordering.
	fix1 := newSOFixture(t)
	fix2 := newSOFixture(t)
	fix3 := newSOFixture(t)

	// SO-1 → IN_PROGRESS (highest priority).
	so1 := createServiceOrder(t, fix1.CustomerID, fix1.VehicleID)
	addWorkToSO(t, so1.ID, []string{fix1.WorkID}, attendantToken(t))
	receiveServiceOrder(t, so1.ID)
	startDiagnosis(t, so1.ID)
	sendForApproval(t, so1.ID)
	acceptServiceOrder(t, so1.ID, fix1.CustomerUserID)
	assert.Equal(t, "IN_PROGRESS", fetchSOStatus(t, so1.ID))

	// SO-2 → AWAITING_APPROVAL.
	so2 := createServiceOrder(t, fix2.CustomerID, fix2.VehicleID)
	addWorkToSO(t, so2.ID, []string{fix2.WorkID}, attendantToken(t))
	receiveServiceOrder(t, so2.ID)
	startDiagnosis(t, so2.ID)
	sendForApproval(t, so2.ID)
	assert.Equal(t, "AWAITING_APPROVAL", fetchSOStatus(t, so2.ID))

	// SO-3 → IN_DIAGNOSIS.
	so3 := createServiceOrder(t, fix3.CustomerID, fix3.VehicleID)
	addWorkToSO(t, so3.ID, []string{fix3.WorkID}, attendantToken(t))
	receiveServiceOrder(t, so3.ID)
	startDiagnosis(t, so3.ID)
	assert.Equal(t, "IN_DIAGNOSIS", fetchSOStatus(t, so3.ID))

	resp := doRequest(t, http.MethodGet, "/v1/service-order?sort_by=status&pageSize=100", nil, attendantToken(t))
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var page struct {
		Items []serviceOrderResponse `json:"items"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&page))
	resp.Body.Close()

	// Extract positions of our three SOs in the result.
	posOf := func(id string) int {
		for i, item := range page.Items {
			if item.ID == id {
				return i
			}
		}
		return -1
	}

	pos1 := posOf(so1.ID)
	pos2 := posOf(so2.ID)
	pos3 := posOf(so3.ID)

	require.GreaterOrEqual(t, pos1, 0, "IN_PROGRESS SO must be in the list")
	require.GreaterOrEqual(t, pos2, 0, "AWAITING_APPROVAL SO must be in the list")
	require.GreaterOrEqual(t, pos3, 0, "IN_DIAGNOSIS SO must be in the list")

	assert.Less(t, pos1, pos2, "IN_PROGRESS should appear before AWAITING_APPROVAL")
	assert.Less(t, pos2, pos3, "AWAITING_APPROVAL should appear before IN_DIAGNOSIS")
}
