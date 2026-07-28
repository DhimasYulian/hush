package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestValidateSort(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name    string
		sorts   []query.Sort
		wantErr error
	}{
		{
			name:  "empty sorts is a no-op",
			sorts: nil,
		},
		{
			name: "valid direct field",
			sorts: []query.Sort{
				{Path: query.Path{"title"}, Direction: query.SortAsc},
			},
		},
		{
			// createdAt is Sortable but was never declared Filterable,
			// confirming the two capabilities are checked independently.
			name: "sortable field that isn't filterable",
			sorts: []query.Sort{
				{Path: query.Path{"createdAt"}, Direction: query.SortDesc},
			},
		},
		{
			name: "unknown field",
			sorts: []query.Sort{
				{Path: query.Path{"nonexistent"}, Direction: query.SortAsc},
			},
			wantErr: ErrUnknownField,
		},
		{
			// views is Filterable but was never declared Sortable.
			name: "filterable field that isn't sortable",
			sorts: []query.Sort{
				{Path: query.Path{"views"}, Direction: query.SortAsc},
			},
			wantErr: ErrUnknownField,
		},
		{
			name: "sort through one relation",
			sorts: []query.Sort{
				{Path: query.Path{"author", "name"}, Direction: query.SortAsc},
			},
		},
		{
			name: "sort through two relations",
			sorts: []query.Sort{
				{Path: query.Path{"author", "profile", "bio"}, Direction: query.SortAsc},
			},
		},
		{
			name: "unknown relation in path",
			sorts: []query.Sort{
				{Path: query.Path{"nonexistent", "name"}, Direction: query.SortAsc},
			},
			wantErr: ErrUnknownField,
		},
		{
			name: "path exceeds absolute max depth",
			sorts: []query.Sort{
				{Path: query.Path{
					"a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "leaf",
				}, Direction: query.SortAsc},
			},
			wantErr: ErrNestingTooDeep,
		},
		{
			name: "second invalid sort in a list is caught",
			sorts: []query.Sort{
				{Path: query.Path{"title"}, Direction: query.SortAsc},
				{Path: query.Path{"nonexistent"}, Direction: query.SortAsc},
			},
			wantErr: ErrUnknownField,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSort(tc.sorts, article)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
