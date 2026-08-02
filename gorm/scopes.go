package gorm

import (
	"github.com/DhimasYulian/hush"
	"gorm.io/gorm"
)

// Scopes builds a GORM scope function that translates a validated hush query
// into GORM clauses. Use it inside a db.Scopes(...) chain:
//
//	db := db.Scopes(gorm.Scopes(schema, q))
//
// schema is the same *hush.Schema used to validate q, so every selected column,
// sort field, relation, and filter value is already schema-safe. nil schema or
// query yields an identity scope.
func Scopes(schema *hush.Schema, q *hush.Query) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if schema == nil || q == nil {
			return db
		}

		db = applySelect(db, db.NamingStrategy, schema, q)
		db = applyFilter(db, db.NamingStrategy, q.Filters)
		db = applySort(db, db.NamingStrategy, schema, q.Sort)
		db = applyGroupBy(db, db.NamingStrategy, schema, q.GroupBy)
		db = applyPagination(db, q.Pagination)
		db = applyPopulates(db, db.NamingStrategy, schema, q.Populates)

		return db
	}
}
