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

func TestNewDB(t *testing.T) {
	t.Parallel()

	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)

	wrapped := sqlx.NewDb(mockDB, "mysql")
	db := NewDB(wrapped)

	assert.NotNil(t, db)
	assert.Same(t, wrapped, db.Raw())
	assert.Implements(t, (*Queryer)(nil), db)
	assert.Implements(t, (*QueryerContext)(nil), db)
}

func TestDBImplementsGet(t *testing.T) {
	t.Parallel()

	db, _ := newTestDB(t)
	assert.Implements(t, (*Queryer)(nil), db)
}

func TestDBImplementsGetContext(t *testing.T) {
	t.Parallel()

	db, _ := newTestDB(t)
	assert.Implements(t, (*QueryerContext)(nil), db)
}

func TestTxImplementsGet(t *testing.T) {
	db, mock := newTestDB(t)
	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	assert.Implements(t, (*Queryer)(nil), tx)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxImplementsGetContext(t *testing.T) {
	db, mock := newTestDB(t)
	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	assert.Implements(t, (*QueryerContext)(nil), tx)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRollback(t *testing.T) {
	db, mock := newTestDB(t)
	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	assert.NoError(t, tx.Rollback())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func runQueryerCases(t *testing.T, fn func(Queryer), expect func(sqlmock.Sqlmock), begin func(*DB) (Commiter, error)) {
	t.Helper()

	t.Run("DB", func(t *testing.T) {
		db, mock := newTestDB(t)
		expect(mock)

		fn(db)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Tx", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectBegin()
		expect(mock)
		mock.ExpectCommit()

		tx, err := begin(db)
		require.NoError(t, err)

		fn(tx)
		assert.NoError(t, tx.Commit())
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func runQueryerContextCases(t *testing.T, fn func(QueryerContext), expect func(sqlmock.Sqlmock), begin func(*DB) (CommiterContext, error)) {
	t.Helper()

	t.Run("DB", func(t *testing.T) {
		db, mock := newTestDB(t)
		expect(mock)

		fn(db)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Tx", func(t *testing.T) {
		db, mock := newTestDB(t)
		mock.ExpectBegin()
		expect(mock)
		mock.ExpectCommit()

		tx, err := begin(db)
		require.NoError(t, err)

		fn(tx)
		assert.NoError(t, tx.Commit())
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestQueryerGet(t *testing.T) {
	fn := func(q Queryer) {
		var id int
		_ = q.Get(&id, sq.Select("id").From("users").Where(sq.Eq{"id": 1}))
	}

	runQueryerCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(selectFromUsersWhere).WithArgs(1)
	}, func(db *DB) (Commiter, error) {
		return db.Begin()
	})
}

func TestQueryerGetContext(t *testing.T) {
	ctx := context.Background()
	fn := func(q QueryerContext) {
		var id int
		_ = q.GetContext(ctx, &id, sq.Select("id").From("users").Where(sq.Eq{"id": 1}))
	}

	runQueryerContextCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(selectFromUsersWhere).WithArgs(1)
	}, func(db *DB) (CommiterContext, error) {
		return db.BeginContext(ctx)
	})
}

func TestQuerySelect(t *testing.T) {
	fn := func(q Queryer) {
		var ids []int
		_ = q.Select(&ids, sq.Select("id").From("users"))
	}

	runQueryerCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(selectFromUsers)
	}, func(db *DB) (Commiter, error) {
		return db.Begin()
	})
}

func TestQuerySelectContext(t *testing.T) {
	ctx := context.Background()
	fn := func(q QueryerContext) {
		var ids []int
		_ = q.SelectContext(ctx, &ids, sq.Select("id").From("users"))
	}

	runQueryerContextCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectQuery(selectFromUsers)
	}, func(db *DB) (CommiterContext, error) {
		return db.BeginContext(ctx)
	})
}

func TestQueryInsert(t *testing.T) {
	fn := func(q Queryer) {
		_, _ = q.Insert(sq.Insert("users").Columns("id").Values(1))
	}

	runQueryerCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(insertIntoUsers).WithArgs(1)
	}, func(db *DB) (Commiter, error) {
		return db.Begin()
	})
}

func TestQueryInsertContext(t *testing.T) {
	ctx := context.Background()
	fn := func(q QueryerContext) {
		_, _ = q.InsertContext(ctx, sq.Insert("users").Columns("id").Values(1))
	}

	runQueryerContextCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(insertIntoUsers).WithArgs(1)
	}, func(db *DB) (CommiterContext, error) {
		return db.BeginContext(ctx)
	})
}

func TestQueryUpdate(t *testing.T) {
	fn := func(q Queryer) {
		_, _ = q.Update(sq.Update("users").Set("name", "leo").Where(sq.Eq{"id": 1}))
	}

	runQueryerCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(updateUsers).WithArgs("leo", 1)
	}, func(db *DB) (Commiter, error) {
		return db.Begin()
	})
}

func TestQueryUpdateContext(t *testing.T) {
	ctx := context.Background()
	fn := func(q QueryerContext) {
		_, _ = q.UpdateContext(ctx, sq.Update("users").Set("name", "leo").Where(sq.Eq{"id": 1}))
	}

	runQueryerContextCases(t, fn, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(updateUsers).WithArgs("leo", 1)
	}, func(db *DB) (CommiterContext, error) {
		return db.BeginContext(ctx)
	})
}

func TestQueryDelete(t *testing.T) {
	doQuery := func(q Queryer) {
		_, _ = q.Delete(sq.Delete("users").Where(sq.Eq{"id": 1}))
	}

	runQueryerCases(t, doQuery, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(deleteFromUsers).WithArgs(1)
	}, func(db *DB) (Commiter, error) {
		return db.Begin()
	})
}

func TestQueryDeleteContext(t *testing.T) {
	ctx := context.Background()
	doQuery := func(q QueryerContext) {
		_, _ = q.DeleteContext(ctx, sq.Delete("users").Where(sq.Eq{"id": 1}))
	}

	runQueryerContextCases(t, doQuery, func(mock sqlmock.Sqlmock) {
		mock.ExpectExec(deleteFromUsers).WithArgs(1)
	}, func(db *DB) (CommiterContext, error) {
		return db.BeginContext(ctx)
	})
}
