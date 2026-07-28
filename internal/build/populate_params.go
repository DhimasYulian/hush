package build

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// CollectPopulateParams filters params to those matching a specific relation
// path and option type (e.g. "fields", "sort", "filters"), rewriting paths
// to be relative to the option root.
func CollectPopulateParams(
	params []query.Param,
	relations []string,
	option string,
) ([]query.Param, error) {
	var result []query.Param

	for _, p := range params {
		parsed, err := ParsePopulatePath(p.Path)
		if err != nil {
			return nil, err
		}

		if parsed.Option != option || !equalRelations(parsed.Relations, relations) {
			continue
		}

		path := make([]string, 1+len(parsed.OptionPath))
		path[0] = option
		copy(path[1:], parsed.OptionPath)

		result = append(result, query.Param{
			Path:  path,
			Value: p.Value,
		})
	}

	return result, nil
}

// equalRelations reports whether two relation path slices are equal.
func equalRelations(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
