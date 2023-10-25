package easql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
)

type queryerContext struct {
	raw RawQueryerContext
}

func (q *queryerContext) GetContext(ctx context.Context, v interface{}, b squirrel.SelectBuilder) error {
	query, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf("error to sql: %w", err)
	}

	if err := q.raw.GetContext(ctx, v, query, args...); err != nil {
		return fmt.Errorf("error get: %w", err)
	}

	return nil
}

func (q *queryerContext) SelectContext(ctx context.Context, v interface{}, b squirrel.SelectBuilder) error {
	query, args, err := b.ToSql()
	if err != nil {
		return fmt.Errorf("error to sql: %w", err)
	}

	if err := q.raw.SelectContext(ctx, v, query, args...); err != nil {
		return fmt.Errorf("error select: %w", err)
	}

	return nil
}

func (q *queryerContext) execQuery(ctx context.Context, builder queryBuilder) (sql.Result, error) {
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error to sql: %w", err)
	}

	res, err := q.raw.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error exec: %w", err)
	}

	return res, nil
}

func (q *queryerContext) InsertContext(ctx context.Context, b squirrel.InsertBuilder) (sql.Result, error) {
	return q.execQuery(ctx, b)
}

func (q *queryerContext) UpdateContext(ctx context.Context, b squirrel.UpdateBuilder) (sql.Result, error) {
	return q.execQuery(ctx, b)
}

func (q *queryerContext) DeleteContext(ctx context.Context, b squirrel.DeleteBuilder) (sql.Result, error) {
	return q.execQuery(ctx, b)
}
