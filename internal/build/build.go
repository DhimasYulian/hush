package build

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// BuildQuery orchestrates building all query sections from parsed params.
// It dispatches to section-specific builders for filters, fields, sort,
// groupBy, pagination, and populate, then assembles the results into a Query.
//
// Params are partitioned once by root section so each section builder sees
// only its own params without repeated full scans.
func BuildQuery(params []query.Param) (*query.Query, error) {
	var (
		filterParams  []query.Param
		fieldParams   []query.Param
		sortParams    []query.Param
		groupByParams []query.Param
		aggParams     []query.Param
		pagParams     []query.Param
		popParams     []query.Param
	)

	for _, p := range params {
		if len(p.Path) == 0 {
			continue
		}

		switch p.Path[0] {
		case "filters":
			filterParams = append(filterParams, p)
		case "fields":
			fieldParams = append(fieldParams, p)
		case "sort":
			sortParams = append(sortParams, p)
		case "groupBy":
			groupByParams = append(groupByParams, p)
		case "aggregations":
			aggParams = append(aggParams, p)
		case "pagination":
			pagParams = append(pagParams, p)
		case "populate":
			popParams = append(popParams, p)
		}
	}

	filters, err := buildFiltersFromParams(filterParams)
	if err != nil {
		return nil, err
	}

	fields, err := BuildFields(fieldParams)
	if err != nil {
		return nil, err
	}

	sorts, err := BuildSort(sortParams)
	if err != nil {
		return nil, err
	}

	groupBy, err := BuildGroupBy(groupByParams)
	if err != nil {
		return nil, err
	}

	aggregations, err := BuildAggregations(aggParams)
	if err != nil {
		return nil, err
	}

	pagination, err := BuildPagination(pagParams)
	if err != nil {
		return nil, err
	}

	populates, populateAll, err := BuildPopulate(popParams)
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
