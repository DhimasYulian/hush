package build

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
)

// operators is an alias for the canonical operator string map in query.
var operators = query.OperatorsByString

// isOperator reports whether segment is a recognized filter comparison operator.
func isOperator(segment string) bool {
	_, ok := operators[segment]
	return ok
}

// isLogical reports whether segment is a logical filter combinator ($and, $or, $not).
func isLogical(segment string) bool {
	switch segment {
	case "$and", "$or", "$not":
		return true
	default:
		return false
	}
}

// validateArity checks that multi-value operators receive the correct number of values.
func validateArity(op query.Operator, values query.Value) error {
	switch op {
	case query.OpBetween:
		if len(values) != 2 {
			return query.QueryError(ErrInvalidFilters, fmt.Sprintf("$between requires exactly 2 values, got %d", len(values)))
		}
	case query.OpIn, query.OpNotIn:
		if len(values) == 0 {
			return query.QueryError(ErrInvalidFilters, fmt.Sprintf("%s requires at least 1 value", op))
		}
	}

	return nil
}
