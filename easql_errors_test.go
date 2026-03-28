package easql

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResult struct{}

func (stubResult) LastInsertId() (int64, error) { return 0, nil }
func (stubResult) RowsAffected() (int64, error) { return 0, nil }

type stubBuilder struct {
	query string
	args  []interface{}
	err   error
}

type contextKey string

func newTestContext() context.Context {
	return context.WithValue(context.Background(), contextKey("key"), "value")
}

func assertSelectBuilderError(t *testing.T, err error) {
	t.Helper()
	assert.EqualError(t, err, "error to sql: select statements must have at least one result column")
}

func (b stubBuilder) ToSql() (string, []interface{}, error) {
	if b.err != nil {
		return "", nil, b.err
	}

	return b.query, b.args, nil
}

type stubRawQueryer struct {
	getErr         error
	selectErr      error
	execErr        error
	execResult     sql.Result
	gotGetQuery    string
	gotGetArgs     []interface{}
	gotSelectQuery string
	gotSelectArgs  []interface{}
	gotExecQuery   string
	gotExecArgs    []interface{}
}

func (r *stubRawQueryer) Get(_ interface{}, query string, args ...interface{}) error {
	r.gotGetQuery = query
	r.gotGetArgs = args
	return r.getErr
}

func (r *stubRawQueryer) Select(_ interface{}, query string, args ...interface{}) error {
	r.gotSelectQuery = query
	r.gotSelectArgs = args
	return r.selectErr
}

func (r *stubRawQueryer) Exec(query string, args ...interface{}) (sql.Result, error) {
	r.gotExecQuery = query
	r.gotExecArgs = args
	return r.execResult, r.execErr
}

type stubRawQueryerContext struct {
	getErr         error
	selectErr      error
	execErr        error
	execResult     sql.Result
	gotGetCtx      context.Context
	gotSelectCtx   context.Context
	gotExecCtx     context.Context
	gotGetQuery    string
	gotGetArgs     []interface{}
	gotSelectQuery string
	gotSelectArgs  []interface{}
	gotExecQuery   string
	gotExecArgs    []interface{}
}

func (r *stubRawQueryerContext) GetContext(ctx context.Context, _ interface{}, query string, args ...interface{}) error {
	r.gotGetCtx = ctx
	r.gotGetQuery = query
	r.gotGetArgs = args
	return r.getErr
}

func (r *stubRawQueryerContext) SelectContext(ctx context.Context, _ interface{}, query string, args ...interface{}) error {
	r.gotSelectCtx = ctx
	r.gotSelectQuery = query
	r.gotSelectArgs = args
	return r.selectErr
}

func (r *stubRawQueryerContext) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	r.gotExecCtx = ctx
	r.gotExecQuery = query
	r.gotExecArgs = args
	return r.execResult, r.execErr
}

func TestDBRawExposesUnderlyingDB(t *testing.T) {
	db, _ := newTestDB(t)
	assert.Same(t, db.raw, db.Raw())
}

func TestDBBeginWrapsError(t *testing.T) {
	db, mock := newTestDB(t)
	wantErr := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(wantErr)

	tx, err := db.Begin()

	assert.Nil(t, tx)
	assert.EqualError(t, err, "error begin: begin failed")
	assert.ErrorIs(t, err, wantErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDBBeginContextWrapsError(t *testing.T) {
	db, mock := newTestDB(t)
	wantErr := errors.New("begin failed")
	mock.ExpectBegin().WillReturnError(wantErr)

	tx, err := db.BeginContext(context.Background())

	assert.Nil(t, tx)
	assert.EqualError(t, err, "error begin: begin failed")
	assert.ErrorIs(t, err, wantErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxCommitWrapsError(t *testing.T) {
	db, mock := newTestDB(t)
	wantErr := errors.New("commit failed")
	mock.ExpectBegin()
	mock.ExpectCommit().WillReturnError(wantErr)

	tx, err := db.Begin()
	require.NoError(t, err)

	err = tx.Commit()

	assert.EqualError(t, err, "error commit: commit failed")
	assert.ErrorIs(t, err, wantErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTxRollbackWrapsError(t *testing.T) {
	db, mock := newTestDB(t)
	wantErr := errors.New("rollback failed")
	mock.ExpectBegin()
	mock.ExpectRollback().WillReturnError(wantErr)

	tx, err := db.Begin()
	require.NoError(t, err)

	err = tx.Rollback()

	assert.EqualError(t, err, "error rollback: rollback failed")
	assert.ErrorIs(t, err, wantErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestQueryerGetPassesSQLAndArgs(t *testing.T) {
	raw := &stubRawQueryer{}
	q := &queryer{raw: raw}
	var got int

	err := q.Get(&got, sq.Select("id").From("users").Where(sq.Eq{"id": 7}))

	assert.NoError(t, err)
	assert.Equal(t, "SELECT id FROM users WHERE id = ?", raw.gotGetQuery)
	assert.Equal(t, []interface{}{7}, raw.gotGetArgs)
}

func TestQueryerGetWrapsRawError(t *testing.T) {
	wantErr := errors.New("get failed")
	q := &queryer{raw: &stubRawQueryer{getErr: wantErr}}
	var got int

	err := q.Get(&got, sq.Select("id").From("users"))

	assert.EqualError(t, err, "error get: get failed")
	assert.ErrorIs(t, err, wantErr)
}

func TestQueryerGetWrapsBuilderError(t *testing.T) {
	q := &queryer{raw: &stubRawQueryer{}}
	var got int

	err := q.Get(&got, sq.SelectBuilder{})

	assertSelectBuilderError(t, err)
}

func TestQueryerSelectPassesSQLAndArgs(t *testing.T) {
	raw := &stubRawQueryer{}
	q := &queryer{raw: raw}
	var got []int

	err := q.Select(&got, sq.Select("id").From("users").Where(sq.Eq{"id": 3}))

	assert.NoError(t, err)
	assert.Equal(t, "SELECT id FROM users WHERE id = ?", raw.gotSelectQuery)
	assert.Equal(t, []interface{}{3}, raw.gotSelectArgs)
}

func TestQueryerSelectWrapsRawError(t *testing.T) {
	wantErr := errors.New("select failed")
	q := &queryer{raw: &stubRawQueryer{selectErr: wantErr}}
	var got []int

	err := q.Select(&got, sq.Select("id").From("users"))

	assert.EqualError(t, err, "error select: select failed")
	assert.ErrorIs(t, err, wantErr)
}

func TestQueryerSelectWrapsBuilderError(t *testing.T) {
	q := &queryer{raw: &stubRawQueryer{}}
	var got []int

	err := q.Select(&got, sq.SelectBuilder{})

	assertSelectBuilderError(t, err)
}

func TestQueryerExecQueryPassesSQLAndArgs(t *testing.T) {
	wantResult := stubResult{}
	raw := &stubRawQueryer{execResult: wantResult}
	q := &queryer{raw: raw}

	res, err := q.execQuery(stubBuilder{query: "UPDATE users SET name = ?", args: []interface{}{"leo"}})

	assert.NoError(t, err)
	assert.Equal(t, wantResult, res)
	assert.Equal(t, "UPDATE users SET name = ?", raw.gotExecQuery)
	assert.Equal(t, []interface{}{"leo"}, raw.gotExecArgs)
}

func TestQueryerExecQueryWrapsBuilderError(t *testing.T) {
	wantErr := errors.New("bad builder")
	q := &queryer{raw: &stubRawQueryer{}}

	res, err := q.execQuery(stubBuilder{err: wantErr})

	assert.Nil(t, res)
	assert.EqualError(t, err, "error to sql: bad builder")
	assert.ErrorIs(t, err, wantErr)
}

func TestQueryerExecQueryWrapsRawError(t *testing.T) {
	wantErr := errors.New("exec failed")
	q := &queryer{raw: &stubRawQueryer{execErr: wantErr}}

	res, err := q.execQuery(stubBuilder{query: "DELETE FROM users", args: []interface{}{}})

	assert.Nil(t, res)
	assert.EqualError(t, err, "error exec: exec failed")
	assert.ErrorIs(t, err, wantErr)
}

func TestQueryerContextGetPassesContextSQLAndArgs(t *testing.T) {
	raw := &stubRawQueryerContext{}
	q := &queryerContext{raw: raw}
	ctx := newTestContext()
	var got int

	err := q.GetContext(ctx, &got, sq.Select("id").From("users").Where(sq.Eq{"id": 9}))

	assert.NoError(t, err)
	assert.Same(t, ctx, raw.gotGetCtx)
	assert.Equal(t, "SELECT id FROM users WHERE id = ?", raw.gotGetQuery)
	assert.Equal(t, []interface{}{9}, raw.gotGetArgs)
}

func TestQueryerContextGetWrapsRawError(t *testing.T) {
	wantErr := errors.New("get context failed")
	q := &queryerContext{raw: &stubRawQueryerContext{getErr: wantErr}}
	var got int

	err := q.GetContext(context.Background(), &got, sq.Select("id").From("users"))

	assert.EqualError(t, err, "error get: get context failed")
	assert.ErrorIs(t, err, wantErr)
}

func TestQueryerContextGetWrapsBuilderError(t *testing.T) {
	q := &queryerContext{raw: &stubRawQueryerContext{}}
	var got int

	err := q.GetContext(context.Background(), &got, sq.SelectBuilder{})

	assertSelectBuilderError(t, err)
}

func TestQueryerContextSelectPassesContextSQLAndArgs(t *testing.T) {
	raw := &stubRawQueryerContext{}
	q := &queryerContext{raw: raw}
	ctx := newTestContext()
	var got []int

	err := q.SelectContext(ctx, &got, sq.Select("id").From("users").Where(sq.Eq{"id": 4}))

	assert.NoError(t, err)
	assert.Same(t, ctx, raw.gotSelectCtx)
	assert.Equal(t, "SELECT id FROM users WHERE id = ?", raw.gotSelectQuery)
	assert.Equal(t, []interface{}{4}, raw.gotSelectArgs)
}

func TestQueryerContextSelectWrapsRawError(t *testing.T) {
	wantErr := errors.New("select context failed")
	q := &queryerContext{raw: &stubRawQueryerContext{selectErr: wantErr}}
	var got []int

	err := q.SelectContext(context.Background(), &got, sq.Select("id").From("users"))

	assert.EqualError(t, err, "error select: select context failed")
	assert.ErrorIs(t, err, wantErr)
}

func TestQueryerContextSelectWrapsBuilderError(t *testing.T) {
	q := &queryerContext{raw: &stubRawQueryerContext{}}
	var got []int

	err := q.SelectContext(context.Background(), &got, sq.SelectBuilder{})

	assertSelectBuilderError(t, err)
}

func TestQueryerContextExecQueryPassesContextSQLAndArgs(t *testing.T) {
	wantResult := stubResult{}
	raw := &stubRawQueryerContext{execResult: wantResult}
	q := &queryerContext{raw: raw}
	ctx := newTestContext()

	res, err := q.execQuery(ctx, stubBuilder{query: "INSERT INTO users(id) VALUES(?)", args: []interface{}{1}})

	assert.NoError(t, err)
	assert.Equal(t, wantResult, res)
	assert.Same(t, ctx, raw.gotExecCtx)
	assert.Equal(t, "INSERT INTO users(id) VALUES(?)", raw.gotExecQuery)
	assert.Equal(t, []interface{}{1}, raw.gotExecArgs)
}

func TestQueryerContextExecQueryWrapsBuilderError(t *testing.T) {
	wantErr := errors.New("bad builder")
	q := &queryerContext{raw: &stubRawQueryerContext{}}

	res, err := q.execQuery(context.Background(), stubBuilder{err: wantErr})

	assert.Nil(t, res)
	assert.EqualError(t, err, "error to sql: bad builder")
	assert.ErrorIs(t, err, wantErr)
}

func TestQueryerContextExecQueryWrapsRawError(t *testing.T) {
	wantErr := errors.New("exec context failed")
	q := &queryerContext{raw: &stubRawQueryerContext{execErr: wantErr}}

	res, err := q.execQuery(context.Background(), stubBuilder{query: "DELETE FROM users", args: []interface{}{}})

	assert.Nil(t, res)
	assert.EqualError(t, err, "error exec: exec context failed")
	assert.ErrorIs(t, err, wantErr)
}
