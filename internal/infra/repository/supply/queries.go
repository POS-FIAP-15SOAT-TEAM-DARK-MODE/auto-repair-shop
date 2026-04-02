package repository

const (
	createSupplyQuery = `INSERT INTO "supply" (id, name, description, unit_price, stock_quantity, version) VALUES ($1, $2, $3, $4, $5, $6)`
)
