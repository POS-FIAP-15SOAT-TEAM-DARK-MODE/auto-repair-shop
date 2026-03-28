package vehicle

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain/mocks"
)

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           string
		mockSetup      func(m *mocks.VehicleService)
		expectedStatus int
	}{
		{
			name: "success",
			body: `{
				"license_plate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customer_id": "123"
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*domain.Vehicle")).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "validation error (missing fields)",
			body: `{
				"license_plate": "ABC1D23",
				"brand": "",
				"model": "",
				"year": 2020,
				"customer_id": ""
			}`,
			mockSetup:      func(m *mocks.VehicleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           `{"license_plate":`,
			mockSetup:      func(m *mocks.VehicleService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error",
			body: `{
				"license_plate": "ABC1D23",
				"brand": "chevrolet",
				"model": "onix",
				"year": 2020,
				"customer_id": "123"
			}`,
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					Create(mock.Anything, mock.AnythingOfType("*domain.Vehicle")).
					Return(errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewVehicleService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			h := HttpHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req

			handlerFunc := h.Create
			handlerFunc(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestHandler_FindByLicensePlate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const plateParam = "plate"

	tests := []struct {
		name           string
		param          string
		mockSetup      func(m *mocks.VehicleService)
		expectedStatus int
	}{
		{
			name:  "success",
			param: "abc-1d23",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					FindByLicensePlate(mock.Anything, "ABC1D23").
					Return(&domain.Vehicle{
						ID:           "1",
						LicensePlate: "ABC1D23",
						Brand:        "chevrolet",
						Model:        "onix",
						Year:         2020,
						CustomerId:   "123",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "not found error",
			param: "ABC1D23",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					FindByLicensePlate(mock.Anything, "ABC1D23").
					Return(nil, domain.ErrVehicleNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "service error",
			param: "ABC1D23",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					FindByLicensePlate(mock.Anything, "ABC1D23").
					Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:  "plate normalization",
			param: "  abc-1d23  ",
			mockSetup: func(m *mocks.VehicleService) {
				m.EXPECT().
					FindByLicensePlate(mock.Anything, "ABC1D23").
					Return(&domain.Vehicle{
						ID:           "1",
						LicensePlate: "ABC1D23",
						Brand:        "fiat",
						Model:        "uno",
						Year:         2010,
						CustomerId:   "456",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := mocks.NewVehicleService(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockService)
			}

			h := HttpHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// set param
			c.Params = gin.Params{
				{Key: plateParam, Value: tt.param},
			}

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			c.Request = req

			handlerFunc := h.FindByLicensePlate
			handlerFunc(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
