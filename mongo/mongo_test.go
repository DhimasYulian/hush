//go:build mongo

package mongo_test

import (
	"context"
	"net/url"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/DhimasYulian/hush"
	hushmongo "github.com/DhimasYulian/hush/mongo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// openMongo connects to a MongoDB server. Point HUSH_TEST_MONGODB_URI at a
// server to override the default (see .github/workflows/ci.yml for the CI
// service container). Build with -tags mongo to run these tests.
func openMongo(t *testing.T) (*mongo.Database, *hush.Schema) {
	t.Helper()

	uri := os.Getenv("HUSH_TEST_MONGODB_URI")
	if uri == "" {
		uri = "mongodb://localhost:27017"
	}

	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	require.NoError(t, err)
	require.NoError(t, client.Ping(ctx, nil))
	t.Cleanup(func() { _ = client.Disconnect(ctx) })

	db := client.Database("hush_test")
	// The test database is persistent across runs and test functions, so drop
	// it to get a clean slate every time.
	require.NoError(t, db.Drop(ctx))

	return db, articleSchema(t)
}

// seedMongo inserts the same author/article graph as the hush/gorm seed. Body
// is left out of most articles so $null/$notNull distinguish missing fields.
func seedMongo(t *testing.T, db *mongo.Database) {
	t.Helper()
	ctx := context.Background()

	profile, err := db.Collection("profiles").InsertOne(ctx, bson.M{"bio": "writes about Go"})
	require.NoError(t, err)

	alice, err := db.Collection("authors").InsertOne(ctx, bson.M{"name": "Alice", "profile_id": profile.InsertedID})
	require.NoError(t, err)
	bob, err := db.Collection("authors").InsertOne(ctx, bson.M{"name": "Bob"})
	require.NoError(t, err)

	now := time.Date(2024, 5, 17, 12, 30, 45, 0, time.UTC)
	_, err = db.Collection("articles").InsertMany(ctx, []any{
		bson.M{"title": "Go for Beginners", "body": "intro", "views": int64(100), "active": true, "status": "published", "createdAt": now, "author_id": alice.InsertedID},
		bson.M{"title": "Advanced Go", "views": int64(200), "active": false, "status": "draft", "createdAt": now.Add(24 * time.Hour), "author_id": alice.InsertedID},
		bson.M{"title": "50%_off sale", "views": int64(300), "active": true, "status": "published", "createdAt": now.Add(48 * time.Hour), "author_id": bob.InsertedID},
		bson.M{"title": "Rust vs Go", "views": int64(50), "active": true, "status": "archived", "createdAt": now.Add(72 * time.Hour), "author_id": bob.InsertedID},
	})
	require.NoError(t, err)
}

// findMongo runs a Find with the translated filter and find options.
func findMongo(t *testing.T, db *mongo.Database, schema *hush.Schema, q *hush.Query) []bson.M {
	t.Helper()

	cur, err := db.Collection("articles").Find(context.Background(), hushmongo.Filter(q), hushmongo.FindOptions(schema, q))
	require.NoError(t, err)
	var out []bson.M
	require.NoError(t, cur.All(context.Background(), &out))
	return out
}

func mongoTitles(articles []bson.M) []string {
	out := make([]string, 0, len(articles))
	for _, a := range articles {
		out = append(out, a["title"].(string))
	}
	sort.Strings(out)
	return out
}

// TestMongoOperatorMatrix is the canonical operator-semantics matrix against a
// real mongod, mirroring the hush/gorm Postgres matrix.
func TestMongoOperatorMatrix(t *testing.T) {
	db, schema := openMongo(t)
	seedMongo(t, db)

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
			name:   "$contains regex metachars are literal",
			values: url.Values{"filters[title][$contains]": {"50%_off"}},
			want:   []string{"50%_off sale"},
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
			q, err := hush.Parse(tt.values, schema)
			require.NoError(t, err)
			want := append([]string(nil), tt.want...)
			sort.Strings(want)
			require.Equal(t, want, mongoTitles(findMongo(t, db, schema, q)))
		})
	}
}

// TestMongoAggregations verifies groupBy + aggregates against a real mongod.
func TestMongoAggregations(t *testing.T) {
	db, schema := openMongo(t)
	seedMongo(t, db)

	q, err := hush.Parse(url.Values{
		"groupBy[0]":                       {"status"},
		"aggregations[cnt][func]":          {"count"},
		"aggregations[total_views][func]":  {"sum"},
		"aggregations[total_views][field]": {"views"},
		"aggregations[avg_views][func]":    {"avg"},
		"aggregations[avg_views][field]":   {"views"},
	}, schema)
	require.NoError(t, err)

	pipe, err := hushmongo.Pipeline(schema, q)
	require.NoError(t, err)

	cur, err := db.Collection("articles").Aggregate(context.Background(), pipe)
	require.NoError(t, err)

	var aggs []struct {
		Status     string  `bson:"_id"`
		Cnt        int64   `bson:"cnt"`
		TotalViews float64 `bson:"total_views"`
		AvgViews   float64 `bson:"avg_views"`
	}
	require.NoError(t, cur.All(context.Background(), &aggs))

	byStatus := map[string]struct {
		Cnt        int64
		TotalViews float64
		AvgViews   float64
	}{}
	for _, a := range aggs {
		byStatus[a.Status] = struct {
			Cnt        int64
			TotalViews float64
			AvgViews   float64
		}{a.Cnt, a.TotalViews, a.AvgViews}
	}

	require.Equal(t, int64(2), byStatus["published"].Cnt)
	require.Equal(t, float64(400), byStatus["published"].TotalViews)
	require.Equal(t, float64(200), byStatus["published"].AvgViews)
	require.Equal(t, int64(1), byStatus["draft"].Cnt)
	require.Equal(t, float64(200), byStatus["draft"].TotalViews)
	require.Equal(t, int64(1), byStatus["archived"].Cnt)
	require.Equal(t, float64(50), byStatus["archived"].TotalViews)
}

// TestMongoPopulate verifies nested $lookup joins against a real mongod.
func TestMongoPopulate(t *testing.T) {
	db, schema := openMongo(t)
	seedMongo(t, db)

	q, err := hush.Parse(url.Values{
		"sort[0]": {"views:desc"},
		"populate[author][populate][profile][fields][0]": {"bio"},
	}, schema)
	require.NoError(t, err)

	pipe, err := hushmongo.Pipeline(schema, q)
	require.NoError(t, err)

	cur, err := db.Collection("articles").Aggregate(context.Background(), pipe)
	require.NoError(t, err)

	var articles []bson.M
	require.NoError(t, cur.All(context.Background(), &articles))

	require.Len(t, articles, 4)
	for _, a := range articles {
		author := a["author"].(bson.A)[0].(bson.M)
		require.NotEmpty(t, author["name"].(string))
		if author["name"].(string) == "Alice" {
			profile := author["profile"].(bson.A)[0].(bson.M)
			require.Equal(t, "writes about Go", profile["bio"].(string))
		}
	}
}

// TestMongoWithCount verifies the limit+1 Find convention and the $facet
// variant against a real mongod.
func TestMongoWithCount(t *testing.T) {
	db, schema := openMongo(t)
	seedMongo(t, db)

	ctx := context.Background()

	t.Run("find fetches limit+1", func(t *testing.T) {
		q, err := hush.Parse(url.Values{
			"sort[0]":               {"views:desc"},
			"pagination[limit]":     {"2"},
			"pagination[withCount]": {"true"},
		}, schema)
		require.NoError(t, err)

		docs := findMongo(t, db, schema, q)
		require.Len(t, docs, 3) // 2 requested + 1 sentinel
	})

	t.Run("facet returns exact count", func(t *testing.T) {
		q, err := hush.Parse(url.Values{
			"filters[status][$eq]":  {"published"},
			"sort[0]":               {"views:desc"},
			"pagination[limit]":     {"2"},
			"pagination[withCount]": {"true"},
		}, schema)
		require.NoError(t, err)

		pipe, err := hushmongo.PipelineFacet(schema, q)
		require.NoError(t, err)

		cur, err := db.Collection("articles").Aggregate(ctx, pipe)
		require.NoError(t, err)

		var facet []bson.M
		require.NoError(t, cur.All(ctx, &facet))

		results := facet[0]["results"].(bson.A)
		total := facet[0]["totalCount"].(bson.A)
		require.Len(t, results, 2)
		require.Equal(t, "50%_off sale", results[0].(bson.M)["title"])
		require.Equal(t, "Go for Beginners", results[1].(bson.M)["title"])
		require.Equal(t, int64(2), countOf(total))
	})
}

// countOf normalizes a decoded $count value (int32 or int64 depending on the
// server) to int64.
func countOf(total bson.A) int64 {
	switch v := total[0].(bson.M)["totalCount"].(type) {
	case int32:
		return int64(v)
	case int64:
		return v
	default:
		return 0
	}
}
