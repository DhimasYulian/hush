# hush Examples

This directory contains integration examples showing how to translate hush queries into different database query formats.

Each example follows the same pattern:

1. Define a schema for the `article` resource
2. Construct realistic `url.Values` with filters, sort, fields, and pagination
3. Parse with `hush.Parse()`
4. Translate the resulting `*hush.Query` into the target format
5. Print the result

## Running

Each example is a standalone `main.go`. Run any example with:

```bash
go run ./examples/<name>
```

Or run all examples:

```bash
go run ./examples/...
```

## Examples

### walk — Filter Walker Pattern (porting contract)

```
go run ./examples/walk
```

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

### goqu — SQL Query Builder

```
go run ./examples/goqu
```

Translates hush queries into [goqu](https://github.com/doug-martin/goqu) expressions. Demonstrates:

- Mapping hush operators to goqu column expressions (`Eq`, `Gt`, `Like`, `ILike`, etc.)
- Combining expressions with `exp.And` / `exp.Or`
- Building SELECT, ORDER BY, LIMIT, and OFFSET clauses
- Binding coerced `Condition.Values` and escaping LIKE wildcards with `hush.EscapeLike`

### gorm — GORM ORM

```
go run ./examples/gorm
```

Translates hush queries into [GORM](https://gorm.io) clause expressions. Demonstrates:

- Mapping hush operators to GORM clauses (`Eq`, `Like`, `In`, `Between`, etc.)
- Building recursive WHERE conditions with `clause.AndConditions` / `clause.OrConditions`
- Using `clause.Expr` with an `ESCAPE` clause so `hush.EscapeLike`-escaped wildcards match literally
- Rendering `$null` / `$notNull` as `IS NULL` / `IS NOT NULL`
- Applying SELECT, ORDER BY, LIMIT, and OFFSET

> Prefer the real adapter: the `hush/gorm` package ships this exact translation
> as a ready-made `db.Scopes(gorm.Scopes(schema, q))` scope, tested against
> SQLite and Postgres.

### mongo — MongoDB Driver

```
go run ./examples/mongo
```

Translates hush queries into `bson.M` filter documents for the [official MongoDB Go driver](https://www.mongodb.com/docs/drivers/go/current/). Demonstrates:

- The most natural 1:1 mapping (hush operators map directly to MongoDB operators)
- Combining filters with `$and` / `$or` / `$nor`
- Using `$regex` with `regexp.QuoteMeta` for literal string pattern matching
- Binding coerced `Condition.Values` with a `hush.Coerce` fallback
- Building projection and sort documents

## Operator Mapping Reference

| hush Operator | SQL (goqu)           | GORM Clause        | MongoDB        |
| ------------- | -------------------- | ------------------ | -------------- |
| `$eq`         | `col.Eq(val)`        | `Eq{val}`          | `field: val`   |
| `$ne`         | `col.Neq(val)`       | `Neq{val}`         | `$ne: val`     |
| `$gt`         | `col.Gt(val)`        | `Gt{val}`          | `$gt: val`     |
| `$gte`        | `col.Gte(val)`       | `Gte{val}`         | `$gte: val`    |
| `$lt`         | `col.Lt(val)`        | `Lt{val}`          | `$lt: val`     |
| `$lte`        | `col.Lte(val)`       | `Lte{val}`         | `$lte: val`    |
| `$in`         | `col.In(vals...)`    | `IN{Values: ...}`  | `$in: [...]`   |
| `$notIn`      | `col.NotIn(vals...)` | `Not(IN{...})`     | `$nin: [...]`  |
| `$between`    | `col.Between(a, b)`  | `Gte + Lte`        | `$gte + $lte`  |
| `$contains`   | `col.Like(%v%)`      | `LIKE ... ESCAPE`  | `$regex: v`    |
| `$containsi`  | `col.ILike(%v%)`     | `LOWER LIKE ...`   | `$regex, $i`   |
| `$startsWith` | `col.Like(v%)`       | `LIKE ... ESCAPE`  | `$regex: ^v`   |
| `$endsWith`   | `col.Like(%v)`       | `LIKE ... ESCAPE`  | `$regex: v$`   |
| `$null`       | `col.IsNull()`       | `IS NULL`          | `field: nil`   |
| `$notNull`    | `col.IsNotNull()`    | `IS NOT NULL`      | `$ne: nil`     |

## Adding Your Own Integration

The `walk/main.go` example is the official porting contract. To integrate hush
with a new database:

1. Copy the `walkFilter` / `walkCondition` type-switch skeleton
2. Replace the string output with your target query builder calls
3. Read condition values through the `valueAt` pattern so you bind coerced
   `Condition.Values` (falling back to `hush.Coerce`)
4. Escape user input for pattern operators with `hush.EscapeLike`
5. Handle `Sort`, `Fields`, and `Pagination` separately (they're straightforward field/value mappings)
