package build

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
)

// Node is a single node in a path tree, holding a segment, optional value,
// and ordered children.
type Node struct {
	Segment  string
	Value    string
	Children map[string]*Node
	Order    []string
}

// NewNode creates a node with the given segment and an empty children map.
func NewNode(segment string) *Node {
	return &Node{
		Segment:  segment,
		Children: make(map[string]*Node),
	}
}

// Child returns the child node for segment, creating it if it doesn't exist.
func (n *Node) Child(segment string) *Node {
	if c, ok := n.Children[segment]; ok {
		return c
	}

	c := NewNode(segment)
	n.Children[segment] = c
	n.Order = append(n.Order, segment)

	return c
}

// OnlyChild returns the single child node, or an error if there isn't exactly one.
func (n *Node) OnlyChild() (*Node, error) {
	if len(n.Children) != 1 {
		return nil, fmt.Errorf("%q must have exactly one child, has %d", n.Segment, len(n.Children))
	}

	return n.Children[n.Order[0]], nil
}

// OrderedChildren returns child nodes in insertion order.
func (n *Node) OrderedChildren() []*Node {
	children := make([]*Node, len(n.Order))
	for i, segment := range n.Order {
		children[i] = n.Children[segment]
	}

	return children
}

// IndexedChildren returns child nodes sorted by their numeric index segments.
// Returns an error if any child segment is not a valid integer.
func (n *Node) IndexedChildren() ([]*Node, error) {
	children := n.OrderedChildren()

	indexed := make([]indexedChild, len(children))
	for i, c := range children {
		idx, err := strconv.Atoi(c.Segment)
		if err != nil {
			return nil, fmt.Errorf("expected numeric index, got %q", c.Segment)
		}
		indexed[i] = indexedChild{index: idx, node: c}
	}

	slices.SortFunc(indexed, func(a, b indexedChild) int {
		return cmp.Compare(a.index, b.index)
	})

	sorted := make([]*Node, len(indexed))
	for i, c := range indexed {
		sorted[i] = c.node
	}

	return sorted, nil
}

// indexedChild pairs a parsed numeric index with its child node.
type indexedChild struct {
	index int
	node  *Node
}
