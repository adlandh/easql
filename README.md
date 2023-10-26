# easql

A sqlx + squirrel wrapper

## Feature

- Provide convenient query/exec APIs.
- Provide same query/exec APIs to DB and Tx.

## Installation

```bash
$ go get github.com/adlandh/easql
```

## Usage

```go 
Get(interface{}, squirrel.SelectBuilder) error
Select(interface{}, squirrel.SelectBuilder) error
Insert(squirrel.InsertBuilder) (sql.Result, error)
Update(squirrel.UpdateBuilder) (sql.Result, error)
Delete(squirrel.DeleteBuilder) (sql.Result, error)
GetContext(context.Context, interface{}, squirrel.SelectBuilder) error
SelectContext(context.Context,interface{}, squirrel.SelectBuilder) error
InsertContext(context.Context,squirrel.InsertBuilder) (sql.Result, error)
UpdateContext(context.Context,squirrel.UpdateBuilder) (sql.Result, error)
DeleteContext(context.Context,squirrel.DeleteBuilder) (sql.Result, error)
```