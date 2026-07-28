package hush

import "github.com/DhimasYulian/hush/internal/query"

// Comparison operators.
const (
	OpEq = query.OpEq
	OpNe = query.OpNe

	OpGt  = query.OpGt
	OpGte = query.OpGte

	OpLt  = query.OpLt
	OpLte = query.OpLte

	OpIn      = query.OpIn
	OpNotIn   = query.OpNotIn
	OpBetween = query.OpBetween

	OpContains   = query.OpContains
	OpContainsi  = query.OpContainsi
	OpStartsWith = query.OpStartsWith
	OpEndsWith   = query.OpEndsWith

	OpNull    = query.OpNull
	OpNotNull = query.OpNotNull
)

// Sort directions.
const (
	SortAsc  = query.SortAsc
	SortDesc = query.SortDesc
)

// Query is the root result of parsing and validating query parameters.
type Query = query.Query

// Filter is the interface implemented by Condition, And, Or, and Not.
type Filter = query.Filter

// Condition is a leaf filter: a field path + operator + value(s).
type Condition = query.Condition

// And combines multiple filters with logical AND.
type And = query.And

// Or combines multiple filters with logical OR.
type Or = query.Or

// Not negates a single filter.
type Not = query.Not

// Sort specifies a field path and direction for ordering.
type Sort = query.Sort

// Pagination holds optional start/limit values.
type Pagination = query.Pagination

// Populate specifies a relation to include, with optional nested options.
type Populate = query.Populate

// Aggregation specifies a computed aggregate function with an output alias.
type Aggregation = query.Aggregation

// Operator is a string-typed filter comparison operator.
type Operator = query.Operator

// SortDirection is a string-typed sort direction ("asc" or "desc").
type SortDirection = query.SortDirection

// Path is a slice of strings representing a dotted/bracket field path.
type Path = query.Path

// Value is a slice of strings representing one or more filter values.
type Value = query.Value

// Field is a string-typed field name.
type Field = query.Field
