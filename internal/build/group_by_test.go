package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestBuildGroupBy(t *testing.T) {
	testCases := []struct {
		name    string
		params  []query.Param
		want    []string
		wantErr error
	}{
		{
			name: "no groupBy",
			want: nil,
		},
		{
			name: "single field",
			params: []query.Param{
				{
					Path:  []string{"groupBy", "0"},
					Value: "status",
				},
			},
			want: []string{"status"},
		},
		{
			name: "multiple fields",
			params: []query.Param{
				{
					Path:  []string{"groupBy", "0"},
					Value: "status",
				},
				{
					Path:  []string{"groupBy", "1"},
					Value: "category",
				},
				{
					Path:  []string{"groupBy", "2"},
					Value: "createdAt",
				},
			},
			want: []string{"status", "category", "createdAt"},
		},
		{
			name: "compact sparse indexes",
			params: []query.Param{
				{
					Path:  []string{"groupBy", "5"},
					Value: "status",
				},
				{
					Path:  []string{"groupBy", "2"},
					Value: "category",
				},
			},
			want: []string{"category", "status"},
		},
		{
			name: "shorthand single field",
			params: []query.Param{
				{
					Path:  []string{"groupBy"},
					Value: "status",
				},
			},
			want: []string{"status"},
		},
		{
			name: "reject non numeric index",
			params: []query.Param{
				{
					Path:  []string{"groupBy", "name"},
					Value: "status",
				},
			},
			wantErr: ErrInvalidGroupBy,
		},
		{
			name: "reject empty index",
			params: []query.Param{
				{
					Path:  []string{"groupBy", ""},
					Value: "status",
				},
			},
			wantErr: ErrInvalidGroupBy,
		},
		{
			name: "reject mixed syntax",
			params: []query.Param{
				{
					Path:  []string{"groupBy"},
					Value: "status",
				},
				{
					Path:  []string{"groupBy", "0"},
					Value: "category",
				},
			},
			wantErr: ErrInvalidGroupBy,
		},
		{
			name: "reject double shorthand",
			params: []query.Param{
				{
					Path:  []string{"groupBy"},
					Value: "status",
				},
				{
					Path:  []string{"groupBy"},
					Value: "category",
				},
			},
			wantErr: ErrInvalidGroupBy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildGroupBy(tc.params)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
