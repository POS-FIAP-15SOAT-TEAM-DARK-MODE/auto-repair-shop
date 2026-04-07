package service_order_history

const (
	selectServiceOrderHistory = `SELECT id, service_order_id, previous_status, new_status, created_at FROM service_order_status_history`
)
