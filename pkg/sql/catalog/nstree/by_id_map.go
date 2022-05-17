package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/google/btree"
)

type byIDMap struct {
	t *btree.BTree
}

func (t byIDMap) upsert(d catalog.NameEntry) (replaced catalog.NameEntry) {
	__antithesis_instrumentation__.Notify(266947)
	replaced, _ = upsert(t.t, makeByIDItem(d).get()).(catalog.NameEntry)
	return replaced
}

func (t byIDMap) get(id descpb.ID) catalog.NameEntry {
	__antithesis_instrumentation__.Notify(266948)
	got, _ := get(t.t, byIDItem{id: id}.get()).(catalog.NameEntry)
	return got
}

func (t byIDMap) delete(id descpb.ID) (removed catalog.NameEntry) {
	__antithesis_instrumentation__.Notify(266949)
	removed, _ = delete(t.t, byIDItem{id: id}.get()).(catalog.NameEntry)
	return removed
}

func (t byIDMap) clear() {
	__antithesis_instrumentation__.Notify(266950)
	clear(t.t)
}

func (t byIDMap) ascend(f EntryIterator) error {
	__antithesis_instrumentation__.Notify(266951)
	return ascend(t.t, func(k interface{}) error {
		__antithesis_instrumentation__.Notify(266952)
		return f(k.(catalog.NameEntry))
	})
}
