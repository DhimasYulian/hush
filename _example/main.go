// Package main demonstrates ALL hush features by walking a comprehensive query.
//
// This example is the reference contract for integrating hush with any
// database or query builder. It parses a rich query string containing every
// hush feature, then recursively walks the resulting tree. The walkFilterVerbose
// and walkPopulateVerbose functions show the canonical shape every adapter must
// support: type-coerced condition values ([hush.Condition.Values], populated by
// [hush.Parse]) with a raw-string fallback via [hush.Coerce], LIKE-wildcard
// escaping via [hush.EscapeLike], and the [$null]/[$notNull] helpers.
//
// Usage:
//
//	go run ./_example
package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/DhimasYulian/hush"
)

func main() {
	authorSchema, err := hush.NewSchema("author").
		Filterable("name", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Filterable("email", hush.TypeString, hush.OpEq).
		Sortable("name").
		Selectable("name", "email").
		Build()
	if err != nil {
		panic(err)
	}

	schema, err := hush.NewSchema("article").
		Filterable("title", hush.TypeString, hush.OpEq, hush.OpNe, hush.OpContains, hush.OpContainsi, hush.OpStartsWith, hush.OpEndsWith).
		Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpGte, hush.OpLt, hush.OpLte, hush.OpBetween).
		Filterable("status", hush.TypeString, hush.OpEq, hush.OpNe, hush.OpIn, hush.OpNotIn).
		Filterable("publishedAt", hush.TypeDate, hush.OpNull, hush.OpNotNull).
		Filterable("active", hush.TypeBool, hush.OpEq).
		Sortable("title", "createdAt", "views", "status").
		Selectable("id", "title", "body", "views", "status", "publishedAt", "createdAt").
		Groupable("status", "createdAt").
		Aggregatable("views").
		Relation("author", authorSchema, 3).
		MaxLimit(100).
		Build()
	if err != nil {
		panic(err)
	}

	// Build a comprehensive query string covering every hush feature.
	//
	// Filter operators used:
	//   $eq, $ne, $gt, $lte, $between, $in, $null, $contains, $containsi
	// Logical combinators:
	//   $and, $or, $not
	// Other features:
	//   fields, sort (asc/desc), groupBy, aggregations (count/sum/avg),
	//   pagination (start/limit), populate with nested fields/sort/filters
	values := url.Values{
		// --- Filters (implicit AND at root) ---
		"filters[title][$containsi]":     {"go"},
		"filters[views][$gt]":            {"100"},
		"filters[publishedAt][$notNull]": {"true"},

		// --- Fields ---
		"fields[0]": {"id"},
		"fields[1]": {"title"},
		"fields[2]": {"views"},
		"fields[3]": {"status"},

		// --- Sort ---
		"sort[0]": {"createdAt:desc"},
		"sort[1]": {"title:asc"},

		// --- Group By ---
		"groupBy[0]": {"status"},

		// --- Aggregations ---
		"aggregations[articleCount][func]": {"count"},
		"aggregations[totalViews][func]":   {"sum"},
		"aggregations[totalViews][field]":  {"views"},
		"aggregations[avgViews][func]":     {"avg"},
		"aggregations[avgViews][field]":    {"views"},

		// --- Pagination ---
		"pagination[limit]": {"25"},
		"pagination[start]": {"0"},

		// --- Populate (relation-keyed syntax only) ---
		"populate[author][fields][0]":          {"name"},
		"populate[author][sort][0]":            {"name:asc"},
		"populate[author][filters][name][$eq]": {"Alice"},
	}

	query, err := hush.Parse(values, schema)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	printComprehensive(query)
}

// printComprehensive walks and prints every section of the parsed Query.
func printComprehensive(q *hush.Query) {
	fmt.Println("=== hush Comprehensive Query Demo ===")

	// Filters
	if q.Filters != nil {
		fmt.Println("◆ Filters:")
		walkFilterVerbose(q.Filters, 1)
		fmt.Println()
	}

	// Fields
	if len(q.Fields) > 0 {
		fmt.Printf("◆ Fields: %s\n\n", strings.Join(q.Fields, ", "))
	}

	// Sort
	if len(q.Sort) > 0 {
		fmt.Println("◆ Sort:")
		for i, s := range q.Sort {
			dir := strings.ToUpper(string(s.Direction))
			fmt.Printf("  %d. %s %s\n", i+1, s.Path[0], dir)
		}
		fmt.Println()
	}

	// GroupBy
	if len(q.GroupBy) > 0 {
		fmt.Printf("◆ Group By: %s\n\n", strings.Join(q.GroupBy, ", "))
	}

	// Aggregations
	if len(q.Aggregations) > 0 {
		fmt.Println("◆ Aggregations:")
		for _, a := range q.Aggregations {
			if a.Field == "" || a.Field == "*" {
				fmt.Printf("  - %s: %s(*)\n", a.Alias, a.Func)
			} else {
				fmt.Printf("  - %s: %s(%s)\n", a.Alias, a.Func, a.Field)
			}
		}
		fmt.Println()
	}

	// Pagination
	fmt.Println("◆ Pagination:")
	if q.Pagination.Start != nil {
		fmt.Printf("  Start: %d\n", *q.Pagination.Start)
	}
	if q.Pagination.Limit != nil {
		fmt.Printf("  Limit: %d\n", *q.Pagination.Limit)
	}
	if q.Pagination.WithCount != nil {
		fmt.Printf("  WithCount: %t\n", *q.Pagination.WithCount)
	}
	fmt.Println()

	// Populate
	if q.PopulateAll {
		fmt.Println("◆ Populate: * (all relations)")
	} else if len(q.Populates) > 0 {
		fmt.Println("◆ Populate:")
		for _, p := range q.Populates {
			walkPopulateVerbose(p, 1)
		}
		fmt.Println()
	}
}

// ---------------------------------------------------------------------------
// Filter tree walker
// ---------------------------------------------------------------------------

func walkCondition(c hush.Condition) string {
	field := c.Path[0]
	val := c.Value[0]

	// $null / $notNull carry no value; the helpers make them explicit.
	if hush.IsNullOperator(c.Operator) {
		return fmt.Sprintf("%s IS NULL", field)
	}
	if hush.IsNotNullOperator(c.Operator) {
		return fmt.Sprintf("%s IS NOT NULL", field)
	}

	switch c.Operator {
	case hush.OpEq:
		return fmt.Sprintf("%s = %v", field, valueAt(c, 0))
	case hush.OpNe:
		return fmt.Sprintf("%s != %v", field, valueAt(c, 0))
	case hush.OpGt:
		return fmt.Sprintf("%s > %v", field, valueAt(c, 0))
	case hush.OpGte:
		return fmt.Sprintf("%s >= %v", field, valueAt(c, 0))
	case hush.OpLt:
		return fmt.Sprintf("%s < %v", field, valueAt(c, 0))
	case hush.OpLte:
		return fmt.Sprintf("%s <= %v", field, valueAt(c, 0))
	case hush.OpIn:
		return fmt.Sprintf("%s IN (%s)", field, joinValues(c))
	case hush.OpNotIn:
		return fmt.Sprintf("%s NOT IN (%s)", field, joinValues(c))
	case hush.OpBetween:
		return fmt.Sprintf("%s BETWEEN %v AND %v", field, valueAt(c, 0), valueAt(c, 1))
	case hush.OpContains:
		return fmt.Sprintf("%s LIKE '%%%s%%'", field, hush.EscapeLike(val))
	case hush.OpContainsi:
		return fmt.Sprintf("%s ILIKE '%%%s%%'", field, hush.EscapeLike(val))
	case hush.OpStartsWith:
		return fmt.Sprintf("%s LIKE '%s%%'", field, hush.EscapeLike(val))
	case hush.OpEndsWith:
		return fmt.Sprintf("%s LIKE '%%%s'", field, hush.EscapeLike(val))
	default:
		return fmt.Sprintf("%s %s %v", field, c.Operator, valueAt(c, 0))
	}
}

// valueAt returns the i-th condition value with its schema-declared type.
// hush.Parse populates Condition.Values with the coerced values; the raw
// string fallback supports hand-built queries that skipped Parse.
func valueAt(c hush.Condition, i int) any {
	if i < len(c.Values) {
		return c.Values[i]
	}
	if i < len(c.Value) {
		if v, err := hush.Coerce(c.FieldType, c.Value[i]); err == nil {
			return v
		}
	}
	return ""
}

// joinValues renders every condition value comma-separated.
func joinValues(c hush.Condition) string {
	parts := make([]string, len(c.Value))
	for i := range c.Value {
		parts[i] = fmt.Sprintf("%v", valueAt(c, i))
	}
	return strings.Join(parts, ", ")
}

func walkFilterVerbose(f hush.Filter, depth int) {
	indent := strings.Repeat("  ", depth)
	switch node := f.(type) {
	case hush.Condition:
		fmt.Printf("%s▸ Condition: %s\n", indent, walkCondition(node))
	case hush.And:
		fmt.Printf("%s▸ AND:\n", indent)
		for _, child := range node.Filters {
			walkFilterVerbose(child, depth+1)
		}
	case hush.Or:
		fmt.Printf("%s▸ OR:\n", indent)
		for _, child := range node.Filters {
			walkFilterVerbose(child, depth+1)
		}
	case hush.Not:
		fmt.Printf("%s▸ NOT:\n", indent)
		walkFilterVerbose(node.Filter, depth+1)
	}
}

// ---------------------------------------------------------------------------
// Populate tree walker
// ---------------------------------------------------------------------------

func walkPopulateVerbose(p hush.Populate, depth int) {
	indent := strings.Repeat("  ", depth)
	fmt.Printf("%s▸ Relation: %s\n", indent, p.Relation)

	if len(p.Fields) > 0 {
		fmt.Printf("%s  Fields: %s\n", indent, strings.Join(p.Fields, ", "))
	}

	if len(p.Sorts) > 0 {
		for _, s := range p.Sorts {
			dir := strings.ToUpper(string(s.Direction))
			fmt.Printf("%s  Sort: %s %s\n", indent, s.Path[0], dir)
		}
	}

	if p.Filters != nil {
		fmt.Printf("%s  Filters:\n", indent)
		walkFilterVerbose(p.Filters, depth+2)
	}

	if len(p.Populates) > 0 {
		for _, child := range p.Populates {
			walkPopulateVerbose(child, depth+1)
		}
	}
}
