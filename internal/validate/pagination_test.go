package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
	"github.com/stretchr/testify/require"
)

func intPtr(v int) *int { return &v }

func TestValidatePagination(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name       string
		pagination query.Pagination
		wantErr    error
	}{
		{
			name:       "zero value pagination is a no-op",
			pagination: query.Pagination{},
		},
		{
			name:       "start with no limit is fine",
			pagination: query.Pagination{Start: intPtr(20)},
		},
		{
			name:       "limit within the schema's default max",
			pagination: query.Pagination{Limit: intPtr(25)},
		},
		{
			name:       "limit exactly at the schema's default max",
			pagination: query.Pagination{Limit: intPtr(100)}, // DefaultMaxLimit
		},
		{
			name:       "limit one past the schema's default max",
			pagination: query.Pagination{Limit: intPtr(101)},
			wantErr:    query.ErrInvalidPagination,
		},
		{
			name:       "start and a valid limit together",
			pagination: query.Pagination{Start: intPtr(50), Limit: intPtr(25)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePagination(tc.pagination, article)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}

// TestValidatePagination_CustomMaxLimit confirms the check uses whatever
// MaxLimit a schema actually configures, not just the default.
func TestValidatePagination_CustomMaxLimit(t *testing.T) {
	tight := &schema.Schema{
		Name:       "tight",
		Filterable: make(map[string]schema.FieldDef),
		Sortable:   make(map[string]struct{}),
		Selectable: make(map[string]struct{}),
		Relations:  make(map[string]schema.RelationDef),
		MaxLimit:   10,
	}

	require.NoError(t, ValidatePagination(query.Pagination{Limit: intPtr(10)}, tight))

	err := ValidatePagination(query.Pagination{Limit: intPtr(11)}, tight)
	require.Error(t, err)
	require.ErrorIs(t, err, query.ErrInvalidPagination)
}
