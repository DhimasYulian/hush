package gorm

import (
	"fmt"
	"strings"

	"github.com/DhimasYulian/hush"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// applySelect builds the SELECT clause from whitelisted fields, groupBy columns,
// and aggregate aliases. GroupBy columns are selected alongside aggregations so
// the result includes the grouped dimension.
func applySelect(db *gorm.DB, naming schema.Namer, sc *hush.Schema, q *hush.Query) *gorm.DB {
	cols, hasAgg := buildSelectColumns(naming, sc, q)
	if len(cols) == 0 {
		return db
	}

	// Aggregate expressions such as "SUM(views) AS total" are raw SQL, so they
	// must be emitted through a clause.Expr; GORM quotes plain []string columns.
	if hasAgg {
		return db.Clauses(clause.Select{Expression: clause.Expr{SQL: strings.Join(cols, ", ")}})
	}

	return db.Select(cols)
}

// buildSelectColumns returns the whitelisted select list and whether it
// contains any aggregate expression.
func buildSelectColumns(naming schema.Namer, sc *hush.Schema, q *hush.Query) (cols []string, hasAgg bool) {
	if len(q.Fields) == 0 && len(q.GroupBy) == 0 && len(q.Aggregations) == 0 {
		return nil, false
	}

	seen := make(map[string]bool, len(q.Fields)+len(q.GroupBy))

	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		cols = append(cols, name)
	}

	for _, f := range q.Fields {
		if sc.Selectable(f) {
			add(naming.ColumnName("", f))
		}
	}

	for _, g := range q.GroupBy {
		if sc.Groupable(g) {
			add(naming.ColumnName("", g))
		}
	}

	for _, a := range q.Aggregations {
		if col := aggregationSelect(naming, a); col != "" {
			cols = append(cols, col)
			hasAgg = true
		}
	}

	return cols, hasAgg
}

// aggregationSelect renders an aggregate as a SQL alias, e.g.
// "SUM(views) AS totalViews". Only the whitelisted count/sum/avg functions are
// emitted; anything else is skipped.
func aggregationSelect(naming schema.Namer, a hush.Aggregation) string {
	target := a.Field
	if target == "" {
		target = "*"
	} else {
		target = naming.ColumnName("", target)
	}

	switch strings.ToUpper(a.Func) {
	case "COUNT":
		return fmt.Sprintf("COUNT(%s) AS %s", target, a.Alias)
	case "SUM":
		return fmt.Sprintf("SUM(%s) AS %s", target, a.Alias)
	case "AVG":
		return fmt.Sprintf("AVG(%s) AS %s", target, a.Alias)
	default:
		return ""
	}
}

// applySort applies ORDER BY for sort entries whose field is sortable in the
// schema; unknown fields are skipped.
func applySort(db *gorm.DB, naming schema.Namer, sc *hush.Schema, sorts []hush.Sort) *gorm.DB {
	for _, s := range sorts {
		if len(s.Path) == 0 {
			continue
		}
		name := s.Path[0]
		if !sc.Sortable(name) {
			continue
		}

		dir := "ASC"
		if s.Direction == hush.SortDesc {
			dir = "DESC"
		}
		db = db.Order(naming.ColumnName("", name) + " " + dir)
	}
	return db
}

// applyGroupBy applies GROUP BY for whitelisted groupable columns.
func applyGroupBy(db *gorm.DB, naming schema.Namer, sc *hush.Schema, groups []hush.Field) *gorm.DB {
	for _, g := range groups {
		if sc.Groupable(g) {
			db = db.Group(naming.ColumnName("", g))
		}
	}
	return db
}

// applyPagination applies OFFSET and LIMIT. When WithCount is true and a limit
// is set, one extra row is fetched so the caller can detect additional rows by
// comparing len(rows) against the requested limit.
func applyPagination(db *gorm.DB, p hush.Pagination) *gorm.DB {
	if p.Limit == nil {
		return db
	}

	limit := *p.Limit
	if p.WithCount != nil && *p.WithCount {
		limit++
	}

	if p.Start != nil {
		db = db.Offset(*p.Start)
	}

	return db.Limit(limit)
}
