# schema

The `schema` package defines the domain types that describe which query operations are allowed for a resource. A schema declares which fields are filterable, sortable, and selectable, which relations are available, and what pagination limits apply.

## Overview

Schemas are the foundation of hush's validation system. They act as a contract between your API and its consumers, ensuring that only declared fields, operators, and paths are used in queries.

Schemas are constructed via the root package's `SchemaBuilder`:

```go
schema, err := hush.NewSchema("article").
    Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
    Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpLt).
    Sortable("title", "createdAt").
    Selectable("title", "body", "createdAt").
    Groupable("title", "createdAt").
    Aggregatable("views", "createdAt").
    MaxLimit(100).
    Build()
```

## Types

### Schema

The root type that defines allowed query operations for a resource.

```go
type Schema struct {
    Name         string
    Filterable   map[string]FieldDef
    Sortable     map[string]struct{}
    Selectable   map[string]struct{}
    Groupable    map[string]struct{}
    Aggregatable map[string]struct{}
    Relations    map[string]RelationDef
    MaxLimit     int
}
```

| Field | Description |
|-------|-------------|
| `Name` | Resource name (e.g. `"article"`, `"user"`) |
| `Filterable` | Fields that can be filtered, mapped to their `FieldDef` |
| `Sortable` | Fields that can be sorted on |
| `Selectable` | Fields that can be selected (field projection) |
| `Groupable` | Fields that can be used in groupBy |
| `Aggregatable` | Fields that can be used in aggregations (sum, avg) |
| `Relations` | Named relations to other schemas |
| `MaxLimit` | Maximum pagination limit (default: 100) |

**Methods:**

- `GetFilterable(name string) (FieldDef, bool)` — returns the field definition if filterable
- `GetSortable(name string) bool` — reports whether the field is sortable
- `GetSelectable(name string) bool` — reports whether the field can be selected
- `GetGroupable(name string) bool` — reports whether the field can be used in groupBy
- `GetAggregatable(name string) bool` — reports whether the field can be used in aggregations
- `GetSelectableFields() []string` — returns all selectable field names in sorted order
- `GetRelations() map[string]RelationDef` — returns a copy of all declared relations
- `GetRelation(name string) (RelationDef, bool)` — returns the named relation definition
- `GetMaxLimit() int` — returns the maximum pagination limit

### FieldDef

Describes a filterable field: its name, type, and allowed operators.

```go
type FieldDef struct {
    Name      string
    Type      FieldType
    Operators map[query.Operator]bool
}
```

| Field | Description |
|-------|-------------|
| `Name` | Field name (e.g. `"title"`, `"createdAt"`) |
| `Type` | Data type of the field (`TypeString`, `TypeNumber`, `TypeBool`, `TypeDate`) |
| `Operators` | Set of operators allowed on this field (e.g. `$eq`, `$containsi`) |

### RelationDef

Describes a named relation to another schema with a max nesting depth.

```go
type RelationDef struct {
    Name               string
    Target             *Schema
    MaxDepth           int
    HiddenFromWildcard bool
}
```

| Field | Description |
|-------|-------------|
| `Name` | Relation name (e.g. `"author"`, `"comments"`) |
| `Target` | The target schema this relation points to |
| `MaxDepth` | Maximum nesting depth for this relation (must be >= 1) |
| `HiddenFromWildcard` | If true, relation is excluded from `populate=*` |

### FieldType

Represents the data type of a filterable field.

```go
type FieldType string
```

| Constant | Value | Description |
|----------|-------|-------------|
| `TypeString` | `"string"` | String fields |
| `TypeNumber` | `"number"` | Numeric fields (int, float) |
| `TypeBool` | `"bool"` | Boolean fields |
| `TypeDate` | `"date"` | Date/time fields |

## Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `DefaultMaxLimit` | `100` | Default maximum pagination limit |
| `AbsoluteMaxLimit` | `1000` | Hard upper bound for pagination limits |

## Error Sentinels

| Sentinel | Description |
|----------|-------------|
| `ErrEmptyName` | Schema or field name is empty |
| `ErrNoOperators` | Filterable field has no allowed operators |
| `ErrUnknownOperator` | Operator is not recognized |
| `ErrDuplicateField` | Field is declared twice in the same schema |
| `ErrDuplicateRelation` | Relation is declared twice in the same schema |
| `ErrNilTarget` | Relation target schema is nil |
| `ErrInvalidMaxDepth` | Relation max depth is less than 1 |
| `ErrInvalidMaxLimit` | Max limit is out of range (must be 1..AbsoluteMaxLimit) |
| `ErrUnknownFieldType` | Field type is not recognized |

## Schema Construction

Schemas are built using the `SchemaBuilder` from the root package. The builder accumulates validation errors and returns them all at once when `Build()` is called:

```go
schema, err := hush.NewSchema("article").
    Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
    Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpLt).
    Sortable("title", "createdAt").
    Selectable("title", "body", "createdAt").
    Groupable("title", "createdAt").
    Aggregatable("title", "createdAt").
    Relation("author", authorSchema, 1).
    MaxLimit(100).
    Build()

if err != nil {
    // err contains all validation errors joined via errors.Join
    fmt.Println(err)
}
```

### Filterable Fields

Declare a field as filterable with its type and allowed operators:

```go
Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi)
Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpLt)
Filterable("published", hush.TypeBool, hush.OpEq)
Filterable("createdAt", hush.TypeDate, hush.OpGt, hush.OpLt, hush.OpBetween)
```

### Sortable Fields

Declare fields that can be sorted on:

```go
Sortable("title", "createdAt", "views")
```

### Selectable Fields

Declare fields that can be selected (field projection):

```go
Selectable("title", "body", "createdAt")
```

### Groupable Fields

Declare fields that can be used in groupBy:

```go
Groupable("title", "createdAt", "status")
```

### Aggregatable Fields

Declare fields that can be used in aggregations (sum, avg):

```go
Aggregatable("views", "salary", "age")
```

### Relations

Declare named relations to other schemas:

```go
Relation("author", authorSchema, 1)                    // depth 1
Relation("comments", commentSchema, 2, hush.Hidden())  // depth 2, hidden from wildcard
```

### Pagination Limits

Set the maximum allowed pagination limit:

```go
MaxLimit(50)  // default is 100, absolute max is 1000
```

## Cross-References

The `FieldDef` type references `query.Operator` from the `query` package. This is the single unidirectional dependency from schema to query types.
