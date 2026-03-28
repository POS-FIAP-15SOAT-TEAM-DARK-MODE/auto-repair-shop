package user

import (
	"context"
	"database/sql"
	"errors"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/infra/db/postgres"
	pgPkg "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/db/postgres"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/logger"
	"go.uber.org/zap"
)

type repo struct{}

func Repository() domain.UserRepository {
	return &repo{}
}

func (u *repo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	db, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	var user domain.User
	err = db.QueryRowContext(ctx, getUserByEmail, email).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidUserCredentials
		}

		return nil, pgPkg.Error(ctx, err)
	}

	return &user, nil
}

func (u *repo) GetRolesByUserId(ctx context.Context, id string) ([]domain.Role, error) {
	db, err := postgres.GetOneTimeTransaction(ctx)
	if err != nil {
		return nil, err
	}

	var roles []domain.Role
	rows, err := db.QueryContext(ctx, getRolesById, id)
	if err != nil {
		return nil, pgPkg.Error(ctx, err)
	}

	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role); err != nil {
			return nil, pgPkg.Error(ctx, err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (u *repo) Create(ctx context.Context, c *domain.User) error {
	tx, err := postgres.GetTransaction(ctx)
	if err != nil {
		return err
	}

	logger.Of(ctx).Debug("Executing query", zap.String("query", createUserQuery), zap.Any("params", c))
	if _, err = tx.ExecContext(ctx, createUserQuery, c.ID, c.Name, c.Email, c.Password); err != nil {
		return pgPkg.Error(ctx, err)
	}
	return nil
}
