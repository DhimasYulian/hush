package mongo

import (
	"strings"

	"github.com/DhimasYulian/hush"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Pipeline builds the aggregation pipeline for a validated query:
//
//	[$match] [$sort] [$skip] [$limit] [$group] [$lookup ...]
//
// $group is emitted when the query has groupBy or aggregations; $lookup stages
// are emitted for populate relations. An error is returned when a populate
// relation is unknown or exceeds the schema's max depth.
func (t Translator) Pipeline(schema *hush.Schema, q *hush.Query) (mongo.Pipeline, error) {
	return t.pipeline(schema, q, true)
}

// pipeline is Pipeline with a knob for the limit bump: the Find path fetches
// limit+1 under WithCount so callers can detect a next page, while the $facet
// variant has an exact $count and should use the exact limit.
func (t Translator) pipeline(schema *hush.Schema, q *hush.Query, bumpLimit bool) (mongo.Pipeline, error) {
	if q == nil {
		return nil, nil
	}

	var pipe mongo.Pipeline

	if f := t.Filter(q); len(f) > 0 {
		pipe = append(pipe, bson.D{{Key: "$match", Value: f}})
	}
	if sort := t.Sort(schema, q); len(sort) > 0 {
		pipe = append(pipe, bson.D{{Key: "$sort", Value: sort}})
	}
	if skip := t.Skip(q); skip != nil {
		pipe = append(pipe, bson.D{{Key: "$skip", Value: *skip}})
	}
	if limit := t.limit(q, bumpLimit); limit != nil {
		pipe = append(pipe, bson.D{{Key: "$limit", Value: *limit}})
	}
	if stage := t.groupStage(schema, q); stage != nil {
		pipe = append(pipe, stage)
	}
	if len(q.Populates) > 0 {
		lookups, err := t.lookups(schema, q.Populates, nil)
		if err != nil {
			return nil, err
		}
		pipe = append(pipe, lookups...)
	}

	return pipe, nil
}

// PipelineFacet builds a pipeline whose single $facet stage returns both the
// query result and an exact total count in one round trip:
//
//	{$facet: { results: [...], totalCount: [{$match}, {$count}] }}
//
// The result is a single document with "results" and "totalCount" keys, so
// WithCount does not need the limit+1 convention used by the Find path.
func (t Translator) PipelineFacet(schema *hush.Schema, q *hush.Query) (mongo.Pipeline, error) {
	inner, err := t.pipeline(schema, q, false)
	if err != nil {
		return nil, err
	}

	results := bson.A{}
	for _, stage := range inner {
		results = append(results, stage)
	}
	if len(results) == 0 {
		// $facet stages must be non-empty.
		results = append(results, bson.D{{Key: "$match", Value: bson.M{}}})
	}

	count := bson.A{}
	if f := t.Filter(q); len(f) > 0 {
		count = append(count, bson.D{{Key: "$match", Value: f}})
	}
	count = append(count, bson.D{{Key: "$count", Value: "totalCount"}})

	return mongo.Pipeline{
		bson.D{{Key: "$facet", Value: bson.M{
			"results":    results,
			"totalCount": count,
		}}},
	}, nil
}

// groupStage builds the $group stage from groupBy fields and aggregations, or
// nil when neither is present. GroupBy fields whitelist against the schema;
// unknown fields are skipped.
func (t Translator) groupStage(schema *hush.Schema, q *hush.Query) bson.D {
	if q == nil || (len(q.GroupBy) == 0 && len(q.Aggregations) == 0) {
		return nil
	}

	groups := make([]hush.Field, 0, len(q.GroupBy))
	for _, g := range q.GroupBy {
		if schema == nil || schema.Groupable(g) {
			groups = append(groups, g)
		}
	}

	var id any
	switch len(groups) {
	case 0:
		id = nil
	case 1:
		id = "$" + t.fieldName(groups[0])
	default:
		doc := bson.M{}
		for _, g := range groups {
			name := t.fieldName(g)
			doc[name] = "$" + name
		}
		id = doc
	}

	group := bson.M{"_id": id}
	for _, a := range q.Aggregations {
		switch strings.ToLower(a.Func) {
		case "count":
			group[a.Alias] = bson.M{"$sum": 1}
		case "sum":
			if schema == nil || schema.Aggregatable(a.Field) {
				group[a.Alias] = bson.M{"$sum": "$" + t.fieldName(a.Field)}
			}
		case "avg":
			if schema == nil || schema.Aggregatable(a.Field) {
				group[a.Alias] = bson.M{"$avg": "$" + t.fieldName(a.Field)}
			}
		default:
			return nil
		}
	}

	return bson.D{{Key: "$group", Value: group}}
}
