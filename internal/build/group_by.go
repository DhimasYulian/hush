package build

import "github.com/DhimasYulian/hush/internal/query"

// BuildGroupBy builds a list of group-by field names from params.
func BuildGroupBy(params []query.Param) ([]query.Field, error) {
	return BuildIndexed(params, ErrInvalidGroupBy, func(value string) (query.Field, error) {
		return value, nil
	})
}
