package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
	"github.com/google/btree"
)

type Database struct {
	schema *Schema

	indexes []index

	entities map[interface{}]*entity
}

func (t *Database) Schema() *Schema {
	__antithesis_instrumentation__.Notify(578369)
	return t.schema
}

func NewDatabase(sc *Schema, indexes [][]Attr) (*Database, error) {
	__antithesis_instrumentation__.Notify(578370)
	t := &Database{
		schema:   sc,
		indexes:  make([]index, len(indexes)+1),
		entities: make(map[interface{}]*entity),
	}

	const degree = 8
	fl := btree.NewFreeList(len(indexes) + 1)
	{
		__antithesis_instrumentation__.Notify(578373)
		var primaryIndex index
		primaryIndex.s = sc
		primaryIndex.tree = btree.NewWithFreeList(degree, fl)
		t.indexes[0] = primaryIndex
	}
	__antithesis_instrumentation__.Notify(578371)
	secondaryIndexes := t.indexes[1:]
	for i, attrs := range indexes {
		__antithesis_instrumentation__.Notify(578374)
		var set ordinalSet
		ords := make([]ordinal, len(attrs))
		for i, a := range attrs {
			__antithesis_instrumentation__.Notify(578376)
			ord, err := sc.getOrdinal(a)
			if err != nil {
				__antithesis_instrumentation__.Notify(578378)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(578379)
			}
			__antithesis_instrumentation__.Notify(578377)
			set = set.add(ord)
			ords[i] = ord
		}
		__antithesis_instrumentation__.Notify(578375)
		spec := indexSpec{mask: set, attrs: ords, s: sc}
		secondaryIndexes[i] = index{
			indexSpec: spec,
			tree:      btree.NewWithFreeList(degree, fl),
		}
	}
	__antithesis_instrumentation__.Notify(578372)
	return t, nil
}

func (t *Database) Insert(v interface{}) error {
	__antithesis_instrumentation__.Notify(578380)
	e, err := toEntity(t.schema, v)
	if err != nil {
		__antithesis_instrumentation__.Notify(578384)
		return err
	} else {
		__antithesis_instrumentation__.Notify(578385)
	}
	__antithesis_instrumentation__.Notify(578381)
	if err := t.insert(e); err != nil {
		__antithesis_instrumentation__.Notify(578386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(578387)
	}
	__antithesis_instrumentation__.Notify(578382)
	for _, v := range e.m {
		__antithesis_instrumentation__.Notify(578388)
		_, isEntity := t.schema.entityTypeSchemas[reflect.TypeOf(v)]
		_, alreadyDefined := t.entities[v]
		if isEntity && func() bool {
			__antithesis_instrumentation__.Notify(578389)
			return !alreadyDefined == true
		}() == true {
			__antithesis_instrumentation__.Notify(578390)
			if err := t.Insert(v); err != nil {
				__antithesis_instrumentation__.Notify(578391)
				return err
			} else {
				__antithesis_instrumentation__.Notify(578392)
			}
		} else {
			__antithesis_instrumentation__.Notify(578393)
		}
	}
	__antithesis_instrumentation__.Notify(578383)
	return nil
}

func (t *Database) insert(e *entity) error {
	__antithesis_instrumentation__.Notify(578394)
	self := e.getComparableValue(t.schema, Self)
	if existing, exists := t.entities[self]; exists {
		__antithesis_instrumentation__.Notify(578397)

		if _, eq := compareEntities(e, existing); !eq {
			__antithesis_instrumentation__.Notify(578399)
			return errors.AssertionFailedf(
				"expected entity %v to equal its already inserted value", self,
			)
		} else {
			__antithesis_instrumentation__.Notify(578400)
		}
		__antithesis_instrumentation__.Notify(578398)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(578401)
	}
	__antithesis_instrumentation__.Notify(578395)
	t.entities[self] = e
	for i := range t.indexes {
		__antithesis_instrumentation__.Notify(578402)
		idx := &t.indexes[i]
		if g := idx.tree.ReplaceOrInsert(&containerItem{
			entity:    e,
			indexSpec: &idx.indexSpec,
		}); g != nil {
			__antithesis_instrumentation__.Notify(578403)
			return errors.AssertionFailedf(
				"expected entity %T(%v) to not exist", self, self,
			)
		} else {
			__antithesis_instrumentation__.Notify(578404)
		}
	}
	__antithesis_instrumentation__.Notify(578396)
	return nil
}

type index struct {
	indexSpec
	tree *btree.BTree
}

type indexSpec struct {
	s     *Schema
	mask  ordinalSet
	attrs []ordinal
}

type entityIterator interface {
	visit(*entity) error
}

func (t *Database) iterate(where *valuesMap, f entityIterator) (err error) {
	__antithesis_instrumentation__.Notify(578405)
	idx, toCheck := t.chooseIndex(where.attrs)
	from, to := getValuesItems(&idx.indexSpec, where, where.attrs)
	defer putValuesItems(from, to)
	idx.tree.AscendRange(from, to, func(i btree.Item) (wantMore bool) {
		__antithesis_instrumentation__.Notify(578408)
		c := i.(*containerItem)

		if where.attrs.without(c.entity.attrs) != 0 {
			__antithesis_instrumentation__.Notify(578412)
			return true
		} else {
			__antithesis_instrumentation__.Notify(578413)
		}
		__antithesis_instrumentation__.Notify(578409)
		var failed bool
		toCheck.forEach(func(a ordinal) (wantMore bool) {
			__antithesis_instrumentation__.Notify(578414)
			_, eq := compareOn(a, (*valuesMap)(c.entity), where)
			failed = !eq
			return !failed
		})
		__antithesis_instrumentation__.Notify(578410)
		if !failed {
			__antithesis_instrumentation__.Notify(578415)
			err = f.visit(c.entity)
		} else {
			__antithesis_instrumentation__.Notify(578416)
		}
		__antithesis_instrumentation__.Notify(578411)
		return err == nil
	})
	__antithesis_instrumentation__.Notify(578406)
	if iterutil.Done(err) {
		__antithesis_instrumentation__.Notify(578417)
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(578418)
	}
	__antithesis_instrumentation__.Notify(578407)
	return err
}

func (t *Database) chooseIndex(m ordinalSet) (_ *index, toCheck ordinalSet) {
	__antithesis_instrumentation__.Notify(578419)

	best, bestOverlap := 0, ordinalSet(0)
	dims := t.indexes[1:]
	for i := range dims {
		__antithesis_instrumentation__.Notify(578421)
		if overlap := dims[i].overlap(m); overlap.len() > bestOverlap.len() {
			__antithesis_instrumentation__.Notify(578422)
			best, bestOverlap = i+1, overlap
		} else {
			__antithesis_instrumentation__.Notify(578423)
		}
	}
	__antithesis_instrumentation__.Notify(578420)
	return &t.indexes[best], m.without(bestOverlap)
}

func (s *indexSpec) overlap(m ordinalSet) ordinalSet {
	__antithesis_instrumentation__.Notify(578424)
	var overlap ordinalSet
	for _, a := range s.attrs {
		__antithesis_instrumentation__.Notify(578426)
		if m.contains(a) {
			__antithesis_instrumentation__.Notify(578427)
			overlap = overlap.add(a)
		} else {
			__antithesis_instrumentation__.Notify(578428)
			break
		}
	}
	__antithesis_instrumentation__.Notify(578425)
	return overlap
}
