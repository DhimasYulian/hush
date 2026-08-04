# build

The `build` package handles phase 2 of the hush pipeline: assembling parsed `Param` values into a structured `*query.Query` tree.

## Overview

The build phase takes the flat list of `Param` values from the parse phase and assembles them into a typed query tree. It dispatches to section-specific builders for filters, fields, sort, groupBy, pagination, and populate.

## Entry Points

### BuildQuery

```go
func BuildQuery(params []query.Param) (*query.Query, error)
```

Orchestrates building all query sections from parsed params. This is the main entry point for the build phase.

**Parameters:**
- `params` -- sorted list of parsed parameters from the parse phase

**Returns:**
- `*query.Query` -- the assembled query tree
- `error` -- build error if any parameter is malformed

**Behavior:**
1. Extracts filter params and builds them into a `Filter` tree
2. Extracts field params and builds them into a `[]Field` slice
3. Extracts sort params and builds them into a `[]Sort` slice
4. Extracts groupBy params and builds them into a `[]Field` slice
5. Extracts aggregation params and builds them into a `[]Aggregation` slice
6. Extracts pagination params and builds them into a `Pagination` struct
7. Extracts populate params and builds them into a `[]Populate` slice
8. Assembles all sections into a `Query` struct

## Filter Building

Filters are the most complex section to build. The process involves:

1. **Insert into tree** -- Filter params are inserted into a path `Tree` data structure
2. **Walk recursively** -- The tree is walked to produce a `query.Filter` tree
3. **Dispatch by segment** -- Each node is dispatched based on its segment value

### Filter Dispatch

The builder recognizes these segment types:

| Segment | Action |
|---------|--------|
| `"filters"` | Root node -- wraps multiple children in implicit AND |
| `"$and"` | Logical AND combinator |
| `"$or"` | Logical OR combinator |
| `"$not"` | Logical NOT combinator |
| `$eq`, `$gt`, etc. | Comparison operator -- builds a `Condition` |
| Other | Field name -- recurses deeper |

### Filter Examples

**Simple equality:**
`filters[title][$eq]=go`
Produces: `Condition{Path: ["title"], Operator: "$eq", Value: ["go"]}`

**Nested relation:**
`filters[author][name][$eq]=alice`
Produces: `Condition{Path: ["author", "name"], Operator: "$eq", Value: ["alice"]}`

**AND logic:**
```
filters[$and][0][title][$eq]=go
filters[$and][1][views][$gt]=100
```
Produces: `And{Filters: [Condition{...}, Condition{...}]}`

**OR logic:**
```
filters[$or][0][author][$eq]=alice
filters[$or][1][author][$eq]=bob
```
Produces: `Or{Filters: [Condition{...}, Condition{...}]}`

**NOT logic:**
`filters[$not][title][$eq]=go`
Produces: `Not{Filter: Condition{...}}`

**Multiple top-level filters (implicit AND):**
```
filters[title][$eq]=go
filters[views][$gt]=100
```
Produces: `And{Filters: [Condition{...}, Condition{...}]}`

### Multi-Value Operators

Operators `$in`, `$notIn`, and `$between` use indexed syntax for multiple values:

```
filters[status][$in][0]=draft
filters[status][$in][1]=published
```
Produces: `Condition{Path: ["status"], Operator: "$in", Value: ["draft", "published"]}`

## Indexed Building

Fields, sort, and shorthand populate use `BuildIndexed`, a generic helper that handles both syntax modes:

### Shorthand Syntax

Single value without index:
`fields=title`
Produces: `["title"]`

### Indexed Syntax

Numeric key array:
```
fields[0]=title
fields[1]=body
```
Produces: `["title", "body"]`

### Mixed Syntax Error

You cannot mix shorthand and indexed syntax. Attempting to do so produces an error.

## Field Building

Fields are built using `BuildIndexed` with a simple string parser. Each param value becomes a field name in the output slice.

## GroupBy Building

GroupBy fields are built using the same `BuildIndexed` mechanism as fields. Each param value becomes a field name in the output slice.

- `groupBy[0]=status` produces `["status"]`
- `groupBy[0]=status&groupBy[1]=category` produces `["status", "category"]`
- Shorthand: `groupBy=status` produces `["status"]`

## Aggregation Building

Aggregation params are keyed by alias, with sub-keys `func` and `field`:

```
aggregations[total][func]=count
aggregations[totalSalary][func]=sum
aggregations[totalSalary][field]=salary
```

**Rules:**
- `func` is required — one of `count`, `sum`, `avg`
- `field` is optional for `count` (defaults to `"*"`), required for `sum`/`avg`
- Each alias must have exactly one `func` and at most one `field`
- Duplicate aliases, unknown keys, and missing required values produce errors

## Sort Building

Sort params use the format `fieldName:direction`:

- `sort[0]=createdAt:desc` produces `Sort{Path: ["createdAt"], Direction: "desc"}`
- `sort[1]=title:asc` produces `Sort{Path: ["title"], Direction: "asc"}`
- `sort[2]=views` (no direction) defaults to `Sort{Path: ["views"], Direction: "asc"}`

Direction defaults to `"asc"` if omitted.

## Pagination Building

Pagination params are parsed into the `Pagination` struct:

- `pagination[start]=0` sets `Start` to `&0`
- `pagination[limit]=25` sets `Limit` to `&25`

Both fields are optional. If a param key is not recognized (e.g. `pagination[page]`), it is ignored.

## Populate Building

Populate supports three syntax modes:

### Wildcard

`populate=*` selects all relations. Must be used alone (no other populate params allowed).

### Indexed (Shorthand)

Simple relation names with numeric indices:
```
populate[0]=author
populate[1]=comments
```
Produces: `[{Relation: "author"}, {Relation: "comments"}]`

### Relation-Keyed

Nested field selection, sorting, filtering, and nested populates:
```
populate[author][fields][0]=name
populate[author][sort][0]=name:asc
populate[author][filters][name][$eq]=Alice
populate[author][populate][0]=profile
```

This uses `PopulateTree` to build a nested relation graph, then flattens it into a slice of `query.Populate`.

### Mixed Syntax Error

You cannot mix indexed and relation-keyed syntax in the same request.

## Internal Types

### Tree

A generic path tree used for filter building. Each node has a segment name, a value, and ordered children. Params are inserted into the tree, then the tree is walked to produce the filter hierarchy.

### Node

A node in the tree with:
- `Segment` -- the path segment name (e.g. `"filters"`, `"title"`, `"$eq"`)
- `Value` -- the leaf value (only set on condition nodes)
- `Children` -- map of child nodes
- `Order` -- insertion order of children

### PopulateTree

A specialized tree for building nested populate relations. Unlike the generic filter Tree, each node carries fields, sorts, filters, and nested populates.

### PopulateNode

A node in the populate tree with:
- `Relation` -- the relation name
- `Fields`, `Sorts`, `Filters` -- relation-level options
- `Children`, `Order` -- nested relations

## Error Sentinels

| Sentinel | Description |
|----------|-------------|
| `ErrInvalidFields` | Fields parameter is malformed |
| `ErrInvalidSort` | Sort parameter is malformed |
| `ErrInvalidFilters` | Filter parameter is malformed |
| `ErrInvalidGroupBy` | GroupBy parameter is malformed |

All build errors are wrapped in a `query.Error` with the appropriate kind and a human-readable message.

## Usage

The build phase is the second stage of the hush pipeline. It is called internally by `hush.Parse`:

```go
// Internal pipeline:
params, err := parse.ParseParams(values)  // phase 1: parse
query, err := build.BuildQuery(params)     // phase 2: build
err := validate.Validate(query, root)      // phase 3: validate
```

Users typically don't call build functions directly -- they use `hush.Parse` which orchestrates all three phases.
