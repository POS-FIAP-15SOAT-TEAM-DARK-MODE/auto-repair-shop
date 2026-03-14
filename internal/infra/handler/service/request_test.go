package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
)

func TestCreateServiceReqDTO_Validate_InvalidPriceSpecialForms(t *testing.T) {
	tests := []struct {
		name        string
		price       string
		expectedErr error
	}{
		{
			name:        "minus only",
			price:       "-",
			expectedErr: domain.ErrInvalidPriceValue,
		},
		{
			name:        "dot only",
			price:       ".",
			expectedErr: domain.ErrInvalidPriceValue,
		},
		{
			name:        "minus dot",
			price:       "-.",
			expectedErr: domain.ErrInvalidPriceValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := &createServiceReqDTO{
				Name:        "Service",
				Description: "Valid description",
				Price:       tt.price,
				Status:      domain.ActiveString,
			}

			err := dto.Validate()
			if err != tt.expectedErr {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestCreateServiceReqDTO_Validate_NormalizesPrice(t *testing.T) {
	dto := &createServiceReqDTO{
		Name:        "Service",
		Description: "Valid description",
		Price:       "1,234.50",
		Status:      domain.ActiveString,
	}

	if err := dto.Validate(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got, want := dto.Price, "1234.50"; got != want {
		t.Fatalf("expected normalized price %q, got %q", want, got)
	}
}

func TestMapListParamsToDomain_Defaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	var captured *domain.ListServiceParams
	router.GET("/services", func(c *gin.Context) {
		captured = mapListParamsToDomain(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if captured.Page != 1 {
		t.Fatalf("expected default page 1, got %d", captured.Page)
	}

	if captured.PageSize != 10 {
		t.Fatalf("expected default page size 10, got %d", captured.PageSize)
	}

	if captured.Status != "" {
		t.Fatalf("expected empty status, got %q", captured.Status)
	}
}

func TestMapListParamsToDomain_CustomParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	var captured *domain.ListServiceParams
	router.GET("/services", func(c *gin.Context) {
		captured = mapListParamsToDomain(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/services?page=3&pageSize=20&status=ACTIVE", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if captured.Page != 3 {
		t.Fatalf("expected page 3, got %d", captured.Page)
	}

	if captured.PageSize != 20 {
		t.Fatalf("expected page size 20, got %d", captured.PageSize)
	}

	if captured.Status != "ACTIVE" {
		t.Fatalf("expected status %q, got %q", "ACTIVE", captured.Status)
	}
}

func TestMapListParamsToDomain_InvalidPageFallsBackToDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	var captured *domain.ListServiceParams
	router.GET("/services", func(c *gin.Context) {
		captured = mapListParamsToDomain(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/services?page=abc&pageSize=-5", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if captured.Page != 1 {
		t.Fatalf("expected fallback page 1, got %d", captured.Page)
	}

	if captured.PageSize != 10 {
		t.Fatalf("expected fallback page size 10, got %d", captured.PageSize)
	}
}

func TestMapListParamsToDomain_InvalidStatusIgnored(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	var captured *domain.ListServiceParams
	router.GET("/services", func(c *gin.Context) {
		captured = mapListParamsToDomain(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/services?status=UNKNOWN", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if captured.Status != "" {
		t.Fatalf("expected empty status for invalid value, got %q", captured.Status)
	}
}

func TestMapListParamsToDomain_InactiveStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	var captured *domain.ListServiceParams
	router.GET("/services", func(c *gin.Context) {
		captured = mapListParamsToDomain(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/services?status=INACTIVE", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if captured.Status != "INACTIVE" {
		t.Fatalf("expected status %q, got %q", "INACTIVE", captured.Status)
	}
}

func TestMapListParamsToDomain_ZeroPageFallsBackToDefault(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	var captured *domain.ListServiceParams
	router.GET("/services", func(c *gin.Context) {
		captured = mapListParamsToDomain(c)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/services?page=0&pageSize=0", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if captured.Page != 1 {
		t.Fatalf("expected fallback page 1 for page=0, got %d", captured.Page)
	}

	if captured.PageSize != 10 {
		t.Fatalf("expected fallback page size 10 for pageSize=0, got %d", captured.PageSize)
	}
}
