package mongo_test

import (
	"net/url"
	"testing"

	"github.com/DhimasYulian/hush"
	hushmongo "github.com/DhimasYulian/hush/mongo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestPipelineGroupBy(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("single group key with count aggregation", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"groupBy[0]":              {"status"},
			"aggregations[cnt][func]": {"count"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$group", Value: bson.M{"_id": "$status", "cnt": bson.M{"$sum": 1}}}},
		}, sentPipeline(t, raw))
	})

	mt.Run("multiple group keys and sum aggregation", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"groupBy[0]":                       {"status"},
			"groupBy[1]":                       {"views"},
			"aggregations[cnt][func]":          {"count"},
			"aggregations[total_views][func]":  {"sum"},
			"aggregations[total_views][field]": {"views"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$group", Value: bson.M{
				"_id": bson.M{"status": "$status", "views": "$views"},
				"cnt": bson.M{"$sum": 1},
				"total_views": bson.M{
					"$sum": "$views",
				},
			}}},
		}, sentPipeline(t, raw))
	})

	mt.Run("aggregations without group keys group over the whole set", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"aggregations[cnt][func]": {"count"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$group", Value: bson.M{"_id": nil, "cnt": bson.M{"$sum": 1}}}},
		}, sentPipeline(t, raw))
	})

	mt.Run("filter feeds the match stage before group", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[status][$eq]":    {"published"},
			"groupBy[0]":              {"status"},
			"aggregations[cnt][func]": {"count"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$match", Value: bson.M{"status": "published"}}},
			bson.D{{Key: "$group", Value: bson.M{"_id": "$status", "cnt": bson.M{"$sum": 1}}}},
		}, sentPipeline(t, raw))
	})
}

func TestPipelineStageOrdering(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("match, sort, skip, limit, group", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[status][$eq]":    {"published"},
			"sort[0]":                 {"views:desc"},
			"pagination[start]":       {"1"},
			"pagination[limit]":       {"5"},
			"pagination[withCount]":   {"false"},
			"groupBy[0]":              {"status"},
			"aggregations[cnt][func]": {"count"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$match", Value: bson.M{"status": "published"}}},
			bson.D{{Key: "$sort", Value: bson.D{{Key: "views", Value: int32(-1)}}}},
			bson.D{{Key: "$skip", Value: int64(1)}},
			bson.D{{Key: "$limit", Value: int64(5)}},
			bson.D{{Key: "$group", Value: bson.M{"_id": "$status", "cnt": bson.M{"$sum": 1}}}},
		}, sentPipeline(t, raw))
	})
}

func TestPipelineLookup(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("project selected related fields", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"populate[author][fields][0]": {"name"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$lookup", Value: bson.M{
				"from":         "authors",
				"localField":   "author_id",
				"foreignField": "_id",
				"as":           "author",
				"pipeline": bson.A{
					bson.D{{Key: "$project", Value: bson.M{"name": 1, "_id": 1}}},
				},
			}}},
		}, sentPipeline(t, raw))
	})

	mt.Run("filter related documents", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"populate[author][filters][name][$eq]": {"Alice"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$lookup", Value: bson.M{
				"from":         "authors",
				"localField":   "author_id",
				"foreignField": "_id",
				"as":           "author",
				"pipeline": bson.A{
					bson.D{{Key: "$match", Value: bson.M{"name": "Alice"}}},
				},
			}}},
		}, sentPipeline(t, raw))
	})
}

func TestPipelineLookupNested(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("nested populate becomes a nested lookup", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"populate[author][populate][profile][fields][0]": {"bio"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregate(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$lookup", Value: bson.M{
				"from":         "authors",
				"localField":   "author_id",
				"foreignField": "_id",
				"as":           "author",
				"pipeline": bson.A{
					bson.D{{Key: "$lookup", Value: bson.M{
						"from":         "profiles",
						"localField":   "profile_id",
						"foreignField": "_id",
						"as":           "profile",
						"pipeline": bson.A{
							bson.D{{Key: "$project", Value: bson.M{"_id": 1, "bio": 1}}},
						},
					}}},
				},
			}}},
		}, sentPipeline(t, raw))
	})
}

func TestPipelineLookupErrors(t *testing.T) {
	schema := articleSchema(t)

	t.Run("unknown relation", func(t *testing.T) {
		q := &hush.Query{Populates: []hush.Populate{{Relation: "nope"}}}
		_, err := hushmongo.Pipeline(schema, q)
		require.ErrorContains(t, err, "unknown relation")
	})

	t.Run("relation exceeds max depth", func(t *testing.T) {
		leaf, err := hush.NewSchema("leaf").Build()
		require.NoError(t, err)
		mid, err := hush.NewSchema("mid").Relation("leaf", leaf, 1).Build()
		require.NoError(t, err)
		top, err := hush.NewSchema("top").Relation("mid", mid, 1).Build()
		require.NoError(t, err)

		// Hand-built query bypassing Parse so the scope-level check is what
		// reports the violation, mirroring hush/gorm.
		q := &hush.Query{Populates: []hush.Populate{{
			Relation: "mid",
			Populates: []hush.Populate{{
				Relation: "leaf",
			}},
		}}}
		_, err = hushmongo.Pipeline(top, q)
		require.ErrorContains(t, err, "exceeds max depth")
	})
}

func TestPipelineFacet(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("results and totalCount in one facet", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[status][$eq]":    {"published"},
			"sort[0]":                 {"views:desc"},
			"pagination[limit]":       {"2"},
			"pagination[withCount]":   {"true"},
			"groupBy[0]":              {"status"},
			"aggregations[cnt][func]": {"count"},
		}, schema)
		require.NoError(mt, err)

		raw := runAggregateFacet(t, mt, schema, q)
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$facet", Value: bson.M{
				"results": bson.A{
					bson.D{{Key: "$match", Value: bson.M{"status": "published"}}},
					bson.D{{Key: "$sort", Value: bson.D{{Key: "views", Value: int32(-1)}}}},
					bson.D{{Key: "$limit", Value: int64(2)}},
					bson.D{{Key: "$group", Value: bson.M{"_id": "$status", "cnt": bson.M{"$sum": 1}}}},
				},
				"totalCount": bson.A{
					bson.D{{Key: "$match", Value: bson.M{"status": "published"}}},
					bson.D{{Key: "$count", Value: "totalCount"}},
				},
			}}},
		}, sentPipeline(t, raw))
	})

	mt.Run("empty query facets an empty match", func(mt *mtest.T) {
		raw := runAggregateFacet(t, mt, schema, &hush.Query{})
		requireRawEqual(t, bson.A{
			bson.D{{Key: "$facet", Value: bson.M{
				"results": bson.A{
					bson.D{{Key: "$match", Value: bson.M{}}},
				},
				"totalCount": bson.A{
					bson.D{{Key: "$count", Value: "totalCount"}},
				},
			}}},
		}, sentPipeline(t, raw))
	})
}
