// Package main demonstrates integrating hush with GORM, a Go ORM.
//
// It translates a parsed hush Query into GORM-compatible clauses:
//   - Filters → clause.Where (Eq, Like, Gt, In, etc.)
//   - Populate → db.Preload (relation eager loading)
//   - Sort → db.Order (OrderByColumn)
//   - Fields → db.Select
//   - GroupBy → db.Group
//   - Pagination → db.Limit / db.Offset
//
// Usage:
//
//	go run ./examples/gorm
package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/DhimasYulian/hush"
	"gorm.io/gorm/clause"
)

func main() {
	authorSchema, err := hush.NewSchema("author").
		Filterable("name", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Sortable("name").
		Selectable("name", "email").
		Build()
	if err != nil {
		panic(err)
	}

	schema, err := hush.NewSchema("article").
		Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpGte, hush.OpLt, hush.OpLte, hush.OpBetween).
		Filterable("status", hush.TypeString, hush.OpEq, hush.OpIn, hush.OpNotNull).
		Sortable("title", "createdAt", "views").
		Selectable("id", "title", "views", "status", "createdAt").
		Groupable("status").
		Aggregatable("views").
		Relation("author", authorSchema, 3).
		MaxLimit(100).
		Build()
	if err != nil {
		panic(err)
	}

	values := url.Values{
		"filters[title][$containsi]":           {"go"},
		"filters[views][$gte]":                 {"50"},
		"filters[status][$eq]":                 {"published"},
		"sort[0]":                              {"createdAt:desc"},
		"fields[0]":                            {"id"},
		"fields[1]":                            {"title"},
		"fields[2]":                            {"views"},
		"groupBy[0]":                           {"status"},
		"aggregations[total][func]":            {"count"},
		"aggregations[totalViews][func]":       {"sum"},
		"aggregations[totalViews][field]":      {"views"},
		"pagination[limit]":                    {"25"},
		"pagination[start]":                    {"0"},
		"populate[author][fields][0]":          {"name"},
		"populate[author][sort][0]":            {"name:asc"},
		"populate[author][filters][name][$eq]": {"Alice"},
	}

	query, err := hush.Parse(values, schema)
	if err != nil {
		fmt.Println("validation error:", err)
		return
	}

	fmt.Println("=== GORM Query Translation ===")

	if query.Filters != nil {
		expr, _ := buildCondition(query.Filters)
		fmt.Printf("// WHERE clause:\n")
		fmt.Printf("db.Where(%#v)\n\n", expr)
	}

	if len(query.Fields) > 0 {
		fmt.Printf("// SELECT:\n")
		fmt.Printf("db.Select(%#v)\n\n", query.Fields)
	}

	if len(query.Sort) > 0 {
		fmt.Println("// ORDER BY:")
		for _, s := range query.Sort {
			dir := "ASC"
			if s.Direction == hush.SortDesc {
				dir = "DESC"
			}
			fmt.Printf("db.Order(%q)\n", fmt.Sprintf("%s %s", s.Path[0], dir))
		}
		fmt.Println()
	}

	if len(query.GroupBy) > 0 {
		fmt.Printf("// GROUP BY:\n")
		fmt.Printf("db.Group(%#v)\n\n", query.GroupBy)
	}

	if len(query.Aggregations) > 0 {
		fmt.Println("// AGGREGATIONS (raw SQL with Select):")
		for _, a := range query.Aggregations {
			if a.Func == "count" {
				fmt.Printf("  db.Select(\"COUNT(%s) AS %s\")\n", a.Field, a.Alias)
			} else {
				fmt.Printf("  db.Select(\"%s(%s) AS %s\")\n", strings.ToUpper(a.Func), a.Field, a.Alias)
			}
		}
		fmt.Println()
	}

	if query.Pagination.Limit != nil {
		fmt.Printf("// PAGINATION:\n")
		fmt.Printf("db.Limit(%d)\n", *query.Pagination.Limit)
		if query.Pagination.Start != nil {
			fmt.Printf("db.Offset(%d)\n", *query.Pagination.Start)
		}
		fmt.Println()
	}

	if len(query.Populates) > 0 {
		fmt.Println("// POPULATE (eager loading):")
		for _, p := range query.Populates {
			printGormPreload(p, "")
		}
		fmt.Println()
	}
}

func printGormPreload(p hush.Populate, prefix string) {
	path := prefix + p.Relation
	fmt.Printf("db.Preload(%q", path)
	if len(p.Fields) > 0 {
		fmt.Printf(", func(db *gorm.DB) *gorm.DB {\n")
		fmt.Printf("    return db.Select(%#v)\n", p.Fields)
		if len(p.Sorts) > 0 {
			for _, s := range p.Sorts {
				fmt.Printf("           .Order(%q)\n", fmt.Sprintf("%s %s", s.Path[0], strings.ToUpper(string(s.Direction))))
			}
		}
		if p.Filters != nil {
			expr, _ := buildCondition(p.Filters)
			fmt.Printf("           .Where(%#v)\n", expr)
		}
		fmt.Printf("  }")
	}
	fmt.Printf(")\n")

	for _, child := range p.Populates {
		printGormPreload(child, path+".")
	}
}

func buildCondition(f hush.Filter) (clause.Expression, []interface{}) {
	switch node := f.(type) {
	case hush.Condition:
		return buildClause(node)
	case hush.And:
		return buildLogicalAnd(node)
	case hush.Or:
		return buildLogicalOr(node)
	case hush.Not:
		return buildLogicalNot(node)
	default:
		return nil, nil
	}
}

func buildClause(c hush.Condition) (clause.Expression, []interface{}) {
	col := clause.Column{Name: c.Path[0]}

	switch c.Operator {
	case hush.OpEq:
		return clause.Eq{Column: col, Value: c.Value[0]}, nil
	case hush.OpNe:
		return clause.Neq{Column: col, Value: c.Value[0]}, nil
	case hush.OpGt:
		return clause.Gt{Column: col, Value: c.Value[0]}, nil
	case hush.OpGte:
		return clause.Gte{Column: col, Value: c.Value[0]}, nil
	case hush.OpLt:
		return clause.Lt{Column: col, Value: c.Value[0]}, nil
	case hush.OpLte:
		return clause.Lte{Column: col, Value: c.Value[0]}, nil

	case hush.OpIn:
		vals := make([]interface{}, len(c.Value))
		for i, v := range c.Value {
			vals[i] = v
		}
		if len(vals) == 1 {
			return clause.IN{Column: col, Values: vals}, nil
		}
		return clause.IN{Column: col, Values: []interface{}{vals}}, nil

	case hush.OpNotIn:
		vals := make([]interface{}, len(c.Value))
		for i, v := range c.Value {
			vals[i] = v
		}
		var inExpr clause.Expression
		if len(vals) == 1 {
			inExpr = clause.IN{Column: col, Values: vals}
		} else {
			inExpr = clause.IN{Column: col, Values: []interface{}{vals}}
		}
		return clause.Not(inExpr), nil

	case hush.OpBetween:
		return clause.And(
			clause.Gte{Column: col, Value: c.Value[0]},
			clause.Lte{Column: col, Value: c.Value[1]},
		), nil

	case hush.OpContains:
		return clause.Like{Column: col, Value: "%" + escapeLike(c.Value[0]) + "%"}, nil
	case hush.OpContainsi:
		return clause.Expr{
			SQL:  "LOWER(?) LIKE LOWER(?)",
			Vars: []interface{}{col, "%" + escapeLike(c.Value[0]) + "%"},
		}, nil
	case hush.OpStartsWith:
		return clause.Like{Column: col, Value: escapeLike(c.Value[0]) + "%"}, nil
	case hush.OpEndsWith:
		return clause.Like{Column: col, Value: "%" + escapeLike(c.Value[0])}, nil

	case hush.OpNull:
		return clause.Eq{Column: col, Value: nil}, nil
	case hush.OpNotNull:
		return clause.Neq{Column: col, Value: nil}, nil

	default:
		return nil, nil
	}
}

func buildLogicalAnd(a hush.And) (clause.Expression, []interface{}) {
	var allArgs []interface{}
	conditions := make([]clause.Expression, len(a.Filters))
	for i, f := range a.Filters {
		expr, args := buildCondition(f)
		conditions[i] = expr
		allArgs = append(allArgs, args...)
	}
	return clause.And(conditions...), allArgs
}

func buildLogicalOr(o hush.Or) (clause.Expression, []interface{}) {
	var allArgs []interface{}
	conditions := make([]clause.Expression, len(o.Filters))
	for i, f := range o.Filters {
		expr, args := buildCondition(f)
		conditions[i] = expr
		allArgs = append(allArgs, args...)
	}
	return clause.Or(conditions...), allArgs
}

func buildLogicalNot(n hush.Not) (clause.Expression, []interface{}) {
	expr, args := buildCondition(n.Filter)
	return clause.Not(expr), args
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}
