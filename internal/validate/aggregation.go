package validate

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// ValidateAggregations checks that each aggregation has a valid func and that
// non-count fields are declared aggregatable in the root schema.
func ValidateAggregations(aggs []query.Aggregation, root *schema.Schema) error {
	for _, a := range aggs {
		switch a.Func {
		case "count":
			if a.Field != "*" && !root.GetAggregatable(a.Field) {
				return query.FieldError(query.ErrInvalidAggregation, a.Field, fmt.Sprintf("field %q is not aggregatable", a.Field))
			}

		case "sum", "avg":
			if !root.GetAggregatable(a.Field) {
				return query.FieldError(query.ErrInvalidAggregation, a.Field, fmt.Sprintf("field %q is not aggregatable", a.Field))
			}

		default:
			return query.QueryError(query.ErrInvalidAggregation, fmt.Sprintf("invalid aggregation func %q", a.Func))
		}
	}

	return nil
}
