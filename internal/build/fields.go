package build

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// BuildFields builds a list of selected field names from params.
func BuildFields(params []query.Param) ([]query.Field, error) {
	return BuildIndexed(params, ErrInvalidFields, func(value string) (query.Field, error) {
		return value, nil
	})
}
