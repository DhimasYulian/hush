package hush

import (
	"errors"
	"fmt"

	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
)

// DefaultMaxLimit is the default maximum pagination limit.
const DefaultMaxLimit = schema.DefaultMaxLimit

// AbsoluteMaxLimit is the hard upper bound for pagination limits.
const AbsoluteMaxLimit = schema.AbsoluteMaxLimit

var allFieldTypes = map[FieldType]bool{
	TypeString: true,
	TypeNumber: true,
	TypeBool:   true,
	TypeDate:   true,
}

// allOperators is the canonical operator set from the query package.
var allOperators = query.AllOperators

// SchemaBuilder constructs a Schema with accumulated validation errors.
// Call Build to finalize and check for errors.
type SchemaBuilder struct {
	schema *schema.Schema
	errs   []error
}

// NewSchema creates a new SchemaBuilder with the given resource name.
func NewSchema(name string) *SchemaBuilder {
	b := &SchemaBuilder{
		schema: &schema.Schema{
			Name:         name,
			Filterable:   make(map[string]FieldDef),
			Sortable:     make(map[string]struct{}),
			Selectable:   make(map[string]struct{}),
			Groupable:    make(map[string]struct{}),
			Aggregatable: make(map[string]struct{}),
			Relations:    make(map[string]RelationDef),
			MaxLimit:     DefaultMaxLimit,
		},
	}

	if name == "" {
		b.fail(ErrEmptyName)
	}

	return b
}

func (b *SchemaBuilder) fail(err error) {
	b.errs = append(b.errs, err)
}

// Filterable declares a field as filterable with a type and set of allowed operators.
func (b *SchemaBuilder) Filterable(name string, t FieldType, ops ...Operator) *SchemaBuilder {
	if name == "" {
		b.fail(ErrEmptyName)
		return b
	}

	if _, exists := b.schema.Filterable[name]; exists {
		b.fail(fmt.Errorf("%w: %q", ErrDuplicateField, name))
		return b
	}

	if !allFieldTypes[t] {
		b.fail(fmt.Errorf("%w: %q on field %q", ErrUnknownFieldType, t, name))
		return b
	}

	if len(ops) == 0 {
		b.fail(fmt.Errorf("%w: %q", ErrNoOperators, name))
		return b
	}

	opSet := make(map[Operator]bool, len(ops))
	for _, op := range ops {
		if !allOperators[op] {
			b.fail(fmt.Errorf("%w: %q on field %q", ErrUnknownOperator, op, name))
			return b
		}
		opSet[op] = true
	}

	b.schema.Filterable[name] = FieldDef{Name: name, Type: t, Operators: opSet}

	return b
}

// Sortable declares fields as sortable.
func (b *SchemaBuilder) Sortable(names ...string) *SchemaBuilder {
	for _, name := range names {
		if name == "" {
			b.fail(ErrEmptyName)
			continue
		}
		b.schema.Sortable[name] = struct{}{}
	}
	return b
}

// Selectable declares fields as selectable (available for field selection).
func (b *SchemaBuilder) Selectable(names ...string) *SchemaBuilder {
	for _, name := range names {
		if name == "" {
			b.fail(ErrEmptyName)
			continue
		}
		b.schema.Selectable[name] = struct{}{}
	}
	return b
}

// Groupable declares fields as groupable (available for groupBy).
func (b *SchemaBuilder) Groupable(names ...string) *SchemaBuilder {
	for _, name := range names {
		if name == "" {
			b.fail(ErrEmptyName)
			continue
		}
		b.schema.Groupable[name] = struct{}{}
	}
	return b
}

// Aggregatable declares fields as aggregatable (available for aggregations).
func (b *SchemaBuilder) Aggregatable(names ...string) *SchemaBuilder {
	for _, name := range names {
		if name == "" {
			b.fail(ErrEmptyName)
			continue
		}
		b.schema.Aggregatable[name] = struct{}{}
	}
	return b
}

// RelationOption configures a relation declaration.
type RelationOption func(*RelationDef)

// Hidden marks a relation as hidden from wildcard populate (populate=*).
func Hidden() RelationOption {
	return func(r *RelationDef) {
		r.HiddenFromWildcard = true
	}
}

// Relation declares a named relation to another schema with a max nesting depth.
func (b *SchemaBuilder) Relation(name string, target *Schema, maxDepth int, opts ...RelationOption) *SchemaBuilder {
	if name == "" {
		b.fail(ErrEmptyName)
		return b
	}

	if _, exists := b.schema.Relations[name]; exists {
		b.fail(fmt.Errorf("%w: %q", ErrDuplicateRelation, name))
		return b
	}

	if target == nil {
		b.fail(fmt.Errorf("%w: relation %q", ErrNilTarget, name))
		return b
	}

	if maxDepth < 1 {
		b.fail(fmt.Errorf("%w: relation %q", ErrInvalidMaxDepth, name))
		return b
	}

	rel := RelationDef{Name: name, Target: target.inner, MaxDepth: maxDepth}
	for _, opt := range opts {
		opt(&rel)
	}

	b.schema.Relations[name] = rel

	return b
}

// MaxLimit sets the maximum allowed pagination limit. Must be between 1 and AbsoluteMaxLimit.
func (b *SchemaBuilder) MaxLimit(n int) *SchemaBuilder {
	if n < 1 || n > AbsoluteMaxLimit {
		b.fail(fmt.Errorf("%w: got %d", ErrInvalidMaxLimit, n))
		return b
	}
	b.schema.MaxLimit = n
	return b
}

// Build finalizes the schema and returns any accumulated validation errors.
func (b *SchemaBuilder) Build() (*Schema, error) {
	if len(b.errs) > 0 {
		return nil, errors.Join(b.errs...)
	}
	return &Schema{inner: b.schema}, nil
}
