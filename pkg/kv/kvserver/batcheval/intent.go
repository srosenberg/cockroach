package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func CollectIntentRows(
	ctx context.Context, reader storage.Reader, usePrefixIter bool, intents []roachpb.Intent,
) ([]roachpb.KeyValue, error) {
	__antithesis_instrumentation__.Notify(97619)
	if len(intents) == 0 {
		__antithesis_instrumentation__.Notify(97622)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(97623)
	}
	__antithesis_instrumentation__.Notify(97620)
	res := make([]roachpb.KeyValue, 0, len(intents))
	for i := range intents {
		__antithesis_instrumentation__.Notify(97624)
		kv, err := readProvisionalVal(ctx, reader, usePrefixIter, &intents[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(97626)
			if errors.HasType(err, (*roachpb.WriteIntentError)(nil)) || func() bool {
				__antithesis_instrumentation__.Notify(97628)
				return errors.HasType(err, (*roachpb.ReadWithinUncertaintyIntervalError)(nil)) == true
			}() == true {
				__antithesis_instrumentation__.Notify(97629)
				log.Fatalf(ctx, "unexpected %T in CollectIntentRows: %+v", err, err)
			} else {
				__antithesis_instrumentation__.Notify(97630)
			}
			__antithesis_instrumentation__.Notify(97627)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(97631)
		}
		__antithesis_instrumentation__.Notify(97625)
		if kv.Value.IsPresent() {
			__antithesis_instrumentation__.Notify(97632)
			res = append(res, kv)
		} else {
			__antithesis_instrumentation__.Notify(97633)
		}
	}
	__antithesis_instrumentation__.Notify(97621)
	return res, nil
}

func readProvisionalVal(
	ctx context.Context, reader storage.Reader, usePrefixIter bool, intent *roachpb.Intent,
) (roachpb.KeyValue, error) {
	__antithesis_instrumentation__.Notify(97634)
	if usePrefixIter {
		__antithesis_instrumentation__.Notify(97638)
		val, _, err := storage.MVCCGetAsTxn(
			ctx, reader, intent.Key, intent.Txn.WriteTimestamp, intent.Txn,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(97641)
			return roachpb.KeyValue{}, err
		} else {
			__antithesis_instrumentation__.Notify(97642)
		}
		__antithesis_instrumentation__.Notify(97639)
		if val == nil {
			__antithesis_instrumentation__.Notify(97643)

			return roachpb.KeyValue{}, nil
		} else {
			__antithesis_instrumentation__.Notify(97644)
		}
		__antithesis_instrumentation__.Notify(97640)
		return roachpb.KeyValue{Key: intent.Key, Value: *val}, nil
	} else {
		__antithesis_instrumentation__.Notify(97645)
	}
	__antithesis_instrumentation__.Notify(97635)
	res, err := storage.MVCCScanAsTxn(
		ctx, reader, intent.Key, intent.Key.Next(), intent.Txn.WriteTimestamp, intent.Txn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(97646)
		return roachpb.KeyValue{}, err
	} else {
		__antithesis_instrumentation__.Notify(97647)
	}
	__antithesis_instrumentation__.Notify(97636)
	if len(res.KVs) > 1 {
		__antithesis_instrumentation__.Notify(97648)
		log.Fatalf(ctx, "multiple key-values returned from single-key scan: %+v", res.KVs)
	} else {
		__antithesis_instrumentation__.Notify(97649)
		if len(res.KVs) == 0 {
			__antithesis_instrumentation__.Notify(97650)

			return roachpb.KeyValue{}, nil
		} else {
			__antithesis_instrumentation__.Notify(97651)
		}
	}
	__antithesis_instrumentation__.Notify(97637)
	return res.KVs[0], nil

}

func acquireUnreplicatedLocksOnKeys(
	res *result.Result,
	txn *roachpb.Transaction,
	scanFmt roachpb.ScanFormat,
	scanRes *storage.MVCCScanResult,
) error {
	__antithesis_instrumentation__.Notify(97652)
	res.Local.AcquiredLocks = make([]roachpb.LockAcquisition, scanRes.NumKeys)
	switch scanFmt {
	case roachpb.BATCH_RESPONSE:
		__antithesis_instrumentation__.Notify(97653)
		var i int
		return storage.MVCCScanDecodeKeyValues(scanRes.KVData, func(key storage.MVCCKey, _ []byte) error {
			__antithesis_instrumentation__.Notify(97657)
			res.Local.AcquiredLocks[i] = roachpb.MakeLockAcquisition(txn, copyKey(key.Key), lock.Unreplicated)
			i++
			return nil
		})
	case roachpb.KEY_VALUES:
		__antithesis_instrumentation__.Notify(97654)
		for i, row := range scanRes.KVs {
			__antithesis_instrumentation__.Notify(97658)
			res.Local.AcquiredLocks[i] = roachpb.MakeLockAcquisition(txn, copyKey(row.Key), lock.Unreplicated)
		}
		__antithesis_instrumentation__.Notify(97655)
		return nil
	default:
		__antithesis_instrumentation__.Notify(97656)
		panic("unexpected scanFormat")
	}
}

func copyKey(k roachpb.Key) roachpb.Key {
	__antithesis_instrumentation__.Notify(97659)
	k2 := make([]byte, len(k))
	copy(k2, k)
	return k2
}
