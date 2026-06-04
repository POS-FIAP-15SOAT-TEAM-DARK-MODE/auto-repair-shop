package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/domain"
	repository "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/auth/repository"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/uow"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(time.Minute)

	u := authDomain.NewUser(uuid.New().String(), "Admin", "admin@example.com", "Secret@123")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"user\"").
		WithArgs(u.ID, u.Name, u.Email, u.Password).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, u)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := uow.NewTransactionalUoW(db)
	repo := repository.NewPostgres(time.Minute)

	u := authDomain.NewUser(uuid.New().String(), "Admin", "admin@example.com", "Secret@123")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"user\"").
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, u)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Create_MissingTx(t *testing.T) {
	repo := repository.NewPostgres(time.Minute)
	u := authDomain.NewUser(uuid.New().String(), "Admin", "admin@example.com", "Secret@123")

	err := repo.Save(context.Background(), u)
	assert.Error(t, err)
}
