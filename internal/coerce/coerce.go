// Package coerce converts raw filter value strings to their schema-declared
// Go types so query adapters can bind values without re-parsing them.
package coerce

import (
	"fmt"
	"strconv"
	"time"

	"github.com/DhimasYulian/hush/internal/query"
)

// Coerce parses a raw filter value string according to the given field type.
// It returns a string for TypeString, a float64 for TypeNumber, a bool for
// TypeBool, and a time.Time for TypeDate.
func Coerce(t query.FieldType, raw string) (any, error) {
	switch t {
	case query.TypeString:
		return raw, nil

	case query.TypeNumber:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("%q is not a valid number", raw)
		}
		return n, nil

	case query.TypeBool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("%q is not a valid bool", raw)
		}
		return b, nil

	case query.TypeDate:
		tm, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return nil, fmt.Errorf("%q is not a valid RFC3339 date", raw)
		}
		return tm, nil

	default:
		return nil, fmt.Errorf("unknown field type %q", t)
	}
}
