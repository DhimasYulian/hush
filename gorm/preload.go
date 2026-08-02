package gorm

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/DhimasYulian/hush"
	"github.com/DhimasYulian/hush/internal/schema"
	"gorm.io/gorm"
	gormschema "gorm.io/gorm/schema"
)

// RelationName converts a hush relation name to the GORM relationship name it
// maps to. GORM keys relationships by the struct field name (e.g. "Author"),
// while hush schemas conventionally declare relations in lowercase
// ("author"). The default mapping CamelCases the hush name; a consumer whose
// GORM field name differs can rename the schema relation to match.
func RelationName(name string) string {
	var sb strings.Builder
	for _, part := range strings.FieldsFunc(name, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	}) {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		sb.WriteString(string(runes))
	}
	return sb.String()
}

// applyPopulates translates populate directives into db.Preload chains.
//
// Validation is eager: the whole populate tree is walked against the schema
// graph at scope time, so unknown relations and depth violations surface even
// when the table is empty. Registration happens per node so each level can
// carry its own Select/Order/filter.
func applyPopulates(db *gorm.DB, naming gormschema.Namer, root *hush.Schema, populates []hush.Populate) *gorm.DB {
	if len(populates) == 0 {
		return db
	}

	validatePopulates(db, root.Inner(), populates, nil)

	for _, p := range populates {
		db = applyPreload(db, naming, root.Inner(), p, nil)
	}

	return db
}

// validatePopulates walks a populate list against the target schema graph,
// reporting unknown relations and max-depth violations to the given db.
func validatePopulates(db *gorm.DB, target *schema.Schema, populates []hush.Populate, prefix []string) {
	for _, p := range populates {
		validatePopulate(db, target, p, prefix)
	}
}

func validatePopulate(db *gorm.DB, target *schema.Schema, p hush.Populate, prefix []string) {
	rel, ok := target.GetRelation(p.Relation)
	if !ok {
		_ = db.AddError(fmt.Errorf("hush/gorm: unknown relation %q", p.Relation))
		return
	}

	if len(prefix)+1 > rel.MaxDepth {
		path := append(append([]string{}, prefix...), p.Relation)
		_ = db.AddError(fmt.Errorf("hush/gorm: relation %q exceeds max depth %d", strings.Join(path, "."), rel.MaxDepth))
		return
	}

	validatePopulates(db, rel.Target, p.Populates, append(append([]string{}, prefix...), p.Relation))
}

// applyPreload registers one Preload for a populate entry. The preload callback
// applies whitelisted Select/Order and the translated filter, and recurses into
// nested populates so GORM receives a single Preload per relation path.
//
// The preload name is always relative (e.g. "Author", "Profile") because a
// callback runs with g already scoped to the parent model; GORM groups the
// nested calls under the parent preload automatically.
func applyPreload(db *gorm.DB, naming gormschema.Namer, target *schema.Schema, p hush.Populate, prefix []string) *gorm.DB {
	rel, ok := target.GetRelation(p.Relation)
	if !ok {
		return db
	}

	name := RelationName(p.Relation)
	relTarget := rel.Target

	preloadFn := func(g *gorm.DB) *gorm.DB {
		if len(p.Fields) > 0 {
			g = preloadSelect(g, naming, relTarget, p.Fields, preloadJoinColumns(g.Statement.Schema, name))
		}

		for _, s := range p.Sorts {
			if len(s.Path) == 0 {
				continue
			}
			name := s.Path[0]
			if !relTarget.GetSortable(name) {
				continue
			}
			dir := "ASC"
			if s.Direction == hush.SortDesc {
				dir = "DESC"
			}
			g = g.Order(naming.ColumnName("", name) + " " + dir)
		}

		if p.Filters != nil {
			expr, err := buildExpr(p.Filters, naming)
			if err != nil {
				_ = g.AddError(err)
			} else {
				g = g.Where(expr)
			}
		}

		for _, child := range p.Populates {
			g = applyPreload(g, naming, relTarget, child, append(prefix, p.Relation))
		}

		return g
	}

	return db.Preload(name, preloadFn)
}

// preloadSelect restricts a preload to whitelisted selectable columns. GORM's
// join columns (joinCols) are always included because they are required to
// reconstruct the association after the restricted projection.
func preloadSelect(g *gorm.DB, naming gormschema.Namer, t *schema.Schema, fields []hush.Field, joinCols []string) *gorm.DB {
	cols := make([]string, 0, len(fields)+len(joinCols))
	seen := make(map[string]bool, cap(cols))

	for _, f := range fields {
		if !t.GetSelectable(f) {
			continue
		}
		col := naming.ColumnName("", f)
		if !seen[col] {
			seen[col] = true
			cols = append(cols, col)
		}
	}

	for _, col := range joinCols {
		if col == "" || seen[col] {
			continue
		}
		seen[col] = true
		cols = append(cols, col)
	}

	if len(cols) == 0 {
		return g
	}

	return g.Select(cols)
}

// preloadJoinColumns returns the columns on the preload target table that GORM
// reads to reconstruct the association: the target's foreign key for
// has-one/has-many, the target's primary key for belongs-to. The parent schema
// carries the relationship definition keyed by the GORM field name.
func preloadJoinColumns(parent *gormschema.Schema, relName string) []string {
	if parent == nil {
		return nil
	}

	rel, ok := parent.Relationships.Relations[relName]
	if !ok {
		return nil
	}

	var cols []string
	for _, ref := range rel.References {
		if ref.PrimaryKey == nil || ref.ForeignKey == nil {
			continue
		}
		if rel.JoinTable != nil {
			cols = append(cols, ref.PrimaryKey.DBName)
			continue
		}
		if ref.OwnPrimaryKey {
			cols = append(cols, ref.ForeignKey.DBName)
		} else {
			cols = append(cols, ref.PrimaryKey.DBName)
		}
	}
	return cols
}
