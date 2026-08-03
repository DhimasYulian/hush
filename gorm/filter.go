package gorm

import (
	"fmt"
	"strings"

	"github.com/DhimasYulian/hush"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// applyFilter translates a filter tree into a single GORM WHERE expression.
// Grouping is preserved through clause.And / clause.Or / clause.Not so that
// nested trees such as (a AND b) OR (c AND d) render with correct parentheses.
func applyFilter(db *gorm.DB, naming schema.Namer, f hush.Filter) *gorm.DB {
	if f == nil {
		return db
	}

	expr, err := buildExpr(f, naming)
	if err != nil {
		_ = db.AddError(err)
		return db
	}

	return db.Where(expr)
}

// buildExpr recursively translates a filter tree into a clause.Expression.
func buildExpr(f hush.Filter, naming schema.Namer) (clause.Expression, error) {
	switch n := f.(type) {
	case hush.Condition:
		return buildCondition(n, naming)

	case hush.And:
		return buildGroup(clause.And, n.Filters, naming)

	case hush.Or:
		return buildGroup(clause.Or, n.Filters, naming)

	case hush.Not:
		child, err := buildExpr(n.Filter, naming)
		if err != nil {
			return nil, err
		}
		return clause.Not(child), nil

	default:
		return nil, fmt.Errorf("hush/gorm: unsupported filter type %T", f)
	}
}

// buildGroup combines a list of sub-filters with the given combinator,
// flattening single-element groups.
func buildGroup(combine func(...clause.Expression) clause.Expression, filters []hush.Filter, naming schema.Namer) (clause.Expression, error) {
	exprs := make([]clause.Expression, 0, len(filters))
	for _, f := range filters {
		expr, err := buildExpr(f, naming)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)
	}

	if len(exprs) == 1 {
		return exprs[0], nil
	}

	return combine(exprs...), nil
}

// buildCondition translates a leaf condition into a clause.Expression. Values
// come from the coerced Condition.Values produced by hush.Parse, so numbers,
// bools, and dates bind with their real types rather than raw strings.
func buildCondition(c hush.Condition, naming schema.Namer) (clause.Expression, error) {
	if len(c.Path) != 1 {
		return nil, fmt.Errorf(
			"hush/gorm: relation-path filter %q requires a join; filter relation columns via populate",
			strings.Join(c.Path, "."),
		)
	}

	col := clause.Column{Name: naming.ColumnName("", c.Path[0])}

	switch c.Operator {
	case hush.OpEq:
		v, err := valueAt(c, 0)
		if err != nil {
			return nil, err
		}
		return clause.Eq{Column: col, Value: v}, nil

	case hush.OpNe:
		v, err := valueAt(c, 0)
		if err != nil {
			return nil, err
		}
		return clause.Neq{Column: col, Value: v}, nil

	case hush.OpGt:
		return comparison(c, func(v any) clause.Expression { return clause.Gt{Column: col, Value: v} })
	case hush.OpGte:
		return comparison(c, func(v any) clause.Expression { return clause.Gte{Column: col, Value: v} })
	case hush.OpLt:
		return comparison(c, func(v any) clause.Expression { return clause.Lt{Column: col, Value: v} })
	case hush.OpLte:
		return comparison(c, func(v any) clause.Expression { return clause.Lte{Column: col, Value: v} })

	case hush.OpIn:
		vals, err := allValues(c)
		if err != nil {
			return nil, err
		}
		return clause.IN{Column: col, Values: vals}, nil

	case hush.OpNotIn:
		vals, err := allValues(c)
		if err != nil {
			return nil, err
		}
		return clause.Not(clause.IN{Column: col, Values: vals}), nil

	case hush.OpBetween:
		lo, err := valueAt(c, 0)
		if err != nil {
			return nil, err
		}
		hi, err := valueAt(c, 1)
		if err != nil {
			return nil, err
		}
		return clause.And(
			clause.Gte{Column: col, Value: lo},
			clause.Lte{Column: col, Value: hi},
		), nil

	case hush.OpContains:
		v, err := stringValue(c)
		if err != nil {
			return nil, err
		}
		return clause.Expr{
			SQL:  "? LIKE ? ESCAPE '\\'",
			Vars: []any{col, "%" + hush.EscapeLike(v) + "%"},
		}, nil

	case hush.OpContainsi:
		v, err := stringValue(c)
		if err != nil {
			return nil, err
		}
		return clause.Expr{
			SQL:  "LOWER(?) LIKE LOWER(?) ESCAPE '\\'",
			Vars: []any{col, "%" + hush.EscapeLike(v) + "%"},
		}, nil

	case hush.OpStartsWith:
		v, err := stringValue(c)
		if err != nil {
			return nil, err
		}
		return clause.Expr{
			SQL:  "? LIKE ? ESCAPE '\\'",
			Vars: []any{col, hush.EscapeLike(v) + "%"},
		}, nil

	case hush.OpEndsWith:
		v, err := stringValue(c)
		if err != nil {
			return nil, err
		}
		return clause.Expr{
			SQL:  "? LIKE ? ESCAPE '\\'",
			Vars: []any{col, "%" + hush.EscapeLike(v)},
		}, nil

	case hush.OpNull:
		return clause.Expr{SQL: "? IS NULL", Vars: []any{col}}, nil

	case hush.OpNotNull:
		return clause.Expr{SQL: "? IS NOT NULL", Vars: []any{col}}, nil

	default:
		return nil, fmt.Errorf("hush/gorm: unsupported operator %q", c.Operator)
	}
}

// comparison builds a single-argument comparison clause from the condition's
// first value.
func comparison(c hush.Condition, build func(v any) clause.Expression) (clause.Expression, error) {
	v, err := valueAt(c, 0)
	if err != nil {
		return nil, err
	}
	return build(v), nil
}

// valueAt returns the i-th value of a condition, preferring the type-coerced
// Condition.Values and falling back to coercing the raw string for hand-built
// queries that skipped hush.Parse.
func valueAt(c hush.Condition, i int) (any, error) {
	if i < len(c.Values) {
		return c.Values[i], nil
	}
	if i < len(c.Value) {
		return hush.Coerce(c.FieldType, c.Value[i])
	}
	return nil, fmt.Errorf("hush/gorm: condition %q %s is missing value %d", c.Path, c.Operator, i)
}

// allValues returns every value of a condition, coercing raw strings when the
// query was not produced by hush.Parse.
func allValues(c hush.Condition) ([]any, error) {
	if len(c.Values) > 0 {
		return c.Values, nil
	}

	out := make([]any, len(c.Value))
	for i, raw := range c.Value {
		v, err := hush.Coerce(c.FieldType, raw)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

// stringValue returns the condition's first value formatted as a string, for
// use with the LIKE-pattern operators.
func stringValue(c hush.Condition) (string, error) {
	v, err := valueAt(c, 0)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", v), nil
}
