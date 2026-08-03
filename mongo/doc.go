// Package mongo translates validated hush queries into MongoDB bson documents
// and aggregation pipelines so a consumer can integrate hush with the official
// MongoDB Go driver without writing any per-operator code.
//
// # Usage
//
// A validated [hush.Query] becomes, at minimum, a filter document and a set of
// find options:
//
//	filter := hushmongo.Filter(q)              // bson.M
//	opts   := hushmongo.FindOptions(schema, q) // *options.FindOptions
//	cursor, err := db.Coll.Find(ctx, filter, opts)
//
// When the query uses groupBy, aggregations, or populate relations, the same
// query translates to an aggregation pipeline:
//
//	pipeline, err := hushmongo.Pipeline(schema, q)
//	if err != nil { /* ... */ }
//	cursor, err := db.Coll.Aggregate(ctx, pipeline)
//
// The pipeline emits, in order:
//
//   - $match — the full [hush.Query.Filters] tree. Every hush operator maps to
//     a MongoDB comparison operator, logical $and/$or/$nor grouping preserves
//     nested trees like (a AND b) OR (c AND d), and string pattern operators
//     use $regex with [regexp.QuoteMeta] so user input matches literally.
//   - $sort — [hush.Query.Sort], skipping columns not sortable in the schema.
//   - $skip / $limit — [hush.Query.Pagination]. When WithCount is true and a
//     limit is set, one extra document is requested so the caller can detect a
//     next page (len(docs) > limit), matching hush/gorm.
//   - $group — [hush.Query.GroupBy] as the _id plus [hush.Query.Aggregations]
//     as $sum/$avg accumulator fields.
//   - $lookup — [hush.Query.Populates] as nested $lookup stages with their own
//     $match/$sort/$project, enforcing each relation's max depth.
//
// [PipelineFacet] returns a pipeline whose single stage is a $facet producing
// { results: [...], totalCount: n } so WithCount can use an exact $count
// instead of the limit+1 convention.
//
// Populate translation needs names the hush schema does not carry (the target
// collection and the foreign key fields). Those come from a [Translator], whose
// zero value applies the default conventions (collection = relation+"s", local
// field = relation+"_id", foreign field = "_id"). Package-level helpers use the
// default Translator; customize via a Translator value.
package mongo
