package mongo

import (
	"github.com/DhimasYulian/hush"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Translator translates a validated hush query into MongoDB bson documents.
// The zero value applies default naming conventions; set the fields to
// customize how relation populates map to $lookup stages.
type Translator struct {
	// CollectionName maps a hush relation name to the MongoDB collection used
	// in a $lookup "from" stage. Default: relation name + "s" ("author" →
	// "authors").
	CollectionName func(rel string) string

	// LocalField maps a hush relation name to the field on the current
	// collection holding the reference to the related document. Default:
	// relation name + "_id" ("author" → "author_id").
	LocalField func(rel string) string

	// ForeignField is the field on the related collection that the local field
	// references. Default: "_id".
	ForeignField string

	// FieldName maps a hush field name to the document field name. Default:
	// identity. Useful when schemas declare logical names that differ from the
	// stored bson keys.
	FieldName func(field string) string
}

// collectionName returns the $lookup target collection for a relation.
func (t Translator) collectionName(rel string) string {
	if t.CollectionName != nil {
		return t.CollectionName(rel)
	}
	return rel + "s"
}

// localField returns the $lookup local field for a relation.
func (t Translator) localField(rel string) string {
	if t.LocalField != nil {
		return t.LocalField(rel)
	}
	return rel + "_id"
}

// foreignField returns the $lookup foreign field.
func (t Translator) foreignField() string {
	if t.ForeignField != "" {
		return t.ForeignField
	}
	return "_id"
}

// fieldName maps a hush field name to the document field name.
func (t Translator) fieldName(field string) string {
	if t.FieldName != nil {
		return t.FieldName(field)
	}
	return field
}

// Filter builds the bson filter document for the query's filter tree using the
// default Translator conventions.
func Filter(q *hush.Query) bson.M {
	return (Translator{}).Filter(q)
}

// Projection builds a projection document from whitelisted selectable fields
// using the default Translator conventions.
func Projection(schema *hush.Schema, q *hush.Query) bson.M {
	return (Translator{}).Projection(schema, q)
}

// Sort builds the ordered sort document using the default Translator
// conventions.
func Sort(schema *hush.Schema, q *hush.Query) bson.D {
	return (Translator{}).Sort(schema, q)
}

// Skip returns the pagination start offset using the default Translator
// conventions, or nil when unset.
func Skip(q *hush.Query) *int64 {
	return (Translator{}).Skip(q)
}

// Limit returns the pagination limit using the default Translator conventions,
// requesting one extra document when WithCount is true.
func Limit(q *hush.Query) *int64 {
	return (Translator{}).Limit(q)
}

// FindOptions bundles projection, sort, skip, and limit into mongo find options
// using the default Translator conventions.
func FindOptions(schema *hush.Schema, q *hush.Query) *options.FindOptions {
	return (Translator{}).FindOptions(schema, q)
}

// Pipeline builds the aggregation pipeline for groupBy, aggregations, and
// populate relations using the default Translator conventions.
func Pipeline(schema *hush.Schema, q *hush.Query) (mongo.Pipeline, error) {
	return (Translator{}).Pipeline(schema, q)
}

// PipelineFacet builds a $facet pipeline that returns results and an exact
// total count using the default Translator conventions.
func PipelineFacet(schema *hush.Schema, q *hush.Query) (mongo.Pipeline, error) {
	return (Translator{}).PipelineFacet(schema, q)
}
