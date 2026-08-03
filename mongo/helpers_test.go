package mongo_test

import (
	"context"
	"testing"

	"github.com/DhimasYulian/hush"
	hushmongo "github.com/DhimasYulian/hush/mongo"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// articleSchema mirrors the hush/gorm test schema so both adapters exercise the
// same operator matrix and relation graph.
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
		Filterable("views", hush.TypeNumber, hush.OpEq, hush.OpNe, hush.OpGt, hush.OpGte, hush.OpLt, hush.OpLte, hush.OpBetween, hush.OpIn, hush.OpNotIn).
		Filterable("active", hush.TypeBool, hush.OpEq, hush.OpNe).
		Filterable("status", hush.TypeString, hush.OpEq, hush.OpNe, hush.OpIn, hush.OpNotIn).
		Filterable("body", hush.TypeString, hush.OpNull, hush.OpNotNull).
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

// mock returns an mtest harness backed by the in-memory mock server, with the
// article collection available on mt.Coll. mtest registers its own cleanup.
func mock(t *testing.T) *mtest.T {
	t.Helper()
	return mtest.New(t, mtest.NewOptions().
		ClientType(mtest.Mock).
		DatabaseName("db").
		CollectionName("articles"))
}

// runFind executes a Find against the mock server with the translated filter
// and options, then returns the raw find command the server received.
func runFind(t *testing.T, mt *mtest.T, schema *hush.Schema, q *hush.Query) bson.Raw {
	t.Helper()

	mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.articles", mtest.FirstBatch, bson.D{}))
	_, err := mt.Coll.Find(context.Background(), hushmongo.Filter(q), hushmongo.FindOptions(schema, q))
	require.NoError(t, err)

	return command(t, mt, "find")
}

// runAggregate executes an Aggregate against the mock server with the
// translated pipeline, then returns the raw aggregate command.
func runAggregate(t *testing.T, mt *mtest.T, schema *hush.Schema, q *hush.Query) bson.Raw {
	t.Helper()

	pipe, err := hushmongo.Pipeline(schema, q)
	require.NoError(t, err)
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.articles", mtest.FirstBatch, bson.D{}))
	_, err = mt.Coll.Aggregate(context.Background(), pipe)
	require.NoError(t, err)

	return command(t, mt, "aggregate")
}

// runAggregateFacet is runAggregate for the $facet pipeline variant.
func runAggregateFacet(t *testing.T, mt *mtest.T, schema *hush.Schema, q *hush.Query) bson.Raw {
	t.Helper()

	pipe, err := hushmongo.PipelineFacet(schema, q)
	require.NoError(t, err)
	mt.AddMockResponses(mtest.CreateCursorResponse(0, "db.articles", mtest.FirstBatch, bson.D{}))
	_, err = mt.Coll.Aggregate(context.Background(), pipe)
	require.NoError(t, err)

	return command(t, mt, "aggregate")
}

// command returns the raw document of the most recent started event with the
// given command name.
func command(t *testing.T, mt *mtest.T, name string) bson.Raw {
	t.Helper()

	for _, evt := range mt.GetAllStartedEvents() {
		if evt.CommandName == name {
			return evt.Command
		}
	}
	t.Fatalf("no %q command was sent", name)
	return nil
}

// sentFilter decodes the filter field from a find command.
func sentFilter(t *testing.T, raw bson.Raw) bson.M {
	t.Helper()
	var out struct {
		Filter bson.M `bson:"filter"`
	}
	require.NoError(t, bson.Unmarshal(raw, &out))
	return out.Filter
}

// sentPipeline decodes the pipeline field from an aggregate command.
func sentPipeline(t *testing.T, raw bson.Raw) bson.A {
	t.Helper()
	var out struct {
		Pipeline bson.A `bson:"pipeline"`
	}
	require.NoError(t, bson.Unmarshal(raw, &out))
	return out.Pipeline
}

// requireRawEqual compares BSON values (documents or pipelines) semantically:
// both sides are canonicalized into order-independent maps with normalized
// numeric types, so key ordering and int vs int32 width differences after a
// mock round trip do not cause false failures.
func requireRawEqual(t *testing.T, want, got any) {
	t.Helper()
	require.Equal(t, canonicalize(want), canonicalize(got))
}

// canonicalize converts BSON values into a canonical form: documents become
// maps, arrays become slices, and all numeric kinds become float64.
func canonicalize(v any) any {
	switch x := v.(type) {
	case bson.M:
		out := make(bson.M, len(x))
		for k, val := range x {
			out[k] = canonicalize(val)
		}
		return out
	case bson.D:
		out := make(bson.M, len(x))
		for _, e := range x {
			out[e.Key] = canonicalize(e.Value)
		}
		return out
	case bson.A:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = canonicalize(val)
		}
		return out
	case int:
		return float64(x)
	case int8:
		return float64(x)
	case int16:
		return float64(x)
	case int32:
		return float64(x)
	case int64:
		return float64(x)
	case uint:
		return float64(x)
	case uint8:
		return float64(x)
	case uint16:
		return float64(x)
	case uint32:
		return float64(x)
	case uint64:
		return float64(x)
	case float32:
		return float64(x)
	case float64:
		return x
	default:
		return v
	}
}
