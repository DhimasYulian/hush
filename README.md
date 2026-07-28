# hush

[![Go Reference](https://pkg.go.dev/badge/github.com/DhimasYulian/hush.svg)](https://pkg.go.dev/github.com/DhimasYulian/hush)
[![Go Report Card](https://goreportcard.com/badge/github.com/DhimasYulian/hush)](https://goreportcard.com/report/github.com/DhimasYulian/hush)
[![GitHub release](https://img.shields.io/github/v/release/DhimasYulian/hush)](https://github.com/DhimasYulian/hush/releases)

A Go library for parsing and validating structured query strings from URL parameters.

hush turns URL query strings like `filters[title][$containsi]=go&sort[0]=createdAt:desc&fields[0]=title` into a typed, validated `Query` object you can use to drive database queries, API filters, or search logic.

## Why?

Building an API means dealing with query parameters like filters, sorting, field selection, pagination. Doing it ad-hoc leads to:

- **Repetitive boilerplate** every endpoint reimplements the same parsing and validation
- **Inconsistent errors** each endpoint reports validation failures differently
- **Security gaps** missing validation leaves your API open to malicious queries
- **Maintenance burden** schema changes require hunting down scattered parsing code

## Install

```bash
go get github.com/DhimasYulian/hush
```

## Quick Start

```go
package main

import (
    "fmt"
    "net/url"

    "github.com/DhimasYulian/hush"
)

func main() {
    // 1. Define a schema (what fields/operators are allowed)
    schema, err := hush.NewSchema("article").
        Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
        Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpLt).
        Sortable("title", "createdAt").
        Selectable("title", "body", "createdAt").
        MaxLimit(100).
        Build()
    if err != nil {
        panic(err)
    }

    // 2. Parse query string values into a validated Query
    values := url.Values{
        "filters[title][$containsi]": {"go"},
        "sort[0]":                    {"createdAt:desc"},
        "fields[0]":                  {"title"},
        "pagination[limit]":          {"25"},
    }

    query, err := hush.Parse(values, schema)
    if err != nil {
        fmt.Println("validation error:", err)
        return
    }

    fmt.Printf("%+v\n", query)
}
```

## How It Works

hush processes query strings through a three-stage pipeline:

```
url.Values  -->  Parse  -->  Build  -->  Validate  -->  *Query
```

1. **Parse** -- Splits URL parameter keys into structured path segments using
   **LHS (left-hand side) bracket notation** — the path structure is encoded in
   the key itself, not the value.
   `filters[title][$containsi]` becomes `["filters", "title", "$containsi"]`.

2. **Build** -- Assembles parsed segments into a typed query tree.
   Produces `Query` with `Filters`, `Fields`, `Sort`, `Pagination`, and `Populates`.

3. **Validate** -- Checks the query against a `Schema` to ensure only declared
   fields, operators, and paths are used. Returns structured errors with context.

## Main Types

| Type          | Description                                                                                                                                           |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Schema`      | Defines which fields are filterable, sortable, selectable, groupable, and aggregatable, and which relations are available. Built via `SchemaBuilder`. |
| `Query`       | The parsed and validated result. Contains `Filters`, `Fields`, `Sort`, `GroupBy`, `Aggregations`, `Pagination`, and `Populates`.                      |
| `Filter`      | Interface for filter nodes: `Condition`, `And`, `Or`, `Not`.                                                                                          |
| `Condition`   | A leaf filter: a field path + operator + value(s).                                                                                                    |
| `Sort`        | A field path and direction for ordering.                                                                                                              |
| `Aggregation` | A computed aggregate function (`count`, `sum`, `avg`) with an alias and optional field.                                                               |
| `Pagination`  | Optional `Start`, `Limit`, and `WithCount` values.                                                                                                    |
| `Populate`    | A relation to include, optionally with its own fields, sorts, filters, and nested populates.                                                          |
| `Error`       | Structured validation error with `Kind`, `Path`, `Field`, `Operator`, and `Message`.                                                                  |

## Supported Operators

| Operator      | Constant       | Description                        |
| ------------- | -------------- | ---------------------------------- |
| `$eq`         | `OpEq`         | Equal                              |
| `$ne`         | `OpNe`         | Not equal                          |
| `$gt`         | `OpGt`         | Greater than                       |
| `$gte`        | `OpGte`        | Greater than or equal              |
| `$lt`         | `OpLt`         | Less than                          |
| `$lte`        | `OpLte`        | Less than or equal                 |
| `$in`         | `OpIn`         | In list                            |
| `$notIn`      | `OpNotIn`      | Not in list                        |
| `$between`    | `OpBetween`    | Between two values                 |
| `$contains`   | `OpContains`   | String contains                    |
| `$containsi`  | `OpContainsi`  | String contains (case-insensitive) |
| `$startsWith` | `OpStartsWith` | String starts with                 |
| `$endsWith`   | `OpEndsWith`   | String ends with                   |
| `$null`       | `OpNull`       | Is null                            |
| `$notNull`    | `OpNotNull`    | Is not null                        |

## Query String Syntax

### Filters

```
filters[field][$operator]=value
filters[relation][field][$operator]=value
filters[$and][0][field][$operator]=value
filters[$or][0][field][$operator]=value
filters[$not][field][$operator]=value
```

### Fields (select specific columns)

```
fields[0]=title
fields[1]=body
```

### Sort

```
sort[0]=createdAt:desc
sort[1]=title:asc
```

Direction defaults to `asc` if omitted.

### Group By

```
groupBy[0]=status
groupBy[1]=category
```

Shorthand (single value): `groupBy=status`

Fields must be declared `Groupable` in the schema.

### Aggregations

```
aggregations[total][func]=count
aggregations[totalSalary][func]=sum
aggregations[totalSalary][field]=salary
aggregations[avgAge][func]=avg
aggregations[avgAge][field]=age
```

Supported functions: `count`, `sum`, `avg`. The `field` parameter is optional for `count` (defaults to `*`), required for `sum` and `avg`. Fields must be declared `Aggregatable` in the schema.

### Pagination

```
pagination[start]=0
pagination[limit]=25
pagination[withCount]=true
```

### Populate (include relations)

```
populate[0]=author
populate[author][fields][0]=name
populate[author][sort][0]=name:asc
populate[author][filters][name][$eq]=Alice
populate[author][populate][0]=profile
```

Wildcard (select all relations):

```
populate=*
```

## Error Handling

hush returns structured errors that you can inspect with `errors.Is` and `errors.As`:

```go
query, err := hush.Parse(values, schema)
if err != nil {
    var qerr *hush.Error
    if errors.As(err, &qerr) {
        fmt.Println("kind:", qerr.Kind)       // e.g. ErrOperatorNotAllowed
        fmt.Println("field:", qerr.Field)      // e.g. "title"
        fmt.Println("operator:", qerr.Operator) // e.g. "$containsi"
        fmt.Println("message:", qerr.Message)
    }
}
```

Parse and build errors fail on the first error. Validation errors accumulate across all sections (filters, fields, sort, populate, pagination) and are joined via `errors.Join`.

## Examples

The [`examples/`](./examples) directory contains integration examples showing how to translate hush queries into different database query formats:

| Example                     | Description                                                         |
| --------------------------- | ------------------------------------------------------------------- |
| [`walk`](./examples/walk)   | Core filter walker pattern — the starting point for any integration |
| [`goqu`](./examples/goqu)   | SQL query builder integration (PostgreSQL, MySQL, SQLite)           |
| [`gorm`](./examples/gorm)   | GORM ORM integration                                                |
| [`mongo`](./examples/mongo) | MongoDB driver integration                                          |

Run any example:

```bash
go run ./examples/walk
go run ./examples/goqu
go run ./examples/gorm
go run ./examples/mongo
```

See [`examples/README.md`](./examples/README.md) for the full operator mapping reference and details on each integration.

## License

hush is released under the [MIT License](http://www.opensource.org/licenses/MIT).
