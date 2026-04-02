package service_order

const (
	insertServiceOrderQuery = `INSERT INTO service_order (id, customer_id, vehicle_id, status, total_amount)
VALUES ($1, $2, $3, $4, $5)`
)
