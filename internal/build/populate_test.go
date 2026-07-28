package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPopulate(t *testing.T) {
	testCases := []struct {
		name            string
		params          []query.Param
		want            []query.Populate
		wantPopulateAll bool
	}{
		{
			name:   "no populate",
			params: []query.Param{},
			want:   nil,
		},
		{
			name: "shorthand",
			params: []query.Param{
				{Path: []string{"populate"}, Value: "author"},
			},
			want: []query.Populate{
				{Relation: "author"},
			},
		},
		{
			name: "wildcard shorthand",
			params: []query.Param{
				{Path: []string{"populate"}, Value: "*"},
			},
			want:            nil,
			wantPopulateAll: true,
		},
		{
			name: "wildcard indexed",
			params: []query.Param{
				{Path: []string{"populate", "0"}, Value: "*"},
			},
			want:            nil,
			wantPopulateAll: true,
		},
		{
			name: "indexed",
			params: []query.Param{
				{Path: []string{"populate", "0"}, Value: "author"},
				{Path: []string{"populate", "1"}, Value: "comments"},
			},
			want: []query.Populate{
				{Relation: "author"},
				{Relation: "comments"},
			},
		},
		{
			name: "indexed respects numeric order regardless of insertion order",
			params: []query.Param{
				{Path: []string{"populate", "1"}, Value: "comments"},
				{Path: []string{"populate", "0"}, Value: "author"},
			},
			want: []query.Populate{
				{Relation: "author"},
				{Relation: "comments"},
			},
		},
		{
			name: "relation with fields",
			params: []query.Param{
				{Path: []string{"populate", "author", "fields", "0"}, Value: "name"},
				{Path: []string{"populate", "author", "fields", "1"}, Value: "email"},
			},
			want: []query.Populate{
				{Relation: "author", Fields: []query.Field{"name", "email"}},
			},
		},
		{
			name: "asterisk nested under a relation option is not the wildcard",
			params: []query.Param{
				{Path: []string{"populate", "author", "fields", "0"}, Value: "*"},
			},
			want: []query.Populate{
				{Relation: "author", Fields: []query.Field{"*"}},
			},
		},
		{
			name: "relation with sort",
			params: []query.Param{
				{Path: []string{"populate", "author", "sort", "0"}, Value: "name:desc"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Sorts: []query.Sort{
						{Path: query.Path{"name"}, Direction: query.SortDesc},
					},
				},
			},
		},
		{
			name: "relation with fields and sort",
			params: []query.Param{
				{Path: []string{"populate", "author", "fields", "0"}, Value: "name"},
				{Path: []string{"populate", "author", "sort", "0"}, Value: "name"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Fields:   []query.Field{"name"},
					Sorts: []query.Sort{
						{Path: query.Path{"name"}, Direction: query.SortAsc},
					},
				},
			},
		},
		{
			name: "relation with a single filter",
			params: []query.Param{
				{Path: []string{"populate", "author", "filters", "name", "$eq"}, Value: "John"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Filters: query.Condition{
						Path:     query.Path{"name"},
						Operator: query.OpEq,
						Value:    query.Value{"John"},
					},
				},
			},
		},
		{
			name: "relation with an implicit and filter",
			params: []query.Param{
				{Path: []string{"populate", "author", "filters", "name", "$eq"}, Value: "John"},
				{Path: []string{"populate", "author", "filters", "age", "$gte"}, Value: "18"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Filters: query.And{
						Filters: []query.Filter{
							query.Condition{
								Path:     query.Path{"name"},
								Operator: query.OpEq,
								Value:    query.Value{"John"},
							},
							query.Condition{
								Path:     query.Path{"age"},
								Operator: query.OpGte,
								Value:    query.Value{"18"},
							},
						},
					},
				},
			},
		},
		{
			name: "relation with fields, sort, and filters together",
			params: []query.Param{
				{Path: []string{"populate", "author", "fields", "0"}, Value: "name"},
				{Path: []string{"populate", "author", "sort", "0"}, Value: "name"},
				{Path: []string{"populate", "author", "filters", "status", "$eq"}, Value: "active"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Fields:   []query.Field{"name"},
					Sorts: []query.Sort{
						{Path: query.Path{"name"}, Direction: query.SortAsc},
					},
					Filters: query.Condition{
						Path:     query.Path{"status"},
						Operator: query.OpEq,
						Value:    query.Value{"active"},
					},
				},
			},
		},
		{
			name: "nested relation with a filter on the inner relation",
			params: []query.Param{
				{
					Path:  []string{"populate", "author", "populate", "profile", "filters", "verified", "$eq"},
					Value: "true",
				},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Populates: []query.Populate{
						{
							Relation: "profile",
							Filters: query.Condition{
								Path:     query.Path{"verified"},
								Operator: query.OpEq,
								Value:    query.Value{"true"},
							},
						},
					},
				},
			},
		},
		{
			name: "logical filter scoped to a relation",
			params: []query.Param{
				{Path: []string{"populate", "author", "filters", "$or", "0", "status", "$eq"}, Value: "draft"},
				{Path: []string{"populate", "author", "filters", "$or", "1", "status", "$eq"}, Value: "published"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Filters: query.Or{
						Filters: []query.Filter{
							query.Condition{
								Path:     query.Path{"status"},
								Operator: query.OpEq,
								Value:    query.Value{"draft"},
							},
							query.Condition{
								Path:     query.Path{"status"},
								Operator: query.OpEq,
								Value:    query.Value{"published"},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple relations preserve insertion order regardless of alphabetical order",
			params: []query.Param{
				{Path: []string{"populate", "zebra", "fields", "0"}, Value: "id"},
				{Path: []string{"populate", "apple", "fields", "0"}, Value: "id"},
			},
			want: []query.Populate{
				{Relation: "zebra", Fields: []query.Field{"id"}},
				{Relation: "apple", Fields: []query.Field{"id"}},
			},
		},
		{
			name: "nested relation via populate",
			params: []query.Param{
				{Path: []string{"populate", "author", "populate", "profile", "fields", "0"}, Value: "bio"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Populates: []query.Populate{
						{Relation: "profile", Fields: []query.Field{"bio"}},
					},
				},
			},
		},
		{
			name: "deeply nested relation",
			params: []query.Param{
				{
					Path:  []string{"populate", "author", "populate", "profile", "populate", "avatar", "fields", "0"},
					Value: "url",
				},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Populates: []query.Populate{
						{
							Relation: "profile",
							Populates: []query.Populate{
								{Relation: "avatar", Fields: []query.Field{"url"}},
							},
						},
					},
				},
			},
		},
		{
			name: "sibling relations each with their own nested populate",
			params: []query.Param{
				{Path: []string{"populate", "author", "populate", "profile", "fields", "0"}, Value: "bio"},
				{Path: []string{"populate", "comments", "fields", "0"}, Value: "body"},
			},
			want: []query.Populate{
				{
					Relation: "author",
					Populates: []query.Populate{
						{Relation: "profile", Fields: []query.Field{"bio"}},
					},
				},
				{Relation: "comments", Fields: []query.Field{"body"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, populateAll, err := BuildPopulate(tc.params)

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.wantPopulateAll, populateAll)
		})
	}
}

func TestBuildPopulate_Deterministic(t *testing.T) {
	params := []query.Param{
		{Path: []string{"populate", "zebra", "fields", "0"}, Value: "id"},
		{Path: []string{"populate", "apple", "fields", "0"}, Value: "id"},
		{Path: []string{"populate", "mango", "fields", "0"}, Value: "id"},
	}

	first, firstAll, err := BuildPopulate(params)
	require.NoError(t, err)

	for i := 0; i < 50; i++ {
		got, gotAll, err := BuildPopulate(params)
		require.NoError(t, err)
		assert.Equal(t, first, got, "buildPopulate must be deterministic across repeated calls")
		assert.Equal(t, firstAll, gotAll)
	}
}

func TestBuildPopulate_Error(t *testing.T) {
	testCases := []struct {
		name   string
		params []query.Param
	}{
		{
			name: "mixing shorthand and relation-keyed",
			params: []query.Param{
				{Path: []string{"populate"}, Value: "author"},
				{Path: []string{"populate", "comments", "fields", "0"}, Value: "name"},
			},
		},
		{
			name: "multiple bare populate values",
			params: []query.Param{
				{Path: []string{"populate"}, Value: "author"},
				{Path: []string{"populate"}, Value: "comments"},
			},
		},
		{
			name: "empty shorthand value",
			params: []query.Param{
				{Path: []string{"populate"}, Value: ""},
			},
		},
		{
			name: "empty indexed value",
			params: []query.Param{
				{Path: []string{"populate", "0"}, Value: ""},
			},
		},
		{
			name: "relation path too short",
			params: []query.Param{
				{Path: []string{"populate", "author"}, Value: ""},
			},
		},
		{
			name: "empty relation segment",
			params: []query.Param{
				{Path: []string{"populate", "", "fields", "0"}, Value: "name"},
			},
		},
		{
			name: "unknown option segment",
			params: []query.Param{
				{Path: []string{"populate", "author", "foo"}, Value: "x"},
			},
		},
		{
			name: "invalid field index within relation",
			params: []query.Param{
				{Path: []string{"populate", "author", "fields", "x"}, Value: "name"},
			},
		},
		{
			name: "invalid sort direction within relation",
			params: []query.Param{
				{Path: []string{"populate", "author", "sort", "0"}, Value: "name:sideways"},
			},
		},
		{
			name: "invalid filter operator within relation",
			params: []query.Param{
				{Path: []string{"populate", "author", "filters", "name", "$foo"}, Value: "John"},
			},
		},
		{
			name: "filter field with more than one operator within relation",
			params: []query.Param{
				{Path: []string{"populate", "author", "filters", "name", "$eq"}, Value: "John"},
				{Path: []string{"populate", "author", "filters", "name", "$ne"}, Value: "Jane"},
			},
		},
		{
			name: "between with wrong arity within relation",
			params: []query.Param{
				{Path: []string{"populate", "author", "filters", "age", "$between", "0"}, Value: "18"},
			},
		},
		{
			name: "wildcard mixed with another populate value",
			params: []query.Param{
				{Path: []string{"populate", "0"}, Value: "*"},
				{Path: []string{"populate", "1"}, Value: "author"},
			},
		},
		{
			name: "wildcard mixed with relation-keyed syntax",
			params: []query.Param{
				{Path: []string{"populate"}, Value: "*"},
				{Path: []string{"populate", "comments", "fields", "0"}, Value: "name"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := BuildPopulate(tc.params)
			require.Error(t, err)
		})
	}
}
