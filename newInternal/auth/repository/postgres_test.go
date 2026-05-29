package repository_test

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := user.Repository()

	u := domain.NewUser("Admin", "admin@example.com", "Secret@123")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"user\"").
		WithArgs(u.ID, u.Name, u.Email, u.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, u)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := user.Repository()

	u := domain.NewUser("Admin", "admin@example.com", "Secret@123")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"user\"").
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, u)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Create_MissingTx(t *testing.T) {
	repo := user.Repository()
	u := domain.NewUser("Admin", "admin@example.com", "Secret@123")

	err := repo.Create(context.Background(), u)
	assert.Error(t, err)
}
