package validate

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// ValidatePopulate validates a list of populate directives, checking that each
// relation exists, nesting depth is within limits, and nested fields/sorts/filters
// are valid against the target schema.
func ValidatePopulate(populates []query.Populate, root *schema.Schema) error {
	return validatePopulateAt(populates, root, 0)
}

func validatePopulateAt(populates []query.Populate, root *schema.Schema, depth int) error {
	for _, pop := range populates {
		if err := validatePopulateEntry(pop, root, depth); err != nil {
			return err
		}
	}
	return nil
}

// validatePopulateEntry validates a single populate entry: relation exists,
// depth is within limits, and nested options are valid.
func validatePopulateEntry(pop query.Populate, root *schema.Schema, depth int) error {
	rel, ok := root.GetRelation(pop.Relation)
	if !ok {
		return query.FieldError(query.ErrInvalidPopulate, pop.Relation, "unknown relation")
	}

	if depth >= rel.MaxDepth {
		return query.FieldError(
			ErrNestingTooDeep,
			pop.Relation,
			fmt.Sprintf("relation %q at depth %d exceeds max depth %d", pop.Relation, depth, rel.MaxDepth),
		)
	}

	if err := ValidateFields(pop.Fields, rel.Target); err != nil {
		return err
	}

	if err := ValidateSort(pop.Sorts, rel.Target); err != nil {
		return err
	}

	if pop.Filters != nil {
		if err := ValidateFilter(pop.Filters, rel.Target); err != nil {
			return err
		}
	}

	return validatePopulateAt(pop.Populates, rel.Target, depth+1)
}
