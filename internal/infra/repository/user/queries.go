package repository

const (
	GetUserByEmail = `SELECT email FROM "user" WHERE email = $1`
	CreateUser     = `INSERT INTO "user" (id, name, email, password_hash) VALUES ($1, $2, $3, $4)`
)
