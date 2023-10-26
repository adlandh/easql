package easql

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

var _ Commiter = (*Tx)(nil)
var _ CommiterContext = (*Tx)(nil)

type Tx struct {
	raw *sqlx.Tx
	Queryer
	QueryerContext
}

func newTx(raw *sqlx.Tx) *Tx {
	return &Tx{
		raw:            raw,
		Queryer:        &queryer{raw: raw},
		QueryerContext: &queryerContext{raw: raw},
	}
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
