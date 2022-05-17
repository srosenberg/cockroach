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
	RegisterReadOnlyCommand(roachpb.Scan, DefaultDeclareIsolatedKeys, Scan)
}

func Scan(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97458)
	args := cArgs.Args.(*roachpb.ScanRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.ScanResponse)

	var res result.Result
	var scanRes storage.MVCCScanResult
	var err error

	avoidExcess := cArgs.EvalCtx.ClusterSettings().Version.IsActive(ctx,
		clusterversion.TargetBytesAvoidExcess)
	opts := storage.MVCCScanOptions{
		Inconsistent: h.ReadConsistency != roachpb.CONSISTENT,
		Txn:          h.Txn,
		Uncertainty:  cArgs.Uncertainty,
		MaxKeys:      h.MaxSpanRequestKeys,
		MaxIntents:   storage.MaxIntentsPerWriteIntentError.Get(&cArgs.EvalCtx.ClusterSettings().SV),
		TargetBytes:  h.TargetBytes,
		TargetBytesAvoidExcess: h.AllowEmpty || func() bool {
			__antithesis_instrumentation__.Notify(97463)
			return avoidExcess == true
		}() == true,
		AllowEmpty:       h.AllowEmpty,
		WholeRowsOfSize:  h.WholeRowsOfSize,
		FailOnMoreRecent: args.KeyLocking != lock.None,
		Reverse:          false,
		MemoryAccount:    cArgs.EvalCtx.GetResponseMemoryAccount(),
	}

	switch args.ScanFormat {
	case roachpb.BATCH_RESPONSE:
		__antithesis_instrumentation__.Notify(97464)
		scanRes, err = storage.MVCCScanToBytes(
			ctx, reader, args.Key, args.EndKey, h.Timestamp, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(97469)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97470)
		}
		__antithesis_instrumentation__.Notify(97465)
		reply.BatchResponses = scanRes.KVData
	case roachpb.KEY_VALUES:
		__antithesis_instrumentation__.Notify(97466)
		scanRes, err = storage.MVCCScan(
			ctx, reader, args.Key, args.EndKey, h.Timestamp, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(97471)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97472)
		}
		__antithesis_instrumentation__.Notify(97467)
		reply.Rows = scanRes.KVs
	default:
		__antithesis_instrumentation__.Notify(97468)
		panic(fmt.Sprintf("Unknown scanFormat %d", args.ScanFormat))
	}
	__antithesis_instrumentation__.Notify(97459)

	reply.NumKeys = scanRes.NumKeys
	reply.NumBytes = scanRes.NumBytes

	if scanRes.ResumeSpan != nil {
		__antithesis_instrumentation__.Notify(97473)
		reply.ResumeSpan = scanRes.ResumeSpan
		reply.ResumeReason = scanRes.ResumeReason
		reply.ResumeNextBytes = scanRes.ResumeNextBytes
	} else {
		__antithesis_instrumentation__.Notify(97474)
	}
	__antithesis_instrumentation__.Notify(97460)

	if h.ReadConsistency == roachpb.READ_UNCOMMITTED {
		__antithesis_instrumentation__.Notify(97475)

		const usePrefixIter = false
		reply.IntentRows, err = CollectIntentRows(ctx, reader, usePrefixIter, scanRes.Intents)
		if err != nil {
			__antithesis_instrumentation__.Notify(97476)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97477)
		}
	} else {
		__antithesis_instrumentation__.Notify(97478)
	}
	__antithesis_instrumentation__.Notify(97461)

	if args.KeyLocking != lock.None && func() bool {
		__antithesis_instrumentation__.Notify(97479)
		return h.Txn != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(97480)
		err = acquireUnreplicatedLocksOnKeys(&res, h.Txn, args.ScanFormat, &scanRes)
		if err != nil {
			__antithesis_instrumentation__.Notify(97481)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97482)
		}
	} else {
		__antithesis_instrumentation__.Notify(97483)
	}
	__antithesis_instrumentation__.Notify(97462)
	res.Local.EncounteredIntents = scanRes.Intents
	return res, nil
}
