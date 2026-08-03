package gorm

import (
	"net/url"
	"testing"
	"time"

	"github.com/DhimasYulian/hush"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Author struct {
	ID        uint
	Name      string
	ProfileID uint
	Profile   Profile
}

type Profile struct {
	ID       uint
	Bio      string
	AuthorID uint
}

type Article struct {
	ID        uint
	Title     string
	Body      *string
	Views     int
	Active    bool
	Status    string
	CreatedAt time.Time
	AuthorID  uint
	Author    Author
}

func strPtr(s string) *string { return &s }

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&Author{}, &Profile{}, &Article{}))
	return db
}

func articleSchema(t *testing.T) *hush.Schema {
	t.Helper()

	profile, err := hush.NewSchema("profile").
		Filterable("bio", hush.TypeString, hush.OpEq).
		Sortable("bio").
		Selectable("id", "bio").
		Build()
	require.NoError(t, err)

	author, err := hush.NewSchema("author").
		Filterable("name", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Sortable("name").
		Selectable("id", "name").
		Relation("profile", profile, 2).
		Build()
	require.NoError(t, err)

	schema, err := hush.NewSchema("article").
		Filterable("title", hush.TypeString, hush.OpEq, hush.OpNe, hush.OpContains, hush.OpContainsi, hush.OpStartsWith, hush.OpEndsWith).
		Filterable("body", hush.TypeString, hush.OpNull, hush.OpNotNull).
		Filterable("views", hush.TypeNumber, hush.OpEq, hush.OpNe, hush.OpGt, hush.OpGte, hush.OpLt, hush.OpLte, hush.OpBetween, hush.OpIn, hush.OpNotIn).
		Filterable("active", hush.TypeBool, hush.OpEq, hush.OpNe).
		Filterable("status", hush.TypeString, hush.OpEq, hush.OpNe, hush.OpIn, hush.OpNotIn).
		Filterable("createdAt", hush.TypeDate, hush.OpGt, hush.OpGte, hush.OpLt, hush.OpLte, hush.OpBetween).
		Sortable("title", "views", "createdAt", "status").
		Selectable("id", "title", "body", "views", "active", "status", "createdAt", "author_id").
		Groupable("status", "views").
		Aggregatable("views").
		Relation("author", author, 2).
		MaxLimit(100).
		Build()
	require.NoError(t, err)

	return schema
}

func seed(t *testing.T, db *gorm.DB) {
	t.Helper()

	alice := Author{Name: "Alice"}
	bob := Author{Name: "Bob"}
	require.NoError(t, db.Create(&alice).Error)
	require.NoError(t, db.Create(&bob).Error)

	now := time.Date(2024, 5, 17, 12, 30, 45, 0, time.UTC)
	articles := []Article{
		{Title: "Go for Beginners", Body: strPtr("intro"), Views: 100, Active: true, Status: "published", CreatedAt: now, AuthorID: alice.ID},
		{Title: "Advanced Go", Views: 200, Active: false, Status: "draft", CreatedAt: now.Add(24 * time.Hour), AuthorID: alice.ID},
		{Title: "50%_off sale", Views: 300, Active: true, Status: "published", CreatedAt: now.Add(48 * time.Hour), AuthorID: bob.ID},
		{Title: "Rust vs Go", Views: 50, Active: true, Status: "archived", CreatedAt: now.Add(72 * time.Hour), AuthorID: bob.ID},
	}
	require.NoError(t, db.Create(&articles).Error)
}

// query builds a hush query from URL values and runs the gorm scope against db.
// The model is always set explicitly so aggregate results can be scanned into a
// destination type that differs from the query's model.
func query(t *testing.T, db *gorm.DB, schema *hush.Schema, values url.Values, dest any) *gorm.DB {
	t.Helper()
	q, err := hush.Parse(values, schema)
	require.NoError(t, err)
	res := db.Model(&Article{}).Scopes(Scopes(schema, q))
	require.NoError(t, res.Find(dest).Error)
	return res
}

func titles(articles []Article) []string {
	out := make([]string, len(articles))
	for i := range articles {
		out[i] = articles[i].Title
	}
	return out
}

func TestOperatorMatrix(t *testing.T) {
	db := openDB(t)
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
			name:   "$gt",
			values: url.Values{"filters[views][$gt]": {"100"}},
			want:   []string{"Advanced Go", "50%_off sale"},
		},
		{
			name:   "$gte",
			values: url.Values{"filters[views][$gte]": {"100"}},
			want:   []string{"Go for Beginners", "Advanced Go", "50%_off sale"},
		},
		{
			name:   "$lt",
			values: url.Values{"filters[views][$lt]": {"100"}},
			want:   []string{"Rust vs Go"},
		},
		{
			name:   "$lte",
			values: url.Values{"filters[views][$lte]": {"100"}},
			want:   []string{"Go for Beginners", "Rust vs Go"},
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
			name:   "$containsi case-insensitive",
			values: url.Values{"filters[title][$containsi]": {"go"}},
			want:   []string{"Go for Beginners", "Advanced Go", "Rust vs Go"},
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
			name:   "LIKE wildcards are literal",
			values: url.Values{"filters[title][$contains]": {"50%_off"}},
			want:   []string{"50%_off sale"},
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var articles []Article
			query(t, db, schema, tt.values, &articles)
			require.Equal(t, tt.want, titles(articles))
		})
	}
}

func TestLogicalGrouping(t *testing.T) {
	db := openDB(t)
	schema := articleSchema(t)
	seed(t, db)

	// (status = published AND views > 300) OR (status = draft AND views > 150).
	// Only the second branch matches (Advanced Go); a wrongly flattened
	// expression would also include 50%_off sale.
	var articles []Article
	query(t, db, schema, url.Values{
		"filters[$or][0][$and][0][status][$eq]": {"published"},
		"filters[$or][0][$and][1][views][$gt]":  {"300"},
		"filters[$or][1][$and][0][status][$eq]": {"draft"},
		"filters[$or][1][$and][1][views][$gt]":  {"150"},
	}, &articles)
	require.Equal(t, []string{"Advanced Go"}, titles(articles))
}

func TestSortAndPagination(t *testing.T) {
	db := openDB(t)
	schema := articleSchema(t)
	seed(t, db)

	t.Run("sort desc", func(t *testing.T) {
		var articles []Article
		query(t, db, schema, url.Values{"sort[0]": {"views:desc"}}, &articles)
		require.Equal(t, []string{"50%_off sale", "Advanced Go", "Go for Beginners", "Rust vs Go"}, titles(articles))
	})

	t.Run("sort whitelist skips unknown", func(t *testing.T) {
		// hush.Parse rejects unknown sortable fields, so a hand-built query is
		// used to exercise the adapter's own whitelist fallback.
		q := &hush.Query{
			Sort: []hush.Sort{
				{Path: hush.Path{"nope"}, Direction: hush.SortAsc},
				{Path: hush.Path{"views"}, Direction: hush.SortAsc},
			},
		}
		var articles []Article
		res := db.Model(&Article{}).Scopes(Scopes(schema, q))
		require.NoError(t, res.Find(&articles).Error)
		require.Equal(t, []string{"Rust vs Go", "Go for Beginners", "Advanced Go", "50%_off sale"}, titles(articles))
	})

	t.Run("limit+offset", func(t *testing.T) {
		var articles []Article
		query(t, db, schema, url.Values{
			"sort[0]":               {"views:desc"},
			"pagination[limit]":     {"2"},
			"pagination[start]":     {"1"},
			"pagination[withCount]": {"false"},
		}, &articles)
		require.Equal(t, []string{"Advanced Go", "Go for Beginners"}, titles(articles))
	})

	t.Run("withCount fetches limit+1", func(t *testing.T) {
		var articles []Article
		res := query(t, db, schema, url.Values{
			"sort[0]":               {"views:desc"},
			"pagination[limit]":     {"2"},
			"pagination[withCount]": {"true"},
		}, &articles)
		require.NoError(t, res.Error)
		require.Len(t, articles, 3) // limit+1 signals has-more
	})

	t.Run("withCount false respects limit", func(t *testing.T) {
		var articles []Article
		res := query(t, db, schema, url.Values{
			"sort[0]":               {"views:desc"},
			"pagination[limit]":     {"2"},
			"pagination[withCount]": {"false"},
		}, &articles)
		require.NoError(t, res.Error)
		require.Len(t, articles, 2)
	})
}

func TestSelectWhitelist(t *testing.T) {
	db := openDB(t)
	schema := articleSchema(t)
	seed(t, db)

	var articles []Article
	query(t, db, schema, url.Values{
		"filters[status][$eq]": {"published"},
		"fields[0]":            {"title"},
	}, &articles)

	require.Len(t, articles, 2)
	for _, a := range articles {
		require.NotEmpty(t, a.Title)
		require.Zero(t, a.Body) // not selected
		require.Zero(t, a.Views)
	}
}

type statusAgg struct {
	Status     string
	Cnt        int
	TotalViews float64
	AvgViews   float64
}

func TestGroupByAggregations(t *testing.T) {
	db := openDB(t)
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

	published := byStatus["published"]
	require.Equal(t, 2, published.Cnt)
	require.Equal(t, float64(400), published.TotalViews)
	require.Equal(t, float64(200), published.AvgViews)

	require.Equal(t, 1, byStatus["draft"].Cnt)
	require.Equal(t, float64(200), byStatus["draft"].TotalViews)
	require.Equal(t, 1, byStatus["archived"].Cnt)
	require.Equal(t, float64(50), byStatus["archived"].TotalViews)
}

func TestPreload(t *testing.T) {
	db := openDB(t)
	schema := articleSchema(t)

	alice := Author{Name: "Alice"}
	bob := Author{Name: "Bob"}
	require.NoError(t, db.Create(&alice).Error)
	require.NoError(t, db.Create(&bob).Error)
	require.NoError(t, db.Create(&Profile{Bio: "writes about Go", AuthorID: alice.ID}).Error)

	now := time.Date(2024, 5, 17, 12, 30, 45, 0, time.UTC)
	require.NoError(t, db.Create(&Article{Title: "A1", Views: 100, Status: "published", CreatedAt: now, AuthorID: alice.ID}).Error)
	require.NoError(t, db.Create(&Article{Title: "A2", Views: 200, Status: "published", CreatedAt: now, AuthorID: bob.ID}).Error)

	t.Run("preload with filter, fields, sort", func(t *testing.T) {
		var articles []Article
		query(t, db, schema, url.Values{
			"filters[status][$eq]":                 {"published"},
			"populate[author][fields][0]":          {"id"},
			"populate[author][fields][1]":          {"name"},
			"populate[author][filters][name][$eq]": {"Alice"},
		}, &articles)

		require.Len(t, articles, 2)
		byTitle := map[string]Article{}
		for _, a := range articles {
			byTitle[a.Title] = a
		}
		require.Equal(t, "Alice", byTitle["A1"].Author.Name)
		require.Empty(t, byTitle["A2"].Author.Name)
	})

	t.Run("nested populate via dotted path", func(t *testing.T) {
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
	})
}

func TestPreloadMaxDepthEnforced(t *testing.T) {
	db := openDB(t)

	// A linear chain s1 -> s2 -> s3 -> leaf where the final relation allows
	// only depth 1, so s1.s2.s3.leaf exceeds it. The query is hand-built to
	// bypass hush.Parse validation; the adapter's eager validation must reject
	// it at scope time even though the table is empty.
	leaf, err := hush.NewSchema("leaf").Build()
	require.NoError(t, err)
	s3, err := hush.NewSchema("s3").Relation("next", leaf, 1).Build()
	require.NoError(t, err)
	s2, err := hush.NewSchema("s2").Relation("next", s3, 2).Build()
	require.NoError(t, err)
	root, err := hush.NewSchema("root").Relation("next", s2, 2).Build()
	require.NoError(t, err)

	q := &hush.Query{
		Populates: []hush.Populate{
			{Relation: "next", Populates: []hush.Populate{
				{Relation: "next", Populates: []hush.Populate{
					{Relation: "next"},
				}},
			}},
		},
	}

	res := Scopes(root, q)(db)
	require.Error(t, res.Error)
	require.Contains(t, res.Error.Error(), "exceeds max depth")
}

func TestNilSchemaOrQueryIsIdentity(t *testing.T) {
	db := openDB(t)

	scope := Scopes(nil, nil)
	require.NotNil(t, scope)
	res := scope(db)
	require.Equal(t, db, res)
}
