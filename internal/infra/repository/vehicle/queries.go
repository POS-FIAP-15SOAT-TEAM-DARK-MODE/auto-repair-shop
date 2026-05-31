package vehicle

const (
	insertNewVehicle = `INSERT INTO "vehicle" (id, license_plate, brand, model, year, customer_id)
						VALUES ($1, $2, $3, $4, $5, $6)`

	selectVehicle = `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle`

	updateVehicle = `UPDATE vehicle SET license_plate = $2, brand = $3, model = $4, year = $5, customer_id = $6 WHERE id = $1`

	deleteVehicle = `DELETE FROM vehicle WHERE id = $1`

	countVehicles = `SELECT COUNT(id) FROM vehicle`
)
