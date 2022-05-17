package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func init() {
	RegisterReadOnlyCommand(roachpb.Get, DefaultDeclareIsolatedKeys, Get)
}

func Get(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96905)
	args := cArgs.Args.(*roachpb.GetRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.GetResponse)

	if h.MaxSpanRequestKeys < 0 || func() bool {
		__antithesis_instrumentation__.Notify(96912)
		return h.TargetBytes < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(96913)

		reply.ResumeSpan = &roachpb.Span{Key: args.Key}
		if h.MaxSpanRequestKeys < 0 {
			__antithesis_instrumentation__.Notify(96915)
			reply.ResumeReason = roachpb.RESUME_KEY_LIMIT
		} else {
			__antithesis_instrumentation__.Notify(96916)
			if h.TargetBytes < 0 {
				__antithesis_instrumentation__.Notify(96917)
				reply.ResumeReason = roachpb.RESUME_BYTE_LIMIT
			} else {
				__antithesis_instrumentation__.Notify(96918)
			}
		}
		__antithesis_instrumentation__.Notify(96914)
		return result.Result{}, nil
	} else {
		__antithesis_instrumentation__.Notify(96919)
	}
	__antithesis_instrumentation__.Notify(96906)

	var val *roachpb.Value
	var intent *roachpb.Intent
	var err error
	val, intent, err = storage.MVCCGet(ctx, reader, args.Key, h.Timestamp, storage.MVCCGetOptions{
		Inconsistent:     h.ReadConsistency != roachpb.CONSISTENT,
		Txn:              h.Txn,
		FailOnMoreRecent: args.KeyLocking != lock.None,
		Uncertainty:      cArgs.Uncertainty,
		MemoryAccount:    cArgs.EvalCtx.GetResponseMemoryAccount(),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(96920)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96921)
	}
	__antithesis_instrumentation__.Notify(96907)
	if val != nil {
		__antithesis_instrumentation__.Notify(96922)

		numBytes := int64(len(val.RawBytes))
		if h.TargetBytes > 0 && func() bool {
			__antithesis_instrumentation__.Notify(96924)
			return h.AllowEmpty == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(96925)
			return numBytes > h.TargetBytes == true
		}() == true {
			__antithesis_instrumentation__.Notify(96926)
			reply.ResumeSpan = &roachpb.Span{Key: args.Key}
			reply.ResumeReason = roachpb.RESUME_BYTE_LIMIT
			reply.ResumeNextBytes = numBytes
			return result.Result{}, nil
		} else {
			__antithesis_instrumentation__.Notify(96927)
		}
		__antithesis_instrumentation__.Notify(96923)
		reply.NumKeys = 1
		reply.NumBytes = numBytes
	} else {
		__antithesis_instrumentation__.Notify(96928)
	}
	__antithesis_instrumentation__.Notify(96908)
	var intents []roachpb.Intent
	if intent != nil {
		__antithesis_instrumentation__.Notify(96929)
		intents = append(intents, *intent)
	} else {
		__antithesis_instrumentation__.Notify(96930)
	}
	__antithesis_instrumentation__.Notify(96909)

	reply.Value = val
	if h.ReadConsistency == roachpb.READ_UNCOMMITTED {
		__antithesis_instrumentation__.Notify(96931)
		var intentVals []roachpb.KeyValue

		const usePrefixIter = true
		intentVals, err = CollectIntentRows(ctx, reader, usePrefixIter, intents)
		if err == nil {
			__antithesis_instrumentation__.Notify(96932)
			switch len(intentVals) {
			case 0:
				__antithesis_instrumentation__.Notify(96933)
			case 1:
				__antithesis_instrumentation__.Notify(96934)
				reply.IntentValue = &intentVals[0].Value
			default:
				__antithesis_instrumentation__.Notify(96935)
				log.Fatalf(ctx, "more than 1 intent on single key: %v", intentVals)
			}
		} else {
			__antithesis_instrumentation__.Notify(96936)
		}
	} else {
		__antithesis_instrumentation__.Notify(96937)
	}
	__antithesis_instrumentation__.Notify(96910)

	var res result.Result
	if args.KeyLocking != lock.None && func() bool {
		__antithesis_instrumentation__.Notify(96938)
		return h.Txn != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(96939)
		return val != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(96940)
		acq := roachpb.MakeLockAcquisition(h.Txn, args.Key, lock.Unreplicated)
		res.Local.AcquiredLocks = []roachpb.LockAcquisition{acq}
	} else {
		__antithesis_instrumentation__.Notify(96941)
	}
	__antithesis_instrumentation__.Notify(96911)
	res.Local.EncounteredIntents = intents
	return res, err
}
