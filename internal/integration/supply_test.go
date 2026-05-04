//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSupply(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Vela de ignição",
		"description":   "Jogo de velas de ignição para motores 1.0 aspirados",
		"unitPrice":     "89.90",
		"stockQuantity": 20,
	}, attendantToken(t))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	s := decodeJSON[supplyResponse](t, resp)
	assert.NotEmpty(t, s.ID)
	assert.Equal(t, "Vela de ignição", s.Name)
	assert.Equal(t, 20, s.StockQuantity)
}

func TestCreateSupply_MissingName(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"description":   "Suprimento sem nome definido aqui",
		"unitPrice":     "50.00",
		"stockQuantity": 10,
	}, attendantToken(t))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateSupply_NegativeStock(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Estoque negativo",
		"description":   "Suprimento com quantidade de estoque negativa",
		"unitPrice":     "50.00",
		"stockQuantity": -1,
	}, attendantToken(t))

	// Domain rejects negative stock → 400 or 422.
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity)
	resp.Body.Close()
}

func TestListSupplies_ReturnsSeedData(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/supplies", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	page := decodeJSON[paginatedResponse[supplyResponse]](t, resp)
	assert.GreaterOrEqual(t, page.TotalItems, int64(5), "seed inserts 5 supplies")
}

func TestListSupplies_Pagination(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/supplies?page=1&pageSize=3", nil, mechanicToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	page := decodeJSON[paginatedResponse[supplyResponse]](t, resp)
	assert.LessOrEqual(t, len(page.Items), 3)
}

func TestUpdateSupply(t *testing.T) {
	createResp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Suprimento para atualizar",
		"description":   "Descrição do suprimento que vai ser atualizado",
		"unitPrice":     "100.00",
		"stockQuantity": 10,
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	created := decodeJSON[supplyResponse](t, createResp)

	resp := doRequest(t, http.MethodPut, "/v1/supplies/"+created.ID, map[string]any{
		"name":          "Suprimento atualizado com sucesso",
		"description":   "Descrição atualizada do suprimento para o teste",
		"unitPrice":     "120.00",
		"stockQuantity": 15,
	}, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	updated := decodeJSON[supplyResponse](t, resp)
	assert.Equal(t, "Suprimento atualizado com sucesso", updated.Name)
	assert.Equal(t, 15, updated.StockQuantity)
}

func TestDeleteSupply(t *testing.T) {
	createResp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Suprimento para deletar",
		"description":   "Este suprimento será deletado no teste de integração",
		"unitPrice":     "50.00",
		"stockQuantity": 5,
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	created := decodeJSON[supplyResponse](t, createResp)

	resp := doRequest(t, http.MethodDelete, "/v1/supplies/"+created.ID, nil, attendantToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestSupply_Unauthorized_Customer(t *testing.T) {
	tok := customerToken(t, "some-customer-id")
	resp := doRequest(t, http.MethodGet, "/v1/supplies", nil, tok)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestMechanicCanCreateAndListSupplies(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/supplies", map[string]any{
		"name":          "Suprimento do mecânico",
		"description":   "Mecânico pode criar suprimentos conforme regra de negócio",
		"unitPrice":     "75.00",
		"stockQuantity": 8,
	}, mechanicToken(t))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	s := decodeJSON[supplyResponse](t, resp)
	assert.NotEmpty(t, s.ID)

	listResp := doRequest(t, http.MethodGet, "/v1/supplies", nil, mechanicToken(t))
	assert.Equal(t, http.StatusOK, listResp.StatusCode)
	listResp.Body.Close()
}
