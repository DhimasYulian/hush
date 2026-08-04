package mongo

import (
	"github.com/DhimasYulian/hush"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Projection returns a projection document from whitelisted selectable fields.
// When fields are requested, _id is excluded so results contain exactly the
// selected columns (SQL "SELECT fields" parity). It returns nil when no fields
// are requested.
func (t Translator) Projection(schema *hush.Schema, q *hush.Query) bson.M {
	if schema == nil || q == nil || len(q.Fields) == 0 {
		return nil
	}

	proj := bson.M{"_id": 0}
	for _, f := range q.Fields {
		if schema.Selectable(f) {
			proj[t.fieldName(f)] = 1
		}
	}
	return proj
}

// Sort returns the ordered sort document for sortable fields; unknown fields
// are skipped.
func (t Translator) Sort(schema *hush.Schema, q *hush.Query) bson.D {
	if schema == nil || q == nil {
		return nil
	}

	sort := make(bson.D, 0, len(q.Sort))
	for _, s := range q.Sort {
		if len(s.Path) == 0 {
			continue
		}
		name := s.Path[0]
		if !schema.Sortable(name) {
			continue
		}
		dir := int32(1)
		if s.Direction == hush.SortDesc {
			dir = -1
		}
		sort = append(sort, bson.E{Key: t.fieldName(name), Value: dir})
	}
	return sort
}

// Skip returns the pagination start offset, or nil when unset.
func (t Translator) Skip(q *hush.Query) *int64 {
	if q == nil || q.Pagination.Start == nil {
		return nil
	}
	start := int64(*q.Pagination.Start)
	return &start
}

// Limit returns the pagination limit. When WithCount is true and a limit is
// set, one extra document is requested so callers can detect a next page by
// comparing len(docs) against the requested limit, matching hush/gorm.
func (t Translator) Limit(q *hush.Query) *int64 {
	return t.limit(q, true)
}

func (t Translator) limit(q *hush.Query, bump bool) *int64 {
	if q == nil || q.Pagination.Limit == nil {
		return nil
	}
	limit := int64(*q.Pagination.Limit)
	if bump && q.Pagination.WithCount != nil && *q.Pagination.WithCount {
		limit++
	}
	return &limit
}

// FindOptions bundles projection, sort, skip, and limit into find options. It
// returns nil when the query needs none of them.
func (t Translator) FindOptions(schema *hush.Schema, q *hush.Query) *options.FindOptions {
	if q == nil {
		return nil
	}

	var (
		proj = t.Projection(schema, q)
		sort = t.Sort(schema, q)
		skip = t.Skip(q)
		lim  = t.Limit(q)
	)
	if len(proj) == 0 && len(sort) == 0 && skip == nil && lim == nil {
		return nil
	}

	opts := options.Find()
	if len(proj) > 0 {
		opts.SetProjection(proj)
	}
	if len(sort) > 0 {
		opts.SetSort(sort)
	}
	if skip != nil {
		opts.SetSkip(*skip)
	}
	if lim != nil {
		opts.SetLimit(*lim)
	}
	return opts
}
