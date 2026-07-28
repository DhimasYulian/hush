# query

The `query` package defines the core query model types for hush. This package contains the structured representation of a parsed query: the `Query` tree (filters, fields, sort, pagination, populate), operator definitions, and the structured `Error` type used across all phases.

## Overview

The query package is the central domain of hush. All other packages (parse, build, validate) depend on these types to represent and validate query parameters.

The main entry point is the `Query` struct, which is the final output of the parse → build → validate pipeline.

## Types

### Query

The root result of parsing and validating query parameters.

```go
type Query struct {
    Filters     Filter
    Fields      []Field
    Sort        []Sort
    GroupBy     []Field
    Aggregations []Aggregation
    Pagination  Pagination
    Populates   []Populate
    PopulateAll bool
}
```

| Field | Description |
|-------|-------------|
| `Filters` | Filter tree (can be `nil` if no filters specified) |
| `Fields` | Fields to select (can be `nil` if all fields selected) |
| `Sort` | Sort specifications (can be `nil` if no sorting) |
| `GroupBy` | Group-by field names (can be `nil` if no grouping) |
| `Aggregations` | Aggregation specifications (can be `nil` if no aggregations) |
| `Pagination` | Pagination parameters (Start, Limit, and WithCount are optional) |
| `Populates` | Relations to include (can be `nil`) |
| `PopulateAll` | Whether wildcard populate (`*`) was specified |

### Filter

Interface implemented by all filter nodes. The concrete types are `Condition`, `And`, `Or`, and `Not`.

```go
type Filter interface {
    isFilter()
}
```

**Implementations:**

- `Condition` — leaf filter: field path + operator + value(s)
- `And` — combines multiple filters with logical AND
- `Or` — combines multiple filters with logical OR
- `Not` — negates a single filter

### Condition

A leaf filter: a field path + operator + value(s).

```go
type Condition struct {
    Path     Path
    Operator Operator
    Value    Value
}
```

| Field | Description |
|-------|-------------|
| `Path` | Field path segments (e.g. `["title"]` or `["author", "name"]`) |
| `Operator` | Comparison operator (e.g. `$eq`, `$containsi`) |
| `Value` | One or more string values (e.g. `["go"]` or `["a", "b"]` for `$in`) |

### And

Combines multiple filters with logical AND.

```go
type And struct {
    Filters []Filter
}
```

### Or

Combines multiple filters with logical OR.

```go
type Or struct {
    Filters []Filter
}
```

### Not

Negates a single filter.

```go
type Not struct {
    Filter Filter
}
```

### Sort

Specifies a field path and direction for ordering.

```go
type Sort struct {
    Path      Path
    Direction SortDirection
}
```

| Field | Description |
|-------|-------------|
| `Path` | Field path segments (e.g. `["createdAt"]`) |
| `Direction` | Sort direction (`SortAsc` or `SortDesc`) |

### Pagination

Holds optional start/limit values and a withCount flag.

```go
type Pagination struct {
	Start     *int
	Limit     *int
	WithCount *bool
}
```

| Field | Description |
|-------|-------------|
| `Start` | Offset into the result set (optional) |
| `Limit` | Maximum number of results (optional). If nil, no limit is applied |
| `WithCount` | Whether to include the total row count in the response. Defaults to `true` when omitted |

Both `Start` and `Limit` are optional. `WithCount` defaults to `true` when the `pagination[withCount]` parameter is not provided.

### Populate

Specifies a relation to include, with optional nested options.

```go
type Populate struct {
    Relation  string
    Fields    []Field
    Filters   Filter
    Sorts     []Sort
    Populates []Populate
}
```

| Field | Description |
|-------|-------------|
| `Relation` | Relation name (e.g. `"author"`) |
| `Fields` | Fields to select on the related resource |
| `Filters` | Filters to apply on the related resource |
| `Sorts` | Sort specifications for the related resource |
| `Populates` | Nested relations to include on the related resource |

### Aggregation

Specifies a computed aggregate function with an output alias.

```go
type Aggregation struct {
    Alias string
    Func  string
    Field string
}
```

| Field | Description |
|-------|-------------|
| `Alias` | Output key name (e.g. `"total"`, `"avgAge"`) |
| `Func` | Aggregate function: `"count"`, `"sum"`, or `"avg"` |
| `Field` | Field to aggregate on. `"*"` for count (default when omitted), specific field name for sum/avg |

### Operator

A string-typed filter comparison operator.

```go
type Operator string
```

### SortDirection

A string-typed sort direction ("asc" or "desc").

```go
type SortDirection string
```

### Path

A slice of strings representing a dotted/bracket field path.

```go
type Path = []string
```

Examples:
- `["title"]` — simple field
- `["author", "name"]` — relation field
- `["filters", "title", "$eq"]` — filter path

### Value

A slice of strings representing one or more filter values.

```go
type Value = []string
```

Single-value operators (e.g. `$eq`) have one element. Multi-value operators (e.g. `$in`, `$between`) have two or more.

### Field

A string-typed field name.

```go
type Field = string
```

### Param

Holds a parsed parameter: path segments and a string value. This is the intermediate representation between parse and build phases.

```go
type Param struct {
    Path  []string
    Value string
}
```

## Operator Constants

### Comparison Operators

| Constant | Value | Description |
|----------|-------|-------------|
| `OpEq` | `$eq` | Equal |
| `OpNe` | `$ne` | Not equal |
| `OpGt` | `$gt` | Greater than |
| `OpGte` | `$gte` | Greater than or equal |
| `OpLt` | `$lt` | Less than |
| `OpLte` | `$lte` | Less than or equal |
| `OpIn` | `$in` | In list (multi-value) |
| `OpNotIn` | `$notIn` | Not in list (multi-value) |
| `OpBetween` | `$between` | Between two values (multi-value) |
| `OpContains` | `$contains` | String contains |
| `OpContainsi` | `$containsi` | String contains (case-insensitive) |
| `OpStartsWith` | `$startsWith` | String starts with |
| `OpEndsWith` | `$endsWith` | String ends with |
| `OpNull` | `$null` | Is null |
| `OpNotNull` | `$notNull` | Is not null |

### Sort Direction Constants

| Constant | Value |
|----------|-------|
| `SortAsc` | `"asc"` |
| `SortDesc` | `"desc"` |

## Lookup Maps

| Map | Description |
|-----|-------------|
| `OperatorsByString` | Maps operator string representations to typed constants (e.g. `"$eq"` → `OpEq`) |
| `AllOperators` | Set of all recognized operators for validation |

## Error Types

### Error

A structured validation error with context about what failed.

```go
type Error struct {
    Kind     error
    Path     Path
    Field    string
    Operator Operator
    Message  string
}
```

| Field | Description |
|-------|-------------|
| `Kind` | The sentinel error (e.g. `ErrInvalidPath`, `ErrUnknownField`) |
| `Path` | The path that caused the error (if applicable) |
| `Field` | The field name (if applicable) |
| `Operator` | The operator (if applicable) |
| `Message` | Human-readable error message |

**Methods:**

- `Error() string` — returns the error message, prefixed with the kind
- `Unwrap() error` — returns the underlying kind error (supports `errors.Is` and `errors.As`)

### Error Constructors

| Constructor | Description |
|-------------|-------------|
| `QueryError(kind error, message string) error` | Creates an Error with a kind and message |
| `PathError(kind error, path Path, message string) error` | Creates an Error with a kind, path, and message |
| `FieldError(kind error, field string, message string) error` | Creates an Error with a kind, field name, and message |
| `OperatorError(kind error, field string, op Operator, message string) error` | Creates an Error with a kind, field, operator, and message |

### Cross-Domain Sentinels

These sentinels are shared by build and validate phases:

| Sentinel | Description |
|----------|-------------|
| `ErrInvalidPopulate` | Populate parameter is malformed |
| `ErrInvalidPagination` | Pagination parameter is malformed |

## Filter Tree Example

A query like:

```
filters[$and][0][title][$eq]=go
filters[$and][1][views][$gt]=100
filters[$or][0][author][$eq]=alice
filters[$or][1][tags][$contains]=golang
```

Produces:

```go
query.And{
    Filters: []query.Filter{
        query.Condition{Path: ["title"], Operator: "$eq", Value: ["go"]},
        query.Condition{Path: ["views"], Operator: "$gt", Value: ["100"]},
        query.Or{
            Filters: []query.Filter{
                query.Condition{Path: ["author"], Operator: "$eq", Value: ["alice"]},
                query.Condition{Path: ["tags"], Operator: "$contains", Value: ["golang"]},
            },
        },
    },
}
```

## Usage

The query package is primarily used internally by hush's pipeline:

1. **Parse phase** produces `Param` values
2. **Build phase** assembles `Param` values into a `Query` tree
3. **Validate phase** checks the `Query` against a `Schema`

Users interact with query types through the root package's re-exports:

```go
query, err := hush.Parse(values, schema)
if err != nil {
    var qerr *hush.Error
    if errors.As(err, &qerr) {
        fmt.Println("kind:", qerr.Kind)
        fmt.Println("field:", qerr.Field)
        fmt.Println("operator:", qerr.Operator)
    }
}
```
