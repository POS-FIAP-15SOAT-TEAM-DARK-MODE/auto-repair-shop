package domain_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestWorkServiceOrderStatus_String(t *testing.T) {
	tests := []struct {
		status domain.WORK_SERVICE_ORDER_STATUS
		want   string
	}{
		{domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START, "AWAITING_START"},
		{domain.WORK_SERVICE_ORDER_STATUS_IN_PROGRESS, "IN_PROGRESS"},
		{domain.WORK_SERVICE_ORDER_STATUS_COMPLETED, "COMPLETED"},
		{domain.WORK_SERVICE_ORDER_STATUS_CANCELLED, "CANCELLED"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.String())
		})
	}
}

func TestWorkServiceOrderHistory_NextStatus(t *testing.T) {
	tests := []struct {
		name     string
		current  domain.WORK_SERVICE_ORDER_STATUS
		wantNext domain.WORK_SERVICE_ORDER_STATUS
		wantErr  error
	}{
		{
			name:     "awaiting_start -> in_progress",
			current:  domain.WORK_SERVICE_ORDER_STATUS_AWAITING_START,
			wantNext: domain.WORK_SERVICE_ORDER_STATUS_IN_PROGRESS,
		},
		{
			name:     "in_progress -> completed",
			current:  domain.WORK_SERVICE_ORDER_STATUS_IN_PROGRESS,
			wantNext: domain.WORK_SERVICE_ORDER_STATUS_COMPLETED,
		},
		{
			name:    "completed is terminal",
			current: domain.WORK_SERVICE_ORDER_STATUS_COMPLETED,
			wantErr: domain.ErrWorkServiceOrderAlreadyCompleted,
		},
		{
			name:    "cancelled cannot advance",
			current: domain.WORK_SERVICE_ORDER_STATUS_CANCELLED,
			wantErr: domain.ErrWorkServiceOrderAlreadyCancelled,
		},
		{
			name:    "invalid status",
			current: domain.WORK_SERVICE_ORDER_STATUS("UNKNOWN"),
			wantErr: domain.ErrWorkServiceOrderInvalidStatus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := domain.WorkServiceOrderHistory{Status: tt.current}
			got, err := h.NextStatus()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantNext, got)
			}
		})
	}
}
