package hush

import (
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFacade(t *testing.T) {
	root, err := NewSchema("article").
		Filterable("title", TypeString, OpContainsi).
		Sortable("createdAt").
		Selectable("title").
		MaxLimit(50).
		Build()
	require.NoError(t, err)

	got, err := Parse(url.Values{
		"filters[title][$containsi]": {"go"},
		"fields[0]":                  {"title"},
		"sort[0]":                    {"createdAt:desc"},
		"pagination[limit]":          {"25"},
	}, root)
	require.NoError(t, err)

	require.Equal(t, &Query{
		Filters: Condition{
			Path:      Path{"title"},
			Operator:  OpContainsi,
			Value:     Value{"go"},
			FieldType: TypeString,
			Values:    []any{"go"},
		},
		Fields: []Field{"title"},
		Sort: []Sort{
			{Path: Path{"createdAt"}, Direction: SortDesc},
		},
		Pagination: Pagination{Limit: intPtr(25), WithCount: boolPtr(true)},
	}, got)
}

func TestParseFacade_ValidationError(t *testing.T) {
	root, err := NewSchema("article").
		Filterable("title", TypeString, OpEq).
		Build()
	require.NoError(t, err)

	_, err = Parse(url.Values{
		"filters[title][$containsi]": {"go"},
	}, root)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrOperatorNotAllowed)

	var queryErr *Error
	require.True(t, errors.As(err, &queryErr))
	require.Equal(t, ErrOperatorNotAllowed, queryErr.Kind)
	require.Equal(t, "title", queryErr.Field)
	require.Equal(t, OpContainsi, queryErr.Operator)
}

func TestParseFacade_SyntaxErrorContext(t *testing.T) {
	root, err := NewSchema("article").Build()
	require.NoError(t, err)

	_, err = Parse(url.Values{
		"filters[title": {"go"},
	}, root)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidSyntax)

	var queryErr *Error
	require.True(t, errors.As(err, &queryErr))
	require.Equal(t, ErrInvalidSyntax, queryErr.Kind)
	require.Equal(t, "missing closing bracket", queryErr.Message)
}

func TestParseFacade_WithGroupBy(t *testing.T) {
	root, err := NewSchema("article").
		Filterable("title", TypeString, OpContainsi).
		Sortable("createdAt").
		Selectable("title", "body").
		Groupable("title", "createdAt").
		MaxLimit(50).
		Build()
	require.NoError(t, err)

	got, err := Parse(url.Values{
		"filters[title][$containsi]": {"go"},
		"fields[0]":                  {"title"},
		"sort[0]":                    {"createdAt:desc"},
		"groupBy[0]":                 {"title"},
		"pagination[limit]":          {"25"},
	}, root)
	require.NoError(t, err)

	require.Equal(t, &Query{
		Filters: Condition{
			Path:      Path{"title"},
			Operator:  OpContainsi,
			Value:     Value{"go"},
			FieldType: TypeString,
			Values:    []any{"go"},
		},
		Fields:  []Field{"title"},
		Sort:    []Sort{{Path: Path{"createdAt"}, Direction: SortDesc}},
		GroupBy: []Field{"title"},
		Pagination: Pagination{
			Limit:     intPtr(25),
			WithCount: boolPtr(true),
		},
	}, got)
}

func TestParseFacade_GroupByValidation(t *testing.T) {
	root, err := NewSchema("article").
		Filterable("title", TypeString, OpContainsi).
		Groupable("title").
		Build()
	require.NoError(t, err)

	_, err = Parse(url.Values{
		"groupBy[0]": {"views"}, // views is not groupable
	}, root)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrUnknownGroupBy)
}

func TestParseFacade_WithAggregations(t *testing.T) {
	root, err := NewSchema("employee").
		Filterable("country", TypeString, OpEq).
		Sortable("country").
		Selectable("country").
		Groupable("country").
		Aggregatable("salary", "age").
		Build()
	require.NoError(t, err)

	got, err := Parse(url.Values{
		"groupBy[0]":                       {"country"},
		"aggregations[total][func]":        {"count"},
		"aggregations[totalSalary][func]":  {"sum"},
		"aggregations[totalSalary][field]": {"salary"},
		"aggregations[avgAge][func]":       {"avg"},
		"aggregations[avgAge][field]":      {"age"},
	}, root)
	require.NoError(t, err)

	require.Equal(t, &Query{
		GroupBy: []Field{"country"},
		Aggregations: []Aggregation{
			{Alias: "avgAge", Func: "avg", Field: "age"},
			{Alias: "totalSalary", Func: "sum", Field: "salary"},
			{Alias: "total", Func: "count", Field: "*"},
		},
		Pagination: Pagination{WithCount: boolPtr(true)},
	}, got)
}

func TestParseFacade_WithPopulate(t *testing.T) {
	author, err := NewSchema("author").
		Filterable("name", TypeString, OpEq, OpContainsi).
		Sortable("name").
		Selectable("name", "email").
		Build()
	require.NoError(t, err)

	article, err := NewSchema("article").
		Filterable("title", TypeString, OpContainsi).
		Sortable("createdAt").
		Selectable("title", "body").
		Relation("author", author, 3).
		MaxLimit(50).
		Build()
	require.NoError(t, err)

	got, err := Parse(url.Values{
		"filters[title][$containsi]":           {"go"},
		"fields[0]":                            {"title"},
		"sort[0]":                              {"createdAt:desc"},
		"pagination[limit]":                    {"25"},
		"populate[author][fields][0]":          {"name"},
		"populate[author][sort][0]":            {"name:asc"},
		"populate[author][filters][name][$eq]": {"Alice"},
	}, article)
	require.NoError(t, err)

	want := &Query{
		Filters: Condition{
			Path:      Path{"title"},
			Operator:  OpContainsi,
			Value:     Value{"go"},
			FieldType: TypeString,
			Values:    []any{"go"},
		},
		Fields: []Field{"title"},
		Sort:   []Sort{{Path: Path{"createdAt"}, Direction: SortDesc}},
		Pagination: Pagination{
			Limit:     intPtr(25),
			WithCount: boolPtr(true),
		},
		Populates: []Populate{
			{
				Relation: "author",
				Fields:   []Field{"name"},
				Sorts:    []Sort{{Path: Path{"name"}, Direction: SortAsc}},
				Filters: Condition{
					Path:      Path{"name"},
					Operator:  OpEq,
					Value:     Value{"Alice"},
					FieldType: TypeString,
					Values:    []any{"Alice"},
				},
			},
		},
	}

	require.Equal(t, want, got)
}

func TestParseFacade_PopulateValidation(t *testing.T) {
	author, err := NewSchema("author").
		Filterable("name", TypeString, OpEq).
		Build()
	require.NoError(t, err)

	article, err := NewSchema("article").
		Relation("author", author, 3).
		Build()
	require.NoError(t, err)

	_, err = Parse(url.Values{
		"populate[0]": {"unknown"},
	}, article)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidPopulate)
}

func TestParseFacade_AggregationValidation(t *testing.T) {
	root, err := NewSchema("employee").
		Filterable("country", TypeString, OpEq).
		Aggregatable("salary").
		Build()
	require.NoError(t, err)

	_, err = Parse(url.Values{
		"aggregations[x][func]":  {"sum"},
		"aggregations[x][field]": {"title"}, // title is not aggregatable
	}, root)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidAggregation)
}
