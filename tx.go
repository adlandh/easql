package easql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Tx struct {
	raw *sqlx.Tx
	Queryer
}

func newTx(raw *sqlx.Tx) *Tx {
	return &Tx{
		raw:     raw,
		Queryer: &queryer{raw: raw},
	}
}

func (db *DB) Begin() (*Tx, error) {
	raw, err := db.raw.Beginx()
	if err != nil {
		return nil, fmt.Errorf("error begin: %w", err)
	}

	return newTx(raw), nil
}

func (tx *Tx) Commit() error {
	if err := tx.raw.Commit(); err != nil {
		return fmt.Errorf("error commit: %w", err)
	}

	return nil
}

func (tx *Tx) Rollback() error {
	if err := tx.raw.Rollback(); err != nil {
		return fmt.Errorf("error rollback: %w", err)
	}

	return nil
}
