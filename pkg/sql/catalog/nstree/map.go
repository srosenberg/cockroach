package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/google/btree"
)

type Map struct {
	byID   byIDMap
	byName byNameMap
}

type EntryIterator func(entry catalog.NameEntry) error

func (dt *Map) Upsert(d catalog.NameEntry) {
	__antithesis_instrumentation__.Notify(267075)
	dt.maybeInitialize()
	if replaced := dt.byName.upsert(d); replaced != nil {
		__antithesis_instrumentation__.Notify(267077)
		dt.byID.delete(replaced.GetID())
	} else {
		__antithesis_instrumentation__.Notify(267078)
	}
	__antithesis_instrumentation__.Notify(267076)
	if replaced := dt.byID.upsert(d); replaced != nil {
		__antithesis_instrumentation__.Notify(267079)
		dt.byName.delete(replaced)
	} else {
		__antithesis_instrumentation__.Notify(267080)
	}
}

func (dt *Map) Remove(id descpb.ID) catalog.NameEntry {
	__antithesis_instrumentation__.Notify(267081)
	dt.maybeInitialize()
	if d := dt.byID.delete(id); d != nil {
		__antithesis_instrumentation__.Notify(267083)
		dt.byName.delete(d)
		return d
	} else {
		__antithesis_instrumentation__.Notify(267084)
	}
	__antithesis_instrumentation__.Notify(267082)
	return nil
}

func (dt *Map) GetByID(id descpb.ID) catalog.NameEntry {
	__antithesis_instrumentation__.Notify(267085)
	if !dt.initialized() {
		__antithesis_instrumentation__.Notify(267087)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267088)
	}
	__antithesis_instrumentation__.Notify(267086)
	return dt.byID.get(id)
}

func (dt *Map) GetByName(parentID, parentSchemaID descpb.ID, name string) catalog.NameEntry {
	__antithesis_instrumentation__.Notify(267089)
	if !dt.initialized() {
		__antithesis_instrumentation__.Notify(267091)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267092)
	}
	__antithesis_instrumentation__.Notify(267090)
	return dt.byName.getByName(parentID, parentSchemaID, name)
}

func (dt *Map) Clear() {
	__antithesis_instrumentation__.Notify(267093)
	if !dt.initialized() {
		__antithesis_instrumentation__.Notify(267095)
		return
	} else {
		__antithesis_instrumentation__.Notify(267096)
	}
	__antithesis_instrumentation__.Notify(267094)
	dt.byID.clear()
	dt.byName.clear()
	btreeSyncPool.Put(dt.byName.t)
	btreeSyncPool.Put(dt.byID.t)
	*dt = Map{}
}

func (dt *Map) IterateByID(f EntryIterator) error {
	__antithesis_instrumentation__.Notify(267097)
	if !dt.initialized() {
		__antithesis_instrumentation__.Notify(267099)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267100)
	}
	__antithesis_instrumentation__.Notify(267098)
	return dt.byID.ascend(f)
}

func (dt *Map) IterateByName(f EntryIterator) error {
	__antithesis_instrumentation__.Notify(267101)
	if !dt.initialized() {
		__antithesis_instrumentation__.Notify(267103)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267104)
	}
	__antithesis_instrumentation__.Notify(267102)
	return dt.byName.ascend(f)
}

func (dt *Map) Len() int {
	__antithesis_instrumentation__.Notify(267105)
	if !dt.initialized() {
		__antithesis_instrumentation__.Notify(267107)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(267108)
	}
	__antithesis_instrumentation__.Notify(267106)
	return dt.byID.len()
}

func (dt Map) initialized() bool {
	__antithesis_instrumentation__.Notify(267109)
	return dt != (Map{})
}

func (dt *Map) maybeInitialize() {
	__antithesis_instrumentation__.Notify(267110)
	if dt.initialized() {
		__antithesis_instrumentation__.Notify(267112)
		return
	} else {
		__antithesis_instrumentation__.Notify(267113)
	}
	__antithesis_instrumentation__.Notify(267111)
	*dt = Map{
		byName: byNameMap{t: btreeSyncPool.Get().(*btree.BTree)},
		byID:   byIDMap{t: btreeSyncPool.Get().(*btree.BTree)},
	}
}
