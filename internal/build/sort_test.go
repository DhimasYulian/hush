package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestBuildSort(t *testing.T) {
	testCases := []struct {
		name    string
		params  []query.Param
		want    []query.Sort
		wantErr error
	}{
		{
			name: "no sort",
			want: nil,
		},
		{
			name: "single sort",
			params: []query.Param{
				{
					Path:  []string{"sort"},
					Value: "name",
				},
			},
			want: []query.Sort{
				{
					Path:      query.Path{"name"},
					Direction: query.SortAsc,
				},
			},
		},
		{
			name: "single descending sort",
			params: []query.Param{
				{
					Path:  []string{"sort"},
					Value: "name:desc",
				},
			},
			want: []query.Sort{
				{
					Path:      query.Path{"name"},
					Direction: query.SortDesc,
				},
			},
		},
		{
			name: "multiple sort",
			params: []query.Param{
				{
					Path:  []string{"sort", "0"},
					Value: "name",
				},
				{
					Path:  []string{"sort", "1"},
					Value: "createdAt:desc",
				},
			},
			want: []query.Sort{
				{
					Path:      query.Path{"name"},
					Direction: query.SortAsc,
				},
				{
					Path:      query.Path{"createdAt"},
					Direction: query.SortDesc,
				},
			},
		},
		{
			name: "compact sparse indexes",
			params: []query.Param{
				{
					Path:  []string{"sort", "5"},
					Value: "name",
				},
				{
					Path:  []string{"sort", "2"},
					Value: "createdAt:desc",
				},
			},
			want: []query.Sort{
				{
					Path:      query.Path{"createdAt"},
					Direction: query.SortDesc,
				},
				{
					Path:      query.Path{"name"},
					Direction: query.SortAsc,
				},
			},
		},
		{
			name: "nested field",
			params: []query.Param{
				{
					Path:  []string{"sort"},
					Value: "author.company.name:desc",
				},
			},
			want: []query.Sort{
				{
					Path:      query.Path{"author", "company", "name"},
					Direction: query.SortDesc,
				},
			},
		},
		{
			name: "reject mixed syntax",
			params: []query.Param{
				{
					Path:  []string{"sort"},
					Value: "name",
				},
				{
					Path:  []string{"sort", "0"},
					Value: "age",
				},
			},
			wantErr: ErrInvalidSort,
		},
		{
			name: "reject non numeric index",
			params: []query.Param{
				{
					Path:  []string{"sort", "name"},
					Value: "id",
				},
			},
			wantErr: ErrInvalidSort,
		},
		{
			name: "reject invalid direction",
			params: []query.Param{
				{
					Path:  []string{"sort"},
					Value: "name:foo",
				},
			},
			wantErr: ErrInvalidSort,
		},
		{
			name: "reject too many separators",
			params: []query.Param{
				{
					Path:  []string{"sort"},
					Value: "name:desc:foo",
				},
			},
			wantErr: ErrInvalidSort,
		},
		{
			name: "reject empty value",
			params: []query.Param{
				{
					Path:  []string{"sort"},
					Value: "",
				},
			},
			wantErr: ErrInvalidSort,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildSort(tc.params)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
