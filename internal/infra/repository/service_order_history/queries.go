package service_order_history

const (
	selectServiceOrderHistory = `SELECT soh.id, soh.service_order_id, soh.previous_status, soh.new_status, soh.created_at FROM service_order_status_history soh`
	countServiceOrderHistory  = `SELECT COUNT(soh.id) FROM service_order_status_history soh`

	selectWorkTimelineByServiceOrderID = `SELECT
    sow.work_id,
    wsosh.previous_status,
    wsosh.new_status,
    wsosh.created_at
FROM service_order_work sow
LEFT JOIN work_service_order_status_history wsosh
       ON wsosh.work_id = sow.work_id
      AND wsosh.service_order_id = sow.service_order_id
WHERE sow.service_order_id = $1
ORDER BY sow.work_id, wsosh.created_at ASC`
)
