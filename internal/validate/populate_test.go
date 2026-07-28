package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
	"github.com/stretchr/testify/require"
)

func TestValidatePopulate(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name    string
		pops    []query.Populate
		wantErr error
	}{
		{
			name: "empty populate is a no-op",
			pops: nil,
		},
		{
			name: "valid single-level populate",
			pops: []query.Populate{
				{Relation: "author"},
			},
		},
		{
			name: "unknown relation",
			pops: []query.Populate{
				{Relation: "nonexistent"},
			},
			wantErr: query.ErrInvalidPopulate,
		},
		{
			name: "valid fields on the related schema",
			pops: []query.Populate{
				{Relation: "author", Fields: []query.Field{"name"}},
			},
		},
		{
			name: "invalid fields on the related schema",
			pops: []query.Populate{
				{Relation: "author", Fields: []query.Field{"nonexistent"}},
			},
			wantErr: ErrUnknownField,
		},
		{
			name: "valid sort on the related schema",
			pops: []query.Populate{
				{Relation: "author", Sorts: []query.Sort{
					{Path: query.Path{"name"}, Direction: query.SortAsc},
				}},
			},
		},
		{
			name: "invalid sort on the related schema",
			pops: []query.Populate{
				{Relation: "author", Sorts: []query.Sort{
					{Path: query.Path{"nonexistent"}, Direction: query.SortAsc},
				}},
			},
			wantErr: ErrUnknownField,
		},
		{
			name: "valid filter on the related schema",
			pops: []query.Populate{
				{Relation: "author", Filters: cond(query.Path{"name"}, query.OpEq, "Jane")},
			},
		},
		{
			name: "invalid filter on the related schema propagates",
			pops: []query.Populate{
				{Relation: "author", Filters: cond(query.Path{"name"}, query.OpContains, "Jane")},
			},
			wantErr: ErrOperatorNotAllowed,
		},
		{
			name: "valid two-level nested populate",
			pops: []query.Populate{
				{
					Relation: "author",
					Populates: []query.Populate{
						{Relation: "profile", Fields: []query.Field{"bio"}},
					},
				},
			},
		},
		{
			name: "invalid field two levels deep is still caught",
			pops: []query.Populate{
				{
					Relation: "author",
					Populates: []query.Populate{
						{Relation: "profile", Fields: []query.Field{"nonexistent"}},
					},
				},
			},
			wantErr: ErrUnknownField,
		},
		{
			name: "unknown relation two levels deep",
			pops: []query.Populate{
				{
					Relation: "author",
					Populates: []query.Populate{
						{Relation: "nonexistent"},
					},
				},
			},
			wantErr: query.ErrInvalidPopulate,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePopulate(tc.pops, article)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TestValidatePopulate_MaxDepth uses its own tight schema chain (separate
// from the shared fixtureSchemas, whose relations are generously capped)
// to exercise RelationDef.MaxDepth precisely at and past its boundary.
func TestValidatePopulate_MaxDepth(t *testing.T) {
	// top --(mid, maxDepth 1)--> mid --(leaf, maxDepth 2)--> leaf
	leaf := &schema.Schema{
		Name:       "leaf",
		Filterable: make(map[string]schema.FieldDef),
		Sortable:   make(map[string]struct{}),
		Selectable: map[string]struct{}{"id": {}},
		Relations:  make(map[string]schema.RelationDef),
	}

	mid := &schema.Schema{
		Name:       "mid",
		Filterable: make(map[string]schema.FieldDef),
		Sortable:   make(map[string]struct{}),
		Selectable: make(map[string]struct{}),
		Relations: map[string]schema.RelationDef{
			"leaf": {Name: "leaf", Target: leaf, MaxDepth: 2},
		},
	}

	top := &schema.Schema{
		Name:       "top",
		Filterable: make(map[string]schema.FieldDef),
		Sortable:   make(map[string]struct{}),
		Selectable: make(map[string]struct{}),
		Relations: map[string]schema.RelationDef{
			"mid": {Name: "mid", Target: mid, MaxDepth: 1},
		},
	}

	testCases := []struct {
		name    string
		pops    []query.Populate
		wantErr error
	}{
		{
			name: "mid at depth 0 is within its own maxDepth of 1",
			pops: []query.Populate{
				{Relation: "mid"},
			},
		},
		{
			name: "leaf at depth 1 is within its own maxDepth of 2",
			pops: []query.Populate{
				{Relation: "mid", Populates: []query.Populate{
					{Relation: "leaf"},
				}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePopulate(tc.pops, top)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TestValidatePopulate_MaxDepthExceeded builds a chain exactly one
// relation too deep for its declared maxDepth, confirming the boundary
// is enforced strictly (depth >= MaxDepth, not >).
func TestValidatePopulate_MaxDepthExceeded(t *testing.T) {
	// top --(a, maxDepth 1)--> mid --(b, maxDepth 1)--> leaf
	leaf := &schema.Schema{
		Name:       "leaf",
		Filterable: make(map[string]schema.FieldDef),
		Sortable:   make(map[string]struct{}),
		Selectable: make(map[string]struct{}),
		Relations:  make(map[string]schema.RelationDef),
	}

	mid := &schema.Schema{
		Name:       "mid",
		Filterable: make(map[string]schema.FieldDef),
		Sortable:   make(map[string]struct{}),
		Selectable: make(map[string]struct{}),
		Relations: map[string]schema.RelationDef{
			"b": {Name: "b", Target: leaf, MaxDepth: 1},
		},
	}

	top := &schema.Schema{
		Name:       "top",
		Filterable: make(map[string]schema.FieldDef),
		Sortable:   make(map[string]struct{}),
		Selectable: make(map[string]struct{}),
		Relations: map[string]schema.RelationDef{
			"a": {Name: "a", Target: mid, MaxDepth: 1},
		},
	}

	// "a" used at depth 0: 0 >= 1 is false, allowed.
	err := ValidatePopulate([]query.Populate{{Relation: "a"}}, top)
	require.NoError(t, err)

	// "b" would be used at depth 1: 1 >= its own maxDepth of 1, rejected.
	err = ValidatePopulate([]query.Populate{
		{Relation: "a", Populates: []query.Populate{
			{Relation: "b"},
		}},
	}, top)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrNestingTooDeep)
}
