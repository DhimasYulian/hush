package validate

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestValidateAggregations(t *testing.T) {
	article, _, _ := fixtureSchemas(t)

	testCases := []struct {
		name    string
		aggs    []query.Aggregation
		wantErr error
	}{
		{
			name: "empty aggregations is a no-op",
			aggs: nil,
		},
		{
			name: "valid count with wildcard field",
			aggs: []query.Aggregation{
				{Alias: "total", Func: "count", Field: "*"},
			},
		},
		{
			name: "valid count with aggregatable field",
			aggs: []query.Aggregation{
				{Alias: "total", Func: "count", Field: "views"},
			},
		},
		{
			name: "valid sum",
			aggs: []query.Aggregation{
				{Alias: "totalViews", Func: "sum", Field: "views"},
			},
		},
		{
			name: "valid avg",
			aggs: []query.Aggregation{
				{Alias: "avgViews", Func: "avg", Field: "views"},
			},
		},
		{
			name: "multiple valid aggregations",
			aggs: []query.Aggregation{
				{Alias: "total", Func: "count", Field: "*"},
				{Alias: "totalViews", Func: "sum", Field: "views"},
				{Alias: "avgViews", Func: "avg", Field: "views"},
			},
		},
		{
			name: "sum on non-aggregatable field",
			aggs: []query.Aggregation{
				{Alias: "x", Func: "sum", Field: "title"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "avg on non-aggregatable field",
			aggs: []query.Aggregation{
				{Alias: "x", Func: "avg", Field: "title"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "count with non-aggregatable explicit field",
			aggs: []query.Aggregation{
				{Alias: "x", Func: "count", Field: "title"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "count on unknown field",
			aggs: []query.Aggregation{
				{Alias: "x", Func: "count", Field: "nonexistent"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateAggregations(tc.aggs, article)

			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
