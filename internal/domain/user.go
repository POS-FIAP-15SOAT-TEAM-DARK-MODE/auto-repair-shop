package domain

import (
	"errors"
	"time"
)

type User struct {
	Id        string
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserRepository interface {
	CreateUser(c *User) (*User, error)
	FindByEmail(email string) (*User, error)
}

// É uma função construtora para criar um novo usuário,
// garantindo que as regras de negócio sejam respeitadas
// (ex: nome, email e senha não podem ser vazios).
func NewUser(id, name, email, password string) (*User, error) {
	if name == "" || email == "" || password == "" {
		return nil, errors.New("name, email and password hash cannot be empty")
	}

	now := time.Now()
	// Aqui a gente cria o Obj User, e já retorna ele pronto pra ser inserido no banco de dados;
	return &User{
		Id:        id,
		Name:      name,
		Email:     email,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
