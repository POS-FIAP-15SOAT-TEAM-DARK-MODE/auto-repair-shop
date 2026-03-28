package vehicle_test

import (
	"context"
	"errors"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/vehicle"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()

	v := &domain.Vehicle{
		ID:           "id-1",
		LicensePlate: "ABC1D23",
		Brand:        "chevrolet",
		Model:        "onix",
		Year:         2020,
		CustomerId:   "cust-1",
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "vehicle" \(id, license_plate, brand, model, year, customer_id\)\s*VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
		WithArgs(v.ID, v.LicensePlate, v.Brand, v.Model, v.Year, v.CustomerId).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, v)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Save_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()

	v := &domain.Vehicle{LicensePlate: "ABC1D23"}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO "vehicle" \(id, license_plate, brand, model, year, customer_id\)\s*VALUES \(\$1, \$2, \$3, \$4, \$5, \$6\)`).
		WillReturnError(errors.New("exec error"))
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, v)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Save_MissingTx(t *testing.T) {
	repo := vehicle.NewVehicleRepository()
	v := &domain.Vehicle{}

	err := repo.Save(context.Background(), v)
	assert.Error(t, err)
}

func TestPostgresRepository_Find_ByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	// make repository use this db for one-time transactions
	postgresdb.ConnectWithDB(db)

	repo := vehicle.NewVehicleRepository()

	id := "veh-1"
	mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM "vehicle"`).
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "license_plate", "brand", "model", "year", "customer_id"}).
			AddRow(id, "ABC1D23", "chevrolet", "onix", 2020, "cust-1"))

	v, err := repo.Find(context.Background(), id)
	assert.NoError(t, err)
	assert.Equal(t, "ABC1D23", v.LicensePlate)
}

func TestPostgresRepository_Find_ByPlate_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)

	repo := vehicle.NewVehicleRepository()

	plate := "ABC1D23"
	mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM "vehicle"`).
		WithArgs(plate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "license_plate", "brand", "model", "year", "customer_id"}).
			AddRow("veh-2", plate, "fiat", "uno", 2010, "cust-2"))

	v, err := repo.Find(context.Background(), plate)
	assert.NoError(t, err)
	assert.Equal(t, "fiat", v.Brand)
}

func TestPostgresRepository_Find_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := vehicle.NewVehicleRepository()

	search := "non-existent"
	mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM "vehicle"`).
		WithArgs(search).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Find(context.Background(), search)
	assert.ErrorIs(t, err, domain.ErrVehicleNotFound)
}

func TestPostgresRepository_Find_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := vehicle.NewVehicleRepository()

	search := "some"
	mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM "vehicle"`).
		WithArgs(search).
		WillReturnError(errors.New("query failed"))

	_, err = repo.Find(context.Background(), search)
	assert.Error(t, err)
}

