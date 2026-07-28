package build

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// buildPopulateRelation handles the relation-keyed syntax where each relation
// can have its own nested fields, sorts, filters, and sub-populates.
//
// Example query string keys:
//
//	populate[author][fields][0]=name
//	populate[author][sort][0]=name:asc
//	populate[author][filters][name][$eq]=Alice
//	populate[author][populate][0]=profile
func buildPopulateRelation(params []query.Param) ([]query.Populate, error) {
	tree := NewPopulateTree()

	for _, param := range params {
		parsed, err := ParsePopulatePath(param.Path)
		if err != nil {
			return nil, err
		}

		tree.Ensure(parsed.Relations)
	}

	for _, node := range tree.Nodes() {
		fieldParams, err := CollectPopulateParams(params, node.Path, "fields")
		if err != nil {
			return nil, err
		}
		if len(fieldParams) > 0 {
			fields, err := BuildFields(fieldParams)
			if err != nil {
				return nil, err
			}
			node.Fields = fields
		}

		sortParams, err := CollectPopulateParams(params, node.Path, "sort")
		if err != nil {
			return nil, err
		}
		if len(sortParams) > 0 {
			sorts, err := BuildSort(sortParams)
			if err != nil {
				return nil, err
			}
			node.Sorts = sorts
		}

		filterParams, err := CollectPopulateParams(params, node.Path, "filters")
		if err != nil {
			return nil, err
		}
		if len(filterParams) > 0 {
			filterTree := NewTree()
			for _, p := range filterParams {
				filterTree.Insert(p.Path, p.Value)
			}
			filters, err := BuildFiltersFromTree(filterTree)
			if err != nil {
				return nil, err
			}
			node.Filters = filters
		}
	}

	return tree.Flatten(), nil
}
