package vehicle_test

import (
	"context"
	"database/sql"
	"errors"
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

func TestPostgresRepository_Find(t *testing.T) {
	tests := []struct {
		name           string
		search         string
		mockSetup      func(mock sqlmock.Sqlmock, search string)
		expectError    bool
		expectNotFound bool
		expectBrand    string
		expectPlate    string
	}{
		{
			name:   "find by plate success",
			search: "ABC1D23",
			mockSetup: func(mock sqlmock.Sqlmock, search string) {
				mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM "vehicle"`).
					WithArgs(search).
					WillReturnRows(sqlmock.NewRows([]string{"id", "license_plate", "brand", "model", "year", "customer_id"}).
						AddRow("veh-2", search, "fiat", "uno", 2010, "cust-2"))
			},
			expectError: false,
			expectBrand: "fiat",
		},
		{
			name:   "not found",
			search: "non-existent",
			mockSetup: func(mock sqlmock.Sqlmock, search string) {
				mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM "vehicle"`).
					WithArgs(search).
					WillReturnError(sql.ErrNoRows)
			},
			expectError:    true,
			expectNotFound: true,
		},
		{
			name:   "query error",
			search: "some",
			mockSetup: func(mock sqlmock.Sqlmock, search string) {
				mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM "vehicle"`).
					WithArgs(search).
					WillReturnError(errors.New("query failed"))
			},
			expectError: true,
		},
	}

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			postgresdb.ConnectWithDB(db)
			repo := vehicle.NewVehicleRepository()

			if tt.mockSetup != nil {
				tt.mockSetup(mock, tt.search)
			}

			v, err := repo.Find(context.Background(), tt.search)
			if tt.expectError {
				if tt.expectNotFound {
					assert.ErrorIs(t, err, domain.ErrVehicleNotFound)
				} else {
					assert.Error(t, err)
				}
				return
			}
			assert.NoError(t, err)
			if tt.expectBrand != "" {
				assert.Equal(t, tt.expectBrand, v.Brand)
			}
			if tt.expectPlate != "" {
				assert.Equal(t, tt.expectPlate, v.LicensePlate)
			}
		})
	}
}

func TestPostgresRepository_Update_MissingTx(t *testing.T) {
	repo := vehicle.NewVehicleRepository()
	v := &domain.Vehicle{}

	err := repo.Update(context.Background(), v)
	assert.Error(t, err)
}

func TestPostgresRepository_Update_Success(t *testing.T) {
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
	mock.ExpectExec(`UPDATE vehicle SET license_plate = \$2, brand = \$3, model = \$4, year = \$5, customer_id = \$6 WHERE id = \$1`).
		WithArgs(v.ID, v.LicensePlate, v.Brand, v.Model, v.Year, v.CustomerId).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Update(ctx, v)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Update_Error(t *testing.T) {
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
	mock.ExpectExec(`UPDATE vehicle SET license_plate = \$2, brand = \$3, model = \$4, year = \$5, customer_id = \$6 WHERE id = \$1`).
		WithArgs(v.ID, v.LicensePlate, v.Brand, v.Model, v.Year, v.CustomerId).
		WillReturnError(errors.New("exec error"))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Update(ctx, v)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Delete_MissingTx(t *testing.T) {
	repo := vehicle.NewVehicleRepository()

	err := repo.Delete(context.Background(), "id-1")
	assert.Error(t, err)
}

func TestPostgresRepository_Delete_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()
	id := "id-1"

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM vehicle WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Delete(ctx, id)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()
	id := "id-1"

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM vehicle WHERE id = \$1`).
		WithArgs(id).
		WillReturnError(errors.New("exec error"))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Delete(ctx, id)
	})

	assert.Error(t, err)
}
