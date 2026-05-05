package vehicle_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
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
		isID           bool
		expectError    bool
		expectNotFound bool
		expectBrand    string
		expectPlate    string
	}{
		{
			name:   "find by plate success",
			search: "ABC1D23",
			mockSetup: func(mock sqlmock.Sqlmock, search string) {
				mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE license_plate =`).
					WithArgs(search).
					WillReturnRows(sqlmock.NewRows([]string{"id", "license_plate", "brand", "model", "year", "customer_id"}).
						AddRow("veh-2", search, "fiat", "uno", 2010, "cust-2"))
			},
			expectError: false,
			expectBrand: "fiat",
		},
		{
			name:   "find by id success",
			search: "ABC1D23",
			isID:   true,
			mockSetup: func(mock sqlmock.Sqlmock, search string) {
				mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE id =`).
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
				mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM  vehicle WHERE license_plate =`).
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
				mock.ExpectQuery(`SELECT id, license_plate, brand, model, year, customer_id FROM  vehicle WHERE license_plate =`).
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

			p := domain.FindVehicleParams{}

			if tt.isID {
				p.ID = tt.search
			} else {
				p.LicensePlate = tt.search
			}

			v, err := repo.Find(context.Background(), p)
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

func TestVehicleRepository_Count_Success_NoFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{}

	query := `SELECT COUNT(id) FROM vehicle`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(10),
		)

	total, err := repo.Count(context.Background(), params)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
}

func TestVehicleRepository_Count_WithCustomerFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{
		CustomerId: "cust-1",
	}

	query := `SELECT COUNT(id) FROM vehicle WHERE customer_id =`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(params.CustomerId).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).AddRow(10),
		)

	total, err := repo.Count(context.Background(), params)

	assert.NoError(t, err)
	assert.Equal(t, int64(10), total)
}

func TestVehicleRepository_Count_TransactionError(t *testing.T) {
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{}

	ctx := context.Background()

	total, err := repo.Count(ctx, params)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if total != 0 {
		t.Fatalf("expected total 0, got %d", total)
	}
}

func TestVehicleRepository_Count_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{}

	expectedErr := errors.New("db error")

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT COUNT").
		WillReturnError(expectedErr)

	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.Count(ctx, params)
		return err
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestVehicleRepository_Count_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{}

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT COUNT").
		WillReturnRows(
			sqlmock.NewRows([]string{"invalid"}),
		)

	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.Count(ctx, params)
		return err
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestVehicleRepository_Search_RowsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{
		Limit: 10,
	}

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{
		"id", "license_plate", "model", "brand", "year", "customer_id",
	}).AddRow("1", "AAA1234", "onix", "chevrolet", 2020, "123")

	rows.RowError(0, errors.New("row error"))

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.Search(ctx, params)
		return err
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestVehicleRepository_Search_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{
		Limit: 10,
	}

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{
		"id", "license_plate", "model", "brand", "year", "customer_id",
	}).AddRow("1", "AAA1234", "onix", "chevrolet", "INVALID_YEAR", "123")

	mock.ExpectQuery("SELECT").
		WillReturnRows(rows)

	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.Search(ctx, params)
		return err
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestVehicleRepository_Search_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{
		Limit: 10,
	}

	expectedErr := errors.New("query error")

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT").
		WillReturnError(expectedErr)

	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		_, err := repo.Search(ctx, params)
		return err
	})

	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func TestVehicleRepository_Search_TransactionError(t *testing.T) {
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{}

	ctx := context.Background()

	result, err := repo.Search(ctx, params)

	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}
}

func TestVehicleRepository_Search_WithCustomerFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{
		CustomerId: "123",
		Limit:      5,
		Offset:     10,
	}

	rows := sqlmock.NewRows([]string{
		"id", "license_plate", "model", "brand", "year", "customer_id",
	}).AddRow("1", "BBB9999", "hb20", "hyundai", 2021, "123")

	query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle WHERE customer_id = $1 LIMIT $2 OFFSET $3`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(params.CustomerId, params.Limit, params.Offset).
		WillReturnRows(rows)

	var result []domain.Vehicle
	result, err = repo.Search(context.Background(), params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 vehicle, got %d", len(result))
	}
}

func TestVehicleRepository_Search_Success_NoFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := vehicle.NewVehicleRepository()

	params := &domain.SearchVehicleParams{
		Limit:  5,
		Offset: 10,
	}

	rows := sqlmock.NewRows([]string{
		"id", "license_plate", "model", "brand", "year", "customer_id",
	}).AddRow("1", "BBB9999", "hb20", "hyundai", 2021, "123")

	query := `SELECT id, license_plate, brand, model, year, customer_id FROM vehicle LIMIT $1 OFFSET $2`
	mock.ExpectQuery(regexp.QuoteMeta(query)).
		WithArgs(params.Limit, params.Offset).
		WillReturnRows(rows)

	var result []domain.Vehicle
	result, err = repo.Search(context.Background(), params)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 vehicle, got %d", len(result))
	}
}
