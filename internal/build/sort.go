package build

import (
	"fmt"
	"strings"

	"github.com/DhimasYulian/hush/internal/query"
)

// BuildSort builds an ordered list of Sort directives from params.
func BuildSort(params []query.Param) ([]query.Sort, error) {
	return BuildIndexed(params, ErrInvalidSort, parseSort)
}

// parseSort parses a single sort value like "createdAt:desc" into a Sort.
func parseSort(value string) (query.Sort, error) {
	if value == "" {
		return query.Sort{}, query.QueryError(ErrInvalidSort, "empty sort")
	}

	parts := strings.Split(value, ":")

	if len(parts) > 2 {
		return query.Sort{}, query.QueryError(ErrInvalidSort, fmt.Sprintf("invalid sort %q", value))
	}

	path := strings.Split(parts[0], ".")

	if len(path) == 0 || path[0] == "" {
		return query.Sort{}, query.QueryError(ErrInvalidSort, "invalid field")
	}

	s := query.Sort{
		Path:      path,
		Direction: query.SortAsc,
	}

	if len(parts) == 1 {
		return s, nil
	}

	switch parts[1] {
	case "asc":
		return s, nil

	case "desc":
		s.Direction = query.SortDesc
		return s, nil

	default:
		return query.Sort{}, query.QueryError(ErrInvalidSort, fmt.Sprintf("invalid direction %q", parts[1]))
	}
}
