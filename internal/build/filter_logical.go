package build

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
)

// buildIndexedLogical builds $and and $or combinators from indexed child nodes.
func buildIndexedLogical(
	n *Node,
	path query.Path,
	construct func([]query.Filter) query.Filter,
) (query.Filter, error) {
	indices, err := n.IndexedChildren()
	if err != nil {
		return nil, err
	}
	if len(indices) == 0 {
		return nil, query.QueryError(ErrInvalidFilters, fmt.Sprintf("%s requires at least one filter", n.Segment))
	}

	children := make([]*Node, len(indices))
	for i, idx := range indices {
		child, err := idx.OnlyChild()
		if err != nil {
			return nil, err
		}
		children[i] = child
	}

	filters, err := buildEach(children, path)
	if err != nil {
		return nil, err
	}

	return construct(filters), nil
}

// buildNot builds a $not combinator from its single child node.
func buildNot(n *Node, path query.Path) (query.Filter, error) {
	child, err := n.OnlyChild()
	if err != nil {
		return nil, err
	}

	filter, err := buildNode(child, path)
	if err != nil {
		return nil, err
	}

	return query.Not{Filter: filter}, nil
}

// buildEach builds each node in a slice, collecting results into a Filter slice.
func buildEach(nodes []*Node, path query.Path) ([]query.Filter, error) {
	filters := make([]query.Filter, len(nodes))

	for i, n := range nodes {
		f, err := buildNode(n, path)
		if err != nil {
			return nil, err
		}
		filters[i] = f
	}

	return filters, nil
}
