package validate

import (
	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// ValidateGroupBy checks that every field in the list is declared as groupable
// in the root schema.
func ValidateGroupBy(fields []query.Field, root *schema.Schema) error {
	for _, f := range fields {
		if !root.GetGroupable(string(f)) {
			return query.FieldError(ErrUnknownGroupBy, string(f), "field is not groupable")
		}
	}
	return nil
}
