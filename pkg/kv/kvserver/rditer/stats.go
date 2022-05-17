package rditer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
)

func ComputeStatsForRange(
	d *roachpb.RangeDescriptor, reader storage.Reader, nowNanos int64,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(114253)
	ms := enginepb.MVCCStats{}
	var err error
	for _, keyRange := range MakeReplicatedKeyRangesExceptLockTable(d) {
		__antithesis_instrumentation__.Notify(114255)
		func() {
			__antithesis_instrumentation__.Notify(114257)
			iter := reader.NewMVCCIterator(storage.MVCCKeyAndIntentsIterKind,
				storage.IterOptions{UpperBound: keyRange.End})
			defer iter.Close()

			var msDelta enginepb.MVCCStats
			if msDelta, err = iter.ComputeStats(keyRange.Start, keyRange.End, nowNanos); err != nil {
				__antithesis_instrumentation__.Notify(114259)
				return
			} else {
				__antithesis_instrumentation__.Notify(114260)
			}
			__antithesis_instrumentation__.Notify(114258)
			ms.Add(msDelta)
		}()
		__antithesis_instrumentation__.Notify(114256)
		if err != nil {
			__antithesis_instrumentation__.Notify(114261)
			return enginepb.MVCCStats{}, err
		} else {
			__antithesis_instrumentation__.Notify(114262)
		}
	}
	__antithesis_instrumentation__.Notify(114254)
	return ms, nil
}
