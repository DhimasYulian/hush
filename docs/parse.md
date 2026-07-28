# parse

The `parse` package handles phase 1 of the hush pipeline: converting URL query string values into structured `Param` values.

## Overview

The parse phase takes raw `net/url.Values` and produces a sorted slice of `Param` values, where each param contains:
- `Path`: the bracket-notation key split into segments
- `Value`: the string value

For example, `filters[title][$eq]=go` becomes:
```go
Param{
    Path:  []string{"filters", "title", "$eq"},
    Value: "go",
}
```

Keys are sorted alphabetically to ensure deterministic ordering.

## Functions

### ParseParams

```go
func ParseParams(values url.Values) ([]query.Param, error)
```

Converts `url.Values` into a sorted slice of `Param` values.

**Parameters:**
- `values` — the raw URL query parameters

**Returns:**
- `[]query.Param` — sorted list of parsed parameters
- `error` — parse error if any key has invalid syntax

**Behavior:**
1. Collects all keys from the `url.Values`
2. Sorts keys alphabetically for deterministic ordering
3. Splits each key into path segments via `ParsePath`
4. Creates one `Param` per value (a single key can have multiple values)

**Example:**

```go
values := url.Values{
    "filters[title][$eq]": {"go"},
    "sort[0]":             {"createdAt:desc"},
    "fields[0]":           {"title"},
}

params, err := parse.ParseParams(values)
// params[0] = Param{Path: ["fields", "0"], Value: "title"}
// params[1] = Param{Path: ["filters", "title", "$eq"], Value: "go"}
// params[2] = Param{Path: ["sort", "0"], Value: "createdAt:desc"}
```

### ParsePath

```go
func ParsePath(key string) ([]string, error)
```

Splits a bracket-notation key into path segments.

**Parameters:**
- `key` — the URL parameter key (e.g. `"filters[name][$eq]"`)

**Returns:**
- `[]string` — path segments (e.g. `["filters", "name", "$eq"]`)
- `error` — parse error if syntax is invalid

**Syntax Rules:**
- Root segment is the part before the first `[` (e.g. `filters` in `filters[name]`)
- Bracket segments are enclosed in `[` and `]` (e.g. `[name]`, `[$eq]`)
- Nested brackets are not allowed (e.g. `filters[[name]]` is invalid)
- Empty segments are not allowed (e.g. `filters[]` is invalid)
- No characters are allowed between closing `]` and opening `[` (e.g. `filters[name][foo][bar]` is ok, but `filters[name]foo[bar]` is not)

**Examples:**

| Input | Output |
|-------|--------|
| `"filters[name][$eq]"` | `["filters", "name", "$eq"]` |
| `"sort[0]"` | `["sort", "0"]` |
| `"fields"` | `["fields"]` |
| `"pagination[limit]"` | `["pagination", "limit"]` |
| `"populate[author][fields][0]"` | `["populate", "author", "fields", "0"]` |

**Error Cases:**

| Input | Error |
|-------|-------|
| `""` | `ErrEmptyKey` |
| `"filters[]"` | `ErrEmptySegment` |
| `"filters[name" ` | `ErrInvalidSyntax` |
| `"filters]name]"` | `ErrInvalidSyntax` |
| `"filters[name]foo[bar]"` | `ErrUnexpectedCharacter` |

## Error Sentinels

| Sentinel | Description |
|----------|-------------|
| `ErrEmptyKey` | Query string key is empty |
| `ErrEmptySegment` | Path segment is empty (e.g. `filters[]`) |
| `ErrInvalidSyntax` | Bracket path has invalid syntax (e.g. unmatched brackets) |
| `ErrUnexpectedCharacter` | Unexpected character in a path (e.g. text between brackets) |

All parse errors are wrapped in a `query.Error` with the appropriate kind and a human-readable message.

## Usage

The parse phase is the first stage of the hush pipeline. It is called internally by `hush.Parse`:

```go
// Internal pipeline:
params, err := parse.ParseParams(values)  // phase 1: parse
query, err := build.BuildQuery(params)     // phase 2: build
err := validate.Validate(query, root)      // phase 3: validate
```

Users typically don't call parse functions directly — they use `hush.Parse` which orchestrates all three phases.

## Bracket Notation Reference

hush uses bracket notation for URL query parameters, similar to PHP/Strapi conventions:

```
root[key1][key2][key3]=value
```

This is equivalent to nested object access:
```json
{
  "root": {
    "key1": {
      "key2": {
        "key3": "value"
      }
    }
  }
}
```

### Common Patterns

| Pattern | Use Case |
|---------|----------|
| `filters[field][$operator]=value` | Filter by field with operator |
| `filters[relation][field][$operator]=value` | Filter on related resource |
| `filters[$and][0][field][$operator]=value` | AND logic |
| `filters[$or][0][field][$operator]=value` | OR logic |
| `fields[0]=fieldName` | Select specific fields |
| `sort[0]=fieldName:asc` | Sort by field |
| `pagination[limit]=25` | Pagination limit |
| `pagination[start]=0` | Pagination offset |
| `pagination[withCount]=true` | Include total row count (default: true) |
| `groupBy[0]=fieldName` | Group by field |
| `aggregations[name][func]=count` | Aggregate function (count, sum, avg) |
| `aggregations[name][field]=fieldName` | Field to aggregate on |
| `populate[0]=relation` | Include relation |
| `populate[relation][fields][0]=name` | Nested field selection |
