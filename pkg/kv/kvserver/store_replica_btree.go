package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/google/btree"
)

type IterationOrder int

const (
	AscendingKeyOrder  = IterationOrder(-1)
	DescendingKeyOrder = IterationOrder(1)
)

type replicaOrPlaceholder struct {
	item rangeKeyItem
	repl *Replica
	ph   *ReplicaPlaceholder
}

func (it *replicaOrPlaceholder) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(125759)
	if it.repl != nil {
		__antithesis_instrumentation__.Notify(125761)
		return it.repl.Desc()
	} else {
		__antithesis_instrumentation__.Notify(125762)
	}
	__antithesis_instrumentation__.Notify(125760)
	return it.ph.Desc()
}

type storeReplicaBTree btree.BTree

func newStoreReplicaBTree() *storeReplicaBTree {
	__antithesis_instrumentation__.Notify(125763)
	return (*storeReplicaBTree)(btree.New(64))
}

func (b *storeReplicaBTree) LookupReplica(ctx context.Context, key roachpb.RKey) *Replica {
	__antithesis_instrumentation__.Notify(125764)
	var repl *Replica
	b.mustDescendLessOrEqual(ctx, key, func(_ context.Context, it replicaOrPlaceholder) error {
		__antithesis_instrumentation__.Notify(125767)
		if it.repl != nil {
			__antithesis_instrumentation__.Notify(125769)
			repl = it.repl
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(125770)
		}
		__antithesis_instrumentation__.Notify(125768)
		return nil
	})
	__antithesis_instrumentation__.Notify(125765)
	if repl == nil || func() bool {
		__antithesis_instrumentation__.Notify(125771)
		return !repl.Desc().ContainsKey(key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(125772)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(125773)
	}
	__antithesis_instrumentation__.Notify(125766)
	return repl
}

func (b *storeReplicaBTree) LookupPrecedingReplica(ctx context.Context, key roachpb.RKey) *Replica {
	__antithesis_instrumentation__.Notify(125774)
	var repl *Replica
	b.mustDescendLessOrEqual(ctx, key, func(_ context.Context, it replicaOrPlaceholder) error {
		__antithesis_instrumentation__.Notify(125776)
		if it.repl != nil && func() bool {
			__antithesis_instrumentation__.Notify(125778)
			return !it.repl.ContainsKey(key.AsRawKey()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(125779)
			repl = it.repl
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(125780)
		}
		__antithesis_instrumentation__.Notify(125777)
		return nil
	})
	__antithesis_instrumentation__.Notify(125775)
	return repl
}

func (b *storeReplicaBTree) VisitKeyRange(
	ctx context.Context,
	startKey, endKey roachpb.RKey,
	order IterationOrder,
	visitor func(context.Context, replicaOrPlaceholder) error,
) error {
	__antithesis_instrumentation__.Notify(125781)
	if endKey.Less(startKey) {
		__antithesis_instrumentation__.Notify(125785)
		return errors.AssertionFailedf("endKey < startKey (%s < %s)", endKey, startKey)
	} else {
		__antithesis_instrumentation__.Notify(125786)
	}
	__antithesis_instrumentation__.Notify(125782)

	b.mustDescendLessOrEqual(ctx, startKey,
		func(_ context.Context, it replicaOrPlaceholder) error {
			__antithesis_instrumentation__.Notify(125787)
			desc := it.Desc()
			if startKey.Less(desc.EndKey) {
				__antithesis_instrumentation__.Notify(125789)

				startKey = it.item.key()
			} else {
				__antithesis_instrumentation__.Notify(125790)

				_ = startKey
			}
			__antithesis_instrumentation__.Notify(125788)
			return iterutil.StopIteration()
		})
	__antithesis_instrumentation__.Notify(125783)

	if order == AscendingKeyOrder {
		__antithesis_instrumentation__.Notify(125791)
		return b.ascendRange(ctx, startKey, endKey, visitor)
	} else {
		__antithesis_instrumentation__.Notify(125792)
	}
	__antithesis_instrumentation__.Notify(125784)

	return b.descendLessOrEqual(ctx, endKey,
		func(ctx context.Context, it replicaOrPlaceholder) error {
			__antithesis_instrumentation__.Notify(125793)
			if it.item.key().Equal(endKey) {
				__antithesis_instrumentation__.Notify(125796)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(125797)
			}
			__antithesis_instrumentation__.Notify(125794)
			desc := it.Desc()

			if desc.EndKey.Compare(startKey) <= 0 {
				__antithesis_instrumentation__.Notify(125798)

				return iterutil.StopIteration()
			} else {
				__antithesis_instrumentation__.Notify(125799)
			}
			__antithesis_instrumentation__.Notify(125795)
			return visitor(ctx, it)
		})
}

func (b *storeReplicaBTree) DeleteReplica(ctx context.Context, repl *Replica) replicaOrPlaceholder {
	__antithesis_instrumentation__.Notify(125800)
	return itemToReplicaOrPlaceholder(b.bt().Delete((*btreeReplica)(repl)))
}

func (b *storeReplicaBTree) DeletePlaceholder(
	ctx context.Context, ph *ReplicaPlaceholder,
) replicaOrPlaceholder {
	__antithesis_instrumentation__.Notify(125801)
	return itemToReplicaOrPlaceholder(b.bt().Delete(ph))
}

func (b *storeReplicaBTree) ReplaceOrInsertReplica(
	ctx context.Context, repl *Replica,
) replicaOrPlaceholder {
	__antithesis_instrumentation__.Notify(125802)
	return itemToReplicaOrPlaceholder(b.bt().ReplaceOrInsert((*btreeReplica)(repl)))
}

func (b *storeReplicaBTree) ReplaceOrInsertPlaceholder(
	ctx context.Context, ph *ReplicaPlaceholder,
) replicaOrPlaceholder {
	__antithesis_instrumentation__.Notify(125803)
	return itemToReplicaOrPlaceholder(b.bt().ReplaceOrInsert(ph))
}

func (b *storeReplicaBTree) mustDescendLessOrEqual(
	ctx context.Context,
	startKey roachpb.RKey,
	visitor func(context.Context, replicaOrPlaceholder) error,
) {
	__antithesis_instrumentation__.Notify(125804)
	if err := b.descendLessOrEqual(ctx, startKey, visitor); err != nil {
		__antithesis_instrumentation__.Notify(125805)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(125806)
	}
}

func (b *storeReplicaBTree) descendLessOrEqual(
	ctx context.Context,
	startKey roachpb.RKey,
	visitor func(context.Context, replicaOrPlaceholder) error,
) error {
	__antithesis_instrumentation__.Notify(125807)
	var err error
	b.bt().DescendLessOrEqual(rangeBTreeKey(startKey), func(it btree.Item) (more bool) {
		__antithesis_instrumentation__.Notify(125810)
		err = visitor(ctx, itemToReplicaOrPlaceholder(it))
		return err == nil
	})
	__antithesis_instrumentation__.Notify(125808)
	if iterutil.Done(err) {
		__antithesis_instrumentation__.Notify(125811)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(125812)
	}
	__antithesis_instrumentation__.Notify(125809)
	return err
}

func (b *storeReplicaBTree) ascendRange(
	ctx context.Context,
	startKey, endKey roachpb.RKey,
	visitor func(context.Context, replicaOrPlaceholder) error,
) error {
	__antithesis_instrumentation__.Notify(125813)
	var err error
	b.bt().AscendRange(rangeBTreeKey(startKey), rangeBTreeKey(endKey), func(it btree.Item) (more bool) {
		__antithesis_instrumentation__.Notify(125816)
		err = visitor(ctx, itemToReplicaOrPlaceholder(it))
		return err == nil
	})
	__antithesis_instrumentation__.Notify(125814)
	if iterutil.Done(err) {
		__antithesis_instrumentation__.Notify(125817)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(125818)
	}
	__antithesis_instrumentation__.Notify(125815)
	return err
}

func (b *storeReplicaBTree) bt() *btree.BTree {
	__antithesis_instrumentation__.Notify(125819)
	return (*btree.BTree)(b)
}

func (b *storeReplicaBTree) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(125820)
	var i int
	b.bt().Ascend(func(bi btree.Item) (more bool) {
		__antithesis_instrumentation__.Notify(125821)
		if i > -1 {
			__antithesis_instrumentation__.Notify(125823)
			w.SafeString("\n")
		} else {
			__antithesis_instrumentation__.Notify(125824)
		}
		__antithesis_instrumentation__.Notify(125822)
		it := itemToReplicaOrPlaceholder(bi)
		desc := it.Desc()
		w.Printf("%v - %v (%s - %s)", []byte(desc.StartKey), []byte(desc.EndKey), desc.StartKey, desc.EndKey)
		i++
		return true
	})
}

func (b *storeReplicaBTree) String() string {
	__antithesis_instrumentation__.Notify(125825)
	return redact.StringWithoutMarkers(b)
}

type rangeKeyItem interface {
	key() roachpb.RKey
}

type rangeBTreeKey roachpb.RKey

var _ rangeKeyItem = rangeBTreeKey{}

func (k rangeBTreeKey) key() roachpb.RKey {
	__antithesis_instrumentation__.Notify(125826)
	return (roachpb.RKey)(k)
}

var _ btree.Item = rangeBTreeKey{}

func (k rangeBTreeKey) Less(i btree.Item) bool {
	__antithesis_instrumentation__.Notify(125827)
	return k.key().Less(i.(rangeKeyItem).key())
}

type btreeReplica Replica

func (r *btreeReplica) key() roachpb.RKey {
	__antithesis_instrumentation__.Notify(125828)
	return r.startKey
}

func (r *btreeReplica) Less(i btree.Item) bool {
	__antithesis_instrumentation__.Notify(125829)
	return r.key().Less(i.(rangeKeyItem).key())
}

func itemToReplicaOrPlaceholder(it btree.Item) replicaOrPlaceholder {
	__antithesis_instrumentation__.Notify(125830)
	if it == nil {
		__antithesis_instrumentation__.Notify(125833)
		return replicaOrPlaceholder{}
	} else {
		__antithesis_instrumentation__.Notify(125834)
	}
	__antithesis_instrumentation__.Notify(125831)
	vit := replicaOrPlaceholder{
		item: it.(rangeKeyItem),
	}
	switch t := it.(type) {
	case *btreeReplica:
		__antithesis_instrumentation__.Notify(125835)
		vit.repl = (*Replica)(t)
	case *ReplicaPlaceholder:
		__antithesis_instrumentation__.Notify(125836)
		vit.ph = t
	}
	__antithesis_instrumentation__.Notify(125832)
	return vit
}
