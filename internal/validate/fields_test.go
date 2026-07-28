package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestValidateFields(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name    string
		fields  []query.Field
		wantErr error
	}{
		{
			name:   "empty fields is a no-op",
			fields: nil,
		},
		{
			name:   "single valid field",
			fields: []query.Field{"title"},
		},
		{
			name:   "multiple valid fields",
			fields: []query.Field{"title", "body"},
		},
		{
			name:    "unknown field",
			fields:  []query.Field{"nonexistent"},
			wantErr: ErrUnknownField,
		},
		{
			// views is Filterable but was never declared Selectable.
			name:    "filterable field that isn't selectable",
			fields:  []query.Field{"views"},
			wantErr: ErrUnknownField,
		},
		{
			name:    "second invalid field in a list is caught",
			fields:  []query.Field{"title", "nonexistent"},
			wantErr: ErrUnknownField,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateFields(tc.fields, article)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
