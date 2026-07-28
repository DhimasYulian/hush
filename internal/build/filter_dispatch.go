package build

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// buildNode recursively dispatches filter tree nodes to the appropriate builder
// based on the node's segment value (operator, logical combinator, or field name).
func buildNode(n *Node, path query.Path) (query.Filter, error) {
	switch n.Segment {
	case "filters":
		return buildRoot(n)
	case "$and":
		return buildIndexedLogical(n, path, func(fs []query.Filter) query.Filter {
			return query.And{Filters: fs}
		})
	case "$or":
		return buildIndexedLogical(n, path, func(fs []query.Filter) query.Filter {
			return query.Or{Filters: fs}
		})
	case "$not":
		return buildNot(n, path)
	default:
		if isOperator(n.Segment) {
			return buildCondition(path, n)
		}
		return buildField(n, path)
	}
}

// buildRoot handles the top-level "filters" node, wrapping multiple
// child filters in an implicit AND.
func buildRoot(n *Node) (query.Filter, error) {
	children := n.OrderedChildren()

	switch len(children) {
	case 0:
		return nil, query.QueryError(ErrInvalidFilters, "filters must not be empty")
	case 1:
		return buildNode(children[0], nil)
	default:
		filters, err := buildEach(children, nil)
		if err != nil {
			return nil, err
		}
		return query.And{Filters: filters}, nil
	}
}

// BuildFiltersFromTree extracts the "filters" subtree and builds it into
// a typed Filter tree. Returns nil if no filter params are present.
func BuildFiltersFromTree(tree *Tree) (query.Filter, error) {
	if len(tree.Root.Children) == 0 {
		return nil, nil
	}

	root, ok := tree.Root.Children["filters"]
	if !ok {
		return nil, query.QueryError(ErrInvalidFilters, "no filters present")
	}

	return buildNode(root, nil)
}
