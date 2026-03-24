package repository

const (
	getUserByEmail  = `SELECT id, name, email, password_hash FROM "user" WHERE email = $1`
	getRolesById    = `SELECT r.name FROM "user_role" ur JOIN "role" r ON r.id = ur.role_id WHERE ur.user_id = $1`
	createUserQuery = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	updateUserQuery = `UPDATE "user" SET name = $1, email = $2 WHERE id = $3`
	deleteUserQuery = `DELETE FROM "user" WHERE id = $1`
)
