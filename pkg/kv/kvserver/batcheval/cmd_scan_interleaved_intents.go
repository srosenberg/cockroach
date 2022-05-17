package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadOnlyCommand(roachpb.ScanInterleavedIntents, declareKeysScanInterleavedIntents, ScanInterleavedIntents)
}

func declareKeysScanInterleavedIntents(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97484)
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
}

func ScanInterleavedIntents(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, response roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97485)
	req := cArgs.Args.(*roachpb.ScanInterleavedIntentsRequest)
	resp := response.(*roachpb.ScanInterleavedIntentsResponse)

	const maxIntentCount = 1000
	const maxIntentBytes = 1 << 20
	iter := reader.NewEngineIterator(storage.IterOptions{
		LowerBound: req.Key,
		UpperBound: req.EndKey,
	})
	defer iter.Close()
	valid, err := iter.SeekEngineKeyGE(storage.EngineKey{Key: req.Key})
	intentCount := 0
	intentBytes := 0

	for ; valid && func() bool {
		__antithesis_instrumentation__.Notify(97487)
		return err == nil == true
	}() == true; valid, err = iter.NextEngineKey() {
		__antithesis_instrumentation__.Notify(97488)
		key, err := iter.EngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(97496)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97497)
		}
		__antithesis_instrumentation__.Notify(97489)
		if !key.IsMVCCKey() {
			__antithesis_instrumentation__.Notify(97498)

			return result.Result{}, errors.New("encountered non-MVCC key during lock table migration")
		} else {
			__antithesis_instrumentation__.Notify(97499)
		}
		__antithesis_instrumentation__.Notify(97490)
		mvccKey, err := key.ToMVCCKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(97500)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97501)
		}
		__antithesis_instrumentation__.Notify(97491)
		if !mvccKey.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(97502)

			continue
		} else {
			__antithesis_instrumentation__.Notify(97503)
		}
		__antithesis_instrumentation__.Notify(97492)

		val := iter.Value()
		meta := enginepb.MVCCMetadata{}
		if err := protoutil.Unmarshal(val, &meta); err != nil {
			__antithesis_instrumentation__.Notify(97504)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97505)
		}
		__antithesis_instrumentation__.Notify(97493)
		if meta.IsInline() {
			__antithesis_instrumentation__.Notify(97506)

			continue
		} else {
			__antithesis_instrumentation__.Notify(97507)
		}
		__antithesis_instrumentation__.Notify(97494)

		if intentCount >= maxIntentCount || func() bool {
			__antithesis_instrumentation__.Notify(97508)
			return intentBytes >= maxIntentBytes == true
		}() == true {
			__antithesis_instrumentation__.Notify(97509)

			resp.ResumeSpan = &roachpb.Span{
				Key:    mvccKey.Key,
				EndKey: req.EndKey,
			}
			break
		} else {
			__antithesis_instrumentation__.Notify(97510)
		}
		__antithesis_instrumentation__.Notify(97495)
		resp.Intents = append(resp.Intents, roachpb.MakeIntent(meta.Txn, mvccKey.Key))
		intentCount++
		intentBytes += len(val)
	}
	__antithesis_instrumentation__.Notify(97486)

	return result.Result{}, nil
}
