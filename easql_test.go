package easql

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

var (
	db   *DB
	mock sqlmock.Sqlmock
	err  error
)

func TestMain(m *testing.M) {
	var raw *sql.DB
	raw, mock, err = sqlmock.New()
	if err != nil {
		return
	}
	db = NewDB(sqlx.NewDb(raw, "mysql"))
	os.Exit(m.Run())
}

func TestNewDB(t *testing.T) {
	t.Parallel()
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestDBImplementsGet(t *testing.T) {
	assert.Implements(t, (*Queryer)(nil), db)
}

func TestDBImplementsGetContext(t *testing.T) {
	assert.Implements(t, (*QueryerContext)(nil), db)
}

func TestTxImplementsGet(t *testing.T) {
	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)
	assert.Implements(t, (*Queryer)(nil), tx)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxImplementsGetContext(t *testing.T) {
	mock.ExpectBegin()
	tx, err := db.Begin()
	assert.NoError(t, err)
	assert.Implements(t, (*QueryerContext)(nil), tx)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRollback(t *testing.T) {
	mock.ExpectBegin()
	mock.ExpectRollback()

	tx, _ := db.Begin()
	assert.NoError(t, tx.Rollback())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testQuery(queryer Queryer, fn func(Queryer)) {
	fn(queryer)
}

func testQueryContext(queryer QueryerContext, fn func(QueryerContext)) {
	fn(queryer)
}

func TestQueryer_Get(t *testing.T) {
	t.Parallel()
	fn := func(q Queryer) {
		var id int
		_ = q.Get(&id, sq.Select("id").From("users").
			Where(sq.Eq{"id": 1}))
	}

	// DB
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(1)
	testQuery(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(1)
	mock.ExpectCommit()
	tx, _ := db.Begin()
	testQuery(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryer_GetContext(t *testing.T) {
	ctx := context.Background()
	fn := func(q QueryerContext) {
		var id int
		_ = q.GetContext(ctx, &id, sq.Select("id").From("users").
			Where(sq.Eq{"id": 1}))
	}

	// DB
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(1)
	testQueryContext(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users WHERE id=?").WithArgs(1)
	mock.ExpectCommit()
	tx, _ := db.BeginContext(ctx)
	testQueryContext(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_Select(t *testing.T) {
	fn := func(q Queryer) {
		var ids []int
		_ = q.Select(&ids, sq.Select("id").From("users"))
	}

	// DB
	mock.ExpectQuery("SELECT id FROM users")
	testQuery(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users")
	mock.ExpectCommit()
	tx, _ := db.Begin()
	testQuery(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_SelectContext(t *testing.T) {
	ctx := context.Background()
	fn := func(q QueryerContext) {
		var ids []int
		_ = q.SelectContext(ctx, &ids, sq.Select("id").From("users"))
	}

	// DB
	mock.ExpectQuery("SELECT id FROM users")
	testQueryContext(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM users")
	mock.ExpectCommit()
	tx, _ := db.BeginContext(ctx)
	testQueryContext(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_Insert(t *testing.T) {
	fn := func(q Queryer) {
		_, _ = q.Insert(sq.Insert("users").Columns("id").
			Values(1))
	}

	expectQuery := func() {
		mock.ExpectExec("INSERT INTO users").WithArgs(1)
	}

	// DB
	expectQuery()
	testQuery(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	expectQuery()
	tx, _ := db.Begin()
	testQuery(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_InsertContext(t *testing.T) {
	ctx := context.Background()
	fn := func(q QueryerContext) {
		_, _ = q.InsertContext(ctx, sq.Insert("users").Columns("id").
			Values(1))
	}

	expectQuery := func() {
		mock.ExpectExec("INSERT INTO users").WithArgs(1)
	}

	// DB
	expectQuery()
	testQueryContext(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	expectQuery()
	tx, _ := db.BeginContext(ctx)
	testQueryContext(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_Update(t *testing.T) {
	fn := func(q Queryer) {
		_, _ = q.Update(sq.Update("users").Set("name", "leo").Where(sq.Eq{"id": 1}))
	}

	expectQuery := func() {
		mock.ExpectExec("UPDATE users SET name").WithArgs("leo", 1)
	}

	// DB
	expectQuery()
	testQuery(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	expectQuery()
	tx, _ := db.Begin()
	testQuery(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_UpdateContext(t *testing.T) {
	ctx := context.Background()

	fn := func(q QueryerContext) {
		_, _ = q.UpdateContext(ctx, sq.Update("users").Set("name", "leo").Where(sq.Eq{"id": 1}))
	}

	expectQuery := func() {
		mock.ExpectExec("UPDATE users SET name").WithArgs("leo", 1)
	}

	// DB
	expectQuery()
	testQueryContext(db, fn)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	expectQuery()
	tx, _ := db.BeginContext(ctx)
	testQueryContext(tx, fn)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_Delete(t *testing.T) {
	doQuery := func(q Queryer) {
		func(q Queryer) {
			_, _ = q.Delete(sq.Delete("users").Where(sq.Eq{"id": 1}))
		}(q)
	}

	expectQuery := func() {
		mock.ExpectExec("DELETE FROM users").WithArgs(1)
	}

	// DB
	expectQuery()
	doQuery(db)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	expectQuery()
	tx, _ := db.Begin()
	doQuery(tx)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQuery_DeleteContext(t *testing.T) {
	ctx := context.Background()
	doQuery := func(q QueryerContext) {
		func(q QueryerContext) {
			_, _ = q.DeleteContext(ctx, sq.Delete("users").Where(sq.Eq{"id": 1}))
		}(q)
	}

	expectQuery := func() {
		mock.ExpectExec("DELETE FROM users").WithArgs(1)
	}

	// DB
	expectQuery()
	doQuery(db)
	assert.NoError(t, mock.ExpectationsWereMet())

	// Transaction
	mock.ExpectBegin()
	expectQuery()
	tx, _ := db.BeginContext(ctx)
	doQuery(tx)
	_ = tx.Commit()
	assert.NoError(t, mock.ExpectationsWereMet())
}
