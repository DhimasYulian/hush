package hush_test

import (
	"fmt"
	"net/url"

	"github.com/DhimasYulian/hush"
)

func ExampleParse() {
	schema, err := hush.NewSchema("article").
		Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Sortable("createdAt").
		Selectable("title", "body").
		MaxLimit(100).
		Build()
	if err != nil {
		panic(err)
	}

	values := url.Values{
		"filters[title][$containsi]": {"go"},
		"sort[0]":                    {"createdAt:desc"},
		"fields[0]":                  {"title"},
		"pagination[limit]":          {"25"},
	}

	query, err := hush.Parse(values, schema)
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("filter field:", query.Filters.(hush.Condition).Path[0])
	fmt.Println("sort dir:", query.Sort[0].Direction)
	fmt.Println("fields:", query.Fields)
	fmt.Println("limit:", *query.Pagination.Limit)
	// Output:
	// filter field: title
	// sort dir: desc
	// fields: [title]
	// limit: 25
}

func ExampleNewSchema() {
	schema, err := hush.NewSchema("article").
		Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
		Filterable("views", hush.TypeNumber, hush.OpGt, hush.OpLt).
		Sortable("title", "createdAt").
		Selectable("title", "body", "createdAt").
		MaxLimit(100).
		Build()
	if err != nil {
		panic(err)
	}

	fmt.Println("name:", schema.Name())
	fmt.Println("max limit:", schema.MaxLimit())

	// Output:
	// name: article
	// max limit: 100
}
