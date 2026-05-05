package work_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/work"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_Save(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := work.Repository()

	w, _ := domain.NewWork("Oil Change", "Description", "100.00", domain.ACTIVE)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"work\"").
		WithArgs(w.ID, w.Name, w.Description, w.Price, true).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, w)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Save_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := work.Repository()

	w, _ := domain.NewWork("Oil Change", "Description", "100.00", domain.ACTIVE)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO \"work\"").WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Save(ctx, w)
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Delete_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := work.Repository()

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM \"work\"").WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Delete(ctx, "id")
	})

	assert.Error(t, err)
}

func TestPostgresRepository_Search_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	mock.ExpectQuery("SELECT s.id, s.name, s.description, s.unit_price, s.status FROM \"work\"").
		WillReturnError(assert.AnError)

	_, err = repo.Search(context.Background(), &domain.SearchWorkParams{Limit: 10})
	assert.Error(t, err)
}

func TestPostgresRepository_Search_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	rows := sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "status"}).
		AddRow("id1", "Work 1", "Desc 1", "invalid-price", true)
	mock.ExpectQuery("SELECT s.id, s.name, s.description, s.unit_price, s.status FROM \"work\"").
		WillReturnRows(rows)

	_, err = repo.Search(context.Background(), &domain.SearchWorkParams{Limit: 10})
	assert.Error(t, err)
}

func TestPostgresRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := work.Repository()

	id := "work-id"

	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM \"work\"").
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
		return repo.Delete(ctx, id)
	})

	assert.NoError(t, err)
}

func TestPostgresRepository_Search(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	params := &domain.SearchWorkParams{Limit: 10, Offset: 0}

	rows := sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "status"}).
		AddRow("id1", "Work 1", "Desc 1", "10.00", true)
	mock.ExpectQuery("SELECT s.id, s.name, s.description, s.unit_price, s.status FROM \"work\"").
		WillReturnRows(rows)

	results, err := repo.Search(context.Background(), params)

	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "id1", results[0].ID)
}

func TestPostgresRepository_Count(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	params := &domain.SearchWorkParams{}

	mock.ExpectQuery("SELECT COUNT\\(s.id\\) FROM \"work\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	total, err := repo.Count(context.Background(), params)

	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
}

func TestPostgresRepository_Count_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	params := &domain.SearchWorkParams{}

	mock.ExpectQuery("SELECT COUNT\\(s.id\\) FROM \"work\"").
		WillReturnError(assert.AnError)

	total, err := repo.Count(context.Background(), params)

	assert.Error(t, err)
	assert.Equal(t, int64(0), total)
}

func TestPostgresRepository_Count_WithStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	params := &domain.SearchWorkParams{Status: "ACTIVE"}

	mock.ExpectQuery("SELECT COUNT\\(s.id\\) FROM \"work\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	total, err := repo.Count(context.Background(), params)

	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
}

func TestPostgresRepository_Save_NoTransaction(t *testing.T) {
	repo := work.Repository()
	w, _ := domain.NewWork("Oil Change", "Description", "100.00", domain.ACTIVE)

	err := repo.Save(context.Background(), w)
	assert.Error(t, err)
}

func TestPostgresRepository_Search_NoTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	mock.ExpectQuery("SELECT s.id, s.name, s.description, s.unit_price, s.status FROM \"work\"").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "status"}))

	_, err = repo.Search(context.Background(), &domain.SearchWorkParams{Limit: 10})
	assert.NoError(t, err)
}

func TestPostgresRepository_FindByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	id := "w1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery(`SELECT s.id, s.name, s.description, s.unit_price, s.status FROM "work" s WHERE s.id = \$1`).
			WithArgs(id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "status"}).
				AddRow(id, "Work 1", "Desc 1", "100.00", domain.ACTIVE))

		w, err := repo.FindByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, id, w.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT s.id, s.name, s.description, s.unit_price, s.status FROM "work" s WHERE s.id = \$1`).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.FindByID(context.Background(), id)
		assert.ErrorIs(t, err, domain.ErrWorkNotFound)
	})
}

func TestPostgresRepository_Search_WithStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := work.Repository()

	params := &domain.SearchWorkParams{Limit: 10, Offset: 0, Status: "ACTIVE"}

	rows := sqlmock.NewRows([]string{"id", "name", "description", "unit_price", "status"}).
		AddRow("id1", "Work 1", "Desc 1", "10.00", true)
	mock.ExpectQuery(`SELECT s.id, s.name, s.description, s.unit_price, s.status FROM "work" s`).
		WillReturnRows(rows)

	results, err := repo.Search(context.Background(), params)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
}
