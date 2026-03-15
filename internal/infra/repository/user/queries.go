package repository

const (
	GetUserByEmail = `SELECT id, name, email, password_hash FROM "user" WHERE email = $1`
	CreateUser     = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
	GetRolesById   = `SELECT r.name FROM "user_role" ur JOIN "role" r ON r.id = ur.role_id WHERE ur.user_id = $1`
)
