package work

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	inmemoryuow "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/in_memory"
	memrepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	appwork "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/work"
	"github.com/shopspring/decimal"
)

func TestCreateWorkHandler_Success_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uow := inmemoryuow.NewInMemoryUoW()
	repo := memrepo.MemoryRepository()
	workService := appwork.Service(uow, repo)
	h := HttpHandler(workService)

	router := gin.New()
	router.POST("/works", h.Create())

	body := map[string]any{
		"name":        "Oil Change",
		"description": "Complete synthetic oil change",
		"price":       "199.90",
		"status":      "ACTIVE",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       string `json:"price"`
		Status      string `json:"status"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.ID == "" {
		t.Errorf("expected non-empty ID in response")
	}

	if resp.Name != body["name"] {
		t.Errorf("expected name %q, got %q", body["name"], resp.Name)
	}

	if resp.Description != body["description"] {
		t.Errorf("expected description %q, got %q", body["description"], resp.Description)
	}

	// Accept possible "199.9" or "199.90" response due to decimal canonicalization
	expectedPrice := "199.90"
	if resp.Price != expectedPrice && resp.Price != "199.9" {
		t.Errorf("expected price %s or %s, got %s", expectedPrice, "199.9", resp.Price)
	}

	if resp.Status != "ACTIVE" {
		t.Errorf("expected status %q, got %q", "ACTIVE", resp.Status)
	}
}

func TestCreateWorkHandler_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWork := domainmocks.NewWorkService(t)
	h := HttpHandler(mockWork)

	router := gin.New()
	router.POST("/works", h.Create())

	body := map[string]any{
		"name":        "Alignment",
		"description": "Wheel alignment",
		"price":       "150.00",
		"status":      "UNKNOWN",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// handler must not call service on DTO validation error
	mockWork.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	var resp struct {
		Error  string   `json:"error"`
		Errors []string `json:"errors"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	expectedMsg := domain.ErrInvalidWorkStatusValue.Error()
	got := resp.Error
	if len(resp.Errors) > 0 {
		got = resp.Errors[0]
	}
	if got != expectedMsg {
		t.Fatalf("expected error %q, got %q", expectedMsg, got)
	}
}

func TestCreateWorkHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWork := domainmocks.NewWorkService(t)
	h := HttpHandler(mockWork)

	router := gin.New()
	router.POST("/works", h.Create())

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewBufferString(`{invalid-json`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// handler must not call service when JSON is invalid
	mockWork.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	var resp struct {
		Error  string   `json:"error"`
		Errors []string `json:"errors"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error == "" && len(resp.Errors) == 0 {
		t.Fatalf("expected non-empty error message")
	}
}

func TestCreateWorkHandler_InvalidPrice_MapToDomainError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// svc will never be called because Validate catches invalid price before MapToDomain is invoked.
	h := HttpHandler(nil)

	router := gin.New()
	router.POST("/works", h.Create())

	body := map[string]any{
		"name":        "Oil Change",
		"description": "Complete synthetic oil change",
		"price":       "invalid-price",
		"status":      "ACTIVE",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp struct {
		Error  string   `json:"error"`
		Errors []string `json:"errors"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error == "" && len(resp.Errors) == 0 {
		t.Fatalf("expected non-empty error message")
	}
}

func TestCreateWorkHandler_Unit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWork := domainmocks.NewWorkService(t)

	createdWork, err := domain.NewWork(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain work: %v", err)
	}

	mockWork.
		EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Work")).
		Return(nil)

	h := HttpHandler(mockWork)

	router := gin.New()
	router.POST("/works", h.Create())

	body := map[string]any{
		"name":        "Oil Change",
		"description": "Complete synthetic oil change",
		"price":       "199.90",
		"status":      "ACTIVE",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       string `json:"price"`
		Status      string `json:"status"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.ID == "" {
		t.Errorf("expected non-empty ID in response")
	}

	if resp.Name != createdWork.Name {
		t.Errorf("expected name %q, got %q", createdWork.Name, resp.Name)
	}

	if resp.Description != createdWork.Description {
		t.Errorf("expected description %q, got %q", createdWork.Description, resp.Description)
	}

	if resp.Price != createdWork.Price.String() {
		t.Errorf("expected price %q, got %q", createdWork.Price.String(), resp.Price)
	}

	if resp.Status != createdWork.Status.String() {
		t.Errorf("expected status %q, got %q", createdWork.Status.String(), resp.Status)
	}
}

func TestCreateWorkHandler_Unit_Success_Inactive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWork := domainmocks.NewWorkService(t)

	createdWork, err := domain.NewWork(
		"Wheel Alignment",
		"Complete wheel alignment",
		"150.00",
		domain.INACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain work: %v", err)
	}

	mockWork.
		EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Work")).
		Return(nil)

	h := HttpHandler(mockWork)

	router := gin.New()
	router.POST("/works", h.Create())

	body := map[string]any{
		"name":        "Wheel Alignment",
		"description": "Complete wheel alignment",
		"price":       "150.00",
		"status":      "INACTIVE",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       string `json:"price"`
		Status      string `json:"status"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.Status != createdWork.Status.String() {
		t.Errorf("expected status %q, got %q", createdWork.Status.String(), resp.Status)
	}
}

func TestCreateWorkHandler_Unit_WorkError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWork := domainmocks.NewWorkService(t)

	expectedErr := errors.New("work create failed")

	mockWork.
		EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Work")).
		Return(expectedErr)

	h := HttpHandler(mockWork)

	router := gin.New()
	router.POST("/works", h.Create())

	body := map[string]any{
		"name":        "Oil Change",
		"description": "Complete synthetic oil change",
		"price":       "199.90",
		"status":      "ACTIVE",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var resp struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error != expectedErr.Error() {
		t.Fatalf("expected error %q, got %q", expectedErr.Error(), resp.Error)
	}
}

func TestListWorkHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWork := domainmocks.NewWorkService(t)

	expectedResponse := &domain.PaginatorResponse[domain.Work]{
		Items: []domain.Work{
			{ID: "1", Name: "Oil Change", Description: "Complete oil change", Price: decimal.NewFromInt(50), Status: domain.ACTIVE},
		},
		TotalItems: 1,
		TotalPages: 1,
		Page:       1,
		PageSize:   10,
	}

	mockWork.EXPECT().
		List(mock.Anything, mock.AnythingOfType("*domain.ListWorkParams")).
		Return(expectedResponse, nil)

	h := HttpHandler(mockWork)

	router := gin.New()
	router.GET("/works", h.List())

	req := httptest.NewRequest(http.MethodGet, "/works?page=1&pageSize=10", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp domain.PaginatorResponse[map[string]any]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.TotalItems != expectedResponse.TotalItems {
		t.Fatalf("expected total items %d, got %d", expectedResponse.TotalItems, resp.TotalItems)
	}

	if resp.Page != expectedResponse.Page {
		t.Fatalf("expected page %d, got %d", expectedResponse.Page, resp.Page)
	}

	if len(resp.Items) != len(expectedResponse.Items) {
		t.Fatalf("expected %d items, got %d", len(expectedResponse.Items), len(resp.Items))
	}
}

func TestListWorkHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockWork := domainmocks.NewWorkService(t)
	expectedErr := errors.New("list work failed")

	mockWork.EXPECT().
		List(mock.Anything, mock.AnythingOfType("*domain.ListWorkParams")).
		Return(nil, expectedErr)

	h := HttpHandler(mockWork)

	router := gin.New()
	router.GET("/works", h.List())

	req := httptest.NewRequest(http.MethodGet, "/works", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var resp struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error != expectedErr.Error() {
		t.Fatalf("expected error %q, got %q", expectedErr.Error(), resp.Error)
	}
}

func TestListWorkHandler_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uow := inmemoryuow.NewInMemoryUoW()
	repo := memrepo.MemoryRepository()
	workService := appwork.Service(uow, repo)
	h := HttpHandler(workService)

	router := gin.New()
	router.POST("/works", h.Create())
	router.GET("/works", h.List())

	createBody := map[string]any{
		"name":        "Oil Change",
		"description": "Complete synthetic oil change",
		"price":       "199.90",
		"status":      "ACTIVE",
	}
	payload, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest(http.MethodPost, "/works", bytes.NewReader(payload))
	createReq.Header.Set("Content-Type", "application/json")
	createRec := httptest.NewRecorder()
	router.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("setup: expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/works?page=1&pageSize=10&status=ACTIVE", nil)
	listRec := httptest.NewRecorder()
	router.ServeHTTP(listRec, listReq)

	if listRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, listRec.Code)
	}

	var resp domain.PaginatorResponse[map[string]any]
	if err := json.Unmarshal(listRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.TotalItems != 1 {
		t.Fatalf("expected 1 total item, got %d", resp.TotalItems)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
}

func TestUpdateWorkHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewWorkService(t)
	h := HttpHandler(mockService)

	mockService.EXPECT().Update(mock.Anything, mock.AnythingOfType("*domain.Work")).Return(nil)

	router := gin.New()
	router.PUT("/works/:id", h.Update())

	body := map[string]any{
		"name":        "Updated Name",
		"description": "Updated Description",
		"price":       "99.99",
		"status":      "ACTIVE",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/works/work-id", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUpdateWorkHandler_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := HttpHandler(nil)

	router := gin.New()
	router.PUT("/works/:id", h.Update())

	body := map[string]any{
		"name":        "Valid Name",
		"description": "Valid Description",
		"price":       "99.99",
		"status":      "INVALID",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/works/work-id", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateWorkHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewWorkService(t)
	h := HttpHandler(mockService)

	mockService.EXPECT().Update(mock.Anything, mock.AnythingOfType("*domain.Work")).Return(errors.New("internal error"))

	router := gin.New()
	router.PUT("/works/:id", h.Update())

	body := map[string]any{
		"name":        "Updated Name",
		"description": "Updated Description",
		"price":       "99.99",
		"status":      "ACTIVE",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/works/work-id", bytes.NewBuffer(jsonBody))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestDeleteWorkHandler_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewWorkService(t)
	h := HttpHandler(mockService)

	mockService.EXPECT().Delete(mock.Anything, "work-id").Return(nil)

	router := gin.New()
	router.DELETE("/works/:id", h.Delete())

	req := httptest.NewRequest(http.MethodDelete, "/works/work-id", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteWorkHandler_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := domainmocks.NewWorkService(t)
	h := HttpHandler(mockService)

	mockService.EXPECT().Delete(mock.Anything, "work-id").Return(errors.New("internal error"))

	router := gin.New()
	router.DELETE("/works/:id", h.Delete())

	req := httptest.NewRequest(http.MethodDelete, "/works/work-id", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestUpdateWorkHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := HttpHandler(nil)

	router := gin.New()
	router.PUT("/works/:id", h.Update())

	req := httptest.NewRequest(http.MethodPut, "/works/work-id", bytes.NewBufferString(`{invalid-json`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateWorkHandler_InvalidPrice(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := HttpHandler(nil)

	router := gin.New()
	router.PUT("/works/:id", h.Update())

	body := map[string]any{
		"name":        "Oil Change",
		"description": "Complete synthetic oil change",
		"price":       "invalid-price",
		"status":      "ACTIVE",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/works/work-id", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdateWorkHandler_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := HttpHandler(nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodPut, "/works/", bytes.NewBufferString(`{"name":"N","description":"D","price":"10.00","status":"ACTIVE"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Update()(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestDeleteWorkHandler_EmptyID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := HttpHandler(nil)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodDelete, "/works/", nil)

	h.Delete()(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreate_InvalidPriceReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &handler{svc: nil}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodPost, "/works", bytes.NewBufferString(
		`{"name":"Oil Change","description":"Complete synthetic oil change","price":"not-a-decimal","status":"ACTIVE"}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")

	h.Create()(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUpdate_InvalidPriceReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &handler{svc: nil}

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request, _ = http.NewRequest(http.MethodPut, "/works/work-id", bytes.NewBufferString(
		`{"name":"Oil Change","description":"Complete synthetic oil change","price":"not-a-decimal","status":"ACTIVE"}`,
	))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{{Key: "id", Value: "work-id"}}

	h.Update()(c)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
