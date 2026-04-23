package repository

const (
	findByIDQuery = `
	SELECT s.id, s.name, s.description, s.unit_price, s.stock_quantity, s.version
	FROM "supply" s
	WHERE s.id = $1
	`
	createSupplyQuery = `INSERT INTO "supply" (id, name, description, unit_price, stock_quantity, version) VALUES ($1, $2, $3, $4, $5, $6)`
)
