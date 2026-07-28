package schema

import (
	"sort"

	"github.com/DhimasYulian/hush/internal/query"
)

// DefaultMaxLimit is the default maximum pagination limit.
const DefaultMaxLimit = 100

// AbsoluteMaxLimit is the hard upper bound for pagination limits.
const AbsoluteMaxLimit = 1000

// FieldType represents the data type of a filterable field.
type FieldType string

const (
	TypeString FieldType = "string"
	TypeNumber FieldType = "number"
	TypeBool   FieldType = "bool"
	TypeDate   FieldType = "date"
)

// FieldDef describes a filterable field: its name, type, and allowed operators.
type FieldDef struct {
	Name      string
	Type      FieldType
	Operators map[query.Operator]bool
}

// RelationDef describes a named relation to another schema with a max nesting depth.
type RelationDef struct {
	Name               string
	Target             *Schema
	MaxDepth           int
	HiddenFromWildcard bool
}

// Schema defines the allowed query operations for a resource.
type Schema struct {
	Name         string
	Filterable   map[string]FieldDef
	Sortable     map[string]struct{}
	Selectable   map[string]struct{}
	Groupable    map[string]struct{}
	Aggregatable map[string]struct{}
	Relations    map[string]RelationDef
	MaxLimit     int
}

// GetFilterable returns the field definition if the field is filterable.
func (s *Schema) GetFilterable(name string) (FieldDef, bool) {
	def, ok := s.Filterable[name]
	return def, ok
}

// GetSortable reports whether the field is sortable.
func (s *Schema) GetSortable(name string) bool {
	_, ok := s.Sortable[name]
	return ok
}

// GetSelectable reports whether the field can be selected.
func (s *Schema) GetSelectable(name string) bool {
	_, ok := s.Selectable[name]
	return ok
}

// GetGroupable reports whether the field can be used in groupBy.
func (s *Schema) GetGroupable(name string) bool {
	_, ok := s.Groupable[name]
	return ok
}

// GetAggregatable reports whether the field can be used in aggregations.
func (s *Schema) GetAggregatable(name string) bool {
	_, ok := s.Aggregatable[name]
	return ok
}

// GetSelectableFields returns all selectable field names in sorted order.
func (s *Schema) GetSelectableFields() []string {
	out := make([]string, 0, len(s.Selectable))
	for name := range s.Selectable {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// GetRelations returns a copy of all declared relations.
func (s *Schema) GetRelations() map[string]RelationDef {
	out := make(map[string]RelationDef, len(s.Relations))
	for name, rel := range s.Relations {
		out[name] = rel
	}
	return out
}

// GetRelation returns the named relation definition if it exists.
func (s *Schema) GetRelation(name string) (RelationDef, bool) {
	rel, ok := s.Relations[name]
	return rel, ok
}

// GetMaxLimit returns the maximum pagination limit allowed by this schema.
func (s *Schema) GetMaxLimit() int {
	return s.MaxLimit
}
