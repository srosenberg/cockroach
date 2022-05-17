package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/google/btree"
)

type ReplicaPlaceholder struct {
	rangeDesc roachpb.RangeDescriptor
	tainted   int32
}

var _ rangeKeyItem = (*ReplicaPlaceholder)(nil)

func (r *ReplicaPlaceholder) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(117807)
	return &r.rangeDesc
}

func (r *ReplicaPlaceholder) key() roachpb.RKey {
	__antithesis_instrumentation__.Notify(117808)
	return r.Desc().StartKey
}

func (r *ReplicaPlaceholder) Less(i btree.Item) bool {
	__antithesis_instrumentation__.Notify(117809)
	return r.Desc().StartKey.Less(i.(rangeKeyItem).key())
}

func (r *ReplicaPlaceholder) String() string {
	__antithesis_instrumentation__.Notify(117810)
	tainted := ""
	if atomic.LoadInt32(&r.tainted) != 0 {
		__antithesis_instrumentation__.Notify(117812)
		tainted = ",tainted"
	} else {
		__antithesis_instrumentation__.Notify(117813)
	}
	__antithesis_instrumentation__.Notify(117811)
	return fmt.Sprintf("range=%d [%s-%s) (placeholder%s)",
		r.Desc().RangeID, r.rangeDesc.StartKey, r.rangeDesc.EndKey, tainted)
}
