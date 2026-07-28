package build

import (
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestBuildFields(t *testing.T) {
	testCases := []struct {
		name    string
		params  []query.Param
		want    []string
		wantErr error
	}{
		{
			name: "no fields",
			want: nil,
		},
		{
			name: "single field",
			params: []query.Param{
				{
					Path:  []string{"fields", "0"},
					Value: "id",
				},
			},
			want: []string{"id"},
		},
		{
			name: "multiple fields",
			params: []query.Param{
				{
					Path:  []string{"fields", "0"},
					Value: "id",
				},
				{
					Path:  []string{"fields", "1"},
					Value: "title",
				},
				{
					Path:  []string{"fields", "2"},
					Value: "createdAt",
				},
			},
			want: []string{
				"id",
				"title",
				"createdAt",
			},
		},
		{
			name: "compact sparse indexes",
			params: []query.Param{
				{
					Path:  []string{"fields", "8"},
					Value: "age",
				},
				{
					Path:  []string{"fields", "5"},
					Value: "name",
				},
			},
			want: []string{"name", "age"},
		},
		{
			name: "reject non numeric index",
			params: []query.Param{
				{
					Path:  []string{"fields", "name"},
					Value: "id",
				},
			},
			wantErr: ErrInvalidFields,
		},
		{
			name: "reject empty index",
			params: []query.Param{
				{
					Path:  []string{"fields", ""},
					Value: "id",
				},
			},
			wantErr: ErrInvalidFields,
		},
		{
			name: "sort by index",
			params: []query.Param{
				{
					Path:  []string{"fields", "5"},
					Value: "name",
				},
				{
					Path:  []string{"fields", "2"},
					Value: "description",
				},
				{
					Path:  []string{"fields", "9"},
					Value: "email",
				},
			},
			want: []string{
				"description",
				"name",
				"email",
			},
		},
		{
			name: "reject mixed syntax",
			params: []query.Param{
				{
					Path:  []string{"fields"},
					Value: "name",
				},
				{
					Path:  []string{"fields", "0"},
					Value: "description",
				},
			},
			wantErr: ErrInvalidFields,
		},
		{
			name: "double shorthand syntax",
			params: []query.Param{
				{
					Path:  []string{"fields"},
					Value: "name",
				},
				{
					Path:  []string{"fields"},
					Value: "description",
				},
			},
			wantErr: ErrInvalidFields,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := BuildFields(tc.params)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}
