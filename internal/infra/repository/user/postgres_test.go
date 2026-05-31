package user_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	postgresdb "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/repository/user"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_GetByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := user.Repository()

	email := "admin@example.com"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password"}).
				AddRow("u1", "Admin", email, "hashed"))

		u, err := repo.GetByEmail(context.Background(), email)
		assert.NoError(t, err)
		assert.Equal(t, "u1", u.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.GetByEmail(context.Background(), email)
		assert.ErrorIs(t, err, domain.ErrInvalidUserCredentials)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(email).
			WillReturnError(assert.AnError)

		_, err := repo.GetByEmail(context.Background(), email)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_GetRolesByUserId(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	postgresdb.ConnectWithDB(db)
	repo := user.Repository()

	uid := "u1"

	t.Run("success", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(uid).
			WillReturnRows(sqlmock.NewRows([]string{"role"}).
				AddRow("ADMIN").AddRow("MECHANIC"))

		roles, err := repo.GetRolesByUserId(context.Background(), uid)
		assert.NoError(t, err)
		assert.Len(t, roles, 2)
		assert.Contains(t, roles, domain.ADMIN)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs(uid).
			WillReturnError(assert.AnError)

		_, err := repo.GetRolesByUserId(context.Background(), uid)
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := user.Repository()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE \"user\"").WithArgs("New Name", "new@e.com", "u1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Update(ctx, "u1", "New Name", "new@e.com")
		})
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE \"user\"").WithArgs("New Name", "new@e.com", "u1").
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Update(ctx, "u1", "New Name", "new@e.com")
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_AssignRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := user.Repository()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO user_role").WithArgs(sqlmock.AnyArg(), "u1", "ADMIN").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.AssignRole(ctx, "u1", domain.ADMIN)
		})
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("INSERT INTO user_role").WithArgs(sqlmock.AnyArg(), "u1", "ADMIN").
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.AssignRole(ctx, "u1", domain.ADMIN)
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_UpdateRole(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := user.Repository()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO user_role").WithArgs(sqlmock.AnyArg(), "u1", "MECHANIC").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.UpdateRole(ctx, "u1", domain.MECHANIC)
		})
		assert.NoError(t, err)
	})

	t.Run("delete error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.UpdateRole(ctx, "u1", domain.MECHANIC)
		})
		assert.Error(t, err)
	})

	t.Run("insert error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM user_role").WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO user_role").WithArgs(sqlmock.AnyArg(), "u1", "MECHANIC").
			WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.UpdateRole(ctx, "u1", domain.MECHANIC)
		})
		assert.Error(t, err)
	})
}

func TestPostgresRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	uowExec := postgresdb.NewTransactionalUoW(db)
	repo := user.Repository()

	t.Run("success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM \"user\"").WithArgs("u1").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, "u1")
		})
		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("DELETE FROM \"user\"").WithArgs("u1").WillReturnError(assert.AnError)
		mock.ExpectRollback()

		err = uowExec.Execute(context.Background(), func(ctx context.Context) error {
			return repo.Delete(ctx, "u1")
		})
		assert.Error(t, err)
	})
}

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
