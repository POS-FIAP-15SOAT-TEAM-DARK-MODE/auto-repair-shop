package service_order

const (
	insertServiceOrderQuery = `INSERT INTO service_order (id, customer_id, vehicle_id, status, total_amount)
VALUES ($1, $2, $3, $4, $5)`

	getServiceOrderHistoryByIDQuery = `SELECT id, previous_status, new_status, created_at FROM service_order_status_history WHERE service_order_id = $1`
)
