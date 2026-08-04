package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// benchmarkSchema mirrors the fixture schema used by the package tests.
func benchmarkSchema() *schema.Schema {
	return &schema.Schema{
		Name: "article",
		Filterable: map[string]schema.FieldDef{
			"title":       {Name: "title", Type: schema.TypeString, Operators: map[query.Operator]bool{query.OpEq: true}},
			"views":       {Name: "views", Type: schema.TypeNumber, Operators: map[query.Operator]bool{query.OpGt: true}},
			"publishedAt": {Name: "publishedAt", Type: schema.TypeDate, Operators: map[query.Operator]bool{query.OpGt: true}},
			"active":      {Name: "active", Type: schema.TypeBool, Operators: map[query.Operator]bool{query.OpEq: true}},
		},
		MaxLimit: 100,
	}
}

// BenchmarkValidateFilter measures validation of a filter tree whose leaves
// span all field types, capturing the per-value parse cost.
func BenchmarkValidateFilter(b *testing.B) {
	root := benchmarkSchema()

	filter := query.And{Filters: []query.Filter{
		query.Condition{Path: query.Path{"title"}, Operator: query.OpEq, Value: query.Value{"hello"}},
		query.Condition{Path: query.Path{"views"}, Operator: query.OpGt, Value: query.Value{"1000"}},
		query.Condition{Path: query.Path{"publishedAt"}, Operator: query.OpGt, Value: query.Value{"2024-01-01T00:00:00Z"}},
		query.Condition{Path: query.Path{"active"}, Operator: query.OpEq, Value: query.Value{"true"}},
	}}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := ValidateFilter(filter, root); err != nil {
			b.Fatal(err)
		}
	}
}
