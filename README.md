# easql

`easql` is a small wrapper around `sqlx` and `squirrel`.

It lets you build SQL with `squirrel` and execute it through the same API on both a regular database handle and a transaction.

## Why use it

- Build queries with `squirrel` instead of hand-writing SQL strings
- Execute reads and writes through a small, consistent API
- Use the same methods on `*easql.DB` and transaction handles
- Use context-aware and non-context variants with matching behavior

## Installation

```bash
go get github.com/adlandh/easql
```

## Overview

`easql` wraps an existing `*sqlx.DB`:

```go
raw := sqlx.NewDb(db, "mysql")
easy := easql.NewDB(raw)
```

The wrapper exposes query helpers for:

- `Get`
- `Select`
- `Insert`
- `Update`
- `Delete`
- `GetContext`
- `SelectContext`
- `InsertContext`
- `UpdateContext`
- `DeleteContext`

`Begin` and `BeginContext` return transaction wrappers that expose the same query methods plus `Commit` and `Rollback`.

## Example

```go
package main

import (
	"context"
	"log"

	sq "github.com/Masterminds/squirrel"
	"github.com/adlandh/easql"
	"github.com/jmoiron/sqlx"
	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

func main() {
	raw, err := sqlx.Connect("mysql", "user:pass@tcp(localhost:3306)/app")
	if err != nil {
		log.Fatal(err)
	}

	db := easql.NewDB(raw)

	var user User
	err = db.Get(&user, sq.Select("id", "name").From("users").Where(sq.Eq{"id": 1}))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	_, err = db.UpdateContext(ctx,
		sq.Update("users").Set("name", "leo").Where(sq.Eq{"id": user.ID}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

## Transactions

Transactions use the same query API as `DB`.

```go
tx, err := db.Begin()
if err != nil {
	return err
}
defer tx.Rollback()

_, err = tx.Insert(
	sq.Insert("users").Columns("id", "name").Values(1, "leo"),
)
if err != nil {
	return err
}

return tx.Commit()
```

For context-aware transaction creation, use `BeginContext(ctx)`.

## API

```go
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
```

## Notes

- `easql` does not create or manage database connections for you; pass in an existing `*sqlx.DB`
- `Raw()` returns the underlying `*sqlx.DB` when you need direct access to `sqlx`
- Errors are wrapped with operation context such as `error begin`, `error get`, or `error exec`
