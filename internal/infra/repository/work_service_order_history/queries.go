package work_service_order_history

const (
	insertWorkSOHistory = `
		INSERT INTO work_service_order_status_history (id, work_id, service_order_id, previous_status, new_status)
		VALUES ($1, $2, $3, $4, $5)`

	searchWorkSOHistory = `
		SELECT wsosh.id, wsosh.work_id, wsosh.service_order_id, wsosh.previous_status, wsosh.new_status, wsosh.created_at
		FROM work_service_order_status_history wsosh`
)
