package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
	"github.com/stretchr/testify/require"
)

func fixtureSchemas(t *testing.T) (article, author, profile *schema.Schema) {
	t.Helper()

	profile = &schema.Schema{
		Name: "profile",
		Filterable: map[string]schema.FieldDef{
			"bio": {Name: "bio", Type: schema.TypeString, Operators: map[query.Operator]bool{query.OpEq: true, query.OpContains: true}},
		},
		Sortable:     map[string]struct{}{"bio": {}},
		Selectable:   map[string]struct{}{"bio": {}},
		Groupable:    map[string]struct{}{},
		Aggregatable: map[string]struct{}{},
		Relations:    map[string]schema.RelationDef{},
	}

	author = &schema.Schema{
		Name: "author",
		Filterable: map[string]schema.FieldDef{
			"name": {Name: "name", Type: schema.TypeString, Operators: map[query.Operator]bool{query.OpEq: true}},
		},
		Sortable:     map[string]struct{}{"name": {}},
		Selectable:   map[string]struct{}{"name": {}},
		Groupable:    map[string]struct{}{},
		Aggregatable: map[string]struct{}{},
		Relations: map[string]schema.RelationDef{
			"profile": {Name: "profile", Target: profile, MaxDepth: 5},
		},
	}

	article = &schema.Schema{
		Name: "article",
		Filterable: map[string]schema.FieldDef{
			"title":       {Name: "title", Type: schema.TypeString, Operators: map[query.Operator]bool{query.OpEq: true, query.OpContains: true}},
			"views":       {Name: "views", Type: schema.TypeNumber, Operators: map[query.Operator]bool{query.OpEq: true, query.OpGt: true}},
			"publishedAt": {Name: "publishedAt", Type: schema.TypeDate, Operators: map[query.Operator]bool{query.OpEq: true, query.OpGt: true, query.OpNull: true, query.OpNotNull: true}},
			"active":      {Name: "active", Type: schema.TypeBool, Operators: map[query.Operator]bool{query.OpEq: true}},
		},
		Sortable:     map[string]struct{}{"title": {}, "createdAt": {}},
		Selectable:   map[string]struct{}{"title": {}, "body": {}},
		Groupable:    map[string]struct{}{"title": {}, "createdAt": {}},
		Aggregatable: map[string]struct{}{"views": {}, "createdAt": {}},
		Relations: map[string]schema.RelationDef{
			"author": {Name: "author", Target: author, MaxDepth: 5},
		},
		MaxLimit: 100,
	}

	return article, author, profile
}

func TestResolvePath(t *testing.T) {
	article, author, profile := fixtureSchemas(t)

	testCases := []struct {
		name       string
		path       query.Path
		wantSchema *schema.Schema
		wantField  string
		wantErr    error
	}{
		{
			name:    "empty path",
			path:    query.Path{},
			wantErr: ErrInvalidPath,
		},
		{
			name:       "direct field on root schema",
			path:       query.Path{"title"},
			wantSchema: article,
			wantField:  "title",
		},
		{
			name:       "one relation hop",
			path:       query.Path{"author", "name"},
			wantSchema: author,
			wantField:  "name",
		},
		{
			name:       "two relation hops",
			path:       query.Path{"author", "profile", "bio"},
			wantSchema: profile,
			wantField:  "bio",
		},
		{
			name:    "unknown top-level relation segment",
			path:    query.Path{"nonexistent", "x"},
			wantErr: ErrUnknownField,
		},
		{
			name:    "leaf-only field used mid-path is not a relation",
			path:    query.Path{"title", "x"},
			wantErr: ErrUnknownField,
		},
		{
			name:    "unknown nested relation segment",
			path:    query.Path{"author", "nonexistent", "x"},
			wantErr: ErrUnknownField,
		},
		{
			// The absolute depth ceiling is a pure length check, independent
			// of whether the schema actually has that many relation levels -
			// so this fires before any relation lookup even happens.
			name: "path exceeds absolute max depth",
			path: query.Path{
				"a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "a", "leaf",
			},
			wantErr: ErrNestingTooDeep,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotSchema, gotField, err := ResolvePath(article, tc.path)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Same(t, tc.wantSchema, gotSchema)
			require.Equal(t, tc.wantField, gotField)
		})
	}
}
