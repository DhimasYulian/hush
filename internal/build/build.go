package build

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// BuildQuery orchestrates building all query sections from parsed params.
// It dispatches to section-specific builders for filters, fields, sort,
// groupBy, pagination, and populate, then assembles the results into a Query.
func BuildQuery(params []query.Param) (*query.Query, error) {
	filters, err := buildFiltersFromParams(params)
	if err != nil {
		return nil, err
	}

	fields, err := BuildFields(paramsFor(params, "fields"))
	if err != nil {
		return nil, err
	}

	sorts, err := BuildSort(paramsFor(params, "sort"))
	if err != nil {
		return nil, err
	}

	groupBy, err := BuildGroupBy(paramsFor(params, "groupBy"))
	if err != nil {
		return nil, err
	}

	aggregations, err := BuildAggregations(paramsFor(params, "aggregations"))
	if err != nil {
		return nil, err
	}

	pagination, err := BuildPagination(paramsFor(params, "pagination"))
	if err != nil {
		return nil, err
	}

	populates, populateAll, err := BuildPopulate(paramsFor(params, "populate"))
	if err != nil {
		return nil, err
	}

	return &query.Query{
		Filters:      filters,
		Fields:       fields,
		Sort:         sorts,
		GroupBy:      groupBy,
		Aggregations: aggregations,
		Pagination:   pagination,
		Populates:    populates,
		PopulateAll:  populateAll,
	}, nil
}

func paramsFor(params []query.Param, root string) []query.Param {
	var result []query.Param

	for _, p := range params {
		if len(p.Path) > 0 && p.Path[0] == root {
			result = append(result, p)
		}
	}

	return result
}

func buildFiltersFromParams(params []query.Param) (query.Filter, error) {
	t := NewTree()

	for _, p := range params {
		if len(p.Path) == 0 || p.Path[0] != "filters" {
			continue
		}
		t.Insert(p.Path, p.Value)
	}

	return BuildFiltersFromTree(t)
}
