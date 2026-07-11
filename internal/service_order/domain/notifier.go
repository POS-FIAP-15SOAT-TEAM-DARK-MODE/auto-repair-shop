package domain

import "context"

//go:generate go run github.com/vektra/mockery/v2@latest --name=StatusNotifier --with-expecter

type StatusNotification struct {
	ServiceOrderID string
	CustomerName   string
	CustomerEmail  string
	// PreviousStatus is nil when the order has just been created.
	PreviousStatus *SERVICE_ORDER_STATUS
	NewStatus      SERVICE_ORDER_STATUS
}

type StatusNotifier interface {
	NotifyStatusChange(ctx context.Context, n StatusNotification) error
}
