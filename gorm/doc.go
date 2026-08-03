// Package gorm translates validated hush queries into GORM clauses so a
// consumer can integrate hush with gorm by importing hush/gorm and using
// db.Scopes without writing any per-operator code.
//
// # Usage
//
//	gorm.Scopes(schema, q)
//
// builds a [GORM scope](https://gorm.io/docs/scopes.html) from a hush schema
// and a validated [hush.Query]:
//
//	db := db.Scopes(gorm.Scopes(schema, q))
//
// The scope applies, in order:
//
//   - SELECT — whitelisted [hush.Query.Fields], [hush.Query.GroupBy] columns,
//     and [hush.Query.Aggregations] rendered as aggregate aliases.
//   - WHERE — the full [hush.Query.Filters] tree. Every hush operator maps to a
//     GORM clause expression and values are bound with their schema-declared
//     type (already coerced by [hush.Parse]). $null and $notNull render as
//     `? IS NULL` / `? IS NOT NULL`. Pattern operators escape LIKE wildcards via
//     [hush.EscapeLike]. Logical $and/$or/$not are grouped so trees like
//     (a AND b) OR (c AND d) stay correct.
//   - ORDER BY — [hush.Query.Sort], skipping columns not sortable in the schema.
//   - GROUP BY — [hush.Query.GroupBy].
//   - LIMIT / OFFSET — [hush.Query.Pagination]. When WithCount is true and a
//     limit is set, the scope fetches limit+1 rows so the caller can detect
//     whether additional rows exist (len(rows) > limit).
//   - PRELOAD — [hush.Query.Populates] via db.Preload with whitelisted Select,
//     Order, and translated Where, enforcing each relation's max depth.
//
// Translation errors are recorded on the statement via db.AddError and surface
// on db.Error when the query runs, matching GORM conventions.
package gorm
