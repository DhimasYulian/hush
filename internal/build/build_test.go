package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name   string
		params []query.Param
		want   *query.Query
	}{
		{
			name:   "empty params",
			params: []query.Param{},
			want:   &query.Query{Pagination: query.Pagination{WithCount: boolPtr(true)}},
		},
		{
			name: "one param per builder",
			params: []query.Param{
				{Path: []string{"fields", "0"}, Value: "title"},
				{Path: []string{"filters", "name", "$eq"}, Value: "John"},
				{Path: []string{"sort", "0"}, Value: "name"},
				{Path: []string{"pagination", "start"}, Value: "0"},
				{Path: []string{"pagination", "limit"}, Value: "10"},
				{Path: []string{"populate", "0"}, Value: "author"},
			},
			want: &query.Query{
				Fields: []query.Field{"title"},
				Filters: query.Condition{
					Path:     query.Path{"name"},
					Operator: query.OpEq,
					Value:    query.Value{"John"},
				},
				Sort: []query.Sort{
					{Path: query.Path{"name"}, Direction: query.SortAsc},
				},
				Pagination: query.Pagination{
					Start:     intPtr(0),
					Limit:     intPtr(10),
					WithCount: boolPtr(true),
				},
				Populates: []query.Populate{
					{Relation: "author"},
				},
			},
		},
		{
			name: "shorthand fields and populate",
			params: []query.Param{
				{Path: []string{"fields"}, Value: "title"},
				{Path: []string{"populate"}, Value: "author"},
			},
			want: &query.Query{
				Fields:     []query.Field{"title"},
				Populates:  []query.Populate{{Relation: "author"}},
				Pagination: query.Pagination{WithCount: boolPtr(true)},
			},
		},
		{
			name: "pagination alone",
			params: []query.Param{
				{Path: []string{"pagination", "start"}, Value: "5"},
			},
			want: &query.Query{
				Pagination: query.Pagination{Start: intPtr(5), WithCount: boolPtr(true)},
			},
		},
		{
			name: "ignores unrecognized top-level params",
			params: []query.Param{
				{Path: []string{"foo"}, Value: "bar"},
				{Path: []string{"filters", "name", "$eq"}, Value: "John"},
			},
			want: &query.Query{
				Filters: query.Condition{
					Path:     query.Path{"name"},
					Operator: query.OpEq,
					Value:    query.Value{"John"},
				},
				Pagination: query.Pagination{WithCount: boolPtr(true)},
			},
		},
		{
			name: "nested populate with its own fields and filters",
			params: []query.Param{
				{Path: []string{"populate", "author", "fields", "0"}, Value: "name"},
				{Path: []string{"populate", "author", "filters", "verified", "$eq"}, Value: "true"},
			},
			want: &query.Query{
				Populates: []query.Populate{
					{
						Relation: "author",
						Fields:   []query.Field{"name"},
						Filters: query.Condition{
							Path:     query.Path{"verified"},
							Operator: query.OpEq,
							Value:    query.Value{"true"},
						},
					},
				},
				Pagination: query.Pagination{WithCount: boolPtr(true)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildQuery(tc.params)

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBuild_Error(t *testing.T) {
	testCases := []struct {
		name   string
		params []query.Param
	}{
		{
			name: "invalid filter operator",
			params: []query.Param{
				{Path: []string{"filters", "name", "$bogus"}, Value: "John"},
			},
		},
		{
			name: "invalid sort direction",
			params: []query.Param{
				{Path: []string{"sort", "0"}, Value: "name:sideways"},
			},
		},
		{
			name: "unknown pagination key",
			params: []query.Param{
				{Path: []string{"pagination", "foo"}, Value: "1"},
			},
		},
		{
			name: "negative pagination value",
			params: []query.Param{
				{Path: []string{"pagination", "start"}, Value: "-1"},
			},
		},
		{
			name: "mixing fields shorthand and indexed syntax",
			params: []query.Param{
				{Path: []string{"fields"}, Value: "title"},
				{Path: []string{"fields", "0"}, Value: "name"},
			},
		},
		{
			name: "mixing populate shorthand and relation-keyed syntax",
			params: []query.Param{
				{Path: []string{"populate"}, Value: "author"},
				{Path: []string{"populate", "comments", "fields", "0"}, Value: "name"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := BuildQuery(tc.params)
			require.Error(t, err)
		})
	}
}
