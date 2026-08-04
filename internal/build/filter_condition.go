package build

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
)

// buildCondition constructs a leaf [query.Condition] from a path and operator node.
// The path is copied so sibling branches that share an underlying slice during
// descent cannot mutate this condition's stored path.
func buildCondition(path query.Path, n *Node) (query.Filter, error) {
	op, ok := operators[n.Segment]
	if !ok {
		return nil, query.PathError(ErrInvalidFilters, path, fmt.Sprintf("unknown operator %q", n.Segment))
	}

	value, err := buildValue(op, n)
	if err != nil {
		return nil, err
	}

	return query.Condition{Path: append([]string(nil), path...), Operator: op, Value: value}, nil
}

func buildValue(op query.Operator, n *Node) (query.Value, error) {
	if op != query.OpIn && op != query.OpNotIn && op != query.OpBetween {
		return query.Value{n.Value}, nil
	}

	children, err := n.IndexedChildren()
	if err != nil {
		return nil, err
	}

	values := make(query.Value, len(children))
	for i, c := range children {
		values[i] = c.Value
	}

	if err := validateArity(op, values); err != nil {
		return nil, err
	}

	return values, nil
}

func buildField(n *Node, path query.Path) (query.Filter, error) {
	path = append(path, n.Segment)

	child, err := n.OnlyChild()
	if err != nil {
		return nil, err
	}

	switch {
	case isOperator(child.Segment):
		return buildCondition(path, child)
	case isLogical(child.Segment):
		return buildNode(child, path)
	default:
		return buildField(child, path)
	}
}
