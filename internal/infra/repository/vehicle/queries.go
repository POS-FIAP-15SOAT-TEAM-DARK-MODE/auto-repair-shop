package vehicle

const (
	insertNewVehicle = `INSERT INTO "vehicle" (id, license_plate, brand, model, year, customer_id)
						VALUES ($1, $2, $3, $4, $5, $6)`
)
