// Package easql is a sqlx + squirrel wrapper
package easql

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

type DB struct {
	raw *sqlx.DB
	Queryer
	QueryerContext
}

var dbLock sync.Once
var dbInstance *DB

func NewDB(raw *sqlx.DB) *DB {
	dbLock.Do(func() {
		dbInstance = &DB{
			raw:            raw,
			Queryer:        &queryer{raw: raw},
			QueryerContext: &queryerContext{raw: raw},
		}
	})

	return dbInstance
}
