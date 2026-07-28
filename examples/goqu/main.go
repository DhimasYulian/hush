// Package main demonstrates integrating hush with goqu, a SQL query builder.
//
// It translates a parsed hush Query into goqu expressions for SELECT, WHERE,
// ORDER BY, GROUP BY, LIMIT/OFFSET, and JOINs (for populate relations).
//
// Usage:
//
//	go run ./examples/goqu
package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/DhimasYulian/hush"
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
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
		// $and[0] → title $containsi "go"
		// $and[1] → $or: views > 100 OR status = "published"
		// $and[2] → status $notNull
		"filters[$and][0][title][$containsi]":   {"go"},
		"filters[$and][1][$or][0][views][$gt]":  {"100"},
		"filters[$and][1][$or][1][status][$eq]": {"published"},
		"filters[$and][2][status][$notNull]":    {"true"},

		"sort[0]":   {"createdAt:desc"},
		"fields[0]": {"id"},
		"fields[1]": {"title"},
		"fields[2]": {"views"},
		"fields[3]": {"status"},

		"groupBy[0]": {"status"},

		"aggregations[articleCount][func]": {"count"},
		"aggregations[totalViews][func]":   {"sum"},
		"aggregations[totalViews][field]":  {"views"},

		"pagination[limit]": {"25"},
		"pagination[start]": {"0"},

		"populate[author][fields][0]": {"name"},
		"populate[author][sort][0]":   {"name:asc"},
	}

	query, err := hush.Parse(values, schema)
	if err != nil {
		fmt.Println("validation error:", err)
		return
	}

	ds := goqu.From("articles")

	if query.Filters != nil {
		expr, err := filterExpr(query.Filters)
		if err != nil {
			fmt.Println("filter error:", err)
			return
		}
		ds = ds.Where(expr)
	}

	if len(query.Fields) > 0 {
		cols := make([]interface{}, len(query.Fields))
		for i, f := range query.Fields {
			cols[i] = goqu.C(f)
		}
		ds = ds.Select(cols...)
	}

	if len(query.Sort) > 0 {
		orderCols := make([]exp.OrderedExpression, len(query.Sort))
		for i, s := range query.Sort {
			col := goqu.C(s.Path[0])
			if s.Direction == hush.SortDesc {
				orderCols[i] = col.Desc()
			} else {
				orderCols[i] = col.Asc()
			}
		}
		ds = ds.Order(orderCols...)
	}

	if len(query.GroupBy) > 0 {
		cols := make([]interface{}, len(query.GroupBy))
		for i, f := range query.GroupBy {
			cols[i] = goqu.C(f)
		}
		ds = ds.GroupBy(cols...)
	}

	if len(query.Aggregations) > 0 {
		cols := make([]interface{}, len(query.Aggregations))
		for i, a := range query.Aggregations {
			if a.Func == "count" {
				cols[i] = goqu.COUNT(a.Field)
			} else if a.Func == "sum" {
				cols[i] = goqu.SUM(a.Field).As(a.Alias)
			} else if a.Func == "avg" {
				cols[i] = goqu.AVG(a.Field).As(a.Alias)
			}
		}
		ds = ds.SelectAppend(cols...)
	}

	if query.Pagination.Limit != nil {
		ds = ds.Limit(uint(*query.Pagination.Limit))
	}
	if query.Pagination.Start != nil {
		ds = ds.Offset(uint(*query.Pagination.Start))
	}

	// Populate: add JOINs for each top-level relation
	for _, p := range query.Populates {
		ds = ds.Join(
			goqu.T(p.Relation),
			goqu.On(goqu.I(fmt.Sprintf("articles.%s_id", p.Relation)).Eq(goqu.I(fmt.Sprintf("%s.id", p.Relation)))),
		)
	}

	sql, _, _ := ds.ToSQL()
	fmt.Println("Generated SQL:")
	fmt.Println(sql)
}

func filterExpr(f hush.Filter) (exp.Expression, error) {
	switch node := f.(type) {
	case hush.Condition:
		return conditionExpr(node)
	case hush.And:
		return andExpr(node)
	case hush.Or:
		return orExpr(node)
	case hush.Not:
		return notExpr(node)
	default:
		return nil, fmt.Errorf("unknown filter type: %T", f)
	}
}

func conditionExpr(c hush.Condition) (exp.Expression, error) {
	col := goqu.C(c.Path[0])
	val := c.Value[0]

	switch c.Operator {
	case hush.OpEq:
		return col.Eq(val), nil
	case hush.OpNe:
		return col.Neq(val), nil
	case hush.OpGt:
		return col.Gt(val), nil
	case hush.OpGte:
		return col.Gte(val), nil
	case hush.OpLt:
		return col.Lt(val), nil
	case hush.OpLte:
		return col.Lte(val), nil

	case hush.OpIn:
		vals := make([]interface{}, len(c.Value))
		for i, v := range c.Value {
			vals[i] = v
		}
		return col.In(vals...), nil
	case hush.OpNotIn:
		vals := make([]interface{}, len(c.Value))
		for i, v := range c.Value {
			vals[i] = v
		}
		return col.NotIn(vals...), nil

	case hush.OpBetween:
		return col.Between(goqu.Range(c.Value[0], c.Value[1])), nil

	case hush.OpContains:
		return col.Like("%" + escapeLike(val) + "%"), nil
	case hush.OpContainsi:
		return col.ILike("%" + escapeLike(val) + "%"), nil
	case hush.OpStartsWith:
		return col.Like(escapeLike(val) + "%"), nil
	case hush.OpEndsWith:
		return col.Like("%" + escapeLike(val)), nil

	case hush.OpNull:
		return col.IsNull(), nil
	case hush.OpNotNull:
		return col.IsNotNull(), nil

	default:
		return nil, fmt.Errorf("unsupported operator: %s", c.Operator)
	}
}

func andExpr(a hush.And) (exp.Expression, error) {
	exprs := make([]exp.Expression, len(a.Filters))
	for i, f := range a.Filters {
		expr, err := filterExpr(f)
		if err != nil {
			return nil, err
		}
		exprs[i] = expr
	}
	return goqu.And(exprs...), nil
}

func orExpr(o hush.Or) (exp.Expression, error) {
	exprs := make([]exp.Expression, len(o.Filters))
	for i, f := range o.Filters {
		expr, err := filterExpr(f)
		if err != nil {
			return nil, err
		}
		exprs[i] = expr
	}
	return goqu.Or(exprs...), nil
}

func notExpr(n hush.Not) (exp.Expression, error) {
	expr, err := filterExpr(n.Filter)
	if err != nil {
		return nil, err
	}
	return goqu.L("NOT (?)", expr), nil
}

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "%", "\\%")
	s = strings.ReplaceAll(s, "_", "\\_")
	return s
}
