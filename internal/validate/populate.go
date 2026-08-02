package validate

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// ValidatePopulate validates a list of populate directives, checking that each
// relation exists, nesting depth is within limits, and nested fields/sorts/filters
// are valid against the target schema. It returns the directives with any
// nested filter conditions enriched with field types and coerced values.
func ValidatePopulate(populates []query.Populate, root *schema.Schema) ([]query.Populate, error) {
	return validatePopulateAt(populates, root, 0)
}

func validatePopulateAt(populates []query.Populate, root *schema.Schema, depth int) ([]query.Populate, error) {
	if populates == nil {
		return nil, nil
	}

	out := make([]query.Populate, len(populates))
	for i, pop := range populates {
		enriched, err := validatePopulateEntry(&pop, root, depth)
		if err != nil {
			return nil, err
		}
		out[i] = enriched
	}
	return out, nil
}

// validatePopulateEntry validates a single populate entry: relation exists,
// depth is within limits, and nested options are valid.
func validatePopulateEntry(pop *query.Populate, root *schema.Schema, depth int) (query.Populate, error) {
	rel, ok := root.GetRelation(pop.Relation)
	if !ok {
		return query.Populate{}, query.FieldError(query.ErrInvalidPopulate, pop.Relation, "unknown relation")
	}

	if depth >= rel.MaxDepth {
		return query.Populate{}, query.FieldError(
			ErrNestingTooDeep,
			pop.Relation,
			fmt.Sprintf("relation %q at depth %d exceeds max depth %d", pop.Relation, depth, rel.MaxDepth),
		)
	}

	if err := ValidateFields(pop.Fields, rel.Target); err != nil {
		return query.Populate{}, err
	}

	if err := ValidateSort(pop.Sorts, rel.Target); err != nil {
		return query.Populate{}, err
	}

	var err error
	pop.Filters, err = ValidateFilter(pop.Filters, rel.Target)
	if err != nil {
		return query.Populate{}, err
	}

	pop.Populates, err = validatePopulateAt(pop.Populates, rel.Target, depth+1)
	if err != nil {
		return query.Populate{}, err
	}

	return *pop, nil
}
