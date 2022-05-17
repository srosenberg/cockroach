package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadOnlyCommand(roachpb.ReverseScan, DefaultDeclareIsolatedKeys, ReverseScan)
}

func ReverseScan(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97415)
	args := cArgs.Args.(*roachpb.ReverseScanRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.ReverseScanResponse)

	var res result.Result
	var scanRes storage.MVCCScanResult
	var err error

	avoidExcess := cArgs.EvalCtx.ClusterSettings().Version.IsActive(ctx,
		clusterversion.TargetBytesAvoidExcess)
	opts := storage.MVCCScanOptions{
		Inconsistent: h.ReadConsistency != roachpb.CONSISTENT,
		Txn:          h.Txn,
		MaxKeys:      h.MaxSpanRequestKeys,
		MaxIntents:   storage.MaxIntentsPerWriteIntentError.Get(&cArgs.EvalCtx.ClusterSettings().SV),
		TargetBytes:  h.TargetBytes,
		TargetBytesAvoidExcess: h.AllowEmpty || func() bool {
			__antithesis_instrumentation__.Notify(97420)
			return avoidExcess == true
		}() == true,
		AllowEmpty:       h.AllowEmpty,
		WholeRowsOfSize:  h.WholeRowsOfSize,
		FailOnMoreRecent: args.KeyLocking != lock.None,
		Reverse:          true,
		MemoryAccount:    cArgs.EvalCtx.GetResponseMemoryAccount(),
	}

	switch args.ScanFormat {
	case roachpb.BATCH_RESPONSE:
		__antithesis_instrumentation__.Notify(97421)
		scanRes, err = storage.MVCCScanToBytes(
			ctx, reader, args.Key, args.EndKey, h.Timestamp, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(97426)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97427)
		}
		__antithesis_instrumentation__.Notify(97422)
		reply.BatchResponses = scanRes.KVData
	case roachpb.KEY_VALUES:
		__antithesis_instrumentation__.Notify(97423)
		scanRes, err = storage.MVCCScan(
			ctx, reader, args.Key, args.EndKey, h.Timestamp, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(97428)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97429)
		}
		__antithesis_instrumentation__.Notify(97424)
		reply.Rows = scanRes.KVs
	default:
		__antithesis_instrumentation__.Notify(97425)
		panic(fmt.Sprintf("Unknown scanFormat %d", args.ScanFormat))
	}
	__antithesis_instrumentation__.Notify(97416)

	reply.NumKeys = scanRes.NumKeys
	reply.NumBytes = scanRes.NumBytes

	if scanRes.ResumeSpan != nil {
		__antithesis_instrumentation__.Notify(97430)
		reply.ResumeSpan = scanRes.ResumeSpan
		reply.ResumeReason = scanRes.ResumeReason
		reply.ResumeNextBytes = scanRes.ResumeNextBytes
	} else {
		__antithesis_instrumentation__.Notify(97431)
	}
	__antithesis_instrumentation__.Notify(97417)

	if h.ReadConsistency == roachpb.READ_UNCOMMITTED {
		__antithesis_instrumentation__.Notify(97432)

		const usePrefixIter = false
		reply.IntentRows, err = CollectIntentRows(ctx, reader, usePrefixIter, scanRes.Intents)
		if err != nil {
			__antithesis_instrumentation__.Notify(97433)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97434)
		}
	} else {
		__antithesis_instrumentation__.Notify(97435)
	}
	__antithesis_instrumentation__.Notify(97418)

	if args.KeyLocking != lock.None && func() bool {
		__antithesis_instrumentation__.Notify(97436)
		return h.Txn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(97437)
		err = acquireUnreplicatedLocksOnKeys(&res, h.Txn, args.ScanFormat, &scanRes)
		if err != nil {
			__antithesis_instrumentation__.Notify(97438)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97439)
		}
	} else {
		__antithesis_instrumentation__.Notify(97440)
	}
	__antithesis_instrumentation__.Notify(97419)
	res.Local.EncounteredIntents = scanRes.Intents
	return res, nil
}
