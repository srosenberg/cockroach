// Package nstree provides a data structure for storing and retrieving
// descriptors namespace entry-like data.
package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/google/btree"
)

type item interface {
	btree.Item
	put()
	value() interface{}
}

const degree = 8

var btreeSyncPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(267130)
		return btree.New(degree)
	},
}

func upsert(t *btree.BTree, toUpsert item) interface{} {
	__antithesis_instrumentation__.Notify(267131)
	if overwritten := t.ReplaceOrInsert(toUpsert); overwritten != nil {
		__antithesis_instrumentation__.Notify(267133)
		overwrittenItem := overwritten.(item)
		defer overwrittenItem.put()
		return overwrittenItem.value()
	} else {
		__antithesis_instrumentation__.Notify(267134)
	}
	__antithesis_instrumentation__.Notify(267132)
	return nil
}

func get(t *btree.BTree, k item) interface{} {
	__antithesis_instrumentation__.Notify(267135)
	defer k.put()
	if got := t.Get(k); got != nil {
		__antithesis_instrumentation__.Notify(267137)
		return got.(item).value()
	} else {
		__antithesis_instrumentation__.Notify(267138)
	}
	__antithesis_instrumentation__.Notify(267136)
	return nil
}

func delete(t *btree.BTree, k item) interface{} {
	__antithesis_instrumentation__.Notify(267139)
	defer k.put()
	if deleted, ok := t.Delete(k).(item); ok {
		__antithesis_instrumentation__.Notify(267141)
		defer deleted.put()
		return deleted.value()
	} else {
		__antithesis_instrumentation__.Notify(267142)
	}
	__antithesis_instrumentation__.Notify(267140)
	return nil
}

func clear(t *btree.BTree) {
	__antithesis_instrumentation__.Notify(267143)
	for t.Len() > 0 {
		__antithesis_instrumentation__.Notify(267144)
		t.DeleteMin().(item).put()
	}
}

func ascend(t *btree.BTree, f func(k interface{}) error) (err error) {
	__antithesis_instrumentation__.Notify(267145)
	t.Ascend(func(i btree.Item) bool {
		__antithesis_instrumentation__.Notify(267148)
		err = f(i.(item).value())
		return err == nil
	})
	__antithesis_instrumentation__.Notify(267146)
	if iterutil.Done(err) {
		__antithesis_instrumentation__.Notify(267149)
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(267150)
	}
	__antithesis_instrumentation__.Notify(267147)
	return err
}
