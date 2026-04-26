package domain

import (
	"context"
	"time"
)

type WORK_SERVICE_ORDER_STATUS string

const (
	WORK_SERVICE_ORDER_STATUS_AWAITING_START WORK_SERVICE_ORDER_STATUS = "AWAITING_START"
	WORK_SERVICE_ORDER_STATUS_IN_PROGRESS    WORK_SERVICE_ORDER_STATUS = "IN_PROGRESS"
	WORK_SERVICE_ORDER_STATUS_COMPLETED      WORK_SERVICE_ORDER_STATUS = "COMPLETED"
	WORK_SERVICE_ORDER_STATUS_CANCELLED      WORK_SERVICE_ORDER_STATUS = "CANCELLED"
)

func (s WORK_SERVICE_ORDER_STATUS) String() string {
	return string(s)
}

var workStatusOrder = []WORK_SERVICE_ORDER_STATUS{
	WORK_SERVICE_ORDER_STATUS_AWAITING_START,
	WORK_SERVICE_ORDER_STATUS_IN_PROGRESS,
	WORK_SERVICE_ORDER_STATUS_COMPLETED,
}

type WorkServiceOrderHistory struct {
	ID             string
	ServiceOrderID string
	WorkID         string
	PreviousStatus *WORK_SERVICE_ORDER_STATUS
	Status         WORK_SERVICE_ORDER_STATUS
	CreatedAt      time.Time
}

func (w WorkServiceOrderHistory) NextStatus() (WORK_SERVICE_ORDER_STATUS, error) {
	if w.Status == WORK_SERVICE_ORDER_STATUS_CANCELLED {
		return "", ErrWorkServiceOrderAlreadyCancelled
	}
	for i, s := range workStatusOrder {
		if s == w.Status {
			if i+1 >= len(workStatusOrder) {
				return "", ErrWorkServiceOrderAlreadyCompleted
			}
			return workStatusOrder[i+1], nil
		}
	}
	return "", ErrWorkServiceOrderInvalidStatus
}

type SearchWorkSOHistoryParams struct {
	ServiceOrderID string
	WorkID         string
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=WorkServiceOrderHistoryRepository --with-expecter
type WorkServiceOrderHistoryRepository interface {
	Insert(ctx context.Context, serviceOrderID, workID string, previousStatus *WORK_SERVICE_ORDER_STATUS, newStatus WORK_SERVICE_ORDER_STATUS) error
	Search(ctx context.Context, params SearchWorkSOHistoryParams) ([]WorkServiceOrderHistory, error)
}
