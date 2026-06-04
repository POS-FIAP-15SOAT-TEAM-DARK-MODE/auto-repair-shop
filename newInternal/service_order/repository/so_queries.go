package repository

const (
	selectSOStatusQuery     = "SELECT status FROM service_order WHERE id = $1"
	insertServiceOrderQuery = `
	INSERT INTO service_order (id, customer_id, vehicle_id, status, total_amount)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (id) DO UPDATE
	SET
		status        = EXCLUDED.status,
		total_amount = EXCLUDED.total_amount,
		updated_at  = NOW();
	`

	insertServiceOrderStatusQuery = `
	INSERT INTO service_order_status_history (id, service_order_id, previous_status, new_status)
	VALUES ($1, $2, $3, $4)
	`

	serviceOrderExistsQuery = `SELECT so.status FROM service_order so WHERE id = $1`

	serviceOrderFindByIDQuery = `
	SELECT
		so.id,
		so.status,
		so.total_amount,
		c.id as cId, c.user_id, c.type,
	    COALESCE(c.cpf, ''), COALESCE(c.cnpj, ''), COALESCE(c.company_name, ''), c.phone,
	    u.name, u.email,
		v.id as vId, v.license_plate, v.brand,
		v.model, v.year
	FROM service_order so
	JOIN customer c ON c.id = so.customer_id
	JOIN "user" u ON u.id = c.user_id
	JOIN vehicle v ON v.id = so.vehicle_id
	WHERE so.id = $1
	`

	countServiceOrderQuery = `SELECT COUNT(so.id) FROM service_order so`

	searchServiceOrderQuery = `
	SELECT
		so.id,
		so.status,
		so.total_amount,
		c.id as cId, c.user_id, c.type,
		COALESCE(c.cpf, ''), COALESCE(c.cnpj, ''), COALESCE(c.company_name, ''), c.phone,
		u.name, u.email,
		v.id as vId, v.license_plate, v.brand,
		v.model, v.year
	FROM service_order so
	JOIN customer c ON c.id = so.customer_id
	JOIN "user" u ON u.id = c.user_id
	JOIN vehicle v ON v.id = so.vehicle_id
	`

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
RETURNING quantity
`

	averageExecutionTimeBaseQuery = `
SELECT
    w.id,
    w.name,
    COALESCE(AVG(EXTRACT(EPOCH FROM (h_end.created_at - h_start.created_at)) / 3600), 0) AS avg_hours
FROM work_service_order_status_history h_start
JOIN work_service_order_status_history h_end
    ON h_start.work_id = h_end.work_id
    AND h_start.service_order_id = h_end.service_order_id
JOIN work w ON w.id = h_start.work_id`
)
