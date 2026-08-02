//go:build postgres

package gorm

import (
	"net/url"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openPostgresDB connects to a Postgres server. Point HUSH_TEST_PG_DSN at a
// database to override the defaults (see .github/workflows/ci.yml for the CI
// service container). Build with -tags postgres to run these tests.
func openPostgresDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("HUSH_TEST_PG_DSN")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=hush_test port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// The test database is persistent across runs and test functions, so drop
	// and recreate the tables to get a clean slate every time.
	require.NoError(t, db.Migrator().DropTable(&Author{}, &Profile{}, &Article{}))
	require.NoError(t, db.AutoMigrate(&Author{}, &Profile{}, &Article{}))
	return db
}

// set returns an order-independent representation of the given titles for
// comparison, since Postgres does not guarantee row order without ORDER BY.
func set(items []string) []string {
	out := append([]string(nil), items...)
	sort.Strings(out)
	return out
}

// TestOperatorMatrixPostgres is the canonical operator-semantics matrix against
// Postgres. It reuses the same schema and seed as the SQLite matrix but
// compares results as sets.
func TestOperatorMatrixPostgres(t *testing.T) {
	db := openPostgresDB(t)
	schema := articleSchema(t)
	seed(t, db)

	tests := []struct {
		name   string
		values url.Values
		want   []string
	}{
		{
			name:   "$eq",
			values: url.Values{"filters[title][$eq]": {"Advanced Go"}},
			want:   []string{"Advanced Go"},
		},
		{
			name:   "$ne",
			values: url.Values{"filters[title][$ne]": {"Advanced Go"}},
			want:   []string{"Go for Beginners", "50%_off sale", "Rust vs Go"},
		},
		{
			name:   "$gt typed",
			values: url.Values{"filters[views][$gt]": {"100"}},
			want:   []string{"Advanced Go", "50%_off sale"},
		},
		{
			name:   "$between inclusive boundaries",
			values: url.Values{"filters[views][$between][0]": {"100"}, "filters[views][$between][1]": {"200"}},
			want:   []string{"Go for Beginners", "Advanced Go"},
		},
		{
			name:   "$in single value",
			values: url.Values{"filters[status][$in][0]": {"published"}},
			want:   []string{"Go for Beginners", "50%_off sale"},
		},
		{
			name:   "$in multiple values",
			values: url.Values{"filters[views][$in][0]": {"100"}, "filters[views][$in][1]": {"300"}},
			want:   []string{"Go for Beginners", "50%_off sale"},
		},
		{
			name:   "$notIn",
			values: url.Values{"filters[status][$notIn][0]": {"draft"}},
			want:   []string{"Go for Beginners", "50%_off sale", "Rust vs Go"},
		},
		{
			name:   "$contains",
			values: url.Values{"filters[title][$contains]": {"Go"}},
			want:   []string{"Go for Beginners", "Advanced Go", "Rust vs Go"},
		},
		{
			name:   "$contains LIKE wildcards are literal",
			values: url.Values{"filters[title][$contains]": {"50%_off"}},
			want:   []string{"50%_off sale"},
		},
		{
			name:   "$containsi case-insensitive",
			values: url.Values{"filters[title][$containsi]": {"go"}},
			want:   []string{"Go for Beginners", "Advanced Go", "Rust vs Go"},
		},
		{
			name:   "$containsi escaped wildcards",
			values: url.Values{"filters[title][$containsi]": {"50%_off"}},
			want:   []string{"50%_off sale"},
		},
		{
			name:   "$startsWith",
			values: url.Values{"filters[title][$startsWith]": {"Adv"}},
			want:   []string{"Advanced Go"},
		},
		{
			name:   "$endsWith",
			values: url.Values{"filters[title][$endsWith]": {"Go"}},
			want:   []string{"Advanced Go", "Rust vs Go"},
		},
		{
			name:   "$null matches missing body",
			values: url.Values{"filters[body][$null]": {"true"}},
			want:   []string{"Advanced Go", "50%_off sale", "Rust vs Go"},
		},
		{
			name:   "$notNull matches present body",
			values: url.Values{"filters[body][$notNull]": {"true"}},
			want:   []string{"Go for Beginners"},
		},
		{
			name:   "bool $eq typed",
			values: url.Values{"filters[active][$eq]": {"false"}},
			want:   []string{"Advanced Go"},
		},
		{
			name:   "date $gte typed",
			values: url.Values{"filters[createdAt][$gte]": {"2024-05-18T12:30:45Z"}},
			want:   []string{"Advanced Go", "50%_off sale", "Rust vs Go"},
		},
		{
			name: "(a AND b) OR (c AND d) grouping",
			values: url.Values{
				"filters[$or][0][$and][0][status][$eq]": {"published"},
				"filters[$or][0][$and][1][views][$gt]":  {"300"},
				"filters[$or][1][$and][0][status][$eq]": {"draft"},
				"filters[$or][1][$and][1][views][$gt]":  {"150"},
			},
			want: []string{"Advanced Go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var articles []Article
			query(t, db, schema, tt.values, &articles)
			require.Equal(t, set(tt.want), set(titles(articles)))
		})
	}
}

// TestPostgresAggregations verifies groupBy + aggregates against Postgres.
func TestPostgresAggregations(t *testing.T) {
	db := openPostgresDB(t)
	schema := articleSchema(t)
	seed(t, db)

	var aggs []statusAgg
	query(t, db, schema, url.Values{
		"groupBy[0]":                       {"status"},
		"aggregations[cnt][func]":          {"count"},
		"aggregations[total_views][func]":  {"sum"},
		"aggregations[total_views][field]": {"views"},
		"aggregations[avg_views][func]":    {"avg"},
		"aggregations[avg_views][field]":   {"views"},
	}, &aggs)

	byStatus := map[string]statusAgg{}
	for _, a := range aggs {
		byStatus[a.Status] = a
	}

	require.Equal(t, 2, byStatus["published"].Cnt)
	require.Equal(t, float64(400), byStatus["published"].TotalViews)
	require.Equal(t, float64(200), byStatus["published"].AvgViews)
	require.Equal(t, 1, byStatus["draft"].Cnt)
	require.Equal(t, float64(200), byStatus["draft"].TotalViews)
	require.Equal(t, 1, byStatus["archived"].Cnt)
	require.Equal(t, float64(50), byStatus["archived"].TotalViews)
}

// TestPostgresPreload verifies nested preload with join-key projection against
// Postgres.
func TestPostgresPreload(t *testing.T) {
	db := openPostgresDB(t)
	schema := articleSchema(t)

	alice := Author{Name: "Alice"}
	bob := Author{Name: "Bob"}
	require.NoError(t, db.Create(&alice).Error)
	require.NoError(t, db.Create(&bob).Error)
	require.NoError(t, db.Create(&Profile{Bio: "writes about Go", AuthorID: alice.ID}).Error)

	now := time.Date(2024, 5, 17, 12, 30, 45, 0, time.UTC)
	require.NoError(t, db.Create(&Article{Title: "A1", Views: 100, Status: "published", CreatedAt: now, AuthorID: alice.ID}).Error)
	require.NoError(t, db.Create(&Article{Title: "A2", Views: 200, Status: "published", CreatedAt: now, AuthorID: bob.ID}).Error)

	var articles []Article
	query(t, db, schema, url.Values{
		"populate[author][populate][profile][fields][0]": {"bio"},
	}, &articles)

	require.Len(t, articles, 2)
	for _, a := range articles {
		require.NotEmpty(t, a.Author.Name)
		if a.Author.Name == "Alice" {
			require.Equal(t, "writes about Go", a.Author.Profile.Bio)
		}
	}
}
