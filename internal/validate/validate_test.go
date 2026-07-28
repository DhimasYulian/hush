package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestValidate_NilGuards(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	t.Run("nil schema", func(t *testing.T) {
		err := Validate(&query.Query{}, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrMissingSchema)
	})

	t.Run("nil query is a no-op", func(t *testing.T) {
		err := Validate(nil, article)
		require.NoError(t, err)
	})
}

func TestValidate_FullyValidQuery(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	q := &query.Query{
		Filters: query.And{Filters: []query.Filter{
			cond(query.Path{"title"}, query.OpEq, "hello"),
			cond(query.Path{"author", "name"}, query.OpEq, "Jane"),
		}},
		Sort: []query.Sort{
			{Path: query.Path{"title"}, Direction: query.SortAsc},
		},
		Fields:  []query.Field{"title", "body"},
		GroupBy: []query.Field{"title"},
		Populates: []query.Populate{
			{Relation: "author", Fields: []query.Field{"name"}},
		},
		Pagination: query.Pagination{Limit: intPtr(25)},
	}

	require.NoError(t, Validate(q, article))
}

func TestValidate_PopulateAll(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	err := Validate(&query.Query{PopulateAll: true}, article)
	require.NoError(t, err)
}

func TestValidate_ZeroValueQuery(t *testing.T) {
	article, _, _ := fixtureSchemas(t)
	require.NoError(t, Validate(&query.Query{}, article))
}

func TestValidate_AccumulatesAcrossSections(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	q := &query.Query{
		Filters: cond(query.Path{"views"}, query.OpContains, "1"), // operator not allowed
		Sort: []query.Sort{
			{Path: query.Path{"nonexistent"}, Direction: query.SortAsc}, // unknown field
		},
		Fields: []query.Field{"nonexistent"}, // unknown field
		Populates: []query.Populate{
			{Relation: "nonexistent"}, // unknown relation
		},
		Pagination: query.Pagination{Limit: intPtr(1000)}, // exceeds default max
	}

	err := Validate(q, article)
	require.Error(t, err)

	require.ErrorIs(t, err, ErrOperatorNotAllowed)
	require.ErrorIs(t, err, ErrUnknownField)
	require.ErrorIs(t, err, query.ErrInvalidPopulate)
	require.ErrorIs(t, err, query.ErrInvalidPagination)
}

func TestValidate_SingleSectionFailures(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name    string
		query   *query.Query
		wantErr error
	}{
		{
			name:    "invalid filter",
			query:   &query.Query{Filters: cond(query.Path{"nonexistent"}, query.OpEq, "x")},
			wantErr: ErrUnknownField,
		},
		{
			name: "invalid sort",
			query: &query.Query{Sort: []query.Sort{
				{Path: query.Path{"views"}, Direction: query.SortAsc}, // filterable, not sortable
			}},
			wantErr: ErrUnknownField,
		},
		{
			name:    "invalid fields",
			query:   &query.Query{Fields: []query.Field{"views"}}, // filterable, not selectable
			wantErr: ErrUnknownField,
		},
		{
			name:    "invalid groupBy",
			query:   &query.Query{GroupBy: []query.Field{"views"}}, // filterable, not groupable
			wantErr: ErrUnknownGroupBy,
		},
		{
			name: "invalid populate",
			query: &query.Query{Populates: []query.Populate{
				{Relation: "author", Fields: []query.Field{"nonexistent"}},
			}},
			wantErr: ErrUnknownField,
		},
		{
			name:    "invalid pagination",
			query:   &query.Query{Pagination: query.Pagination{Limit: intPtr(9999)}},
			wantErr: query.ErrInvalidPagination,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.query, article)
			require.Error(t, err)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}
