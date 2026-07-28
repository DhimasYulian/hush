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

### walk — Filter Walker Pattern

```
go run ./examples/walk
```

**Teaches the core pattern** for integrating hush with any database layer. Demonstrates:

- Type-switching on `hush.Filter` to handle `Condition`, `And`, `Or`, `Not`
- Recursive walking with depth tracking
- Translating `hush.Condition` operators to human-readable strings
- A verbose tree printer for debugging

This is the starting point if you're integrating hush with a database or ORM not covered here. The `walkFilter` function can be adapted to generate SQL, MongoDB bson.M, or any other query format.

### goqu — SQL Query Builder

```
go run ./examples/goqu
```

Translates hush queries into [goqu](https://github.com/doug-martin/goqu) expressions. Demonstrates:

- Mapping hush operators to goqu column expressions (`Eq`, `Gt`, `Like`, `ILike`, etc.)
- Combining expressions with `exp.And` / `exp.Or`
- Building SELECT, ORDER BY, LIMIT, and OFFSET clauses
- Escaping LIKE wildcards in user input

### gorm — GORM ORM

```
go run ./examples/gorm
```

Translates hush queries into [GORM](https://gorm.io) clause expressions. Demonstrates:

- Mapping hush operators to GORM clauses (`Eq`, `Like`, `In`, `Between`, etc.)
- Building recursive WHERE conditions with `clause.AndConditions` / `clause.OrConditions`
- Using `clause.Expr` for case-insensitive LIKE (`LOWER() LIKE LOWER()`)
- Applying SELECT, ORDER BY, LIMIT, and OFFSET

### mongo — MongoDB Driver

```
go run ./examples/mongo
```

Translates hush queries into `bson.M` filter documents for the [official MongoDB Go driver](https://www.mongodb.com/docs/drivers/go/current/). Demonstrates:

- The most natural 1:1 mapping (hush operators map directly to MongoDB operators)
- Combining filters with `$and` / `$or` / `$nor`
- Using `$regex` for string pattern matching
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
| `$contains`   | `col.Like(%v%)`      | `Like{val}`        | `$regex: val`  |
| `$containsi`  | `col.ILike(%v%)`     | `LOWER LIKE LOWER` | `$regex, $i`   |
| `$startsWith` | `col.Like(v%)`       | `Like{val%}`       | `$regex: ^val` |
| `$endsWith`   | `col.Like(%v)`       | `Like{%val}`       | `$regex: val$` |
| `$null`       | `col.IsNull()`       | `Eq{nil}`          | `field: nil`   |
| `$notNull`    | `col.IsNotNull()`    | `Neq{nil}`         | `$ne: nil`     |

## Adding Your Own Integration

The `walk/main.go` example shows the universal pattern. To integrate hush with a new database:

1. Copy the `walkFilter` / `walkCondition` type-switch skeleton
2. Replace the string output with your target query builder calls
3. Handle `Sort`, `Fields`, and `Pagination` separately (they're straightforward field/value mappings)
