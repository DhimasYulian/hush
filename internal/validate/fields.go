package validate

import (
	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// ValidateFields checks that every field in the list is declared as selectable
// in the root schema.
func ValidateFields(fields []query.Field, root *schema.Schema) error {
	for _, f := range fields {
		if !root.GetSelectable(f) {
			return query.FieldError(ErrUnknownField, f, "field is not selectable")
		}
	}
	return nil
}
