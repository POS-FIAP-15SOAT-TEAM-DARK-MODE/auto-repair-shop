package service_order_history

const (
	selectServiceOrderHistory = `SELECT soh.id, soh.service_order_id, soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`
	countServiceOrderHistory  = `SELECT COUNT(soh.id) FROM service_order_status_history soh`
)
