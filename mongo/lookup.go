package mongo

import (
	"fmt"
	"strings"

	"github.com/DhimasYulian/hush"
	"github.com/DhimasYulian/hush/internal/schema"
	"go.mongodb.org/mongo-driver/bson"
)

// lookups translates populate directives into $lookup stages.
//
// Validation is eager: the whole populate tree is walked against the schema
// graph up front, so unknown relations and depth violations are reported even
// when the collection is empty.
func (t Translator) lookups(root *hush.Schema, populates []hush.Populate, prefix []string) ([]bson.D, error) {
	if len(populates) == 0 {
		return nil, nil
	}
	if root == nil {
		return nil, fmt.Errorf("hush/mongo: cannot translate populate %q without a schema", populates[0].Relation)
	}
	return t.lookupsInner(root.Inner(), populates, prefix)
}

func (t Translator) lookupsInner(target *schema.Schema, populates []hush.Populate, prefix []string) ([]bson.D, error) {
	out := make([]bson.D, 0, len(populates))
	for _, p := range populates {
		stage, err := t.lookupInner(target, p, prefix)
		if err != nil {
			return nil, err
		}
		out = append(out, stage)
	}
	return out, nil
}

// lookupInner builds a single $lookup stage for a populate entry. The related
// documents are filtered, projected, and sorted by an inner pipeline, which
// recurses into nested populates so each relation level is a nested $lookup.
func (t Translator) lookupInner(target *schema.Schema, p hush.Populate, prefix []string) (bson.D, error) {
	rel, ok := target.GetRelation(p.Relation)
	if !ok {
		return nil, fmt.Errorf("hush/mongo: unknown relation %q", p.Relation)
	}

	if len(prefix)+1 > rel.MaxDepth {
		path := append(append([]string{}, prefix...), p.Relation)
		return nil, fmt.Errorf("hush/mongo: relation %q exceeds max depth %d", strings.Join(path, "."), rel.MaxDepth)
	}

	as := p.Relation
	doc := bson.D{
		{Key: "from", Value: t.collectionName(p.Relation)},
		{Key: "localField", Value: t.localField(p.Relation)},
		{Key: "foreignField", Value: t.foreignField()},
		{Key: "as", Value: as},
	}

	relTarget := rel.Target
	pipeline := bson.A{}

	if p.Filters != nil {
		if f := t.buildFilter(p.Filters); len(f) > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$match", Value: f}})
		}
	}

	if len(p.Fields) > 0 {
		cols := t.lookupProjectFields(p, relTarget)
		if len(cols) > 0 {
			proj := bson.M{}
			for _, c := range cols {
				proj[c] = 1
			}
			pipeline = append(pipeline, bson.D{{Key: "$project", Value: proj}})
		}
	}

	for _, s := range p.Sorts {
		if len(s.Path) == 0 {
			continue
		}
		name := s.Path[0]
		if !relTarget.GetSortable(name) {
			continue
		}
		dir := int32(1)
		if s.Direction == hush.SortDesc {
			dir = -1
		}
		pipeline = append(pipeline, bson.D{{Key: "$sort", Value: bson.D{{Key: t.fieldName(name), Value: dir}}}})
	}

	if len(p.Populates) > 0 {
		children, err := t.lookupsInner(relTarget, p.Populates, append(prefix, p.Relation))
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			pipeline = append(pipeline, child)
		}
	}

	if len(pipeline) > 0 {
		doc = append(doc, bson.E{Key: "pipeline", Value: pipeline})
	}

	return bson.D{{Key: "$lookup", Value: doc}}, nil
}

// lookupProjectFields returns the whitelisted selectable fields plus the join
// fields that must survive projection for the $lookup to reconstruct the
// relation: the foreign field (matched against the parent) and the local field
// of every immediate nested populate (matched by the next $lookup level).
func (t Translator) lookupProjectFields(p hush.Populate, target *schema.Schema) []string {
	cols := make([]string, 0, len(p.Fields)+len(p.Populates)+1)
	seen := make(map[string]bool, cap(cols))

	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		cols = append(cols, name)
	}

	for _, f := range p.Fields {
		if target.GetSelectable(f) {
			add(t.fieldName(f))
		}
	}

	add(t.foreignField())
	for _, child := range p.Populates {
		add(t.localField(child.Relation))
	}

	return cols
}
