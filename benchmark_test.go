package hush

import (
	"net/url"
	"testing"
)

// BenchmarkParse measures the full parse → build → validate pipeline over a
// query with filters, fields, sort, pagination, and a populated relation.
func BenchmarkParse(b *testing.B) {
	author, err := NewSchema("author").
		Filterable("name", TypeString, OpEq, OpContainsi).
		Sortable("name").
		Selectable("name", "email").
		Build()
	if err != nil {
		b.Fatal(err)
	}

	article, err := NewSchema("article").
		Filterable("title", TypeString, OpContainsi).
		Filterable("views", TypeNumber, OpGt).
		Sortable("createdAt").
		Selectable("title", "body").
		Relation("author", author, 3).
		MaxLimit(50).
		Build()
	if err != nil {
		b.Fatal(err)
	}

	values := url.Values{
		"filters[title][$containsi]":           {"go"},
		"filters[views][$gt]":                  {"10"},
		"fields[0]":                            {"title"},
		"sort[0]":                              {"createdAt:desc"},
		"pagination[limit]":                    {"25"},
		"populate[author][fields][0]":          {"name"},
		"populate[author][sort][0]":            {"name:asc"},
		"populate[author][filters][name][$eq]": {"Alice"},
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := Parse(values, article); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkEscapeLike measures escaping of wildcard-heavy LIKE values.
func BenchmarkEscapeLike(b *testing.B) {
	input := `50%_off\_\a\b %_`

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		EscapeLike(input)
	}
}
