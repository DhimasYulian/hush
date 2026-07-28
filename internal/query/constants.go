package query

// Comparison operators.
const (
	OpEq Operator = "$eq"
	OpNe Operator = "$ne"

	OpGt  Operator = "$gt"
	OpGte Operator = "$gte"

	OpLt  Operator = "$lt"
	OpLte Operator = "$lte"

	OpIn      Operator = "$in"
	OpNotIn   Operator = "$notIn"
	OpBetween Operator = "$between"

	OpContains   Operator = "$contains"
	OpContainsi  Operator = "$containsi"
	OpStartsWith Operator = "$startsWith"
	OpEndsWith   Operator = "$endsWith"

	OpNull    Operator = "$null"
	OpNotNull Operator = "$notNull"
)

// Sort directions.
const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// OperatorsByString maps operator string representations (e.g. "$eq") to their
// typed constants. Used by the build phase to parse filter path segments.
var OperatorsByString = map[string]Operator{
	"$eq":         OpEq,
	"$ne":         OpNe,
	"$lt":         OpLt,
	"$lte":        OpLte,
	"$gt":         OpGt,
	"$gte":        OpGte,
	"$in":         OpIn,
	"$notIn":      OpNotIn,
	"$between":    OpBetween,
	"$contains":   OpContains,
	"$containsi":  OpContainsi,
	"$startsWith": OpStartsWith,
	"$endsWith":   OpEndsWith,
	"$null":       OpNull,
	"$notNull":    OpNotNull,
}

// AllOperators is the set of all recognized operators. Used by the schema
// builder to validate that declared operators are known.
var AllOperators = map[Operator]bool{
	OpEq:         true,
	OpNe:         true,
	OpGt:         true,
	OpGte:        true,
	OpLt:         true,
	OpLte:        true,
	OpIn:         true,
	OpNotIn:      true,
	OpBetween:    true,
	OpContains:   true,
	OpContainsi:  true,
	OpStartsWith: true,
	OpEndsWith:   true,
	OpNull:       true,
	OpNotNull:    true,
}
