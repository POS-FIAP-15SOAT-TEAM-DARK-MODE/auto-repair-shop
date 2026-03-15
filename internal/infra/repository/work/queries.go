package work

const (
	upsertQuery = `
		INSERT INTO "service" (id, name, description, unit_price, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE
		SET
			name        = EXCLUDED.name,
			description = EXCLUDED.description,
			unit_price  = EXCLUDED.unit_price,
			status      = EXCLUDED.status,
			updated_at  = NOW();
	`
)
