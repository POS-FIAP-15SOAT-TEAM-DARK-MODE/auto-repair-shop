package service_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	domainmocks "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
	inmemoryuow "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/in_memory"
	handler "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/handler/service"
	memrepo "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service"
	appservice "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/services/service"
)

func TestCreateServiceHandler_Success_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	uow := inmemoryuow.NewInMemoryUoW()
	repo := memrepo.MemoryRepository()
	srvc := appservice.Service(uow, repo)
	h := handler.Handler(srvc)

	router := gin.New()
	router.POST("/services", h.Create)

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

	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Status      string  `json:"status"`
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

	if resp.Price != 199.90 {
		t.Errorf("expected price %f, got %f", 199.90, resp.Price)
	}

	if resp.Status != "ACTIVE" {
		t.Errorf("expected status %q, got %q", "ACTIVE", resp.Status)
	}
}

func TestCreateServiceHandler_InvalidStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceService(t)
	h := handler.Handler(mockSvc)

	router := gin.New()
	router.POST("/services", h.Create)

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

	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// handler must not call service on DTO validation error
	mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	var resp struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error != domain.ErrInvalidStatusValue.Error() {
		t.Fatalf("expected error %q, got %q", domain.ErrInvalidStatusValue.Error(), resp.Error)
	}
}

func TestCreateServiceHandler_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceService(t)
	h := handler.Handler(mockSvc)

	router := gin.New()
	router.POST("/services", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewBufferString(`{invalid-json`))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// handler must not call service when JSON is invalid
	mockSvc.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)

	var resp struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestCreateServiceHandler_InvalidPrice_MapToDomainError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// svc will never be called because MapToDomain fails before service.Create is invoked.
	h := handler.Handler(nil)

	router := gin.New()
	router.POST("/services", h.Create)

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

	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	var resp struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal error response: %v", err)
	}

	if resp.Error == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestCreateServiceHandler_Unit_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceService(t)

	createdService, err := domain.NewService(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	mockSvc.
		EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Service")).
		Return(createdService, nil)

	h := handler.Handler(mockSvc)

	router := gin.New()
	router.POST("/services", h.Create)

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

	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Status      string  `json:"status"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.ID != createdService.ID {
		t.Errorf("expected ID %q, got %q", createdService.ID, resp.ID)
	}

	if resp.Name != createdService.Name {
		t.Errorf("expected name %q, got %q", createdService.Name, resp.Name)
	}

	if resp.Description != createdService.Description {
		t.Errorf("expected description %q, got %q", createdService.Description, resp.Description)
	}

	price, _ := createdService.Price.Float64()
	if resp.Price != price {
		t.Errorf("expected price %f, got %f", price, resp.Price)
	}

	if resp.Status != createdService.Status.String() {
		t.Errorf("expected status %q, got %q", createdService.Status.String(), resp.Status)
	}
}

func TestCreateServiceHandler_Unit_Success_Inactive(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceService(t)

	createdService, err := domain.NewService(
		"Wheel Alignment",
		"Complete wheel alignment",
		"150.00",
		domain.INACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	mockSvc.
		EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Service")).
		Return(createdService, nil)

	h := handler.Handler(mockSvc)

	router := gin.New()
	router.POST("/services", h.Create)

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

	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var resp struct {
		ID          string  `json:"id"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Status      string  `json:"status"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if resp.Status != createdService.Status.String() {
		t.Errorf("expected status %q, got %q", createdService.Status.String(), resp.Status)
	}
}

func TestCreateServiceHandler_Unit_ServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockSvc := domainmocks.NewServiceService(t)

	expectedErr := errors.New("service create failed")

	createdService, err := domain.NewService(
		"Oil Change",
		"Complete synthetic oil change",
		"199.90",
		domain.ACTIVE,
	)
	if err != nil {
		t.Fatalf("failed to create domain service: %v", err)
	}

	mockSvc.
		EXPECT().
		Create(mock.Anything, mock.AnythingOfType("*domain.Service")).
		Return(createdService, expectedErr)

	h := handler.Handler(mockSvc)

	router := gin.New()
	router.POST("/services", h.Create)

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

	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(payload))
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
