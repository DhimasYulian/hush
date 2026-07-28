package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestBuildPagination(t *testing.T) {
	testCases := []struct {
		name    string
		params  []query.Param
		want    query.Pagination
		wantErr error
	}{
		{
			name: "no pagination defaults withCount to true",
			want: query.Pagination{
				WithCount: boolPtr(true),
			},
		},
		{
			name: "start only",
			params: []query.Param{
				{
					Path:  []string{"pagination", "start"},
					Value: "10",
				},
			},
			want: query.Pagination{
				Start:     intPtr(10),
				WithCount: boolPtr(true),
			},
		},
		{
			name: "limit only",
			params: []query.Param{
				{
					Path:  []string{"pagination", "limit"},
					Value: "20",
				},
			},
			want: query.Pagination{
				Limit:     intPtr(20),
				WithCount: boolPtr(true),
			},
		},
		{
			name: "start and limit",
			params: []query.Param{
				{
					Path:  []string{"pagination", "start"},
					Value: "10",
				},
				{
					Path:  []string{"pagination", "limit"},
					Value: "20",
				},
			},
			want: query.Pagination{
				Start:     intPtr(10),
				Limit:     intPtr(20),
				WithCount: boolPtr(true),
			},
		},
		{
			name: "withCount true",
			params: []query.Param{
				{
					Path:  []string{"pagination", "withCount"},
					Value: "true",
				},
			},
			want: query.Pagination{
				WithCount: boolPtr(true),
			},
		},
		{
			name: "withCount false",
			params: []query.Param{
				{
					Path:  []string{"pagination", "withCount"},
					Value: "false",
				},
			},
			want: query.Pagination{
				WithCount: boolPtr(false),
			},
		},
		{
			name: "withCount numeric 1",
			params: []query.Param{
				{
					Path:  []string{"pagination", "withCount"},
					Value: "1",
				},
			},
			want: query.Pagination{
				WithCount: boolPtr(true),
			},
		},
		{
			name: "withCount numeric 0",
			params: []query.Param{
				{
					Path:  []string{"pagination", "withCount"},
					Value: "0",
				},
			},
			want: query.Pagination{
				WithCount: boolPtr(false),
			},
		},
		{
			name: "all fields together",
			params: []query.Param{
				{
					Path:  []string{"pagination", "start"},
					Value: "0",
				},
				{
					Path:  []string{"pagination", "limit"},
					Value: "50",
				},
				{
					Path:  []string{"pagination", "withCount"},
					Value: "false",
				},
			},
			want: query.Pagination{
				Start:     intPtr(0),
				Limit:     intPtr(50),
				WithCount: boolPtr(false),
			},
		},
		{
			name: "reject invalid key",
			params: []query.Param{
				{
					Path:  []string{"pagination", "page"},
					Value: "1",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
		{
			name: "reject invalid integer",
			params: []query.Param{
				{
					Path:  []string{"pagination", "start"},
					Value: "abc",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
		{
			name: "reject invalid withCount value",
			params: []query.Param{
				{
					Path:  []string{"pagination", "withCount"},
					Value: "abc",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
		{
			name: "reject duplicate start",
			params: []query.Param{
				{
					Path:  []string{"pagination", "start"},
					Value: "0",
				},
				{
					Path:  []string{"pagination", "start"},
					Value: "10",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
		{
			name: "reject duplicate limit",
			params: []query.Param{
				{
					Path:  []string{"pagination", "limit"},
					Value: "10",
				},
				{
					Path:  []string{"pagination", "limit"},
					Value: "20",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
		{
			name: "reject duplicate withCount",
			params: []query.Param{
				{
					Path:  []string{"pagination", "withCount"},
					Value: "true",
				},
				{
					Path:  []string{"pagination", "withCount"},
					Value: "false",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
		{
			name: "reject negative start",
			params: []query.Param{
				{
					Path:  []string{"pagination", "start"},
					Value: "-1",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
		{
			name: "reject negative limit",
			params: []query.Param{
				{
					Path:  []string{"pagination", "limit"},
					Value: "-100",
				},
			},
			wantErr: query.ErrInvalidPagination,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildPagination(tc.params)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
