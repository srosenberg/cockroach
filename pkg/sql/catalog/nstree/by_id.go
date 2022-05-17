package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/google/btree"
)

func (t byIDMap) len() int {
	__antithesis_instrumentation__.Notify(266940)
	return t.t.Len()
}

type byIDItem struct {
	id descpb.ID
	v  interface{}
}

func makeByIDItem(d interface{ GetID() descpb.ID }) byIDItem {
	__antithesis_instrumentation__.Notify(266941)
	return byIDItem{id: d.GetID(), v: d}
}

var _ btree.Item = (*byIDItem)(nil)

func (b *byIDItem) Less(thanItem btree.Item) bool {
	__antithesis_instrumentation__.Notify(266942)
	than := thanItem.(*byIDItem)
	return b.id < than.id
}

var byIDItemPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(266943); return new(byIDItem) },
}

func (b byIDItem) get() *byIDItem {
	__antithesis_instrumentation__.Notify(266944)
	alloc := byIDItemPool.Get().(*byIDItem)
	*alloc = b
	return alloc
}

func (b *byIDItem) value() interface{} {
	__antithesis_instrumentation__.Notify(266945)
	return b.v
}

func (b *byIDItem) put() {
	__antithesis_instrumentation__.Notify(266946)
	*b = byIDItem{}
	byIDItemPool.Put(b)
}
