package hush

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParse_EnrichesConditionTypes(t *testing.T) {
	root, err := NewSchema("article").
		Filterable("title", TypeString, OpEq, OpContains, OpStartsWith, OpEndsWith).
		Filterable("views", TypeNumber, OpGt, OpLte, OpBetween, OpIn).
		Filterable("active", TypeBool, OpEq).
		Filterable("publishedAt", TypeDate, OpGt, OpNull, OpNotNull).
		Filterable("status", TypeString, OpNotIn).
		MaxLimit(100).
		Build()
	require.NoError(t, err)

	got, err := Parse(url.Values{
		"filters[$and][0][title][$eq]":         {"go lang"},
		"filters[$and][1][views][$between][0]": {"1"},
		"filters[$and][1][views][$between][1]": {"10"},
		"filters[$and][2][views][$in][0]":      {"1"},
		"filters[$and][2][views][$in][1]":      {"2"},
		"filters[$and][2][views][$in][2]":      {"3"},
		"filters[$and][3][active][$eq]":        {"true"},
		"filters[$and][4][publishedAt][$gt]":   {"2024-05-17T12:30:45Z"},
		"filters[$and][5][publishedAt][$null]": {"true"},
		"filters[$and][6][status][$notIn][0]":  {"draft"},
		"filters[$and][7][title][$contains]":   {"50%_off"},
	}, root)
	require.NoError(t, err)

	and, ok := got.Filters.(And)
	require.True(t, ok)
	require.Len(t, and.Filters, 8)

	byKey := make(map[string]Condition)
	for _, f := range and.Filters {
		c, ok := f.(Condition)
		require.True(t, ok)
		byKey[c.Path[0]+string(c.Operator)] = c
	}

	wantTypes := map[string]FieldType{
		"title$eq": TypeString, "title$contains": TypeString,
		"views$between": TypeNumber, "views$in": TypeNumber,
		"active$eq": TypeBool, "publishedAt$gt": TypeDate,
		"publishedAt$null": TypeDate, "status$notIn": TypeString,
	}
	for key, want := range wantTypes {
		require.Equal(t, want, byKey[key].FieldType, key)
	}

	require.Equal(t, []any{float64(1), float64(10)}, byKey["views$between"].Values)
	require.Equal(t, []any{float64(1), float64(2), float64(3)}, byKey["views$in"].Values)
	require.Nil(t, byKey["publishedAt$null"].Values)
	require.Equal(t, []any{"50%_off"}, byKey["title$contains"].Values)
}

func TestParse_TypedConditionValues(t *testing.T) {
	root, err := NewSchema("article").
		Filterable("title", TypeString, OpContainsi).
		Filterable("views", TypeNumber, OpGte).
		Filterable("active", TypeBool, OpEq).
		Filterable("publishedAt", TypeDate, OpGt).
		Filterable("status", TypeString, OpIn).
		MaxLimit(100).
		Build()
	require.NoError(t, err)

	got, err := Parse(url.Values{
		"filters[title][$containsi]": {"go"},
		"filters[views][$gte]":       {"50"},
		"filters[active][$eq]":       {"true"},
		"filters[publishedAt][$gt]":  {"2024-05-17T12:30:45Z"},
		"filters[status][$in][0]":    {"published"},
		"filters[status][$in][1]":    {"draft"},
	}, root)
	require.NoError(t, err)

	and := got.Filters.(And)
	byKey := map[string]Condition{}
	for _, f := range and.Filters {
		c := f.(Condition)
		byKey[c.Path[0]] = c
	}

	title := byKey["title"]
	require.Equal(t, TypeString, title.FieldType)
	require.Equal(t, []any{"go"}, title.Values)

	views := byKey["views"]
	require.Equal(t, TypeNumber, views.FieldType)
	require.Equal(t, []any{float64(50)}, views.Values)

	active := byKey["active"]
	require.Equal(t, TypeBool, active.FieldType)
	require.Equal(t, []any{true}, active.Values)

	published := byKey["publishedAt"]
	require.Equal(t, TypeDate, published.FieldType)
	wantTime, err := time.Parse(time.RFC3339, "2024-05-17T12:30:45Z")
	require.NoError(t, err)
	require.Equal(t, []any{wantTime}, published.Values)

	status := byKey["status"]
	require.Equal(t, TypeString, status.FieldType)
	require.Equal(t, []any{"published", "draft"}, status.Values)
}

func TestParse_NullOperatorsCarryNoValues(t *testing.T) {
	root, err := NewSchema("article").
		Filterable("publishedAt", TypeDate, OpNull, OpNotNull).
		MaxLimit(100).
		Build()
	require.NoError(t, err)

	got, err := Parse(url.Values{
		"filters[publishedAt][$notNull]": {"true"},
	}, root)
	require.NoError(t, err)

	c := got.Filters.(Condition)
	require.Equal(t, OpNotNull, c.Operator)
	require.Equal(t, TypeDate, c.FieldType)
	require.Nil(t, c.Values)
	require.Equal(t, Value{"true"}, c.Value)
}

func TestCoerceExport(t *testing.T) {
	v, err := Coerce(TypeNumber, "42")
	require.NoError(t, err)
	require.Equal(t, float64(42), v)

	_, err = Coerce(TypeNumber, "nope")
	require.Error(t, err)
}
