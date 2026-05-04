package customer_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/customer"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := customer.Repository()

	id := "c1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "user_id", "type", "cpf", "cnpj", "company_name", "phone", "name", "email",
			}).AddRow("c1", "u1", "INDIVIDUAL", "12345678901", "", "", "11999999999", "João", "joao@e.com"))

		c, err := repo.GetByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, id, c.ID)
		assert.Equal(t, "João", c.User.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetByID(context.Background(), id)
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(id).
			WillReturnError(assert.AnError)

		_, err := repo.GetByID(context.Background(), id)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_GetByDocument(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := customer.Repository()

	doc := "12345678901"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(doc).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "user_id", "type", "cpf", "cnpj", "company_name", "phone", "name", "email",
			}).AddRow("c1", "u1", "INDIVIDUAL", doc, "", "", "11999999999", "João", "joao@e.com"))

		c, err := repo.GetByDocument(context.Background(), doc)
		assert.NoError(t, err)
		assert.Equal(t, doc, c.CPF)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(doc).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetByDocument(context.Background(), doc)
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(doc).
			WillReturnError(assert.AnError)

		_, err := repo.GetByDocument(context.Background(), doc)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_GetByUserID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := customer.Repository()

	uid := "u1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(uid).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "user_id", "type", "cpf", "cnpj", "company_name", "phone", "name", "email",
			}).AddRow("c1", uid, "INDIVIDUAL", "123", "", "", "119", "João", "j@e.com"))

		c, err := repo.GetByUserID(context.Background(), uid)
		assert.NoError(t, err)
		assert.Equal(t, uid, c.UserID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(uid).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetByUserID(context.Background(), uid)
		assert.ErrorIs(t, err, domain.ErrCustomerNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(uid).
			WillReturnError(assert.AnError)

		_, err := repo.GetByUserID(context.Background(), uid)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := customer.Repository()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE customer").WithArgs("11888888888", "c1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Update(ctx, "c1", "11888888888")
		})
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE customer").WithArgs("11888888888", "c1").
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Update(ctx, "c1", "11888888888")
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := customer.Repository()

	id := "c1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("DELETE FROM customer").WithArgs(id).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, id)
		})
		assert.NoError(t, err)
	})

	t.Run("has service orders", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, id)
		})
		assert.ErrorIs(t, err, domain.ErrCustomerHasServiceOrders)
	})

	t.Run("count error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT").WithArgs(id).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, id)
		})
		assert.Error(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT COUNT").WithArgs(id).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
		mock.ExpectExec("DELETE FROM customer").WithArgs(id).WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, id)
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := customer.Repository()

	c, _ := domain.NewCustomer(
		"João Silva",
		"joao@example.com",
		"Secret@123",
		domain.IndividualCustomerType,
		"111.444.777-35",
		"",
		"11999999999",
	)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO customer").
		WithArgs(c.ID, c.UserID, c.Type, c.CPF, nil, nil, c.Phone).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, &c)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Create_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := customer.Repository()

	c, _ := domain.NewCustomer(
		"João Silva", "j@e.com", "Secret@123",
		domain.IndividualCustomerType, "11144477735", "", "11999999999",
	)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO customer").
		WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Create(ctx, &c)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Create_MissingTx(t *testing.T) {
	repo := customer.Repository()
	c := &domain.Customer{}

	err := repo.Create(context.Background(), c)
	assert.Error(t, err)
}
