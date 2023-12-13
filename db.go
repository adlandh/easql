// Package easql is a sqlx + squirrel wrapper
package easql

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var _ Beginner = (*DB)(nil)
var _ BeginnerContext = (*DB)(nil)

type DB struct {
	raw *sqlx.DB
	Queryer
	QueryerContext
}

func NewDB(raw *sqlx.DB) *DB {
	return &DB{
		raw:            raw,
		Queryer:        &queryer{raw: raw},
		QueryerContext: &queryerContext{raw: raw},
	}
}

func (db *DB) Begin() (Commiter, error) {
	raw, err := db.raw.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error begin: %w", err)
	}

	return newTx(raw), nil
}

func (db *DB) BeginContext(ctx context.Context) (CommiterContext, error) {
	raw, err := db.raw.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error begin: %w", err)
	}

	return newTx(raw), nil
}
