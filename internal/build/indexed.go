package build

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/DhimasYulian/hush/internal/query"
)

type indexedItem[T any] struct {
	index int
	value T
}

// BuildIndexed is a generic builder for ordered numeric-index arrays.
// It handles both shorthand syntax (single value) and indexed syntax
// (e.g. "fields[0]=name"), producing an ordered result.
func BuildIndexed[T any](
	params []query.Param,
	sentinel error,
	parseValue func(string) (T, error),
) ([]T, error) {
	if len(params) == 0 {
		return nil, nil
	}

	if len(params) == 1 && len(params[0].Path) == 1 {
		value, err := parseValue(params[0].Value)
		if err != nil {
			return nil, err
		}

		return []T{value}, nil
	}

	items := make([]indexedItem[T], 0, len(params))

	shorthandCount := 0

	for _, p := range params {
		if len(p.Path) == 1 {
			shorthandCount++
		}
	}

	if shorthandCount > 0 && shorthandCount != len(params) {
		return nil, query.QueryError(sentinel, "cannot mix shorthand and indexed syntax")
	}

	if shorthandCount > 1 && shorthandCount == len(params) {
		return nil, query.QueryError(sentinel, "shorthand syntax does not support multiple values")
	}

	for _, p := range params {
		if len(p.Path) != 2 {
			return nil, query.PathError(sentinel, p.Path, "invalid path")
		}

		index, err := strconv.Atoi(p.Path[1])
		if err != nil {
			return nil, query.PathError(sentinel, p.Path, fmt.Sprintf("invalid index %q", p.Path[1]))
		}

		value, err := parseValue(p.Value)
		if err != nil {
			return nil, err
		}

		items = append(items, indexedItem[T]{index: index, value: value})
	}

	sort.Slice(items, func(a, b int) bool {
		return items[a].index < items[b].index
	})

	result := make([]T, len(items))
	for i, item := range items {
		result[i] = item.value
	}

	return result, nil
}
