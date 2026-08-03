package validate

import (
	"fmt"
	"strconv"
	"time"

	"github.com/DhimasYulian/hush/internal/coerce"
	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// maxLogicalDepth caps nested $and/$or/$not depth to prevent stack overflow.
const maxLogicalDepth = 10

// ValidateFilter validates a filter tree against the schema, checking that all
// fields are filterable, operators are allowed, and value types match. It
// returns a copy of the tree whose leaf [query.Condition]s carry their
// schema-declared FieldType and type-coerced Values.
func ValidateFilter(f query.Filter, root *schema.Schema) (query.Filter, error) {
	return validateFilterAt(f, root, 0)
}

// validateFilterAt recursively validates a filter node at a given depth and
// returns the node enriched with field types and coerced values.
func validateFilterAt(f query.Filter, s *schema.Schema, depth int) (query.Filter, error) {
	if f == nil {
		return nil, nil
	}

	if depth > maxLogicalDepth {
		return nil, query.QueryError(ErrNestingTooDeep, fmt.Sprintf("filter nesting exceeds max depth %d", maxLogicalDepth))
	}

	switch n := f.(type) {
	case query.Condition:
		return validateCondition(n, s)

	case query.And:
		filters, err := validateFilterList(n.Filters, s, depth+1)
		if err != nil {
			return nil, err
		}
		return query.And{Filters: filters}, nil

	case query.Or:
		filters, err := validateFilterList(n.Filters, s, depth+1)
		if err != nil {
			return nil, err
		}
		return query.Or{Filters: filters}, nil

	case query.Not:
		child, err := validateFilterAt(n.Filter, s, depth+1)
		if err != nil {
			return nil, err
		}
		return query.Not{Filter: child}, nil

	default:
		return nil, query.QueryError(ErrUnknownFilterNode, fmt.Sprintf("%T", f))
	}
}

func validateFilterList(filters []query.Filter, s *schema.Schema, depth int) ([]query.Filter, error) {
	out := make([]query.Filter, len(filters))
	for i, f := range filters {
		enriched, err := validateFilterAt(f, s, depth)
		if err != nil {
			return nil, err
		}
		out[i] = enriched
	}
	return out, nil
}

// validateCondition validates a single filter condition: field exists, operator
// is allowed, and value types match the field type. It returns the condition
// enriched with its schema-declared field type and coerced values.
func validateCondition(c query.Condition, s *schema.Schema) (query.Condition, error) {
	target, field, err := ResolvePath(s, c.Path)
	if err != nil {
		return query.Condition{}, err
	}

	def, ok := target.GetFilterable(field)
	if !ok {
		return query.Condition{}, query.PathError(ErrUnknownField, c.Path, fmt.Sprintf("%q is not filterable", field))
	}

	if !def.Operators[c.Operator] {
		return query.Condition{}, query.OperatorError(
			ErrOperatorNotAllowed,
			field,
			c.Operator,
			fmt.Sprintf("operator %q is not allowed on field %q", c.Operator, field),
		)
	}

	values, err := validateValue(def.Type, c.Operator, c.Value, field)
	if err != nil {
		return query.Condition{}, err
	}

	c.FieldType = def.Type
	c.Values = values

	return c, nil
}

// validateValue checks that all values in a condition match the expected field
// type and returns the coerced values. NULL operators treat their value as a
// bool presence flag, so they carry no coerced data.
func validateValue(t schema.FieldType, op query.Operator, values query.Value, field string) ([]any, error) {
	if op == query.OpNull || op == query.OpNotNull {
		t = schema.TypeBool
	}

	for _, v := range values {
		if err := validateValueType(t, v); err != nil {
			return nil, query.FieldError(ErrInvalidValue, field, fmt.Sprintf("field %q: %s", field, err))
		}
	}

	if op == query.OpNull || op == query.OpNotNull {
		return nil, nil
	}

	out := make([]any, len(values))
	for i, v := range values {
		coerced, err := coerce.Coerce(t, v)
		if err != nil {
			return nil, query.FieldError(ErrInvalidValue, field, fmt.Sprintf("field %q: %s", field, err))
		}
		out[i] = coerced
	}

	return out, nil
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
