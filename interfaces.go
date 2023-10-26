package easql

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
)

type Queryer interface {
	Get(interface{}, squirrel.SelectBuilder) error
	Select(interface{}, squirrel.SelectBuilder) error
	Insert(squirrel.InsertBuilder) (sql.Result, error)
	Update(squirrel.UpdateBuilder) (sql.Result, error)
	Delete(squirrel.DeleteBuilder) (sql.Result, error)
}

type QueryerContext interface {
	GetContext(context.Context, interface{}, squirrel.SelectBuilder) error
	SelectContext(context.Context, interface{}, squirrel.SelectBuilder) error
	InsertContext(context.Context, squirrel.InsertBuilder) (sql.Result, error)
	UpdateContext(context.Context, squirrel.UpdateBuilder) (sql.Result, error)
	DeleteContext(context.Context, squirrel.DeleteBuilder) (sql.Result, error)
}

type RawQueryer interface {
	Get(interface{}, string, ...interface{}) error
	Select(interface{}, string, ...interface{}) error
	Exec(string, ...interface{}) (sql.Result, error)
}

type RawQueryerContext interface {
	GetContext(context.Context, interface{}, string, ...interface{}) error
	SelectContext(context.Context, interface{}, string, ...interface{}) error
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

type queryBuilder interface {
	ToSql() (string, []interface{}, error)
}

type Commiter interface {
	Commit() error
	Rollback() error
	Queryer
}

type CommiterContext interface {
	Commit() error
	Rollback() error
	QueryerContext
}

type Beginner interface {
	Begin() (Commiter, error)
}

type BeginnerContext interface {
	BeginContext(context.Context) (CommiterContext, error)
}
