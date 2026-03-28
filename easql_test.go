package easql

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	selectFromUsers      = "SELECT id FROM users"
	selectFromUsersWhere = selectFromUsers + " WHERE id = ?"
	insertIntoUsers      = "INSERT INTO users"
	deleteFromUsers      = "DELETE FROM users WHERE id = ?"
	updateUsers          = "UPDATE users SET name = ?"
)

func newTestDB(t *testing.T) (*DB, sqlmock.Sqlmock) {
	t.Helper()

	raw, mock, err := sqlmock.New()
	require.NoError(t, err)

	return NewDB(sqlx.NewDb(raw, "mysql")), mock
}

func runDBCase(t *testing.T, expect func(sqlmock.Sqlmock), run func(*DB)) {
	t.Helper()

	db, mock := newTestDB(t)
	expect(mock)
	run(db)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func runTxCase(t *testing.T, expect func(sqlmock.Sqlmock), run func(Commiter), begin func(*DB) (Commiter, error)) {
	t.Helper()

	db, mock := newTestDB(t)
	mock.ExpectBegin()
	expect(mock)
	mock.ExpectCommit()

	tx, err := begin(db)
	require.NoError(t, err)
	defer func() {
		_ = tx.Rollback()
	}()

	run(tx)
	assert.NoError(t, tx.Commit())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func runDBAndTxCase(t *testing.T, expect func(sqlmock.Sqlmock), run func(Queryer)) {
	t.Helper()

	t.Run("DB", func(t *testing.T) {
		runDBCase(t, expect, func(db *DB) {
			run(db)
		})
	})

	t.Run("Tx", func(t *testing.T) {
		runTxCase(t, expect, func(tx Commiter) {
			run(tx)
		}, func(db *DB) (Commiter, error) {
			return db.Begin()
		})
	})
}

func runDBAndTxContextCase(t *testing.T, expect func(sqlmock.Sqlmock), run func(QueryerContext), ctx context.Context) {
	t.Helper()

	t.Run("DB", func(t *testing.T) {
		runDBCase(t, expect, func(db *DB) {
			run(db)
		})
	})

	t.Run("Tx", func(t *testing.T) {
		runTxCase(t, expect, func(tx Commiter) {
			run(tx.(CommiterContext))
		}, func(db *DB) (Commiter, error) {
			ctxTx, err := db.BeginContext(ctx)
			if err != nil {
				return nil, err
			}

			return ctxTx.(Commiter), nil
		})
	})
}

func TestNewDB(t *testing.T) {
	t.Parallel()

	db, _ := newTestDB(t)

	assert.NotNil(t, db)
	assert.Same(t, db.raw, db.Raw())
	assert.Implements(t, (*Queryer)(nil), db)
	assert.Implements(t, (*QueryerContext)(nil), db)
}

func TestImplementsInterfaces(t *testing.T) {
	t.Parallel()

	t.Run("DB", func(t *testing.T) {
		db, _ := newTestDB(t)
		assert.Implements(t, (*Queryer)(nil), db)
		assert.Implements(t, (*QueryerContext)(nil), db)
	})

	t.Run("Tx", func(t *testing.T) {
		runDBCase(t, func(mock sqlmock.Sqlmock) {
			mock.ExpectBegin()
		}, func(db *DB) {
			tx, err := db.Begin()
			require.NoError(t, err)
			defer func() {
				_ = tx.Rollback()
			}()
			assert.Implements(t, (*Queryer)(nil), tx)
			assert.Implements(t, (*QueryerContext)(nil), tx)
		})
	})
}

func TestRollback(t *testing.T) {
	runDBCase(t, func(mock sqlmock.Sqlmock) {
		mock.ExpectBegin()
		mock.ExpectRollback()
	}, func(db *DB) {
		tx, err := db.Begin()
		require.NoError(t, err)
		defer func() {
			_ = tx.Rollback()
		}()
		assert.NoError(t, tx.Rollback())
	})
}

func TestQueryerMethods(t *testing.T) {
	queryerCases := []struct {
		name   string
		expect func(sqlmock.Sqlmock)
		run    func(Queryer)
	}{
		{
			name: "Get",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectFromUsersWhere).WithArgs(1)
			},
			run: func(q Queryer) {
				var id int
				_ = q.Get(&id, sq.Select("id").From("users").Where(sq.Eq{"id": 1}))
			},
		},
		{
			name: "Select",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectFromUsers)
			},
			run: func(q Queryer) {
				var ids []int
				_ = q.Select(&ids, sq.Select("id").From("users"))
			},
		},
		{
			name: "Insert",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(insertIntoUsers).WithArgs(1)
			},
			run: func(q Queryer) {
				_, _ = q.Insert(sq.Insert("users").Columns("id").Values(1))
			},
		},
		{
			name: "Update",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateUsers).WithArgs("leo", 1)
			},
			run: func(q Queryer) {
				_, _ = q.Update(sq.Update("users").Set("name", "leo").Where(sq.Eq{"id": 1}))
			},
		},
		{
			name: "Delete",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(deleteFromUsers).WithArgs(1)
			},
			run: func(q Queryer) {
				_, _ = q.Delete(sq.Delete("users").Where(sq.Eq{"id": 1}))
			},
		},
	}

	for _, tc := range queryerCases {
		t.Run(tc.name, func(t *testing.T) {
			runDBAndTxCase(t, tc.expect, tc.run)
		})
	}
}

func TestQueryerContextMethods(t *testing.T) {
	ctx := context.Background()
	queryerContextCases := []struct {
		name   string
		expect func(sqlmock.Sqlmock)
		run    func(QueryerContext)
	}{
		{
			name: "GetContext",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectFromUsersWhere).WithArgs(1)
			},
			run: func(q QueryerContext) {
				var id int
				_ = q.GetContext(ctx, &id, sq.Select("id").From("users").Where(sq.Eq{"id": 1}))
			},
		},
		{
			name: "SelectContext",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(selectFromUsers)
			},
			run: func(q QueryerContext) {
				var ids []int
				_ = q.SelectContext(ctx, &ids, sq.Select("id").From("users"))
			},
		},
		{
			name: "InsertContext",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(insertIntoUsers).WithArgs(1)
			},
			run: func(q QueryerContext) {
				_, _ = q.InsertContext(ctx, sq.Insert("users").Columns("id").Values(1))
			},
		},
		{
			name: "UpdateContext",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(updateUsers).WithArgs("leo", 1)
			},
			run: func(q QueryerContext) {
				_, _ = q.UpdateContext(ctx, sq.Update("users").Set("name", "leo").Where(sq.Eq{"id": 1}))
			},
		},
		{
			name: "DeleteContext",
			expect: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(deleteFromUsers).WithArgs(1)
			},
			run: func(q QueryerContext) {
				_, _ = q.DeleteContext(ctx, sq.Delete("users").Where(sq.Eq{"id": 1}))
			},
		},
	}

	for _, tc := range queryerContextCases {
		t.Run(tc.name, func(t *testing.T) {
			runDBAndTxContextCase(t, tc.expect, tc.run, ctx)
		})
	}
}
