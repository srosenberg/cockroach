package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/google/btree"
)

type byNameMap struct {
	t *btree.BTree
}

func (t byNameMap) upsert(d catalog.NameEntry) (replaced catalog.NameEntry) {
	__antithesis_instrumentation__.Notify(266965)
	replaced, _ = upsert(t.t, makeByNameItem(d).get()).(catalog.NameEntry)
	return replaced
}

func (t byNameMap) getByName(parentID, parentSchemaID descpb.ID, name string) catalog.NameEntry {
	__antithesis_instrumentation__.Notify(266966)
	got, _ := get(t.t, byNameItem{
		parentID:       parentID,
		parentSchemaID: parentSchemaID,
		name:           name,
	}.get()).(catalog.NameEntry)
	return got
}

func (t byNameMap) delete(d catalog.NameKey) {
	__antithesis_instrumentation__.Notify(266967)
	delete(t.t, makeByNameItem(d).get())
}

func (t byNameMap) clear() {
	__antithesis_instrumentation__.Notify(266968)
	clear(t.t)
}

func (t byNameMap) ascend(f EntryIterator) error {
	__antithesis_instrumentation__.Notify(266969)
	return ascend(t.t, func(k interface{}) error {
		__antithesis_instrumentation__.Notify(266970)
		return f(k.(catalog.NameEntry))
	})
}
