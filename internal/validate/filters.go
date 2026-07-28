package validate

import (
	"fmt"
	"strconv"
	"time"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// maxLogicalDepth caps nested $and/$or/$not depth to prevent stack overflow.
const maxLogicalDepth = 10

// ValidateFilter validates a filter tree against the schema, checking that all
// fields are filterable, operators are allowed, and value types match.
func ValidateFilter(f query.Filter, root *schema.Schema) error {
	return validateFilterAt(f, root, 0)
}

// validateFilterAt recursively validates a filter node at a given depth.
func validateFilterAt(f query.Filter, s *schema.Schema, depth int) error {
	if f == nil {
		return nil
	}

	if depth > maxLogicalDepth {
		return query.QueryError(ErrNestingTooDeep, fmt.Sprintf("filter nesting exceeds max depth %d", maxLogicalDepth))
	}

	switch n := f.(type) {
	case query.Condition:
		return validateCondition(n, s)

	case query.And:
		return validateFilterList(n.Filters, s, depth+1)

	case query.Or:
		return validateFilterList(n.Filters, s, depth+1)

	case query.Not:
		return validateFilterAt(n.Filter, s, depth+1)

	default:
		return query.QueryError(ErrUnknownFilterNode, fmt.Sprintf("%T", f))
	}
}

func validateFilterList(filters []query.Filter, s *schema.Schema, depth int) error {
	for _, f := range filters {
		if err := validateFilterAt(f, s, depth); err != nil {
			return err
		}
	}
	return nil
}

// validateCondition validates a single filter condition: field exists, operator
// is allowed, and value types match the field type.
func validateCondition(c query.Condition, s *schema.Schema) error {
	target, field, err := ResolvePath(s, c.Path)
	if err != nil {
		return err
	}

	def, ok := target.GetFilterable(field)
	if !ok {
		return query.PathError(ErrUnknownField, c.Path, fmt.Sprintf("%q is not filterable", field))
	}

	if !def.Operators[c.Operator] {
		return query.OperatorError(
			ErrOperatorNotAllowed,
			field,
			c.Operator,
			fmt.Sprintf("operator %q is not allowed on field %q", c.Operator, field),
		)
	}

	if err := validateValue(def.Type, c.Operator, c.Value, field); err != nil {
		return err
	}

	return nil
}

// validateValue checks that all values in a condition match the expected field type.
func validateValue(t schema.FieldType, op query.Operator, values query.Value, field string) error {
	if op == query.OpNull || op == query.OpNotNull {
		t = schema.TypeBool
	}

	for _, v := range values {
		if err := validateValueType(t, v); err != nil {
			return query.FieldError(ErrInvalidValue, field, fmt.Sprintf("field %q: %s", field, err))
		}
	}

	return nil
}

// validateValueType checks that a single string value parses correctly for the
// given field type (string, number, bool, or RFC3339 date).
func validateValueType(t schema.FieldType, v string) error {
	switch t {
	case schema.TypeString:
		return nil

	case schema.TypeNumber:
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			return fmt.Errorf("%q is not a valid number", v)
		}
		return nil

	case schema.TypeBool:
		if _, err := strconv.ParseBool(v); err != nil {
			return fmt.Errorf("%q is not a valid bool", v)
		}
		return nil

	case schema.TypeDate:
		if _, err := time.Parse(time.RFC3339, v); err != nil {
			return fmt.Errorf("%q is not a valid RFC3339 date", v)
		}
		return nil

	default:
		return fmt.Errorf("unknown field type %q", t)
	}
}
