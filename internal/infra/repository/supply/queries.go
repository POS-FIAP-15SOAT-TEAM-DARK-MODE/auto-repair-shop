package repository

const (
	createSupplyQuery  = `INSERT INTO "supply" (id, name, description, unit_price, stock_quantity, version) VALUES ($1, $2, $3, $4, $5, $6)`
	listSuppliesQuery  = `SELECT id, name, description, unit_price, stock_quantity, version FROM "supply" LIMIT $1 OFFSET $2`
	countSuppliesQuery = `SELECT COUNT(*) FROM "supply"`
)
