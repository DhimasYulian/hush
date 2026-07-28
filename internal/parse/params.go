package parse

import (
	"net/url"
	"sort"

	"github.com/DhimasYulian/hush/internal/query"
)

// ParseParams converts url.Values into a sorted slice of Param values.
// Keys are sorted alphabetically to ensure deterministic ordering.
func ParseParams(values url.Values) ([]query.Param, error) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	params := make([]query.Param, 0)

	for _, key := range keys {
		path, err := ParsePath(key)
		if err != nil {
			return nil, err
		}

		for _, value := range values[key] {
			params = append(params, query.Param{
				Path:  path,
				Value: value,
			})
		}
	}

	return params, nil
}
