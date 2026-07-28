package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestBuildAggregations(t *testing.T) {
	testCases := []struct {
		name    string
		params  []query.Param
		want    []query.Aggregation
		wantErr error
	}{
		{
			name: "no aggregations",
			want: nil,
		},
		{
			name: "count without field defaults to *",
			params: []query.Param{
				{Path: []string{"aggregations", "total", "func"}, Value: "count"},
			},
			want: []query.Aggregation{
				{Alias: "total", Func: "count", Field: "*"},
			},
		},
		{
			name: "count with explicit field",
			params: []query.Param{
				{Path: []string{"aggregations", "total", "func"}, Value: "count"},
				{Path: []string{"aggregations", "total", "field"}, Value: "id"},
			},
			want: []query.Aggregation{
				{Alias: "total", Func: "count", Field: "id"},
			},
		},
		{
			name: "sum with field",
			params: []query.Param{
				{Path: []string{"aggregations", "totalSalary", "func"}, Value: "sum"},
				{Path: []string{"aggregations", "totalSalary", "field"}, Value: "salary"},
			},
			want: []query.Aggregation{
				{Alias: "totalSalary", Func: "sum", Field: "salary"},
			},
		},
		{
			name: "avg with field",
			params: []query.Param{
				{Path: []string{"aggregations", "avgAge", "func"}, Value: "avg"},
				{Path: []string{"aggregations", "avgAge", "field"}, Value: "age"},
			},
			want: []query.Aggregation{
				{Alias: "avgAge", Func: "avg", Field: "age"},
			},
		},
		{
			name: "multiple aggregations",
			params: []query.Param{
				{Path: []string{"aggregations", "total", "func"}, Value: "count"},
				{Path: []string{"aggregations", "totalSalary", "func"}, Value: "sum"},
				{Path: []string{"aggregations", "totalSalary", "field"}, Value: "salary"},
				{Path: []string{"aggregations", "avgAge", "func"}, Value: "avg"},
				{Path: []string{"aggregations", "avgAge", "field"}, Value: "age"},
			},
			want: []query.Aggregation{
				{Alias: "total", Func: "count", Field: "*"},
				{Alias: "totalSalary", Func: "sum", Field: "salary"},
				{Alias: "avgAge", Func: "avg", Field: "age"},
			},
		},
		{
			name: "field before func is ok",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "field"}, Value: "salary"},
				{Path: []string{"aggregations", "x", "func"}, Value: "sum"},
			},
			want: []query.Aggregation{
				{Alias: "x", Func: "sum", Field: "salary"},
			},
		},
		{
			name: "reject invalid func",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "func"}, Value: "min"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject missing func",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "field"}, Value: "salary"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject missing field for sum",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "func"}, Value: "sum"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject missing field for avg",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "func"}, Value: "avg"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject duplicate alias func",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "func"}, Value: "count"},
				{Path: []string{"aggregations", "x", "func"}, Value: "sum"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject duplicate alias field",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "func"}, Value: "sum"},
				{Path: []string{"aggregations", "x", "field"}, Value: "a"},
				{Path: []string{"aggregations", "x", "field"}, Value: "b"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject unknown key",
			params: []query.Param{
				{Path: []string{"aggregations", "x", "order"}, Value: "asc"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject empty alias",
			params: []query.Param{
				{Path: []string{"aggregations", "", "func"}, Value: "count"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
		{
			name: "reject invalid path length",
			params: []query.Param{
				{Path: []string{"aggregations", "x"}, Value: "count"},
			},
			wantErr: query.ErrInvalidAggregation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildAggregations(tc.params)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
