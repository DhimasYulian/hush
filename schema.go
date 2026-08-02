package hush

import (
	"github.com/DhimasYulian/hush/internal/schema"
)

// FieldType represents the data type of a filterable field.
type FieldType = schema.FieldType

const (
	TypeString = schema.TypeString
	TypeNumber = schema.TypeNumber
	TypeBool   = schema.TypeBool
	TypeDate   = schema.TypeDate
)

// FieldDef describes a filterable field: its name, type, and allowed operators.
type FieldDef = schema.FieldDef

// RelationDef describes a named relation to another schema with a max nesting depth.
type RelationDef = schema.RelationDef

// Schema defines the allowed query operations for a resource. Built via SchemaBuilder.
type Schema struct {
	inner *schema.Schema
}

// Name returns the schema's resource name.
func (s *Schema) Name() string {
	return s.inner.Name
}

// Filterable returns the field definition if the field is filterable.
func (s *Schema) Filterable(name string) (FieldDef, bool) {
	return s.inner.GetFilterable(name)
}

// Sortable reports whether the field is sortable.
func (s *Schema) Sortable(name string) bool {
	return s.inner.GetSortable(name)
}

// Selectable reports whether the field can be selected.
func (s *Schema) Selectable(name string) bool {
	return s.inner.GetSelectable(name)
}

// SelectableFields returns all selectable field names in sorted order.
func (s *Schema) SelectableFields() []string {
	return s.inner.GetSelectableFields()
}

// Groupable reports whether the field can be used in groupBy.
func (s *Schema) Groupable(name string) bool {
	return s.inner.GetGroupable(name)
}

// Aggregatable reports whether the field can be used in aggregations.
func (s *Schema) Aggregatable(name string) bool {
	return s.inner.GetAggregatable(name)
}

// Relations returns a copy of all declared relations.
func (s *Schema) Relations() map[string]RelationDef {
	return s.inner.GetRelations()
}

// Relation returns the named relation definition if it exists.
func (s *Schema) Relation(name string) (RelationDef, bool) {
	return s.inner.GetRelation(name)
}

// MaxLimit returns the maximum pagination limit allowed by this schema.
func (s *Schema) MaxLimit() int {
	return s.inner.GetMaxLimit()
}

// Inner returns the underlying schema definition. It is exposed for
// integrations (such as the hush/gorm adapter) that need to walk relations and
// their nested targets directly.
func (s *Schema) Inner() *schema.Schema {
	return s.inner
}
