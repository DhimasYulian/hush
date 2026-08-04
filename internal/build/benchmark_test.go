package build

import (
	"fmt"
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
)

// BenchmarkBuildQuery_PopulateMany exercises the relation-keyed populate
// builder with many relations, each carrying fields, sort, and filters. This
// path was previously quadratic (a full param rescan per node per option).
func BenchmarkBuildQuery_PopulateMany(b *testing.B) {
	const relations = 20

	params := make([]query.Param, 0, relations*3+3)
	for i := 0; i < relations; i++ {
		rel := fmt.Sprintf("rel%02d", i)
		params = append(params,
			query.Param{Path: []string{"populate", rel, "fields", "0"}, Value: "name"},
			query.Param{Path: []string{"populate", rel, "sort", "0"}, Value: "name:desc"},
			query.Param{Path: []string{"populate", rel, "filters", "status", "$eq"}, Value: "active"},
		)
	}
	params = append(params,
		query.Param{Path: []string{"filters", "title", "$eq"}, Value: "go"},
		query.Param{Path: []string{"fields", "0"}, Value: "title"},
		query.Param{Path: []string{"sort", "0"}, Value: "createdAt:desc"},
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := BuildQuery(params); err != nil {
			b.Fatal(err)
		}
	}
}
