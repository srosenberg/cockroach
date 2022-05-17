package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type rangeIDReplicaMap syncutil.IntMap

func (m *rangeIDReplicaMap) Load(rangeID roachpb.RangeID) (*Replica, bool) {
	__antithesis_instrumentation__.Notify(125837)
	val, ok := (*syncutil.IntMap)(m).Load(int64(rangeID))
	return (*Replica)(val), ok
}

func (m *rangeIDReplicaMap) LoadOrStore(
	rangeID roachpb.RangeID, repl *Replica,
) (_ *Replica, loaded bool) {
	__antithesis_instrumentation__.Notify(125838)
	val, loaded := (*syncutil.IntMap)(m).LoadOrStore(int64(rangeID), unsafe.Pointer(repl))
	return (*Replica)(val), loaded
}

func (m *rangeIDReplicaMap) Delete(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(125839)
	(*syncutil.IntMap)(m).Delete(int64(rangeID))
}

func (m *rangeIDReplicaMap) Range(f func(*Replica)) {
	__antithesis_instrumentation__.Notify(125840)
	v := func(k int64, v unsafe.Pointer) bool {
		__antithesis_instrumentation__.Notify(125842)
		f((*Replica)(v))
		return true
	}
	__antithesis_instrumentation__.Notify(125841)
	(*syncutil.IntMap)(m).Range(v)
}
