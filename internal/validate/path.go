package validate

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// maxPathDepth caps relation traversal depth to prevent infinite loops.
const maxPathDepth = 10

// ResolvePath walks the schema's relation graph to find the target schema
// and final field name for a given path. For example, path ["author", "name"]
// traverses the "author" relation and returns (authorSchema, "name", nil).
func ResolvePath(root *schema.Schema, path query.Path) (*schema.Schema, string, error) {
	if len(path) == 0 {
		return nil, "", query.QueryError(ErrInvalidPath, "path must not be empty")
	}

	if len(path)-1 > maxPathDepth {
		return nil, "", query.PathError(ErrNestingTooDeep, path, fmt.Sprintf("path exceeds max depth of %d", maxPathDepth))
	}

	cur := root

	for _, seg := range path[:len(path)-1] {
		rel, ok := cur.GetRelation(seg)
		if !ok {
			return nil, "", query.PathError(ErrUnknownField, path, fmt.Sprintf("unknown relation %q", seg))
		}
		cur = rel.Target
	}

	return cur, path[len(path)-1], nil
}
