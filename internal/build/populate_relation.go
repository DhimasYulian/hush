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
//
// Params are parsed once and bucketed by their leaf relation node and option
// (fields/sort/filters), avoiding repeated full scans of the param list.
func buildPopulateRelation(params []query.Param) ([]query.Populate, error) {
	tree := NewPopulateTree()
	buckets := make(map[*PopulateNode]map[string][]query.Param)

	for _, param := range params {
		parsed, err := ParsePopulatePath(param.Path)
		if err != nil {
			return nil, err
		}

		node := tree.Ensure(parsed.Relations)

		if parsed.Option == "" {
			continue
		}

		opts := buckets[node]
		if opts == nil {
			opts = make(map[string][]query.Param)
			buckets[node] = opts
		}

		relParam := query.Param{
			Path:  append([]string{parsed.Option}, parsed.OptionPath...),
			Value: param.Value,
		}
		opts[parsed.Option] = append(opts[parsed.Option], relParam)
	}

	for _, node := range tree.Nodes() {
		opts := buckets[node]
		if opts == nil {
			continue
		}

		if fieldParams := opts["fields"]; len(fieldParams) > 0 {
			fields, err := BuildFields(fieldParams)
			if err != nil {
				return nil, err
			}
			node.Fields = fields
		}

		if sortParams := opts["sort"]; len(sortParams) > 0 {
			sorts, err := BuildSort(sortParams)
			if err != nil {
				return nil, err
			}
			node.Sorts = sorts
		}

		if filterParams := opts["filters"]; len(filterParams) > 0 {
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
