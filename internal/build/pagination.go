package build

import (
	"fmt"
	"strconv"

	"github.com/DhimasYulian/hush/internal/query"
)

// BuildPagination builds a Pagination from params with "start", "limit", and
// "withCount" keys. When "withCount" is omitted it defaults to true.
func BuildPagination(params []query.Param) (query.Pagination, error) {
	var pagination query.Pagination

	intFields := map[string]**int{
		"start": &pagination.Start,
		"limit": &pagination.Limit,
	}

	for _, p := range params {
		if len(p.Path) != 2 {
			return pagination, query.PathError(query.ErrInvalidPagination, p.Path, "invalid pagination path")
		}

		key := p.Path[1]

		if key == "withCount" {
			if pagination.WithCount != nil {
				return pagination, query.PathError(query.ErrInvalidPagination, p.Path, "duplicate withCount")
			}

			value, err := strconv.ParseBool(p.Value)
			if err != nil {
				return pagination, query.PathError(query.ErrInvalidPagination, p.Path, fmt.Sprintf("invalid withCount %q", p.Value))
			}

			pagination.WithCount = &value
			continue
		}

		field, ok := intFields[key]
		if !ok {
			return pagination, query.PathError(query.ErrInvalidPagination, p.Path, fmt.Sprintf("unknown pagination key %q", key))
		}

		if *field != nil {
			return pagination, query.PathError(query.ErrInvalidPagination, p.Path, fmt.Sprintf("duplicate %s", key))
		}

		value, err := strconv.Atoi(p.Value)
		if err != nil {
			return pagination, query.PathError(query.ErrInvalidPagination, p.Path, fmt.Sprintf("invalid %s %q", key, p.Value))
		}

		if value < 0 {
			return pagination, query.PathError(query.ErrInvalidPagination, p.Path, fmt.Sprintf("%s must be non-negative", key))
		}

		*field = &value
	}

	if pagination.WithCount == nil {
		v := true
		pagination.WithCount = &v
	}

	return pagination, nil
}
