package db_test

import (
	"testing"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/newInternal/pkg/db"
	"github.com/stretchr/testify/assert"
)

func TestQueryBuilder_Base(t *testing.T) {
	qb := db.QueryBuilder("SELECT * FROM table")
	query, args := qb.Build()

	assert.Equal(t, "SELECT * FROM table", query)
	assert.Empty(t, args)
}

func TestQueryBuilder_Conditions(t *testing.T) {
	qb := db.QueryBuilder("SELECT * FROM table")
	qb.Add("name =", "John")
	qb.Add("age >", 20)
	qb.Add("empty", "") // Should be ignored
	qb.Add("nil", nil)  // Should be ignored
	query, args := qb.Build()

	assert.Equal(t, "SELECT * FROM table WHERE name = $1 AND age > $2", query)
	assert.Equal(t, []any{"John", 20}, args)
}

func TestQueryBuilder_Pagination(t *testing.T) {
	qb := db.QueryBuilder("SELECT * FROM table")
	qb.AddPagination(10, 20)
	query, args := qb.Build()

	assert.Equal(t, "SELECT * FROM table LIMIT $1 OFFSET $2", query)
	assert.Equal(t, []any{int64(10), int64(20)}, args)
}

func TestQueryBuilder_OrderBy(t *testing.T) {
	t.Run("DefaultASC", func(t *testing.T) {
		qb := db.QueryBuilder("SELECT * FROM table")
		qb.OrderBy("created_at", "")
		query, _ := qb.Build()
		assert.Equal(t, "SELECT * FROM table ORDER BY created_at ASC", query)
	})

	t.Run("DESC", func(t *testing.T) {
		qb := db.QueryBuilder("SELECT * FROM table")
		qb.OrderBy("id", db.DESC)
		query, _ := qb.Build()
		assert.Equal(t, "SELECT * FROM table ORDER BY id DESC", query)
	})
}

func TestQueryBuilder_Complex(t *testing.T) {
	qb := db.QueryBuilder("SELECT * FROM table")
	qb.Add("status =", "ACTIVE")
	qb.OrderBy("id", db.ASC)
	qb.AddPagination(10, 0) // Offset 0 should not be added
	query, args := qb.Build()

	assert.Equal(t, "SELECT * FROM table WHERE status = $1 ORDER BY id ASC LIMIT $2", query)
	assert.Equal(t, []any{"ACTIVE", int64(10)}, args)
}
