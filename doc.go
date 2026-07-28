// Package hush parses, builds, and validates structured query strings from
// URL parameters.
//
// # Overview
//
// hush processes URL query strings through a three-stage pipeline:
//
//	Parse → Build → Validate → *Query
//
// Given a URL query string like:
//
//	filters[title][$containsi]=go&sort[0]=createdAt:desc&fields[0]=title
//
// hush produces a typed [Query] value with structured Filters, Fields, Sort,
// Pagination, and Populates — validated against a user-defined [Schema].
//
// # Entrypoints
//
// The two main entrypoints are:
//
//   - [NewSchema] — creates a [SchemaBuilder] for defining which fields,
//     operators, relations, and limits are allowed. Call [SchemaBuilder.Build]
//     to produce a [Schema].
//   - [Parse] — takes [net/url.Values] and a [Schema], runs the full
//     pipeline, and returns a validated [*Query].
//
// # Schema Definition
//
// Use [NewSchema] and the builder methods to declare what is allowed:
//
//	schema, err := hush.NewSchema("article").
//	    Filterable("title", hush.TypeString, hush.OpEq, hush.OpContainsi).
//	    Filterable("views", hush.TypeNumber, hush.OpGt).
//	    Sortable("title", "createdAt").
//	    Selectable("title", "body", "createdAt").
//	    Relation("author", authorSchema, 5).
//	    MaxLimit(100).
//	    Build()
//
// # Query String Syntax
//
// The expected query string format uses bracket notation:
//
//	filters[field][$operator]=value          — filter conditions
//	filters[$and][0][field][$operator]=value — AND combinators
//	filters[$or][0][field][$operator]=value  — OR combinators
//	filters[$not][field][$operator]=value    — NOT combinator
//	fields[0]=fieldName                      — field selection
//	sort[0]=fieldName:desc                   — sorting (asc or desc)
//	groupBy[0]=fieldName                     — group by field
//	aggregations[name][func]=count           — aggregate function (count, sum, avg)
//	aggregations[name][field]=fieldName      — field to aggregate on
//	pagination[start]=0                      — offset
//	pagination[limit]=25                     — page size
//	pagination[withCount]=true               — include total row count (default true)
//	populate[0]=relationName                 — include relations
//	populate[relation][fields][0]=name       — nested field selection
//	populate=*                               — populate all relations
//
// # Error Handling
//
// Parse and build errors fail fast (first error). Validation errors accumulate
// across all query sections and are returned together via [errors.Join].
//
// All errors can be inspected with [errors.Is] against the exported sentinel
// values (e.g. [ErrUnknownField], [ErrOperatorNotAllowed]). Structured error
// context is available via [errors.As] with the [*Error] type.
package hush
