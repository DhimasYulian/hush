package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestValidateGroupBy(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name    string
		fields  []query.Field
		wantErr error
	}{
		{
			name:   "empty groupBy is a no-op",
			fields: nil,
		},
		{
			name:   "single valid groupable field",
			fields: []query.Field{"title"},
		},
		{
			name:   "multiple valid groupable fields",
			fields: []query.Field{"title", "createdAt"},
		},
		{
			name:    "unknown field",
			fields:  []query.Field{"nonexistent"},
			wantErr: ErrUnknownGroupBy,
		},
		{
			// views is Filterable but was never declared Groupable.
			name:    "filterable field that isn't groupable",
			fields:  []query.Field{"views"},
			wantErr: ErrUnknownGroupBy,
		},
		{
			// body is Selectable but was never declared Groupable.
			name:    "selectable field that isn't groupable",
			fields:  []query.Field{"body"},
			wantErr: ErrUnknownGroupBy,
		},
		{
			name:    "second invalid field in a list is caught",
			fields:  []query.Field{"title", "nonexistent"},
			wantErr: ErrUnknownGroupBy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateGroupBy(tc.fields, article)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
