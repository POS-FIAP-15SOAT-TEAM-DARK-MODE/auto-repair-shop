package service_order_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	service_order "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/service_order"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	so := buildServiceOrder()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT status FROM service_order WHERE id = \\$1").
		WithArgs(so.ID).
		WillReturnRows(sqlmock.NewRows([]string{"status"})) // empty rows, same as ErrNoRows

	mock.ExpectExec("INSERT INTO service_order").
		WithArgs(so.ID, so.Customer.ID, so.Vehicle.ID, so.Status.String(), so.TotalAmount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO service_order_status_history").
		WithArgs(sqlmock.AnyArg(), so.ID, nil, so.Status.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, so)
	})

	assert.NoError(t, err)
}

func buildFullServiceOrder() *domain.ServiceOrder {
	customer := &domain.Customer{
		ID:     "customer-id",
		UserID: "user-id",
		Type:   domain.IndividualCustomerType,
		CPF:    "12345678901",
		Phone:  "11999999999",
		User: &domain.User{
			Name:  "João Silva",
			Email: "joao@example.com",
		},
	}
	vehicle := &domain.Vehicle{
		ID:           "vehicle-id",
		LicensePlate: "ABC-1234",
		Brand:        "Toyota",
		Model:        "Corolla",
		Year:         2020,
	}
	so := domain.NewServiceOrder(customer, vehicle)
	so.ID = "so-id"
	so.TotalAmount = decimal.NewFromFloat(150.50)
	return so
}

func TestPostgresRepository_Save_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	so := buildFullServiceOrder()
	so.Status = domain.SERVICE_ORDER_STATUS_RECEIVED

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT status FROM service_order WHERE id = \\$1").
		WithArgs(so.ID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("NEW"))

	mock.ExpectExec("INSERT INTO service_order").
		WithArgs(so.ID, so.Customer.ID, so.Vehicle.ID, so.Status.String(), so.TotalAmount).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO service_order_status_history").
		WithArgs(sqlmock.AnyArg(), so.ID, "NEW", so.Status.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, so)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Save_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	so := buildFullServiceOrder()

	t.Run("select status error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT status FROM service_order WHERE id = \\$1").
			WithArgs(so.ID).
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Save(ctx, so)
		})
		assert.Error(t, err)
	})

	t.Run("insert so error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT status FROM service_order WHERE id = \\$1").
			WithArgs(so.ID).
			WillReturnRows(sqlmock.NewRows([]string{"status"}))
		mock.ExpectExec("INSERT INTO service_order").WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Save(ctx, so)
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_ExistsByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := service_order.Repository()

	id := "so-id"

	t.Run("exists", func(t *testing.T) {
		mock.ExpectQuery("SELECT so.status FROM service_order so WHERE id = \\$1").
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("NEW"))

		exists, status, err := repo.ExistsByID(context.Background(), id)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, domain.SERVICE_ORDER_STATUS_NEW, status)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT so.status FROM service_order so WHERE id = \\$1").
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		exists, _, err := repo.ExistsByID(context.Background(), id)
		assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
		assert.False(t, exists)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT so.status FROM service_order so WHERE id = \\$1").
			WithArgs(id).
			WillReturnError(assert.AnError)

		exists, _, err := repo.ExistsByID(context.Background(), id)
		assert.Error(t, err)
		assert.False(t, exists)
	})
}

func TestPostgresRepository_ListWorksByServiceOrderID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := service_order.Repository()

	id := "so-id"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT w.id, w.name, w.description, sow.unit_price, w.status").
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "status"}).
				AddRow("w1", "Work 1", "Desc 1", 100.0, true))

		works, err := repo.ListWorksByServiceOrderID(context.Background(), id)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assert.Len(t, works, 1)
		assert.Equal(t, "w1", works[0].ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT w.id, w.name, w.description, sow.unit_price, w.status").
			WithArgs(id).
			WillReturnError(assert.AnError)

		_, err := repo.ListWorksByServiceOrderID(context.Background(), id)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_AddWorkLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO service_order_work").
		WithArgs(sqlmock.AnyArg(), "so1", "w1", decimal.NewFromInt(100)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.AddWorkLink(ctx, "so1", "w1", decimal.NewFromInt(100))
	})
	assert.NoError(t, err)
}

func TestPostgresRepository_RemoveWorkLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM service_order_work").
			WithArgs("so1", "w1").
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.RemoveWorkLink(ctx, "so1", "w1")
		})
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM service_order_work").
			WithArgs("so1", "w1").
			WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.RemoveWorkLink(ctx, "so1", "w1")
		})
		assert.ErrorIs(t, err, domain.ErrServiceOrderWorkNotFound)
	})
}

func TestPostgresRepository_ListSuppliesByServiceOrderID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := service_order.Repository()

	id := "so-id"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT s.id, s.name, s.description, sos.unit_price, sos.quantity, s.version").
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "quantity", "version"}).
				AddRow("s1", "Supply 1", "Desc 1", 10.0, 5, 1))

		supplies, err := repo.ListSuppliesByServiceOrderID(context.Background(), id)
		assert.NoError(t, err)
		assert.Len(t, supplies, 1)
		assert.Equal(t, "s1", supplies[0].ID)
		assert.Equal(t, 5, supplies[0].StockQuantity)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT s.id, s.name, s.description, sos.unit_price, sos.quantity, s.version").
			WithArgs(id).
			WillReturnError(assert.AnError)

		_, err := repo.ListSuppliesByServiceOrderID(context.Background(), id)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_AddSupplyLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO service_order_supply").
		WithArgs(sqlmock.AnyArg(), "so1", "s1", 5, decimal.NewFromInt(10)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.AddSupplyLink(ctx, "so1", "s1", 5, decimal.NewFromInt(10))
	})
	assert.NoError(t, err)
}

func TestPostgresRepository_RemoveSupplyLink(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := service_order.Repository()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("DELETE FROM service_order_supply").
			WithArgs("so1", "s1").
			WillReturnRows(sqlmock.NewRows([]string{"quantity"}).AddRow(5))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			qty, err := repo.RemoveSupplyLink(ctx, "so1", "s1")
			assert.NoError(t, err)
			assert.Equal(t, 5, qty)
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("DELETE FROM service_order_supply").
			WithArgs("so1", "s1").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			_, err = repo.RemoveSupplyLink(ctx, "so1", "s1")
			return err
		})
		assert.ErrorIs(t, err, domain.ErrServiceOrderSupplyNotFound)
	})
}

func TestPostgresRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := service_order.Repository()

	id := "so-id"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "status", "total_amount",
				"cId", "user_id", "type",
				"cpf", "cnpj", "company_name", "phone",
				"name", "email",
				"vId", "license_plate", "brand",
				"model", "year",
			}).AddRow(
				id, "NEW", 150.50,
				"c1", "u1", "INDIVIDUAL",
				"12345678901", "", "", "11999999999",
				"João Silva", "joao@example.com",
				"v1", "ABC-1234", "Toyota",
				"Corolla", 2020,
			))

		so, err := repo.FindByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, id, so.ID)
		assert.Equal(t, domain.SERVICE_ORDER_STATUS_NEW, so.Status)
		assert.Equal(t, "João Silva", so.Customer.User.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindByID(context.Background(), id)
		assert.ErrorIs(t, err, domain.ErrServiceOrderNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(id).
			WillReturnError(assert.AnError)

		_, err := repo.FindByID(context.Background(), id)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_SearchAndCount(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := service_order.Repository()

	params := &domain.ServiceOrderFilterParams{
		Status:     "NEW",
		CustomerID: "c1",
		VehicleID:  "v1",
		Limit:      10,
		Offset:     0,
	}

	t.Run("Count", func(t *testing.T) {
		mock.ExpectQuery("SELECT COUNT").WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		count, err := repo.Count(context.Background(), params)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Search", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{
			"id", "status", "total_amount",
			"cId", "user_id", "type",
			"cpf", "cnpj", "company_name", "phone",
			"name", "email",
			"vId", "license_plate", "brand",
			"model", "year",
		}).AddRow(
			"so1", "NEW", 150.50,
			"c1", "u1", "INDIVIDUAL",
			"12345678901", "", "", "11999999999",
			"João Silva", "joao@example.com",
			"v1", "ABC-1234", "Toyota",
			"Corolla", 2020,
		))
		items, err := repo.Search(context.Background(), params)
		assert.NoError(t, err)
		assert.Len(t, items, 1)
	})
}

func TestPostgresRepository_AverageExecutionTimeInHours(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := service_order.Repository()

	mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "avg_hours"}).
		AddRow("w1", "Work 1", 2.5))

	res, err := repo.AverageExecutionTimeInHours(context.Background(), []string{"w1"})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, 2.5, res[0].AverageHours)
}
