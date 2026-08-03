# hush Example

This directory contains the official porting contract for integrating hush with
a database or query builder that does not yet have a first-class adapter.

## Running

```bash
go run ./_example
```

## Filter Walker Pattern (porting contract)

**This is the official reference contract for integrating hush with any
database or query builder.** It teaches the core pattern every adapter must
support:

- Type-switching on `hush.Filter` to handle `Condition`, `And`, `Or`, `Not`
- Recursive walking with depth tracking
- Translating `hush.Condition` operators to target expressions
- Reading values through `valueAt`, which uses the type-coerced
  [hush.Condition.Values] populated by `hush.Parse` and falls back to
  [hush.Coerce] for hand-built queries
- Escaping LIKE wildcards in user input with [hush.EscapeLike]
- Detecting the value-less `$null` / `$notNull` operators via
  [hush.IsNullOperator] / [hush.IsNotNullOperator]

To integrate hush with a new database, copy this example and replace the string
output with your target query builder calls.

> GORM is covered by a first-class adapter instead of a print-style example:
> [`hush/gorm`](./gorm) ships the complete operator matrix as a ready-made
> `db.Scopes(gorm.Scopes(schema, q))` scope, tested against SQLite and Postgres.
> See the [README](./README.md#gorm-adapter) for usage and the runnable
> [`gorm.Scopes` example](https://pkg.go.dev/github.com/DhimasYulian/hush/gorm#example-Scopes).

> MongoDB is covered by a first-class adapter instead of a print-style example:
> [`hush/mongo`](./mongo) ships the complete operator matrix as ready-made
> translation functions — `mongo.Filter`, `mongo.FindOptions`, and
> `mongo.Pipeline` / `mongo.PipelineFacet` — tested against `mtest` and a real
> mongod. See the [README](./README.md#mongodb-adapter) for usage and the
> runnable [`mongo.Pipeline` example](https://pkg.go.dev/github.com/DhimasYulian/hush/mongo#example-Pipeline).

## Adding Your Own Integration

The `_example/main.go` example is the official porting contract. To integrate
hush with a new database:

1. Copy the `walkFilterVerbose` / `walkCondition` type-switch skeleton
2. Replace the string output with your target query builder calls
3. Read condition values through the `valueAt` pattern so you bind coerced
   `Condition.Values` (falling back to `hush.Coerce`)
4. Escape user input for pattern operators with `hush.EscapeLike`
5. Handle `Sort`, `Fields`, and `Pagination` separately (they're straightforward field/value mappings)
