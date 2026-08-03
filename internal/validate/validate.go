package validate

import (
	"errors"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// Validate checks all query sections against the schema and returns accumulated
// errors via errors.Join. Returns nil if the query is valid or nil.
//
// Validation also enriches the query in place: every leaf [query.Condition]
// (including those inside populate subtrees) gains its schema-declared
// FieldType and type-coerced Values.
func Validate(q *query.Query, root *schema.Schema) error {
	if root == nil {
		return ErrMissingSchema
	}

	if q == nil {
		return nil
	}

	var errs []error

	if err := validateFilterSection(q, root); err != nil {
		errs = append(errs, err)
	}

	if err := ValidateSort(q.Sort, root); err != nil {
		errs = append(errs, err)
	}

	if err := ValidateFields(q.Fields, root); err != nil {
		errs = append(errs, err)
	}

	if err := ValidateGroupBy(q.GroupBy, root); err != nil {
		errs = append(errs, err)
	}

	if err := ValidateAggregations(q.Aggregations, root); err != nil {
		errs = append(errs, err)
	}

	if err := validatePopulateSection(q, root); err != nil {
		errs = append(errs, err)
	}

	if err := ValidatePagination(q.Pagination, root); err != nil {
		errs = append(errs, err)
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.Join(errs...)
}

// validateFilterSection validates and enriches the root filter tree, assigning
// the enriched tree back into the query on success.
func validateFilterSection(q *query.Query, root *schema.Schema) error {
	filters, err := ValidateFilter(q.Filters, root)
	if err != nil {
		return err
	}
	q.Filters = filters
	return nil
}

// validatePopulateSection validates and enriches populate subtrees, skipping
// relation validation when populate=* (wildcard) is used, since the build phase
// already ensures it's used alone. The enriched tree is assigned back into the
// query on success.
func validatePopulateSection(q *query.Query, root *schema.Schema) error {
	if q.PopulateAll {
		return nil
	}

	populates, err := ValidatePopulate(q.Populates, root)
	if err != nil {
		return err
	}
	q.Populates = populates
	return nil
}
