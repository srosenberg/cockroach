package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/google/btree"
)

type Set struct {
	t *btree.BTree
}

func (s *Set) Add(components catalog.NameKey) {
	__antithesis_instrumentation__.Notify(267114)
	s.maybeInitialize()
	item := makeByNameItem(components).get()
	item.v = item
	upsert(s.t, item)
}

func (s *Set) Contains(components catalog.NameKey) bool {
	__antithesis_instrumentation__.Notify(267115)
	if !s.initialized() {
		__antithesis_instrumentation__.Notify(267117)
		return false
	} else {
		__antithesis_instrumentation__.Notify(267118)
	}
	__antithesis_instrumentation__.Notify(267116)
	return get(s.t, makeByNameItem(components).get()) != nil
}

func (s *Set) Clear() {
	__antithesis_instrumentation__.Notify(267119)
	if !s.initialized() {
		__antithesis_instrumentation__.Notify(267121)
		return
	} else {
		__antithesis_instrumentation__.Notify(267122)
	}
	__antithesis_instrumentation__.Notify(267120)
	clear(s.t)
	btreeSyncPool.Put(s.t)
	*s = Set{}
}

func (s *Set) Empty() bool {
	__antithesis_instrumentation__.Notify(267123)
	return !s.initialized() || func() bool {
		__antithesis_instrumentation__.Notify(267124)
		return s.t.Len() == 0 == true
	}() == true
}

func (s *Set) maybeInitialize() {
	__antithesis_instrumentation__.Notify(267125)
	if s.initialized() {
		__antithesis_instrumentation__.Notify(267127)
		return
	} else {
		__antithesis_instrumentation__.Notify(267128)
	}
	__antithesis_instrumentation__.Notify(267126)
	*s = Set{
		t: btreeSyncPool.Get().(*btree.BTree),
	}
}

func (s Set) initialized() bool {
	__antithesis_instrumentation__.Notify(267129)
	return s != (Set{})
}
