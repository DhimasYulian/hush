package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildFilters(t *testing.T) {
	testCases := []struct {
		name   string
		params []query.Param
		want   query.Filter
	}{
		{
			name: "no filters",
			params: []query.Param{
				{
					Path:  []string{"sort", "0"},
					Value: "name",
				},
			},
			want: nil,
		},
		{
			name: "single filter",
			params: []query.Param{
				{
					Path:  []string{"filters", "name", "$eq"},
					Value: "John",
				},
			},
			want: query.Condition{
				Path:     query.Path{"name"},
				Operator: query.OpEq,
				Value:    query.Value{"John"},
			},
		},
		{
			name: "implicit and preserves insertion order",
			params: []query.Param{
				{
					Path:  []string{"filters", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"filters", "age", "$gte"},
					Value: "18",
				},
			},
			want: query.And{
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
		{
			name: "implicit and preserves insertion order regardless of alphabetical order",
			params: []query.Param{
				{
					Path:  []string{"filters", "zebra", "$eq"},
					Value: "z",
				},
				{
					Path:  []string{"filters", "apple", "$eq"},
					Value: "a",
				},
			},
			want: query.And{
				Filters: []query.Filter{
					query.Condition{
						Path:     query.Path{"zebra"},
						Operator: query.OpEq,
						Value:    query.Value{"z"},
					},
					query.Condition{
						Path:     query.Path{"apple"},
						Operator: query.OpEq,
						Value:    query.Value{"a"},
					},
				},
			},
		},
		{
			name: "ignore non filter params",
			params: []query.Param{
				{
					Path:  []string{"fields", "0"},
					Value: "id",
				},
				{
					Path:  []string{"filters", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"pagination", "start"},
					Value: "0",
				},
			},
			want: query.Condition{
				Path:     query.Path{"name"},
				Operator: query.OpEq,
				Value:    query.Value{"John"},
			},
		},
		{
			name: "nested field path",
			params: []query.Param{
				{
					Path:  []string{"filters", "author", "profile", "name", "$eq"},
					Value: "Jane",
				},
			},
			want: query.Condition{
				Path:     query.Path{"author", "profile", "name"},
				Operator: query.OpEq,
				Value:    query.Value{"Jane"},
			},
		},
		{
			name: "explicit and",
			params: []query.Param{
				{
					Path:  []string{"filters", "$and", "0", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"filters", "$and", "1", "age", "$gte"},
					Value: "18",
				},
			},
			want: query.And{
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
		{
			name: "explicit and respects numeric index order regardless of insertion order",
			params: []query.Param{
				{
					Path:  []string{"filters", "$and", "1", "age", "$gte"},
					Value: "18",
				},
				{
					Path:  []string{"filters", "$and", "0", "name", "$eq"},
					Value: "John",
				},
			},
			want: query.And{
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
		{
			name: "or",
			params: []query.Param{
				{
					Path:  []string{"filters", "$or", "0", "status", "$eq"},
					Value: "draft",
				},
				{
					Path:  []string{"filters", "$or", "1", "status", "$eq"},
					Value: "published",
				},
			},
			want: query.Or{
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
		{
			name: "not",
			params: []query.Param{
				{
					Path:  []string{"filters", "$not", "status", "$eq"},
					Value: "archived",
				},
			},
			want: query.Not{
				Filter: query.Condition{
					Path:     query.Path{"status"},
					Operator: query.OpEq,
					Value:    query.Value{"archived"},
				},
			},
		},
		{
			name: "logical scoped under a field path",
			params: []query.Param{
				{
					Path:  []string{"filters", "author", "$or", "0", "$eq"},
					Value: "Jane",
				},
				{
					Path:  []string{"filters", "author", "$or", "1", "$eq"},
					Value: "John",
				},
			},
			want: query.Or{
				Filters: []query.Filter{
					query.Condition{
						Path:     query.Path{"author"},
						Operator: query.OpEq,
						Value:    query.Value{"Jane"},
					},
					query.Condition{
						Path:     query.Path{"author"},
						Operator: query.OpEq,
						Value:    query.Value{"John"},
					},
				},
			},
		},
		{
			name: "nested and inside or",
			params: []query.Param{
				{
					Path:  []string{"filters", "$or", "0", "$and", "0", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"filters", "$or", "0", "$and", "1", "age", "$gte"},
					Value: "18",
				},
				{
					Path:  []string{"filters", "$or", "1", "status", "$eq"},
					Value: "vip",
				},
			},
			want: query.Or{
				Filters: []query.Filter{
					query.And{
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
					query.Condition{
						Path:     query.Path{"status"},
						Operator: query.OpEq,
						Value:    query.Value{"vip"},
					},
				},
			},
		},
		{
			name: "in operator",
			params: []query.Param{
				{
					Path:  []string{"filters", "status", "$in", "1"},
					Value: "published",
				},
				{
					Path:  []string{"filters", "status", "$in", "0"},
					Value: "draft",
				},
			},
			want: query.Condition{
				Path:     query.Path{"status"},
				Operator: query.OpIn,
				Value:    query.Value{"draft", "published"},
			},
		},
		{
			name: "notIn operator preserves numeric order regardless of insertion",
			params: []query.Param{
				{
					Path:  []string{"filters", "status", "$notIn", "1"},
					Value: "archived",
				},
				{
					Path:  []string{"filters", "status", "$notIn", "0"},
					Value: "deleted",
				},
			},
			want: query.Condition{
				Path:     query.Path{"status"},
				Operator: query.OpNotIn,
				Value:    query.Value{"deleted", "archived"},
			},
		},
		{
			name: "between operator",
			params: []query.Param{
				{
					Path:  []string{"filters", "age", "$between", "0"},
					Value: "18",
				},
				{
					Path:  []string{"filters", "age", "$between", "1"},
					Value: "65",
				},
			},
			want: query.Condition{
				Path:     query.Path{"age"},
				Operator: query.OpBetween,
				Value:    query.Value{"18", "65"},
			},
		},
		{
			name: "null operator",
			params: []query.Param{
				{
					Path:  []string{"filters", "deletedAt", "$null"},
					Value: "true",
				},
			},
			want: query.Condition{
				Path:     query.Path{"deletedAt"},
				Operator: query.OpNull,
				Value:    query.Value{"true"},
			},
		},
		{
			name: "containsi operator",
			params: []query.Param{
				{
					Path:  []string{"filters", "title", "$containsi"},
					Value: "hello",
				},
			},
			want: query.Condition{
				Path:     query.Path{"title"},
				Operator: query.OpContainsi,
				Value:    query.Value{"hello"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildFiltersFromParams(tc.params)

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestBuildFilters_Deterministic(t *testing.T) {
	params := []query.Param{
		{Path: []string{"filters", "name", "$eq"}, Value: "John"},
		{Path: []string{"filters", "age", "$gte"}, Value: "18"},
		{Path: []string{"filters", "status", "$eq"}, Value: "active"},
	}

	first, err := buildFiltersFromParams(params)
	require.NoError(t, err)

	for i := 0; i < 50; i++ {
		got, err := buildFiltersFromParams(params)
		require.NoError(t, err)
		assert.Equal(t, first, got, "buildFilters must be deterministic across repeated calls")
	}
}

func TestBuildFilters_Error(t *testing.T) {
	testCases := []struct {
		name   string
		params []query.Param
	}{
		{
			name: "invalid operator",
			params: []query.Param{
				{
					Path:  []string{"filters", "name", "$foo"},
					Value: "John",
				},
			},
		},
		{
			name: "between missing value",
			params: []query.Param{
				{
					Path:  []string{"filters", "age", "$between", "0"},
					Value: "18",
				},
			},
		},
		{
			name: "between too many values",
			params: []query.Param{
				{
					Path:  []string{"filters", "age", "$between", "0"},
					Value: "18",
				},
				{
					Path:  []string{"filters", "age", "$between", "1"},
					Value: "40",
				},
				{
					Path:  []string{"filters", "age", "$between", "2"},
					Value: "65",
				},
			},
		},
		{
			name: "in with no values",
			params: []query.Param{
				{
					Path:  []string{"filters", "status", "$in"},
					Value: "",
				},
			},
		},
		{
			name: "field with more than one child",
			params: []query.Param{
				{
					Path:  []string{"filters", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"filters", "name", "$ne"},
					Value: "Jane",
				},
			},
		},
		{
			name: "not with more than one child",
			params: []query.Param{
				{
					Path:  []string{"filters", "$not", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"filters", "$not", "age", "$gte"},
					Value: "18",
				},
			},
		},
		{
			name: "and with non-numeric index",
			params: []query.Param{
				{
					Path:  []string{"filters", "$and", "first", "name", "$eq"},
					Value: "John",
				},
			},
		},
		{
			name: "empty filters object",
			params: []query.Param{
				{
					Path:  []string{"filters"},
					Value: "",
				},
			},
		},
		{
			name: "unknown operator nested under field",
			params: []query.Param{
				{
					Path:  []string{"filters", "author", "profile", "$bogus"},
					Value: "x",
				},
			},
		},
		{
			name: "and with zero children",
			params: []query.Param{
				{
					Path:  []string{"filters", "$and"},
					Value: "",
				},
			},
		},
		{
			name: "or with zero children",
			params: []query.Param{
				{
					Path:  []string{"filters", "$or"},
					Value: "",
				},
			},
		},
		{
			name: "not with zero children",
			params: []query.Param{
				{
					Path:  []string{"filters", "$not"},
					Value: "",
				},
			},
		},
		{
			name: "or with non-numeric index",
			params: []query.Param{
				{
					Path:  []string{"filters", "$or", "first", "status", "$eq"},
					Value: "draft",
				},
			},
		},
		{
			name: "in with non-numeric index",
			params: []query.Param{
				{
					Path:  []string{"filters", "status", "$in", "first"},
					Value: "draft",
				},
			},
		},
		{
			name: "between with non-numeric index",
			params: []query.Param{
				{
					Path:  []string{"filters", "age", "$between", "low"},
					Value: "18",
				},
			},
		},
		{
			name: "between with three values via mixed indices",
			params: []query.Param{
				{
					Path:  []string{"filters", "age", "$between", "0"},
					Value: "18",
				},
				{
					Path:  []string{"filters", "age", "$between", "5"},
					Value: "65",
				},
				{
					Path:  []string{"filters", "age", "$between", "9"},
					Value: "99",
				},
			},
		},
		{
			name: "notIn with zero values",
			params: []query.Param{
				{
					Path:  []string{"filters", "status", "$notIn"},
					Value: "",
				},
			},
		},
		{
			name: "and containing an invalid operator",
			params: []query.Param{
				{
					Path:  []string{"filters", "$and", "0", "name", "$foo"},
					Value: "John",
				},
			},
		},
		{
			name: "not wrapping an invalid operator",
			params: []query.Param{
				{
					Path:  []string{"filters", "$not", "name", "$foo"},
					Value: "John",
				},
			},
		},
		{
			name: "logical keyword used as a bare field value",
			params: []query.Param{
				{
					Path:  []string{"filters", "$and"},
					Value: "oops",
				},
			},
		},
		{
			name: "field node with two operator children",
			params: []query.Param{
				{
					Path:  []string{"filters", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"filters", "name", "$startsWith"},
					Value: "Jo",
				},
			},
		},
		{
			name: "and index containing two competing filters",
			params: []query.Param{
				{
					Path:  []string{"filters", "$and", "0", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"filters", "$and", "0", "age", "$gte"},
					Value: "18",
				},
			},
		},
		{
			name: "deeply nested invalid operator inside or/and",
			params: []query.Param{
				{
					Path:  []string{"filters", "$or", "0", "$and", "0", "name", "$notAnOperator"},
					Value: "John",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := buildFiltersFromParams(tc.params)
			require.Error(t, err)
		})
	}
}
