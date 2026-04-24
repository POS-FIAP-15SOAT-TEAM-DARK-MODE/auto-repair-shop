package supply

const (
	upsertQuery = `
    INSERT INTO "supply" (id, name, description, unit_price, stock_quantity, version)
    VALUES ($1, $2, $3, $4, $5, $6)
    ON CONFLICT (id) DO UPDATE
    SET
        name           = EXCLUDED.name,
        description    = EXCLUDED.description,
        unit_price     = EXCLUDED.unit_price,
        stock_quantity = EXCLUDED.stock_quantity,
        version        = EXCLUDED.version,
        updated_at     = NOW()
`
	countSuppliesQuery = `SELECT COUNT(s.id) FROM "supply" s`
	searchQuery        = `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s`
	findByIDQuery      = `
	SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version
	FROM "supply" s
	WHERE s.id = $1
	`
)
