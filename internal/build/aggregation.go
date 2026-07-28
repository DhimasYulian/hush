package build

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
)

var validAggFuncs = map[string]bool{
	"count": true,
	"sum":   true,
	"avg":   true,
}

// BuildAggregations builds a list of Aggregation values from params.
//
// Expected syntax:
//
//	aggregations[alias][func]=count|sum|avg
//	aggregations[alias][field]=fieldName
func BuildAggregations(params []query.Param) ([]query.Aggregation, error) {
	if len(params) == 0 {
		return nil, nil
	}

	type aggBuilder struct {
		alias string
		func_ string
		field string
	}

	byAlias := make(map[string]*aggBuilder)
	var order []string

	for _, p := range params {
		if len(p.Path) != 3 || p.Path[0] != "aggregations" {
			return nil, query.PathError(query.ErrInvalidAggregation, p.Path, "invalid aggregation path")
		}

		alias := p.Path[1]
		if alias == "" {
			return nil, query.PathError(query.ErrInvalidAggregation, p.Path, "empty aggregation alias")
		}

		key := p.Path[2]

		ab, ok := byAlias[alias]
		if !ok {
			ab = &aggBuilder{alias: alias}
			byAlias[alias] = ab
			order = append(order, alias)
		}

		switch key {
		case "func":
			if ab.func_ != "" {
				return nil, query.PathError(query.ErrInvalidAggregation, p.Path, fmt.Sprintf("duplicate func for alias %q", alias))
			}
			if !validAggFuncs[p.Value] {
				return nil, query.PathError(query.ErrInvalidAggregation, p.Path, fmt.Sprintf("invalid aggregation func %q", p.Value))
			}
			ab.func_ = p.Value

		case "field":
			if ab.field != "" {
				return nil, query.PathError(query.ErrInvalidAggregation, p.Path, fmt.Sprintf("duplicate field for alias %q", alias))
			}
			ab.field = p.Value

		default:
			return nil, query.PathError(query.ErrInvalidAggregation, p.Path, fmt.Sprintf("unknown aggregation key %q", key))
		}
	}

	result := make([]query.Aggregation, 0, len(order))

	for _, alias := range order {
		ab := byAlias[alias]

		if ab.func_ == "" {
			return nil, query.QueryError(query.ErrInvalidAggregation, fmt.Sprintf("missing func for aggregation %q", alias))
		}

		if ab.func_ == "count" && ab.field == "" {
			ab.field = "*"
		}

		if ab.func_ != "count" && ab.field == "" {
			return nil, query.QueryError(query.ErrInvalidAggregation, fmt.Sprintf("missing field for aggregation %q", alias))
		}

		result = append(result, query.Aggregation{
			Alias: alias,
			Func:  ab.func_,
			Field: ab.field,
		})
	}

	return result, nil
}
