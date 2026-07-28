package validate

import (
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// ValidatePagination checks that the pagination limit does not exceed the
// schema's configured maximum.
func ValidatePagination(p query.Pagination, root *schema.Schema) error {
	if p.Limit == nil {
		return nil
	}

	if *p.Limit > root.GetMaxLimit() {
		return query.QueryError(query.ErrInvalidPagination, fmt.Sprintf("limit %d exceeds max %d", *p.Limit, root.GetMaxLimit()))
	}

	return nil
}
