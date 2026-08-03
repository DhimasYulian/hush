package mongo

import (
	"fmt"
	"regexp"

	"github.com/DhimasYulian/hush"
	"go.mongodb.org/mongo-driver/bson"
)

// Filter returns the bson filter document for the query's filter tree, or nil
// when the query carries no filters.
func (t Translator) Filter(q *hush.Query) bson.M {
	if q == nil || q.Filters == nil {
		return nil
	}
	return t.buildFilter(q.Filters)
}

// buildFilter translates a filter tree node into a bson document. And/Or groups
// always render as explicit $and/$or arrays so conditions on the same field
// never collide on a shared bson.M key.
func (t Translator) buildFilter(f hush.Filter) bson.M {
	switch n := f.(type) {
	case hush.Condition:
		return t.buildCondition(n)
	case hush.And:
		return t.buildGroup("$and", n.Filters)
	case hush.Or:
		return t.buildGroup("$or", n.Filters)
	case hush.Not:
		return bson.M{"$nor": bson.A{t.buildFilter(n.Filter)}}
	default:
		return nil
	}
}

// buildGroup combines child filters with the given logical operator, dropping
// empty children and flattening single-child groups.
func (t Translator) buildGroup(op string, filters []hush.Filter) bson.M {
	docs := bson.A{}
	for _, f := range filters {
		if d := t.buildFilter(f); len(d) > 0 {
			docs = append(docs, d)
		}
	}
	switch len(docs) {
	case 0:
		return nil
	case 1:
		return docs[0].(bson.M)
	default:
		return bson.M{op: docs}
	}
}

// buildCondition translates a leaf condition into a bson document. Values come
// from the coerced Condition.Values produced by hush.Parse, so numbers, bools,
// and dates bind with their real types rather than raw strings.
func (t Translator) buildCondition(c hush.Condition) bson.M {
	if len(c.Path) != 1 {
		return nil
	}

	field := t.fieldName(c.Path[0])

	switch c.Operator {
	case hush.OpEq:
		return bson.M{field: t.valueAt(c, 0)}
	case hush.OpNe:
		return bson.M{field: bson.M{"$ne": t.valueAt(c, 0)}}
	case hush.OpGt:
		return bson.M{field: bson.M{"$gt": t.valueAt(c, 0)}}
	case hush.OpGte:
		return bson.M{field: bson.M{"$gte": t.valueAt(c, 0)}}
	case hush.OpLt:
		return bson.M{field: bson.M{"$lt": t.valueAt(c, 0)}}
	case hush.OpLte:
		return bson.M{field: bson.M{"$lte": t.valueAt(c, 0)}}

	case hush.OpIn:
		return bson.M{field: bson.M{"$in": t.allValues(c)}}
	case hush.OpNotIn:
		return bson.M{field: bson.M{"$nin": t.allValues(c)}}

	case hush.OpBetween:
		return bson.M{field: bson.M{"$gte": t.valueAt(c, 0), "$lte": t.valueAt(c, 1)}}

	// LIKE patterns map to $regex. QuoteMeta makes the user's literal text
	// match literally instead of as a regex pattern.
	case hush.OpContains:
		return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(t.stringValue(c))}}
	case hush.OpContainsi:
		return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(t.stringValue(c)), "$options": "i"}}
	case hush.OpStartsWith:
		return bson.M{field: bson.M{"$regex": "^" + regexp.QuoteMeta(t.stringValue(c))}}
	case hush.OpEndsWith:
		return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(t.stringValue(c)) + "$"}}

	case hush.OpNull:
		return bson.M{field: nil}
	case hush.OpNotNull:
		return bson.M{field: bson.M{"$ne": nil}}

	default:
		return nil
	}
}

// valueAt returns the i-th value of a condition, preferring the type-coerced
// Condition.Values and falling back to coercing the raw string for hand-built
// queries that skipped hush.Parse.
func (t Translator) valueAt(c hush.Condition, i int) any {
	if i < len(c.Values) {
		return c.Values[i]
	}
	if i < len(c.Value) {
		if v, err := hush.Coerce(c.FieldType, c.Value[i]); err == nil {
			return v
		}
		return c.Value[i]
	}
	return nil
}

// allValues returns every value of a condition, coercing raw strings when the
// query was not produced by hush.Parse.
func (t Translator) allValues(c hush.Condition) []any {
	if len(c.Values) > 0 {
		return c.Values
	}
	out := make([]any, len(c.Value))
	for i := range c.Value {
		out[i] = t.valueAt(c, i)
	}
	return out
}

// stringValue returns the condition's first value formatted as a string, for
// use with the $regex pattern operators.
func (t Translator) stringValue(c hush.Condition) string {
	return fmt.Sprintf("%v", t.valueAt(c, 0))
}
