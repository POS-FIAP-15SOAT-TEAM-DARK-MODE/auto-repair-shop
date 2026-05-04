//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateWork(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Teste de compressão",
		"description": "Verificação de compressão dos cilindros do motor",
		"price":       "350.00",
		"status":      "ACTIVE",
	}, attendantToken(t))

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	w := decodeJSON[workResponse](t, resp)
	assert.NotEmpty(t, w.ID)
	assert.Equal(t, "Teste de compressão", w.Name)
	assert.Equal(t, "ACTIVE", w.Status)
}

func TestCreateWork_InvalidPrice(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Trabalho Inválido",
		"description": "Descrição de trabalho inválido aqui",
		"price":       "not-a-number",
		"status":      "ACTIVE",
	}, attendantToken(t))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateWork_NegativePrice(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Preço negativo",
		"description": "Trabalho com preço negativo para teste",
		"price":       "-100.00",
		"status":      "ACTIVE",
	}, attendantToken(t))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateWork_EmptyName(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "",
		"description": "Descrição sem nome aqui para testar",
		"price":       "100.00",
		"status":      "ACTIVE",
	}, attendantToken(t))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateWork_InvalidStatus(t *testing.T) {
	resp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Work Bad Status",
		"description": "Uma descrição suficiente para o teste",
		"price":       "100.00",
		"status":      "UNKNOWN_STATUS",
	}, attendantToken(t))

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	resp.Body.Close()
}

func TestListWorks_ReturnsSeedData(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/works", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	page := decodeJSON[paginatedResponse[workResponse]](t, resp)
	assert.GreaterOrEqual(t, page.TotalItems, int64(5), "seed inserts 5 works")
}

func TestListWorks_Pagination(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/works?page=1&pageSize=2", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	page := decodeJSON[paginatedResponse[workResponse]](t, resp)
	assert.LessOrEqual(t, len(page.Items), 2)
	assert.Equal(t, int64(2), page.PageSize)
}

func TestListWorks_FilterByStatus(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/works?status=ACTIVE", nil, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	page := decodeJSON[paginatedResponse[workResponse]](t, resp)
	for _, w := range page.Items {
		assert.Equal(t, "ACTIVE", w.Status)
	}
}

func TestListWorks_Unauthorized_Customer(t *testing.T) {
	// CUSTOMER role cannot list works.
	tok, _ := func() (string, error) {
		return "", nil
	}()
	// Use an arbitrary customer user ID for the token.
	tok = customerToken(t, "some-customer-id")
	resp := doRequest(t, http.MethodGet, "/v1/works", nil, tok)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

func TestUpdateWork(t *testing.T) {
	// Create a work to update.
	createResp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Trabalho para atualizar",
		"description": "Descrição original do trabalho para teste",
		"price":       "200.00",
		"status":      "ACTIVE",
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	created := decodeJSON[workResponse](t, createResp)

	// Update it.
	resp := doRequest(t, http.MethodPut, "/v1/works/"+created.ID, map[string]any{
		"name":        "Trabalho atualizado",
		"description": "Descrição atualizada para o mesmo trabalho",
		"price":       "250.00",
		"status":      "ACTIVE",
	}, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	updated := decodeJSON[workResponse](t, resp)
	assert.Equal(t, "Trabalho atualizado", updated.Name)
	assert.Equal(t, "250", updated.Price)
}

func TestDeactivateWork(t *testing.T) {
	createResp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Trabalho para desativar",
		"description": "Descrição do trabalho que será desativado",
		"price":       "100.00",
		"status":      "ACTIVE",
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	created := decodeJSON[workResponse](t, createResp)

	resp := doRequest(t, http.MethodPut, "/v1/works/"+created.ID, map[string]any{
		"name":        created.Name,
		"description": "Descrição do trabalho que será desativado",
		"price":       created.Price,
		"status":      "INACTIVE",
	}, attendantToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	updated := decodeJSON[workResponse](t, resp)
	assert.Equal(t, "INACTIVE", updated.Status)
}

func TestDeleteWork(t *testing.T) {
	createResp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Trabalho para deletar",
		"description": "Descrição do trabalho a ser deletado agora",
		"price":       "150.00",
		"status":      "ACTIVE",
	}, attendantToken(t))
	require.Equal(t, http.StatusCreated, createResp.StatusCode)
	created := decodeJSON[workResponse](t, createResp)

	resp := doRequest(t, http.MethodDelete, "/v1/works/"+created.ID, nil, attendantToken(t))
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp.Body.Close()
}

func TestMechanicCanListWorks(t *testing.T) {
	resp := doRequest(t, http.MethodGet, "/v1/works", nil, mechanicToken(t))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestMechanicCannotCreateWork(t *testing.T) {
	// Works CREATE is AttendantRoles only (ATTENDANT, ADMIN).
	resp := doRequest(t, http.MethodPost, "/v1/works", map[string]any{
		"name":        "Mechanic Work",
		"description": "Mechanic não deve criar trabalhos na API",
		"price":       "100.00",
		"status":      "ACTIVE",
	}, mechanicToken(t))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}
