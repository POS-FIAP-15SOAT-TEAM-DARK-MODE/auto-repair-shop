package service_order

const (
	insertServiceOrderQuery = `INSERT INTO service_order (id, customer_id, vehicle_id, status, total_amount)
VALUES ($1, $2, $3, $4, $5)`

	serviceOrderExistsQuery = `SELECT so.status FROM service_order so WHERE id = $1`

	listWorksByServiceOrderQuery = `
SELECT w.id, w.name, w.description, sow.unit_price, w.status
FROM service_order_work sow
JOIN work w ON w.id = sow.work_id
WHERE sow.service_order_id = $1
ORDER BY sow.id ASC
`

	insertServiceOrderWorkQuery = `
INSERT INTO service_order_work (id, service_order_id, work_id, unit_price)
VALUES ($1, $2, $3, $4)
ON CONFLICT (service_order_id, work_id) DO NOTHING
`

	deleteServiceOrderWorkQuery = `
DELETE FROM service_order_work
WHERE service_order_id = $1 AND work_id = $2
`

	listSuppliesByServiceOrderQuery = `
SELECT s.id, s.name, s.description, sos.unit_price, sos.quantity, s.version
FROM service_order_supply sos
JOIN supply s ON s.id = sos.supply_id
WHERE sos.service_order_id = $1
ORDER BY sos.id ASC
`

	insertServiceOrderSuppliesQuery = `
INSERT INTO service_order_supply (id, service_order_id, supply_id, quantity, unit_price)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (service_order_id, supply_id) DO NOTHING
`

	deleteServiceOrderSuppliesQuery = `
DELETE FROM service_order_supply
WHERE service_order_id = $1 AND supply_id = $2
`
)
