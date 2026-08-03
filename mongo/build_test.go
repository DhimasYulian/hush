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

func TestProjection(t *testing.T) {
	schema := articleSchema(t)

	q, err := hush.Parse(url.Values{
		"fields[0]": {"title"},
		"fields[1]": {"views"},
	}, schema)
	require.NoError(t, err)

	require.Equal(t, bson.M{"_id": 0, "title": 1, "views": 1}, hushmongo.Projection(schema, q))

	// Non-selectable fields are skipped (hand-built query exercising the
	// whitelist); only _id:0 remains.
	q = &hush.Query{Fields: []hush.Field{"nope"}}
	require.Equal(t, bson.M{"_id": 0}, hushmongo.Projection(schema, q))

	q = &hush.Query{}
	require.Nil(t, hushmongo.Projection(schema, q))
}

func TestSort(t *testing.T) {
	schema := articleSchema(t)

	q, err := hush.Parse(url.Values{
		"sort[0]": {"createdAt:desc"},
		"sort[1]": {"title:asc"},
	}, schema)
	require.NoError(t, err)
	require.Equal(t, bson.D{{Key: "createdAt", Value: int32(-1)}, {Key: "title", Value: int32(1)}}, hushmongo.Sort(schema, q))

	// Unknown fields are skipped (hand-built query exercising the whitelist).
	q = &hush.Query{Sort: []hush.Sort{
		{Path: hush.Path{"nope"}, Direction: hush.SortAsc},
		{Path: hush.Path{"views"}, Direction: hush.SortDesc},
	}}
	require.Equal(t, bson.D{{Key: "views", Value: int32(-1)}}, hushmongo.Sort(schema, q))
}

func TestPagination(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("skip and exact limit", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"pagination[limit]":     {"2"},
			"pagination[start]":     {"1"},
			"pagination[withCount]": {"false"},
		}, schema)
		require.NoError(mt, err)

		raw := runFind(t, mt, schema, q)
		var sent struct {
			Skip  *int64 `bson:"skip"`
			Limit *int64 `bson:"limit"`
		}
		require.NoError(t, bson.Unmarshal(raw, &sent))
		require.Equal(t, int64(1), *sent.Skip)
		require.Equal(t, int64(2), *sent.Limit)
	})

	mt.Run("withCount fetches limit+1", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"pagination[limit]":     {"2"},
			"pagination[withCount]": {"true"},
		}, schema)
		require.NoError(mt, err)

		require.Equal(t, int64(3), *hushmongo.Limit(q))
		require.Nil(t, hushmongo.Skip(q))

		raw := runFind(t, mt, schema, q)
		var sent struct {
			Limit *int64 `bson:"limit"`
		}
		require.NoError(t, bson.Unmarshal(raw, &sent))
		require.Equal(t, int64(3), *sent.Limit)
	})

	mt.Run("no pagination sends no skip or limit", func(mt *mtest.T) {
		raw := runFind(t, mt, schema, &hush.Query{})
		var sent struct {
			Skip  *int64 `bson:"skip"`
			Limit *int64 `bson:"limit"`
		}
		require.NoError(t, bson.Unmarshal(raw, &sent))
		require.Nil(t, sent.Skip)
		require.Nil(t, sent.Limit)
	})
}

func TestFindOptionsBundlesAllClauses(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("find command carries filter, projection, sort, skip, limit", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[status][$eq]":  {"published"},
			"sort[0]":               {"views:desc"},
			"fields[0]":             {"title"},
			"pagination[limit]":     {"2"},
			"pagination[withCount]": {"false"},
		}, schema)
		require.NoError(mt, err)

		raw := runFind(t, mt, schema, q)
		require.Equal(t, bson.M{"status": "published"}, sentFilter(t, raw))

		var sent struct {
			Projection bson.M `bson:"projection"`
			Sort       bson.D `bson:"sort"`
			Limit      int64  `bson:"limit"`
		}
		require.NoError(t, bson.Unmarshal(raw, &sent))
		require.Equal(t, bson.M{"_id": int32(0), "title": int32(1)}, sent.Projection)
		require.Equal(t, bson.D{{Key: "views", Value: int32(-1)}}, sent.Sort)
		require.Equal(t, int64(2), sent.Limit)
	})

	mt.Run("FindOptions is nil when nothing applies", func(mt *mtest.T) {
		require.Nil(t, hushmongo.FindOptions(schema, &hush.Query{}))
	})
}
