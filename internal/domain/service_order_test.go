package domain_test

import (
	"testing"
	"time"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestServiceOrderStatus_String(t *testing.T) {
	tests := []struct {
		status domain.SERVICE_ORDER_STATUS
		want   string
	}{
		{domain.SERVICE_ORDER_STATUS_NEW, "NEW"},
		{domain.SERVICE_ORDER_STATUS_RECEIVED, "RECEIVED"},
		{domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, "IN_DIAGNOSIS"},
		{domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, "AWAITING_APPROVAL"},
		{domain.SERVICE_ORDER_STATUS_REJECTED, "REJECTED"},
		{domain.SERVICE_ORDER_STATUS_IN_PROGRESS, "IN_PROGRESS"},
		{domain.SERVICE_ORDER_STATUS_COMPLETED, "COMPLETED"},
		{domain.SERVICE_ORDER_STATUS_DELIVERED, "DELIVERED"},
		{domain.SERVICE_ORDER_STATUS_CANCELLED, "CANCELLED"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.String())
		})
	}
}

func TestServiceOrderStatus_IsCancelable(t *testing.T) {
	tests := []struct {
		status     domain.SERVICE_ORDER_STATUS
		cancelable bool
	}{
		{domain.SERVICE_ORDER_STATUS_NEW, true},
		{domain.SERVICE_ORDER_STATUS_RECEIVED, true},
		{domain.SERVICE_ORDER_STATUS_IN_DIAGNOSIS, true},
		{domain.SERVICE_ORDER_STATUS_AWAITING_APPROVAL, true},
		{domain.SERVICE_ORDER_STATUS_IN_PROGRESS, true},
		{domain.SERVICE_ORDER_STATUS_REJECTED, false},
		{domain.SERVICE_ORDER_STATUS_COMPLETED, false},
		{domain.SERVICE_ORDER_STATUS_DELIVERED, false},
		{domain.SERVICE_ORDER_STATUS_CANCELLED, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.cancelable, tt.status.IsCancelable())
		})
	}
}

func TestStringToServiceOrderStatus(t *testing.T) {
	tests := []struct {
		input string
		want  domain.SERVICE_ORDER_STATUS
	}{
		{"NEW", domain.SERVICE_ORDER_STATUS_NEW},
		{"IN_PROGRESS", domain.SERVICE_ORDER_STATUS_IN_PROGRESS},
		{"DELIVERED", domain.SERVICE_ORDER_STATUS_DELIVERED},
		{"", domain.SERVICE_ORDER_STATUS("")},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.Equal(t, tt.want, domain.StringToServiceOrderStatus(tt.input))
		})
	}
}

func TestNewServiceOrder(t *testing.T) {
	customer := &domain.Customer{ID: "cust-1"}
	vehicle := &domain.Vehicle{ID: "veh-1"}

	so := domain.NewServiceOrder(customer, vehicle)

	assert.NotEmpty(t, so.ID)
	assert.Equal(t, domain.SERVICE_ORDER_STATUS_NEW, so.Status)
	assert.Equal(t, customer, so.Customer)
	assert.Equal(t, vehicle, so.Vehicle)
	assert.True(t, so.TotalAmount.Equal(decimal.Zero))
}

func TestNewHistoryServiceOrderID(t *testing.T) {
	id1 := domain.NewHistoryServiceOrderID()
	id2 := domain.NewHistoryServiceOrderID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestServiceOrder_PrepareForSum(t *testing.T) {
	so := &domain.ServiceOrder{}
	// Should not panic when called multiple times
	so.PrepareForSum()
	so.PrepareForSum()
}

func TestServiceOrder_ResetPricing(t *testing.T) {
	work := domain.Work{Price: decimal.NewFromInt(100)}
	so := &domain.ServiceOrder{}
	so.SumWorkValue(work)

	assert.True(t, so.TotalAmount.Equal(decimal.NewFromInt(100)))

	so.ResetPricing()
	assert.True(t, so.TotalAmount.Equal(decimal.Zero))
}

func TestServiceOrder_SumWorkValue(t *testing.T) {
	tests := []struct {
		name      string
		works     []domain.Work
		wantTotal string
	}{
		{
			name:      "single work",
			works:     []domain.Work{{Price: decimal.NewFromInt(150)}},
			wantTotal: "150",
		},
		{
			name: "multiple works",
			works: []domain.Work{
				{Price: decimal.NewFromInt(100)},
				{Price: decimal.NewFromFloat(50.50)},
			},
			wantTotal: "150.5",
		},
		{
			name:      "zero price work",
			works:     []domain.Work{{Price: decimal.Zero}},
			wantTotal: "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := &domain.ServiceOrder{}
			for _, w := range tt.works {
				so.SumWorkValue(w)
			}
			assert.Equal(t, tt.wantTotal, so.TotalAmount.String())
		})
	}
}

func TestServiceOrder_SumSupplyValue(t *testing.T) {
	tests := []struct {
		name      string
		supply    domain.Supply
		wantTotal string
	}{
		{
			name: "supply with quantity 2 and price 10",
			supply: domain.Supply{
				UnitPrice:     decimal.NewFromInt(10),
				StockQuantity: 2,
			},
			wantTotal: "20",
		},
		{
			name: "supply with quantity 1",
			supply: domain.Supply{
				UnitPrice:     decimal.NewFromFloat(9.99),
				StockQuantity: 1,
			},
			wantTotal: "9.99",
		},
		{
			name: "supply with quantity 0",
			supply: domain.Supply{
				UnitPrice:     decimal.NewFromInt(100),
				StockQuantity: 0,
			},
			wantTotal: "0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			so := &domain.ServiceOrder{}
			so.SumSupplyValue(tt.supply)
			assert.Equal(t, tt.wantTotal, so.TotalAmount.String())
		})
	}
}

func TestServiceOrder_ResetPricing_Complex(t *testing.T) {
	so := &domain.ServiceOrder{}
	so.SumWorkValue(domain.Work{Price: decimal.NewFromInt(100)})
	so.SumSupplyValue(domain.Supply{UnitPrice: decimal.NewFromInt(10), StockQuantity: 5})

	assert.True(t, so.TotalAmount.Equal(decimal.NewFromInt(150)))

	so.ResetPricing()
	assert.True(t, so.TotalAmount.Equal(decimal.Zero))
}

func TestServiceOrderHistory_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  *domain.SearchServiceOrderHistoryParams
		wantErr error
	}{
		{
			name:    "valid params",
			params:  &domain.SearchServiceOrderHistoryParams{ID: "some-id"},
			wantErr: nil,
		},
		{
			name:    "empty id",
			params:  &domain.SearchServiceOrderHistoryParams{ID: ""},
			wantErr: domain.ErrServiceOrderIDRequired,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestNewServiceOrderHistory(t *testing.T) {
	prev := domain.SERVICE_ORDER_STATUS_NEW
	next := domain.SERVICE_ORDER_STATUS_RECEIVED
	now := time.Now()

	item := domain.NewServiceOrderHistory(prev, next, now)

	assert.Equal(t, prev, item.PreviousStatus)
	assert.Equal(t, next, item.NewStatus)
	assert.Equal(t, now, item.CreatedAt)
}
