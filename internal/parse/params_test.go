package parse

import (
	"net/url"
	"testing"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/stretchr/testify/require"
)

func TestParseParams(t *testing.T) {
	testCases := []struct {
		name    string
		values  url.Values
		want    []query.Param
		wantErr bool
	}{
		{
			name: "single parameter",
			values: url.Values{
				"filters[name][$eq]": {"John"},
			},
			want: []query.Param{
				{
					Path:  []string{"filters", "name", "$eq"},
					Value: "John",
				},
			},
		},
		{
			name: "multiple parameters",
			values: url.Values{
				"filters[name][$eq]": {"John"},
				"sort[0]":            {"createdAt:desc"},
				"pagination[limit]":  {"25"},
			},
			want: []query.Param{
				{
					Path:  []string{"filters", "name", "$eq"},
					Value: "John",
				},
				{
					Path:  []string{"pagination", "limit"},
					Value: "25",
				},
				{
					Path:  []string{"sort", "0"},
					Value: "createdAt:desc",
				},
			},
		},
		{
			name: "duplicate values",
			values: url.Values{
				"fields[0]": {"id", "title"},
			},
			want: []query.Param{
				{
					Path:  []string{"fields", "0"},
					Value: "id",
				},
				{
					Path:  []string{"fields", "0"},
					Value: "title",
				},
			},
		},
		{
			name: "empty value",
			values: url.Values{
				"populate": {""},
			},
			want: []query.Param{
				{
					Path:  []string{"populate"},
					Value: "",
				},
			},
		},
		{
			name: "invalid path",
			values: url.Values{
				"filters[][name]": {"John"},
			},
			wantErr: true,
		},
		{
			name:   "empty values",
			values: url.Values{},
			want:   []query.Param{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseParams(tc.values)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParseParams_Deterministic(t *testing.T) {
	values := url.Values{
		"filters[name][$eq]":  {"John"},
		"filters[age][$gte]":  {"18"},
		"sort[0]":             {"name"},
		"pagination[limit]":   {"25"},
		"populate[author][0]": {"profile"},
		"fields[0]":           {"id"},
		"fields[1]":           {"title"},
	}

	first, err := ParseParams(values)
	require.NoError(t, err)

	for i := 0; i < 50; i++ {
		got, err := ParseParams(values)
		require.NoError(t, err)
		require.Equal(t, first, got, "ParseParams must be deterministic across repeated calls")
	}
}
