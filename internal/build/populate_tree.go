package build

import (
	"github.com/DhimasYulian/hush/internal/query"
)

// PopulateTree is a specialized tree for building nested populate relations.
// Unlike the generic filter Tree, each node carries fields, sorts, and filters.
type PopulateTree struct {
	Children map[string]*PopulateNode
	Order    []string
}

// PopulateNode represents a single relation in a populate tree, with optional
// field selection, sorting, filtering, and nested populates.
type PopulateNode struct {
	Relation string

	Fields  []query.Field
	Sorts   []query.Sort
	Filters query.Filter

	Children map[string]*PopulateNode
	Order    []string
}

// NewPopulateTree creates an empty populate tree.
func NewPopulateTree() *PopulateTree {
	return &PopulateTree{Children: make(map[string]*PopulateNode)}
}

func NewPopulateNode(relation string) *PopulateNode {
	return &PopulateNode{
		Relation: relation,
		Children: make(map[string]*PopulateNode),
	}
}

// Ensure creates intermediate nodes for the given relation path and returns
// the leaf node. Existing nodes are reused.
func (t *PopulateTree) Ensure(relations []string) *PopulateNode {
	children, order := t.Children, &t.Order

	var node *PopulateNode

	for _, relation := range relations {
		n, ok := children[relation]
		if !ok {
			n = NewPopulateNode(relation)
			children[relation] = n
			*order = append(*order, relation)
		}

		node = n
		children = n.Children
		order = &n.Order
	}

	return node
}

// Nodes returns all nodes in depth-first insertion order.
func (t *PopulateTree) Nodes() []*PopulateNode {
	var nodes []*PopulateNode
	collectNodes(t.Children, t.Order, &nodes)
	return nodes
}

func collectNodes(children map[string]*PopulateNode, order []string, nodes *[]*PopulateNode) {
	for _, relation := range order {
		node := children[relation]
		*nodes = append(*nodes, node)
		collectNodes(node.Children, node.Order, nodes)
	}
}

// Flatten converts the tree into a flat slice of [query.Populate] with nested
// children, preserving insertion order.
func (t *PopulateTree) Flatten() []query.Populate {
	return flattenPopulate(t.Children, t.Order)
}

func flattenPopulate(children map[string]*PopulateNode, order []string) []query.Populate {
	if len(order) == 0 {
		return nil
	}

	result := make([]query.Populate, len(order))
	for i, relation := range order {
		node := children[relation]
		result[i] = query.Populate{
			Relation:  node.Relation,
			Fields:    node.Fields,
			Sorts:     node.Sorts,
			Filters:   node.Filters,
			Populates: flattenPopulate(node.Children, node.Order),
		}
	}

	return result
}
