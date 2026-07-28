package build

// Tree is a path-based trie used to group filter parameters by their
// field path segments before recursive dispatch.
type Tree struct {
	Root *Node
}

// NewTree creates an empty filter tree with a root node.
func NewTree() *Tree {
	return &Tree{Root: NewNode("")}
}

// Insert adds a path/value pair to the tree, creating intermediate nodes as needed.
func (t *Tree) Insert(path []string, value string) {
	n := t.Root
	for _, segment := range path {
		n = n.Child(segment)
	}
	n.Value = value
}
