package easql

import (
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
)

const errorSQL = "error to sql: %w"

type queryer struct {
	raw RawQueryer
}

func (q *queryer) Get(v interface{}, b squirrel.SelectBuilder) error {
	query, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf(errorSQL, err)
	}

	if err := q.raw.Get(v, query, args...); err != nil {
		return fmt.Errorf("error get: %w", err)
	}

	return nil
}

func (q *queryer) Select(v interface{}, b squirrel.SelectBuilder) error {
	query, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf(errorSQL, err)
	}

	if err := q.raw.Select(v, query, args...); err != nil {
		return fmt.Errorf("error select: %w", err)
	}

	return nil
}

func (q *queryer) execQuery(builder queryBuilder) (sql.Result, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf(errorSQL, err)
	}

	res, err := q.raw.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error exec: %w", err)
	}

	return res, nil
}

func (q *queryer) Insert(b squirrel.InsertBuilder) (sql.Result, error) {
	return q.execQuery(b)
}

func (q *queryer) Update(b squirrel.UpdateBuilder) (sql.Result, error) {
	return q.execQuery(b)
}

func (q *queryer) Delete(b squirrel.DeleteBuilder) (sql.Result, error) {
	return q.execQuery(b)
}
