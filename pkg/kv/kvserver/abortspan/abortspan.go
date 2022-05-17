package abortspan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type AbortSpan struct {
	rangeID roachpb.RangeID
}

func New(rangeID roachpb.RangeID) *AbortSpan {
	__antithesis_instrumentation__.Notify(94060)
	return &AbortSpan{
		rangeID: rangeID,
	}
}

func fillUUID(b byte) uuid.UUID {
	__antithesis_instrumentation__.Notify(94061)
	var ret uuid.UUID
	for i := range ret.GetBytes() {
		__antithesis_instrumentation__.Notify(94063)
		ret[i] = b
	}
	__antithesis_instrumentation__.Notify(94062)
	return ret
}

var txnIDMin = fillUUID('\x00')
var txnIDMax = fillUUID('\xff')

func MinKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(94064)
	return keys.AbortSpanKey(rangeID, txnIDMin)
}

func (sc *AbortSpan) min() roachpb.Key {
	__antithesis_instrumentation__.Notify(94065)
	return MinKey(sc.rangeID)
}

func MaxKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(94066)
	return keys.AbortSpanKey(rangeID, txnIDMax)
}

func (sc *AbortSpan) max() roachpb.Key {
	__antithesis_instrumentation__.Notify(94067)
	return MaxKey(sc.rangeID)
}

func (sc *AbortSpan) ClearData(e storage.Engine) error {
	__antithesis_instrumentation__.Notify(94068)

	iter := e.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{UpperBound: sc.max()})
	defer iter.Close()
	b := e.NewUnindexedBatch(true)
	defer b.Close()
	err := b.ClearIterRange(iter, sc.min(), sc.max())
	if err != nil {
		__antithesis_instrumentation__.Notify(94070)
		return err
	} else {
		__antithesis_instrumentation__.Notify(94071)
	}
	__antithesis_instrumentation__.Notify(94069)
	return b.Commit(false)
}

func (sc *AbortSpan) Get(
	ctx context.Context, reader storage.Reader, txnID uuid.UUID, entry *roachpb.AbortSpanEntry,
) (bool, error) {
	__antithesis_instrumentation__.Notify(94072)

	key := keys.AbortSpanKey(sc.rangeID, txnID)
	ok, err := storage.MVCCGetProto(ctx, reader, key, hlc.Timestamp{}, entry, storage.MVCCGetOptions{})
	return ok, err
}

func (sc *AbortSpan) Iterate(
	ctx context.Context, reader storage.Reader, f func(roachpb.Key, roachpb.AbortSpanEntry) error,
) error {
	__antithesis_instrumentation__.Notify(94073)
	_, err := storage.MVCCIterate(ctx, reader, sc.min(), sc.max(), hlc.Timestamp{}, storage.MVCCScanOptions{},
		func(kv roachpb.KeyValue) error {
			__antithesis_instrumentation__.Notify(94075)
			var entry roachpb.AbortSpanEntry
			if _, err := keys.DecodeAbortSpanKey(kv.Key, nil); err != nil {
				__antithesis_instrumentation__.Notify(94078)
				return err
			} else {
				__antithesis_instrumentation__.Notify(94079)
			}
			__antithesis_instrumentation__.Notify(94076)
			if err := kv.Value.GetProto(&entry); err != nil {
				__antithesis_instrumentation__.Notify(94080)
				return err
			} else {
				__antithesis_instrumentation__.Notify(94081)
			}
			__antithesis_instrumentation__.Notify(94077)
			return f(kv.Key, entry)
		})
	__antithesis_instrumentation__.Notify(94074)
	return err
}

func (sc *AbortSpan) Del(
	ctx context.Context, reader storage.ReadWriter, ms *enginepb.MVCCStats, txnID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(94082)
	key := keys.AbortSpanKey(sc.rangeID, txnID)
	return storage.MVCCDelete(ctx, reader, ms, key, hlc.Timestamp{}, nil)
}

func (sc *AbortSpan) Put(
	ctx context.Context,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	txnID uuid.UUID,
	entry *roachpb.AbortSpanEntry,
) error {
	__antithesis_instrumentation__.Notify(94083)
	log.VEventf(ctx, 2, "writing abort span entry for %s", txnID.Short())
	key := keys.AbortSpanKey(sc.rangeID, txnID)
	return storage.MVCCPutProto(ctx, readWriter, ms, key, hlc.Timestamp{}, nil, entry)
}

func (sc *AbortSpan) CopyTo(
	ctx context.Context,
	r storage.Reader,
	w storage.ReadWriter,
	ms *enginepb.MVCCStats,
	ts hlc.Timestamp,
	newRangeID roachpb.RangeID,
) error {
	__antithesis_instrumentation__.Notify(94084)
	var abortSpanCopyCount, abortSpanSkipCount int

	threshold := ts.Add(-kvserverbase.TxnCleanupThreshold.Nanoseconds(), 0)
	var scratch [64]byte
	if err := sc.Iterate(ctx, r, func(k roachpb.Key, entry roachpb.AbortSpanEntry) error {
		__antithesis_instrumentation__.Notify(94086)
		if entry.Timestamp.Less(threshold) {
			__antithesis_instrumentation__.Notify(94089)

			abortSpanSkipCount++
			return nil
		} else {
			__antithesis_instrumentation__.Notify(94090)
		}
		__antithesis_instrumentation__.Notify(94087)

		abortSpanCopyCount++
		var txnID uuid.UUID
		txnID, err := keys.DecodeAbortSpanKey(k, scratch[:0])
		if err != nil {
			__antithesis_instrumentation__.Notify(94091)
			return err
		} else {
			__antithesis_instrumentation__.Notify(94092)
		}
		__antithesis_instrumentation__.Notify(94088)
		return storage.MVCCPutProto(ctx, w, ms,
			keys.AbortSpanKey(newRangeID, txnID),
			hlc.Timestamp{}, nil, &entry,
		)
	}); err != nil {
		__antithesis_instrumentation__.Notify(94093)
		return errors.Wrap(err, "AbortSpan.CopyTo")
	} else {
		__antithesis_instrumentation__.Notify(94094)
	}
	__antithesis_instrumentation__.Notify(94085)
	log.Eventf(ctx, "abort span: copied %d entries, skipped %d", abortSpanCopyCount, abortSpanSkipCount)
	return nil
}
