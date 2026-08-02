package query

// Query is the root result of parsing and validating query parameters.
type Query struct {
	Filters Filter

	Populates   []Populate
	PopulateAll bool
	Fields      []Field
	Sort        []Sort
	GroupBy     []Field

	Aggregations []Aggregation

	Pagination Pagination
}

// Filter is the interface implemented by Condition, And, Or, and Not.
type Filter interface {
	isFilter()
}

// Condition is a leaf filter: a field path + operator + value(s).
//
// FieldType is the schema-declared type of the filtered field and Values holds
// the raw [Value] strings coerced to that type. Both are populated by Parse
// once the schema is available, so adapters do not need to resolve types
// themselves. Values is nil when the operator carries no value (e.g. $null).
type Condition struct {
	Path      Path
	Operator  Operator
	Value     Value
	FieldType FieldType
	Values    []any
}

func (c Condition) isFilter() {}

// And combines multiple filters with logical AND.
type And struct {
	Filters []Filter
}

func (a And) isFilter() {}

// Or combines multiple filters with logical OR.
type Or struct {
	Filters []Filter
}

func (o Or) isFilter() {}

// Not negates a single filter.
type Not struct {
	Filter Filter
}

func (n Not) isFilter() {}

// Sort specifies a field path and direction for ordering.
type Sort struct {
	Path      Path
	Direction SortDirection
}

// Pagination holds optional start/limit values and a withCount flag.
type Pagination struct {
	Start     *int
	Limit     *int
	WithCount *bool
}

// Populate specifies a relation to include, with optional nested options.
type Populate struct {
	Relation string

	Fields  []Field
	Filters Filter
	Sorts   []Sort

	Populates []Populate
}

// Aggregation specifies a computed aggregate function with an output alias.
type Aggregation struct {
	Alias string
	Func  string
	Field string
}

// FieldType represents the data type of a filterable field.
type FieldType string

const (
	TypeString FieldType = "string"
	TypeNumber FieldType = "number"
	TypeBool   FieldType = "bool"
	TypeDate   FieldType = "date"
)

// Operator is a string-typed filter comparison operator.
type Operator string

// SortDirection is a string-typed sort direction ("asc" or "desc").
type SortDirection string

// Path is a slice of strings representing a dotted/bracket field path.
type Path = []string

// Value is a slice of strings representing one or more filter values.
type Value = []string

// Field is a string-typed field name.
type Field = string

// Param holds a parsed parameter: path segments and a string value.
type Param struct {
	Path  []string
	Value string
}
