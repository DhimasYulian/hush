# validate

The `validate` package handles phase 3 of the hush pipeline: checking a built `*query.Query` against a `*schema.Schema` to ensure all fields, operators, paths, and nesting depths are allowed.

## Overview

The validate phase is the final gatekeeper in the hush pipeline. It ensures that the query produced by the build phase conforms to the schema declared by the API. Validation errors accumulate across all sections (filters, fields, sort, populate, pagination) and are joined via `errors.Join`.

## Entry Points

### Validate

```go
func Validate(q *query.Query, root *schema.Schema) error
```

Validates all query sections against the schema and returns accumulated errors. Returns nil if the query is valid or if the query is nil.

**Parameters:**
- `q` -- the query to validate (may be nil)
- `root` -- the schema to validate against (must not be nil)

**Returns:**
- `error` -- nil if valid, or joined errors from all invalid sections

**Behavior:**
1. Returns `ErrMissingSchema` if `root` is nil
2. Returns nil if `q` is nil (empty query is valid)
3. Validates filters, sort, fields, groupBy, aggregations, populate, and pagination
4. Accumulates all errors across sections
5. Returns `errors.Join(errs...)` if any errors found

## Validation Rules

### Filters

Filters are validated recursively:

- **Field check** -- Each field in a filter path must be declared `Filterable` in the schema
- **Operator check** -- The operator must be allowed for that field (per `FieldDef.Operators`)
- **Value type check** -- The value type must match the field type (e.g. numeric operators on number fields)
- **Relation resolution** -- Relation paths are resolved via `Schema.Relations`
- **Depth limit** -- Nesting depth is capped by `RelationDef.MaxDepth`

**Example:**
```
filters[title][$eq]=go
```
- Field `title` must be declared `Filterable`
- Operator `$eq` must be in `title`'s allowed operators
- Value `"go"` must be valid for `TypeString`

### Fields

Each field must be declared `Selectable` in the schema:

```
fields[0]=title
fields[1]=body
```
- `title` must be in `Schema.Selectable`
- `body` must be in `Schema.Selectable`

### GroupBy

Each field must be declared `Groupable` in the schema:

```
groupBy[0]=status
groupBy[1]=category
```
- `status` must be in `Schema.Groupable`
- `category` must be in `Schema.Groupable`

### Aggregations

Each aggregation's func must be `count`, `sum`, or `avg`. The field must be declared `Aggregatable` in the schema (except `count` with wildcard `*`):

```
aggregations[total][func]=count          — valid (wildcard field)
aggregations[totalSalary][func]=sum
aggregations[totalSalary][field]=salary  — salary must be in Schema.Aggregatable
```

### Sort

Each sort path must be declared `Sortable` in the schema:

```
sort[0]=createdAt:desc
```
- `createdAt` must be in `Schema.Sortable`

### Populate

Each relation must be declared in `Schema.Relations`, and nesting depth is capped by `RelationDef.MaxDepth`:

```
populate[author][fields][0]=name
```
- `author` must be in `Schema.Relations`
- Nesting depth must not exceed `RelationDef.MaxDepth`

When `PopulateAll` is true (wildcard `populate=*`), relation validation is skipped since the build phase already ensures wildcard is used alone.

### Pagination

The limit must not exceed `Schema.MaxLimit`:

```
pagination[limit]=25
```
- `25` must be <= `Schema.MaxLimit` (default 100, absolute max 1000)

## Validation Strategy

The validate phase uses an **accumulating error** strategy rather than fail-fast:

1. Each section (filters, fields, groupBy, aggregations, sort, populate, pagination) is validated independently
2. Errors from each section are collected
3. All errors are returned together via `errors.Join`

This means a single request can report multiple validation errors at once, making it easier for API consumers to fix all issues in one round trip.

## Error Sentinels

| Sentinel | Description |
|----------|-------------|
| `ErrInvalidPath` | Filter or sort path is invalid |
| `ErrUnknownField` | Field is not declared in the schema |
| `ErrOperatorNotAllowed` | Operator is not allowed on a field |
| `ErrNestingTooDeep` | Relation nesting exceeds the allowed depth |
| `ErrInvalidValue` | Filter value does not match the field type |
| `ErrUnknownFilterNode` | Filter node type is not recognized |
| `ErrMissingSchema` | The root schema is nil |
| `ErrUnknownGroupBy` | GroupBy field is not declared in the schema |

All validation errors are wrapped in a `query.Error` with the appropriate kind, path, field, operator, and a human-readable message.

## Error Context

Validation errors carry structured context via `query.Error`:

```go
type Error struct {
    Kind     error       // sentinel error (e.g. ErrUnknownField)
    Path     []string    // path that caused the error
    Field    string      // field name (if applicable)
    Operator Operator    // operator (if applicable)
    Message  string      // human-readable message
}
```

This allows API consumers to inspect errors programmatically:

```go
query, err := hush.Parse(values, schema)
if err != nil {
    var qerr *hush.Error
    if errors.As(err, &qerr) {
        switch qerr.Kind {
        case hush.ErrUnknownField:
            // handle unknown field
        case hush.ErrOperatorNotAllowed:
            // handle operator not allowed
        case hush.ErrNestingTooDeep:
            // handle nesting too deep
        }
    }
}
```

## Path Resolution

Filter and sort paths are resolved through the schema's relation graph:

1. Start at the root schema
2. For each path segment, check if it's a direct field or a relation
3. If it's a relation, follow the relation to the target schema
4. Continue resolving until you reach the leaf field
5. Check that the leaf field is declared with the appropriate capability (filterable, sortable, selectable)

The depth of relation nesting is capped by `RelationDef.MaxDepth`. For example, if `author` has `MaxDepth: 1`, you can filter on `author[name]` but not `author[profile][bio]`.

## Usage

The validate phase is the third and final stage of the hush pipeline. It is called internally by `hush.Parse`:

```go
// Internal pipeline:
params, err := parse.ParseParams(values)  // phase 1: parse
query, err := build.BuildQuery(params)     // phase 2: build
err := validate.Validate(query, root)      // phase 3: validate
```

Users typically don't call validate functions directly -- they use `hush.Parse` which orchestrates all three phases.
