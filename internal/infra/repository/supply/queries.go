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
	deleteQuery = `DELETE FROM "supply" s where s.id = $1`

	// decrementStockQuery atomically decrements stock only when enough quantity is available,
	// preventing race conditions without a separate SELECT FOR UPDATE.
	decrementStockQuery = `
	UPDATE supply
	SET stock_quantity = stock_quantity - $2,
	    updated_at     = NOW()
	WHERE id = $1
	  AND stock_quantity >= $2
	`
	restoreStockQuery = `
	UPDATE supply
	SET stock_quantity = stock_quantity + $2,
	    updated_at     = NOW()
	WHERE id = $1
	`
)
