package gorm_test

import (
	"fmt"
	"net/url"

	"github.com/DhimasYulian/hush"
	hushgorm "github.com/DhimasYulian/hush/gorm"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Article and Author are plain GORM models. The hush schema below governs which
// of their columns clients may filter, sort, select, group, aggregate, and
// preload — everything is whitelisted up front, so a validated hush.Query is
// already safe to pass to the scope.
type Article struct {
	ID       uint
	Title    string
	Views    int
	Status   string
	AuthorID uint
	Author   Author
}

type Author struct {
	ID   uint
	Name string
}

func ExampleScopes() {
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

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		fmt.Println("db:", err)
		return
	}
	if err := db.AutoMigrate(&Author{}, &Article{}); err != nil {
		fmt.Println("migrate:", err)
		return
	}

	alice := Author{Name: "Alice"}
	bob := Author{Name: "Bob"}
	db.Create(&alice)
	db.Create(&bob)
	db.Create(&[]Article{
		{Title: "Go for Beginners", Views: 100, Status: "published", AuthorID: alice.ID},
		{Title: "Advanced Go", Views: 200, Status: "published", AuthorID: bob.ID},
		{Title: "Rust vs Go", Views: 50, Status: "draft", AuthorID: bob.ID},
	})

	run := func(values url.Values, dest any) {
		q, err := hush.Parse(values, articleSchema)
		if err != nil {
			fmt.Println("parse:", err)
			return
		}
		res := db.Model(&Article{}).Scopes(hushgorm.Scopes(articleSchema, q)).Find(dest)
		if res.Error != nil {
			fmt.Println("find:", res.Error)
			return
		}
	}

	fmt.Println("withCount defaults to true, so a plain limit 2 would fetch 3 rows;")
	fmt.Println("set it to false for an exact limit")
	var filtered []Article
	run(url.Values{
		"filters[title][$containsi]": {"go"},
		"sort[0]":                    {"views:desc"},
		"fields[0]":                  {"title"},
		"pagination[limit]":          {"2"},
		"pagination[withCount]":      {"false"},
	}, &filtered)
	for _, a := range filtered {
		fmt.Println(a.Title)
	}

	fmt.Println("withCount fetches limit+1 so callers can detect a next page")
	var counted []Article
	run(url.Values{
		"sort[0]":               {"views:desc"},
		"pagination[limit]":     {"2"},
		"pagination[withCount]": {"true"},
	}, &counted)
	fmt.Printf("rows=%d hasMore=%v\n", len(counted), len(counted) > 2)

	fmt.Println("group by status with count and sum aggregations")
	type statusSummary struct {
		Status     string
		Cnt        int
		TotalViews float64
	}
	var summaries []statusSummary
	run(url.Values{
		"groupBy[0]":                       {"status"},
		"sort[0]":                          {"status:asc"},
		"aggregations[cnt][func]":          {"count"},
		"aggregations[total_views][func]":  {"sum"},
		"aggregations[total_views][field]": {"views"},
	}, &summaries)
	for _, s := range summaries {
		fmt.Printf("%s: %d rows, %v views\n", s.Status, s.Cnt, s.TotalViews)
	}

	fmt.Println("preload the author relation with a whitelisted select")
	var withAuthor []Article
	run(url.Values{
		"sort[0]":                     {"views:desc"},
		"populate[author][fields][0]": {"name"},
	}, &withAuthor)
	for _, a := range withAuthor {
		fmt.Printf("%s / %s\n", a.Title, a.Author.Name)
	}

	// Output:
	// withCount defaults to true, so a plain limit 2 would fetch 3 rows;
	// set it to false for an exact limit
	// Advanced Go
	// Go for Beginners
	// withCount fetches limit+1 so callers can detect a next page
	// rows=3 hasMore=true
	// group by status with count and sum aggregations
	// draft: 1 rows, 50 views
	// published: 2 rows, 300 views
	// preload the author relation with a whitelisted select
	// Advanced Go / Bob
	// Go for Beginners / Alice
	// Rust vs Go / Bob
}
