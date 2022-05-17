package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/gc"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

var QueryResolvedTimestampIntentCleanupAge = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.query_resolved_timestamp.intent_cleanup_age",
	"minimum intent age that QueryResolvedTimestamp requests will consider for async intent cleanup",
	10*time.Second,
	settings.NonNegativeDuration,
)

func init() {
	RegisterReadOnlyCommand(roachpb.QueryResolvedTimestamp, DefaultDeclareKeys, QueryResolvedTimestamp)
}

func QueryResolvedTimestamp(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97193)
	args := cArgs.Args.(*roachpb.QueryResolvedTimestampRequest)
	reply := resp.(*roachpb.QueryResolvedTimestampResponse)

	closedTS := cArgs.EvalCtx.GetClosedTimestamp(ctx)

	st := cArgs.EvalCtx.ClusterSettings()
	maxEncounteredIntents := gc.MaxIntentsPerCleanupBatch.Get(&st.SV)
	maxEncounteredIntentKeyBytes := gc.MaxIntentKeyBytesPerCleanupBatch.Get(&st.SV)
	intentCleanupAge := QueryResolvedTimestampIntentCleanupAge.Get(&st.SV)
	intentCleanupThresh := cArgs.EvalCtx.Clock().Now().Add(-intentCleanupAge.Nanoseconds(), 0)
	minIntentTS, encounteredIntents, err := computeMinIntentTimestamp(
		reader, args.Span(), maxEncounteredIntents, maxEncounteredIntentKeyBytes, intentCleanupThresh,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(97196)
		return result.Result{}, errors.Wrapf(err, "computing minimum intent timestamp")
	} else {
		__antithesis_instrumentation__.Notify(97197)
	}
	__antithesis_instrumentation__.Notify(97194)

	reply.ResolvedTS = closedTS
	if !minIntentTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(97198)
		reply.ResolvedTS.Backward(minIntentTS.Prev())
	} else {
		__antithesis_instrumentation__.Notify(97199)
	}
	__antithesis_instrumentation__.Notify(97195)

	var res result.Result
	res.Local.EncounteredIntents = encounteredIntents
	return res, nil
}

func computeMinIntentTimestamp(
	reader storage.Reader,
	span roachpb.Span,
	maxEncounteredIntents int64,
	maxEncounteredIntentKeyBytes int64,
	intentCleanupThresh hlc.Timestamp,
) (hlc.Timestamp, []roachpb.Intent, error) {
	__antithesis_instrumentation__.Notify(97200)
	ltStart, _ := keys.LockTableSingleKey(span.Key, nil)
	ltEnd, _ := keys.LockTableSingleKey(span.EndKey, nil)
	iter := reader.NewEngineIterator(storage.IterOptions{LowerBound: ltStart, UpperBound: ltEnd})
	defer iter.Close()

	var meta enginepb.MVCCMetadata
	var minTS hlc.Timestamp
	var encountered []roachpb.Intent
	var encounteredKeyBytes int64
	for valid, err := iter.SeekEngineKeyGE(storage.EngineKey{Key: ltStart}); ; valid, err = iter.NextEngineKey() {
		__antithesis_instrumentation__.Notify(97202)
		if err != nil {
			__antithesis_instrumentation__.Notify(97209)
			return hlc.Timestamp{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(97210)
			if !valid {
				__antithesis_instrumentation__.Notify(97211)
				break
			} else {
				__antithesis_instrumentation__.Notify(97212)
			}
		}
		__antithesis_instrumentation__.Notify(97203)
		engineKey, err := iter.EngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(97213)
			return hlc.Timestamp{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(97214)
		}
		__antithesis_instrumentation__.Notify(97204)
		lockedKey, err := keys.DecodeLockTableSingleKey(engineKey.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(97215)
			return hlc.Timestamp{}, nil, errors.Wrapf(err, "decoding LockTable key: %v", lockedKey)
		} else {
			__antithesis_instrumentation__.Notify(97216)
		}
		__antithesis_instrumentation__.Notify(97205)

		if err := protoutil.Unmarshal(iter.UnsafeValue(), &meta); err != nil {
			__antithesis_instrumentation__.Notify(97217)
			return hlc.Timestamp{}, nil, errors.Wrapf(err, "unmarshaling mvcc meta: %v", lockedKey)
		} else {
			__antithesis_instrumentation__.Notify(97218)
		}
		__antithesis_instrumentation__.Notify(97206)
		if meta.Txn == nil {
			__antithesis_instrumentation__.Notify(97219)
			return hlc.Timestamp{}, nil,
				errors.AssertionFailedf("nil transaction in LockTable. Key: %v,"+"mvcc meta: %v",
					lockedKey, meta)
		} else {
			__antithesis_instrumentation__.Notify(97220)
		}
		__antithesis_instrumentation__.Notify(97207)

		if minTS.IsEmpty() {
			__antithesis_instrumentation__.Notify(97221)
			minTS = meta.Txn.WriteTimestamp
		} else {
			__antithesis_instrumentation__.Notify(97222)
			minTS.Backward(meta.Txn.WriteTimestamp)
		}
		__antithesis_instrumentation__.Notify(97208)

		oldEnough := meta.Txn.WriteTimestamp.Less(intentCleanupThresh)
		intentFitsByCount := int64(len(encountered)) < maxEncounteredIntents
		intentFitsByBytes := encounteredKeyBytes < maxEncounteredIntentKeyBytes
		if oldEnough && func() bool {
			__antithesis_instrumentation__.Notify(97223)
			return intentFitsByCount == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(97224)
			return intentFitsByBytes == true
		}() == true {
			__antithesis_instrumentation__.Notify(97225)
			encountered = append(encountered, roachpb.MakeIntent(meta.Txn, lockedKey))
			encounteredKeyBytes += int64(len(lockedKey))
		} else {
			__antithesis_instrumentation__.Notify(97226)
		}
	}
	__antithesis_instrumentation__.Notify(97201)
	return minTS, encountered, nil
}
