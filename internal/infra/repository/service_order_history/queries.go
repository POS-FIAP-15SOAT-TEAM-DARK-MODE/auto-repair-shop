package service_order_history

const (
	selectServiceOrderHistory     = `SELECT soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`
	selectWorkServiceOrderHistory = `SELECT wsosh.work_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at FROM work_service_order_status_history wsosh`
)
