package supply

const (
	upsertQuery = `
    INSERT INTO "supply" (id, name, description, unit_price, stock_quantity)
    VALUES ($1, $2, $3, $4, $5)
    ON CONFLICT (id) DO UPDATE
    SET
        name           = EXCLUDED.name,
        description    = EXCLUDED.description,
        unit_price     = EXCLUDED.unit_price,
        stock_quantity = EXCLUDED.stock_quantity,
        updated_at     = NOW()
`
	countSuppliesQuery = `SELECT COUNT(s.id) FROM "supply" s`
	searchQuery        = `SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version FROM "supply" s`
	updateQuery        = `
    UPDATE "supply"
    SET
        name           = $2,
        description    = $3,
        unit_price     = $4,
        stock_quantity = $5,
        updated_at     = NOW()
    WHERE id = $1
`
)
