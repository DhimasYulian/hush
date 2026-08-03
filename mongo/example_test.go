package mongo_test

import (
	"fmt"
	"net/url"

	"github.com/DhimasYulian/hush"
	hushmongo "github.com/DhimasYulian/hush/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

// ExamplePipeline translates a validated hush.Query into MongoDB filter,
// find-option, and aggregation structures with zero per-operator code. The
// schema whitelists everything a client may touch, so the translated documents
// are safe to hand to the driver.
func ExamplePipeline() {
	authorSchema, err := hush.NewSchema("author").
		Filterable("name", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Sortable("name").
		Selectable("id", "name").
		Build()
	if err != nil {
		fmt.Println("schema:", err)
		return
	}

	articleSchema, err := hush.NewSchema("article").
		Filterable("title", hush.TypeString, hush.OpContainsi).
		Filterable("views", hush.TypeNumber, hush.OpGt).
		Filterable("status", hush.TypeString, hush.OpEq, hush.OpIn).
		Sortable("title", "views", "status").
		Selectable("id", "title", "views", "status").
		Groupable("status").
		Aggregatable("views").
		Relation("author", authorSchema, 2).
		MaxLimit(100).
		Build()
	if err != nil {
		fmt.Println("schema:", err)
		return
	}

	q, err := hush.Parse(url.Values{
		"filters[status][$eq]":        {"published"},
		"filters[views][$gt]":         {"50"},
		"sort[0]":                     {"views:desc"},
		"fields[0]":                   {"title"},
		"pagination[limit]":           {"2"},
		"pagination[withCount]":       {"true"},
		"populate[author][fields][0]": {"name"},
	}, articleSchema)
	if err != nil {
		fmt.Println("parse:", err)
		return
	}

	fmt.Println("filter for Find or $match:")
	fmt.Printf("%v\n", hushmongo.Filter(q))

	fmt.Println("find options (projection, sort, skip, limit):")
	fmt.Printf("%v\n", hushmongo.Projection(articleSchema, q))
	fmt.Printf("%v\n", hushmongo.Sort(articleSchema, q))
	skip, limit := hushmongo.Skip(q), hushmongo.Limit(q)
	if limit != nil {
		fmt.Printf("skip=%v limit=%d (limit+1 because withCount)\n", skip, *limit)
	} else {
		fmt.Printf("skip=%v limit=<nil>\n", skip)
	}
	fmt.Println("aggregation pipeline:")
	pipe, err := hushmongo.Pipeline(articleSchema, q)
	if err != nil {
		fmt.Println("pipeline:", err)
		return
	}
	for _, stage := range pipe {
		fmt.Printf("%v\n", stage)
	}

	fmt.Println("facet pipeline (results + exact totalCount):")
	facet, err := hushmongo.PipelineFacet(articleSchema, q)
	if err != nil {
		fmt.Println("pipeline:", err)
		return
	}
	fmt.Println("results:")
	for _, s := range facet[0][0].Value.(bson.M)["results"].(bson.A) {
		fmt.Printf("%v\n", s)
	}
	fmt.Println("totalCount:")
	for _, s := range facet[0][0].Value.(bson.M)["totalCount"].(bson.A) {
		fmt.Printf("%v\n", s)
	}

	// Output:
	// filter for Find or $match:
	// map[$and:[map[status:published] map[views:map[$gt:50]]]]
	// find options (projection, sort, skip, limit):
	// map[_id:0 title:1]
	// [{views -1}]
	// skip=<nil> limit=3 (limit+1 because withCount)
	// aggregation pipeline:
	// [{$match map[$and:[map[status:published] map[views:map[$gt:50]]]]}]
	// [{$sort [{views -1}]}]
	// [{$limit 3}]
	// [{$lookup [{from authors} {localField author_id} {foreignField _id} {as author} {pipeline [[{$project map[_id:1 name:1]}]]}]}]
	// facet pipeline (results + exact totalCount):
	// results:
	// [{$match map[$and:[map[status:published] map[views:map[$gt:50]]]]}]
	// [{$sort [{views -1}]}]
	// [{$limit 2}]
	// [{$lookup [{from authors} {localField author_id} {foreignField _id} {as author} {pipeline [[{$project map[_id:1 name:1]}]]}]}]
	// totalCount:
	// [{$match map[$and:[map[status:published] map[views:map[$gt:50]]]]}]
	// [{$count totalCount}]
}
