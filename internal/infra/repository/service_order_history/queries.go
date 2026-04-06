package service_order

const (
	selectServiceOrderHistory = `SELECT id, previous_status, new_status, created_at FROM service_order_status_history`
)
