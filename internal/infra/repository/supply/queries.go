package repository

const (
	createSupplyQuery  = `INSERT INTO "supply" (id, name, description, unit_price, stock_quantity, version) VALUES ($1, $2, $3, $4, $5, $6)`
	listSuppliesQuery  = `SELECT id, name, description, unit_price, stock_quantity, version FROM "supply" ORDER BY name ASC`
	getSupplyByIDQuery = `SELECT id, name, description, unit_price, stock_quantity, version FROM "supply" WHERE id = $1`
)
