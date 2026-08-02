// Package main demonstrates integrating hush with the official MongoDB Go driver.
//
// It translates a parsed hush Query into:
//   - bson.M filter documents → Collection.Find()
//   - bson.D projection, sort
//   - Aggregation pipeline $lookup stages for populate relations
//
// Usage:
//
//	go run ./examples/mongo
package main

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/DhimasYulian/hush"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	authorSchema, err := hush.NewSchema("author").
		Filterable("name", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Sortable("name").
		Selectable("name", "email").
		Build()
	if err != nil {
		panic(err)
	}

	schema, err := hush.NewSchema("article").
		Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpGte, hush.OpLt, hush.OpLte, hush.OpBetween).
		Filterable("status", hush.TypeString, hush.OpEq, hush.OpIn, hush.OpNotNull).
		Sortable("title", "createdAt", "views").
		Selectable("id", "title", "views", "status", "createdAt").
		Aggregatable("views").
		Relation("author", authorSchema, 3).
		MaxLimit(100).
		Build()
	if err != nil {
		panic(err)
	}

	values := url.Values{
		"filters[title][$containsi]":           {"go"},
		"filters[views][$gte]":                 {"50"},
		"filters[status][$eq]":                 {"published"},
		"sort[0]":                              {"createdAt:desc"},
		"fields[0]":                            {"title"},
		"fields[1]":                            {"views"},
		"aggregations[total][func]":            {"count"},
		"aggregations[totalViews][func]":       {"sum"},
		"aggregations[totalViews][field]":      {"views"},
		"pagination[limit]":                    {"25"},
		"pagination[start]":                    {"0"},
		"populate[author][fields][0]":          {"name"},
		"populate[author][filters][name][$eq]": {"Alice"},
	}

	query, err := hush.Parse(values, schema)
	if err != nil {
		fmt.Println("validation error:", err)
		return
	}

	fmt.Println("=== MongoDB Query Options ===")

	// Filter
	filter := bson.M{}
	if query.Filters != nil {
		filter = buildBSON(query.Filters)
	}
	fmt.Println("Filter:", formatBSON(filter))

	// Projection
	projection := bson.M{}
	if len(query.Fields) > 0 {
		for _, f := range query.Fields {
			projection[f] = 1
		}
	}
	if len(projection) > 0 {
		fmt.Println("Projection:", formatBSON(projection))
	}

	// Sort
	sort := bson.M{}
	if len(query.Sort) > 0 {
		for _, s := range query.Sort {
			if s.Direction == hush.SortDesc {
				sort[s.Path[0]] = -1
			} else {
				sort[s.Path[0]] = 1
			}
		}
		fmt.Println("Sort:", formatBSON(sort))
	}

	// Aggregations as $group pipeline stage
	if len(query.Aggregations) > 0 {
		fmt.Println("\n  Aggregation pipeline (instead of Find):")
		fmt.Println("  db.Collection(\"articles\").Aggregate(ctx, mongo.Pipeline{")
		fmt.Println("    {\"$match\": filter},")

		groupFields := bson.M{}
		for _, a := range query.Aggregations {
			if a.Func == "count" {
				groupFields[a.Alias] = bson.M{"$sum": 1}
			} else if a.Func == "sum" {
				groupFields[a.Alias] = bson.M{"$sum": "$" + a.Field}
			} else if a.Func == "avg" {
				groupFields[a.Alias] = bson.M{"$avg": "$" + a.Field}
			}
		}
		group := bson.M{"$group": bson.M{"_id": nil}}
		group["$group"].(bson.M)["_id"] = nil
		for k, v := range groupFields {
			group["$group"].(bson.M)[k] = v
		}
		fmt.Printf("    %s,\n", formatBSON(group))
		fmt.Println("  })")
	}

	// Pagination
	if query.Pagination.Start != nil {
		fmt.Printf("Skip: %d\n", *query.Pagination.Start)
	}
	if query.Pagination.Limit != nil {
		fmt.Printf("Limit: %d\n", *query.Pagination.Limit)
	}

	// Populate using $lookup in aggregation pipeline
	if len(query.Populates) > 0 {
		fmt.Println("\n  Populate ($lookup stages for aggregation):")
		fmt.Println("  db.Collection(\"articles\").Aggregate(ctx, mongo.Pipeline{")
		fmt.Println("    {\"$match\": filter},")
		fmt.Println("    {\"$skip\": skip},")
		fmt.Println("    {\"$limit\": limit},")
		for _, p := range query.Populates {
			fmt.Printf("    %s,\n", formatBSON(buildLookup(p)))
		}
		fmt.Println("  })")
	}
}

// buildLookup creates a $lookup aggregation stage for a Populate relation.
func buildLookup(p hush.Populate) bson.M {
	lookup := bson.M{
		"$lookup": bson.M{
			"from":         p.Relation + "s",
			"localField":   p.Relation + "_id",
			"foreignField": "_id",
			"as":           p.Relation,
		},
	}

	if p.Filters != nil || len(p.Fields) > 0 {
		pipeline := bson.A{}

		if p.Filters != nil {
			pipeline = append(pipeline, bson.M{"$match": buildBSON(p.Filters)})
		}
		if len(p.Fields) > 0 {
			proj := bson.M{"_id": 0}
			for _, f := range p.Fields {
				proj[f] = 1
			}
			pipeline = append(pipeline, bson.M{"$project": proj})
		}

		lookup["$lookup"].(bson.M)["let"] = bson.M{}
		lookup["$lookup"].(bson.M)["pipeline"] = pipeline
		lookup["$lookup"].(bson.M)["as"] = p.Relation
	}

	for _, child := range p.Populates {
		_ = child // nested populates would add more $lookup stages
	}

	return lookup
}

func buildBSON(f hush.Filter) bson.M {
	switch node := f.(type) {
	case hush.Condition:
		return conditionBSON(node)
	case hush.And:
		return andBSON(node)
	case hush.Or:
		return orBSON(node)
	case hush.Not:
		return notBSON(node)
	default:
		return bson.M{}
	}
}

func conditionBSON(c hush.Condition) bson.M {
	field := c.Path[0]

	switch c.Operator {
	case hush.OpEq:
		return bson.M{field: valueAt(c, 0)}
	case hush.OpNe:
		return bson.M{field: bson.M{"$ne": valueAt(c, 0)}}
	case hush.OpGt:
		return bson.M{field: bson.M{"$gt": valueAt(c, 0)}}
	case hush.OpGte:
		return bson.M{field: bson.M{"$gte": valueAt(c, 0)}}
	case hush.OpLt:
		return bson.M{field: bson.M{"$lt": valueAt(c, 0)}}
	case hush.OpLte:
		return bson.M{field: bson.M{"$lte": valueAt(c, 0)}}

	case hush.OpIn:
		return bson.M{field: bson.M{"$in": values(c)}}
	case hush.OpNotIn:
		return bson.M{field: bson.M{"$nin": values(c)}}

	case hush.OpBetween:
		return bson.M{field: bson.M{"$gte": valueAt(c, 0), "$lte": valueAt(c, 1)}}

	// LIKE patterns map to regular expressions. Like EscapeLike, QuoteMeta
	// makes the user's literal text match literally instead of as a pattern.
	case hush.OpContains:
		return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(c.Value[0]), "$options": ""}}
	case hush.OpContainsi:
		return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(c.Value[0]), "$options": "i"}}
	case hush.OpStartsWith:
		return bson.M{field: bson.M{"$regex": "^" + regexp.QuoteMeta(c.Value[0]), "$options": ""}}
	case hush.OpEndsWith:
		return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(c.Value[0]) + "$", "$options": ""}}

	case hush.OpNull:
		return bson.M{field: nil}
	case hush.OpNotNull:
		return bson.M{field: bson.M{"$ne": nil}}

	default:
		return bson.M{}
	}
}

// valueAt returns the i-th condition value with its schema-declared type,
// falling back to hush.Coerce for hand-built queries that skipped Parse.
func valueAt(c hush.Condition, i int) any {
	if i < len(c.Values) {
		return c.Values[i]
	}
	if i < len(c.Value) {
		if v, err := hush.Coerce(c.FieldType, c.Value[i]); err == nil {
			return v
		}
	}
	return nil
}

// values returns every condition value with its schema-declared type.
func values(c hush.Condition) []any {
	out := make([]any, len(c.Value))
	for i := range c.Value {
		out[i] = valueAt(c, i)
	}
	return out
}

func andBSON(a hush.And) bson.M {
	filters := make([]bson.M, len(a.Filters))
	for i, f := range a.Filters {
		filters[i] = buildBSON(f)
	}
	return bson.M{"$and": filters}
}

func orBSON(o hush.Or) bson.M {
	filters := make([]bson.M, len(o.Filters))
	for i, f := range o.Filters {
		filters[i] = buildBSON(f)
	}
	return bson.M{"$or": filters}
}

func notBSON(n hush.Not) bson.M {
	switch child := n.Filter.(type) {
	case hush.Condition:
		field := child.Path[0]
		return bson.M{field: bson.M{"$not": bson.M{operatorToMongo(child.Operator): valueAt(child, 0)}}}
	default:
		return bson.M{"$nor": []bson.M{buildBSON(n.Filter)}}
	}
}

func operatorToMongo(op hush.Operator) string {
	switch op {
	case hush.OpEq:
		return "$eq"
	case hush.OpNe:
		return "$ne"
	case hush.OpGt:
		return "$gt"
	case hush.OpGte:
		return "$gte"
	case hush.OpLt:
		return "$lt"
	case hush.OpLte:
		return "$lte"
	case hush.OpIn:
		return "$in"
	case hush.OpNotIn:
		return "$nin"
	default:
		return "$eq"
	}
}

func formatBSON(m bson.M) string {
	b, err := bson.MarshalExtJSON(m, false, false)
	if err != nil {
		return fmt.Sprintf("%v", m)
	}
	return string(b)
}
