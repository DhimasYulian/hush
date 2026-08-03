package mongo_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/DhimasYulian/hush"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestFilterOperatorMatrix(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	tests := []struct {
		name   string
		values url.Values
		want   bson.M
	}{
		{
			name:   "$eq",
			values: url.Values{"filters[title][$eq]": {"Advanced Go"}},
			want:   bson.M{"title": "Advanced Go"},
		},
		{
			name:   "$ne",
			values: url.Values{"filters[title][$ne]": {"Advanced Go"}},
			want:   bson.M{"title": bson.M{"$ne": "Advanced Go"}},
		},
		{
			name:   "$gt typed",
			values: url.Values{"filters[views][$gt]": {"100"}},
			want:   bson.M{"views": bson.M{"$gt": float64(100)}},
		},
		{
			name:   "$gte typed",
			values: url.Values{"filters[views][$gte]": {"100"}},
			want:   bson.M{"views": bson.M{"$gte": float64(100)}},
		},
		{
			name:   "$lt typed",
			values: url.Values{"filters[views][$lt]": {"100"}},
			want:   bson.M{"views": bson.M{"$lt": float64(100)}},
		},
		{
			name:   "$lte typed",
			values: url.Values{"filters[views][$lte]": {"100"}},
			want:   bson.M{"views": bson.M{"$lte": float64(100)}},
		},
		{
			name:   "$in single value",
			values: url.Values{"filters[status][$in][0]": {"published"}},
			want:   bson.M{"status": bson.M{"$in": bson.A{"published"}}},
		},
		{
			name:   "$in multiple values",
			values: url.Values{"filters[views][$in][0]": {"100"}, "filters[views][$in][1]": {"300"}},
			want:   bson.M{"views": bson.M{"$in": bson.A{float64(100), float64(300)}}},
		},
		{
			name:   "$notIn",
			values: url.Values{"filters[status][$notIn][0]": {"draft"}},
			want:   bson.M{"status": bson.M{"$nin": bson.A{"draft"}}},
		},
		{
			name:   "$between",
			values: url.Values{"filters[views][$between][0]": {"100"}, "filters[views][$between][1]": {"200"}},
			want:   bson.M{"views": bson.M{"$gte": float64(100), "$lte": float64(200)}},
		},
		{
			name:   "$contains",
			values: url.Values{"filters[title][$contains]": {"Go"}},
			want:   bson.M{"title": bson.M{"$regex": "Go"}},
		},
		{
			name:   "$containsi case-insensitive",
			values: url.Values{"filters[title][$containsi]": {"go"}},
			want:   bson.M{"title": bson.M{"$regex": "go", "$options": "i"}},
		},
		{
			name:   "$contains regex metachars are literal",
			values: url.Values{"filters[title][$contains]": {"50%_off"}},
			want:   bson.M{"title": bson.M{"$regex": "50%_off"}},
		},
		{
			name:   "$startsWith anchored",
			values: url.Values{"filters[title][$startsWith]": {"Adv"}},
			want:   bson.M{"title": bson.M{"$regex": "^Adv"}},
		},
		{
			name:   "$endsWith anchored",
			values: url.Values{"filters[title][$endsWith]": {"Go"}},
			want:   bson.M{"title": bson.M{"$regex": "Go$"}},
		},
		{
			name:   "$null",
			values: url.Values{"filters[body][$null]": {"true"}},
			want:   bson.M{"body": nil},
		},
		{
			name:   "$notNull",
			values: url.Values{"filters[body][$notNull]": {"true"}},
			want:   bson.M{"body": bson.M{"$ne": nil}},
		},
		{
			name:   "bool $eq typed",
			values: url.Values{"filters[active][$eq]": {"false"}},
			want:   bson.M{"active": false},
		},
		{
			name:   "date $gte typed",
			values: url.Values{"filters[createdAt][$gte]": {"2024-05-18T12:30:45Z"}},
			want: bson.M{"createdAt": bson.M{
				"$gte": primitive.DateTime(time.Date(2024, 5, 18, 12, 30, 45, 0, time.UTC).UnixMilli()),
			}},
		},
	}

	for _, tt := range tests {
		tt := tt
		mt.Run(tt.name, func(mt *mtest.T) {
			q, err := hush.Parse(tt.values, schema)
			require.NoError(mt, err)

			raw := runFind(t, mt, schema, q)
			require.Equal(mt, tt.want, sentFilter(t, raw))
		})
	}
}

func TestFilterLogicalGrouping(t *testing.T) {
	mt := mock(t)
	schema := articleSchema(t)

	mt.Run("implicit AND at root flattens to a single $and group", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[status][$eq]": {"published"},
			"filters[views][$gt]":  {"100"},
		}, schema)
		require.NoError(mt, err)

		raw := runFind(t, mt, schema, q)
		require.Equal(mt, bson.M{"$and": bson.A{
			bson.M{"status": "published"},
			bson.M{"views": bson.M{"$gt": float64(100)}},
		}}, sentFilter(t, raw))
	})

	mt.Run("(a AND b) OR (c AND d) keeps grouping", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[$or][0][$and][0][status][$eq]": {"published"},
			"filters[$or][0][$and][1][views][$gt]":  {"300"},
			"filters[$or][1][$and][0][status][$eq]": {"draft"},
			"filters[$or][1][$and][1][views][$gt]":  {"150"},
		}, schema)
		require.NoError(mt, err)

		raw := runFind(t, mt, schema, q)
		require.Equal(mt, bson.M{"$or": bson.A{
			bson.M{"$and": bson.A{
				bson.M{"status": "published"},
				bson.M{"views": bson.M{"$gt": float64(300)}},
			}},
			bson.M{"$and": bson.A{
				bson.M{"status": "draft"},
				bson.M{"views": bson.M{"$gt": float64(150)}},
			}},
		}}, sentFilter(t, raw))
	})

	mt.Run("same-field conditions use $and to avoid key collision", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[$and][0][views][$gt]": {"10"},
			"filters[$and][1][views][$lt]": {"100"},
		}, schema)
		require.NoError(mt, err)

		raw := runFind(t, mt, schema, q)
		require.Equal(mt, bson.M{"$and": bson.A{
			bson.M{"views": bson.M{"$gt": float64(10)}},
			bson.M{"views": bson.M{"$lt": float64(100)}},
		}}, sentFilter(t, raw))
	})

	mt.Run("$not maps to $nor", func(mt *mtest.T) {
		q, err := hush.Parse(url.Values{
			"filters[$not][status][$eq]": {"draft"},
		}, schema)
		require.NoError(mt, err)

		raw := runFind(t, mt, schema, q)
		require.Equal(mt, bson.M{"$nor": bson.A{
			bson.M{"status": "draft"},
		}}, sentFilter(t, raw))
	})

	mt.Run("hand-built query falls back to hush.Coerce", func(mt *mtest.T) {
		q := &hush.Query{
			Filters: hush.Condition{
				Path:      hush.Path{"views"},
				Operator:  hush.OpGt,
				Value:     hush.Value{"50"},
				FieldType: hush.TypeNumber,
			},
		}
		raw := runFind(t, mt, schema, q)
		require.Equal(mt, bson.M{"views": bson.M{"$gt": float64(50)}}, sentFilter(t, raw))
	})

	mt.Run("nil filter yields an empty filter doc", func(mt *mtest.T) {
		raw := runFind(t, mt, schema, &hush.Query{})
		require.Empty(mt, sentFilter(t, raw))
	})
}
