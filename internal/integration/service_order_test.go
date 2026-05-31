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
