package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/google/btree"
)

type byNameItem struct {
	parentID, parentSchemaID descpb.ID
	name                     string
	v                        interface{}
}

func makeByNameItem(d catalog.NameKey) byNameItem {
	__antithesis_instrumentation__.Notify(266953)
	return byNameItem{
		parentID:       d.GetParentID(),
		parentSchemaID: d.GetParentSchemaID(),
		name:           d.GetName(),
		v:              d,
	}
}

var _ btree.Item = (*byNameItem)(nil)

func (b *byNameItem) Less(thanItem btree.Item) bool {
	__antithesis_instrumentation__.Notify(266954)
	than := thanItem.(*byNameItem)
	if b.parentID != than.parentID {
		__antithesis_instrumentation__.Notify(266957)
		return b.parentID < than.parentID
	} else {
		__antithesis_instrumentation__.Notify(266958)
	}
	__antithesis_instrumentation__.Notify(266955)
	if b.parentSchemaID != than.parentSchemaID {
		__antithesis_instrumentation__.Notify(266959)
		return b.parentSchemaID < than.parentSchemaID
	} else {
		__antithesis_instrumentation__.Notify(266960)
	}
	__antithesis_instrumentation__.Notify(266956)
	return b.name < than.name
}

func (b *byNameItem) value() interface{} {
	__antithesis_instrumentation__.Notify(266961)
	return b.v
}

var byNameItemPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(266962); return new(byNameItem) },
}

func (b byNameItem) get() *byNameItem {
	__antithesis_instrumentation__.Notify(266963)
	item := byNameItemPool.Get().(*byNameItem)
	*item = b
	return item
}

func (b *byNameItem) put() {
	__antithesis_instrumentation__.Notify(266964)
	*b = byNameItem{}
	byNameItemPool.Put(b)
}
