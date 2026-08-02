package hush

import "github.com/DhimasYulian/hush/internal/coerce"

// Coerce parses a raw filter value string according to the given field type,
// returning a string for TypeString, a float64 for TypeNumber, a bool for
// TypeBool, and a time.Time for TypeDate. The parsed Query already carries
// these coerced values on each Condition.Values, so most adapters never need
// to call this directly.
func Coerce(t FieldType, raw string) (any, error) {
	return coerce.Coerce(t, raw)
}
