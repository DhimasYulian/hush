package validate

import (
	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// ValidateSort checks that every sort path resolves to a sortable field.
func ValidateSort(sorts []query.Sort, root *schema.Schema) error {
	for _, s := range sorts {
		if err := validateSortEntry(s, root); err != nil {
			return err
		}
	}
	return nil
}

// validateSortEntry validates a single sort directive by resolving its path
// and checking that the target field is sortable.
func validateSortEntry(s query.Sort, root *schema.Schema) error {
	target, field, err := ResolvePath(root, s.Path)
	if err != nil {
		return err
	}

	if !target.GetSortable(field) {
		return query.PathError(ErrUnknownField, s.Path, "field is not sortable")
	}

	return nil
}
