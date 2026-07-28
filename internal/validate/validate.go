package validate

import (
	"errors"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// Validate checks all query sections against the schema and returns accumulated
// errors via errors.Join. Returns nil if the query is valid or nil.
func Validate(q *query.Query, root *schema.Schema) error {
	if root == nil {
		return ErrMissingSchema
	}

	if q == nil {
		return nil
	}

	var errs []error

	if err := ValidateFilter(q.Filters, root); err != nil {
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

// validatePopulateSection skips relation validation when populate=* (wildcard)
// is used, since the build phase already ensures it's used alone.
func validatePopulateSection(q *query.Query, root *schema.Schema) error {
	if q.PopulateAll {
		return nil
	}

	return ValidatePopulate(q.Populates, root)
}
