package hush

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuilder_Valid(t *testing.T) {
	author, err := NewSchema("author").
		Filterable("name", TypeString, OpEq, OpContains).
		Build()
	require.NoError(t, err)

	s, err := NewSchema("article").
		Filterable("title", TypeString, OpEq, OpContains).
		Filterable("views", TypeNumber, OpGt, OpGte).
		Sortable("title", "createdAt").
		Selectable("title", "body").
		Groupable("title", "createdAt").
		Aggregatable("title", "createdAt").
		Relation("author", author, 2).
		MaxLimit(50).
		Build()
	require.NoError(t, err)

	def, ok := s.Filterable("title")
	require.True(t, ok)
	require.Equal(t, TypeString, def.Type)
	require.True(t, def.Operators[OpEq])
	require.True(t, def.Operators[OpContains])
	require.False(t, def.Operators[OpGt])

	viewsDef, ok := s.Filterable("views")
	require.True(t, ok)
	require.Equal(t, TypeNumber, viewsDef.Type)

	require.True(t, s.Sortable("createdAt"))
	require.False(t, s.Sortable("views"))

	require.True(t, s.Selectable("body"))
	require.False(t, s.Selectable("views"))

	require.True(t, s.Groupable("title"))
	require.True(t, s.Groupable("createdAt"))
	require.False(t, s.Groupable("views"))

	require.True(t, s.Aggregatable("title"))
	require.True(t, s.Aggregatable("createdAt"))
	require.False(t, s.Aggregatable("body"))

	rel, ok := s.Relation("author")
	require.True(t, ok)
	require.Equal(t, 2, rel.MaxDepth)
	require.Same(t, author.inner, rel.Target)

	require.Equal(t, 50, s.MaxLimit())
}

func TestBuilder_Relation_Hidden(t *testing.T) {
	author, err := NewSchema("author").Build()
	require.NoError(t, err)

	s, err := NewSchema("article").
		Relation("author", author, 2).
		Relation("auditLog", author, 1, Hidden()).
		Build()
	require.NoError(t, err)

	visible, ok := s.Relation("author")
	require.True(t, ok)
	require.False(t, visible.HiddenFromWildcard)

	hidden, ok := s.Relation("auditLog")
	require.True(t, ok)
	require.True(t, hidden.HiddenFromWildcard)
}

func TestBuilder_DefaultMaxLimit(t *testing.T) {
	s, err := NewSchema("article").Build()
	require.NoError(t, err)
	require.Equal(t, DefaultMaxLimit, s.MaxLimit())
}

func TestBuilder_Errors(t *testing.T) {
	author, err := NewSchema("author").Filterable("name", TypeString, OpEq).Build()
	require.NoError(t, err)

	testCases := []struct {
		name    string
		build   func() (*Schema, error)
		wantErr error
	}{
		{
			name:    "empty schema name",
			build:   func() (*Schema, error) { return NewSchema("").Build() },
			wantErr: ErrEmptyName,
		},
		{
			name: "empty filterable field name",
			build: func() (*Schema, error) {
				return NewSchema("article").Filterable("", TypeString, OpEq).Build()
			},
			wantErr: ErrEmptyName,
		},
		{
			name: "filterable with no operators",
			build: func() (*Schema, error) {
				return NewSchema("article").Filterable("title", TypeString).Build()
			},
			wantErr: ErrNoOperators,
		},
		{
			name: "filterable with unknown operator",
			build: func() (*Schema, error) {
				return NewSchema("article").Filterable("title", TypeString, Operator("$bogus")).Build()
			},
			wantErr: ErrUnknownOperator,
		},
		{
			name: "filterable with unknown field type",
			build: func() (*Schema, error) {
				return NewSchema("article").Filterable("title", FieldType("bogus"), OpEq).Build()
			},
			wantErr: ErrUnknownFieldType,
		},
		{
			name: "duplicate filterable field",
			build: func() (*Schema, error) {
				return NewSchema("article").
					Filterable("title", TypeString, OpEq).
					Filterable("title", TypeString, OpContains).
					Build()
			},
			wantErr: ErrDuplicateField,
		},
		{
			name: "empty sortable name",
			build: func() (*Schema, error) {
				return NewSchema("article").Sortable("").Build()
			},
			wantErr: ErrEmptyName,
		},
		{
			name: "empty selectable name",
			build: func() (*Schema, error) {
				return NewSchema("article").Selectable("").Build()
			},
			wantErr: ErrEmptyName,
		},
		{
			name: "empty groupable name",
			build: func() (*Schema, error) {
				return NewSchema("article").Groupable("").Build()
			},
			wantErr: ErrEmptyName,
		},
		{
			name: "empty aggregatable name",
			build: func() (*Schema, error) {
				return NewSchema("article").Aggregatable("").Build()
			},
			wantErr: ErrEmptyName,
		},
		{
			name: "empty relation name",
			build: func() (*Schema, error) {
				return NewSchema("article").Relation("", author, 1).Build()
			},
			wantErr: ErrEmptyName,
		},
		{
			name: "duplicate relation",
			build: func() (*Schema, error) {
				return NewSchema("article").
					Relation("author", author, 1).
					Relation("author", author, 2).
					Build()
			},
			wantErr: ErrDuplicateRelation,
		},
		{
			name: "nil relation target",
			build: func() (*Schema, error) {
				return NewSchema("article").Relation("author", nil, 1).Build()
			},
			wantErr: ErrNilTarget,
		},
		{
			name: "relation max depth zero",
			build: func() (*Schema, error) {
				return NewSchema("article").Relation("author", author, 0).Build()
			},
			wantErr: ErrInvalidMaxDepth,
		},
		{
			name: "relation max depth negative",
			build: func() (*Schema, error) {
				return NewSchema("article").Relation("author", author, -1).Build()
			},
			wantErr: ErrInvalidMaxDepth,
		},
		{
			name: "max limit zero",
			build: func() (*Schema, error) {
				return NewSchema("article").MaxLimit(0).Build()
			},
			wantErr: ErrInvalidMaxLimit,
		},
		{
			name: "max limit above absolute ceiling",
			build: func() (*Schema, error) {
				return NewSchema("article").MaxLimit(AbsoluteMaxLimit + 1).Build()
			},
			wantErr: ErrInvalidMaxLimit,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := tc.build()
			require.Nil(t, s)
			require.Error(t, err)
			require.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestBuilder_AccumulatesMultipleErrors(t *testing.T) {
	// Each call below fails for a different reason. NewSchema("") contributes
	// ErrEmptyName; the rest are independent, unrelated failures, so all
	// four sentinels should surface from the single joined Build error.
	_, err := NewSchema("").
		Filterable("title", TypeString). // no operators given
		Relation("author", nil, 1).
		MaxLimit(0).
		Build()

	require.Error(t, err)
	require.ErrorIs(t, err, ErrEmptyName)
	require.ErrorIs(t, err, ErrNoOperators)
	require.ErrorIs(t, err, ErrNilTarget)
	require.ErrorIs(t, err, ErrInvalidMaxLimit)
}
