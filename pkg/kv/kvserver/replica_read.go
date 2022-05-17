package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/kr/pretty"
)

func (r *Replica) executeReadOnlyBatch(
	ctx context.Context, ba *roachpb.BatchRequest, g *concurrency.Guard,
) (br *roachpb.BatchResponse, _ *concurrency.Guard, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120135)
	r.readOnlyCmdMu.RLock()
	defer r.readOnlyCmdMu.RUnlock()

	st, err := r.checkExecutionCanProceed(ctx, ba, g)
	if err != nil {
		__antithesis_instrumentation__.Notify(120144)
		return nil, g, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120145)
	}
	__antithesis_instrumentation__.Notify(120136)

	ui := uncertainty.ComputeInterval(&ba.Header, st, r.Clock().MaxOffset())

	rec := NewReplicaEvalContext(r, g.LatchSpans())

	rw := r.store.Engine().NewReadOnly(storage.StandardDurability)
	if !rw.ConsistentIterators() {
		__antithesis_instrumentation__.Notify(120146)

		panic("expected consistent iterators")
	} else {
		__antithesis_instrumentation__.Notify(120147)
	}
	__antithesis_instrumentation__.Notify(120137)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(120148)
		rw = spanset.NewReadWriterAt(rw, g.LatchSpans(), ba.Timestamp)
	} else {
		__antithesis_instrumentation__.Notify(120149)
	}
	__antithesis_instrumentation__.Notify(120138)
	defer rw.Close()

	var result result.Result
	br, result, pErr = r.executeReadOnlyBatchWithServersideRefreshes(ctx, rw, rec, ba, ui, g)

	if isConcurrencyRetryError(pErr) {
		__antithesis_instrumentation__.Notify(120150)
		if g.EvalKind == concurrency.OptimisticEval {
			__antithesis_instrumentation__.Notify(120152)

			if !g.CheckOptimisticNoLatchConflicts() {
				__antithesis_instrumentation__.Notify(120153)
				return nil, g, roachpb.NewError(roachpb.NewOptimisticEvalConflictsError())
			} else {
				__antithesis_instrumentation__.Notify(120154)
			}
		} else {
			__antithesis_instrumentation__.Notify(120155)
		}
		__antithesis_instrumentation__.Notify(120151)
		pErr = maybeAttachLease(pErr, &st.Lease)
		return nil, g, pErr
	} else {
		__antithesis_instrumentation__.Notify(120156)
	}
	__antithesis_instrumentation__.Notify(120139)

	if g.EvalKind == concurrency.OptimisticEval {
		__antithesis_instrumentation__.Notify(120157)
		if pErr == nil {
			__antithesis_instrumentation__.Notify(120158)

			latchSpansRead, lockSpansRead, err := r.collectSpansRead(ba, br)
			if err != nil {
				__antithesis_instrumentation__.Notify(120160)
				return nil, g, roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(120161)
			}
			__antithesis_instrumentation__.Notify(120159)
			if ok := g.CheckOptimisticNoConflicts(latchSpansRead, lockSpansRead); !ok {
				__antithesis_instrumentation__.Notify(120162)
				return nil, g, roachpb.NewError(roachpb.NewOptimisticEvalConflictsError())
			} else {
				__antithesis_instrumentation__.Notify(120163)
			}
		} else {
			__antithesis_instrumentation__.Notify(120164)

			return nil, g, roachpb.NewError(roachpb.NewOptimisticEvalConflictsError())
		}
	} else {
		__antithesis_instrumentation__.Notify(120165)
	}
	__antithesis_instrumentation__.Notify(120140)

	intents := result.Local.DetachEncounteredIntents()
	if pErr == nil {
		__antithesis_instrumentation__.Notify(120166)
		pErr = r.handleReadOnlyLocalEvalResult(ctx, ba, result.Local)
	} else {
		__antithesis_instrumentation__.Notify(120167)
	}
	__antithesis_instrumentation__.Notify(120141)

	ec, g := endCmds{repl: r, g: g, st: st}, nil
	ec.done(ctx, ba, br, pErr)

	if len(intents) > 0 {
		__antithesis_instrumentation__.Notify(120168)
		log.Eventf(ctx, "submitting %d intents to asynchronous processing", len(intents))

		allowSyncProcessing := ba.ReadConsistency == roachpb.CONSISTENT
		if err := r.store.intentResolver.CleanupIntentsAsync(ctx, intents, allowSyncProcessing); err != nil {
			__antithesis_instrumentation__.Notify(120169)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(120170)
		}
	} else {
		__antithesis_instrumentation__.Notify(120171)
	}
	__antithesis_instrumentation__.Notify(120142)

	if pErr != nil {
		__antithesis_instrumentation__.Notify(120172)
		log.VErrEventf(ctx, 3, "%v", pErr.String())
	} else {
		__antithesis_instrumentation__.Notify(120173)
		log.Event(ctx, "read completed")
	}
	__antithesis_instrumentation__.Notify(120143)
	return br, nil, pErr
}

type evalContextWithAccount struct {
	batcheval.EvalContext
	memAccount *mon.BoundAccount
}

var evalContextWithAccountPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(120174)
		return &evalContextWithAccount{}
	},
}

func newEvalContextWithAccount(
	ctx context.Context, evalCtx batcheval.EvalContext, mon *mon.BytesMonitor,
) *evalContextWithAccount {
	__antithesis_instrumentation__.Notify(120175)
	ec := evalContextWithAccountPool.Get().(*evalContextWithAccount)
	ec.EvalContext = evalCtx
	if ec.memAccount != nil {
		__antithesis_instrumentation__.Notify(120177)
		ec.memAccount.Init(ctx, mon)
	} else {
		__antithesis_instrumentation__.Notify(120178)
		acc := mon.MakeBoundAccount()
		ec.memAccount = &acc
	}
	__antithesis_instrumentation__.Notify(120176)
	return ec
}

func (e *evalContextWithAccount) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(120179)
	e.memAccount.Close(ctx)

	*e.memAccount = mon.BoundAccount{}
	evalContextWithAccountPool.Put(e)
}
func (e evalContextWithAccount) GetResponseMemoryAccount() *mon.BoundAccount {
	__antithesis_instrumentation__.Notify(120180)
	return e.memAccount
}

func (r *Replica) executeReadOnlyBatchWithServersideRefreshes(
	ctx context.Context,
	rw storage.ReadWriter,
	rec batcheval.EvalContext,
	ba *roachpb.BatchRequest,
	ui uncertainty.Interval,
	g *concurrency.Guard,
) (br *roachpb.BatchResponse, res result.Result, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120181)
	log.Event(ctx, "executing read-only batch")

	var rootMonitor *mon.BytesMonitor

	if ba.AdmissionHeader.SourceLocation != roachpb.AdmissionHeader_LOCAL || func() bool {
		__antithesis_instrumentation__.Notify(120186)
		return ba.AdmissionHeader.NoMemoryReservedAtSource == true
	}() == true {
		__antithesis_instrumentation__.Notify(120187)

		rootMonitor = r.store.getRootMemoryMonitorForKV()
	} else {
		__antithesis_instrumentation__.Notify(120188)
	}
	__antithesis_instrumentation__.Notify(120182)
	var boundAccount *mon.BoundAccount
	if rootMonitor != nil {
		__antithesis_instrumentation__.Notify(120189)
		evalCtx := newEvalContextWithAccount(ctx, rec, rootMonitor)
		boundAccount = evalCtx.memAccount

		defer evalCtx.close(ctx)
		rec = evalCtx
	} else {
		__antithesis_instrumentation__.Notify(120190)
	}
	__antithesis_instrumentation__.Notify(120183)

	for retries := 0; ; retries++ {
		__antithesis_instrumentation__.Notify(120191)
		if retries > 0 {
			__antithesis_instrumentation__.Notify(120193)

			boundAccount.Clear(ctx)
			log.VEventf(ctx, 2, "server-side retry of batch")
		} else {
			__antithesis_instrumentation__.Notify(120194)
		}
		__antithesis_instrumentation__.Notify(120192)
		br, res, pErr = evaluateBatch(ctx, kvserverbase.CmdIDKey(""), rw, rec, nil, ba, ui, true)

		if pErr == nil || func() bool {
			__antithesis_instrumentation__.Notify(120195)
			return retries > 0 == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(120196)
			return !canDoServersideRetry(ctx, pErr, ba, br, g, nil) == true
		}() == true {
			__antithesis_instrumentation__.Notify(120197)
			break
		} else {
			__antithesis_instrumentation__.Notify(120198)
		}
	}
	__antithesis_instrumentation__.Notify(120184)

	if pErr != nil {
		__antithesis_instrumentation__.Notify(120199)

		res.Local = result.LocalResult{
			EncounteredIntents: res.Local.DetachEncounteredIntents(),
			Metrics:            res.Local.Metrics,
		}
		return nil, res, pErr
	} else {
		__antithesis_instrumentation__.Notify(120200)
	}
	__antithesis_instrumentation__.Notify(120185)
	return br, res, nil
}

func (r *Replica) handleReadOnlyLocalEvalResult(
	ctx context.Context, ba *roachpb.BatchRequest, lResult result.LocalResult,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(120201)

	{
		__antithesis_instrumentation__.Notify(120205)
		lResult.Reply = nil
	}
	__antithesis_instrumentation__.Notify(120202)

	if lResult.AcquiredLocks != nil {
		__antithesis_instrumentation__.Notify(120206)

		log.Eventf(ctx, "acquiring %d unreplicated locks", len(lResult.AcquiredLocks))
		for i := range lResult.AcquiredLocks {
			__antithesis_instrumentation__.Notify(120208)
			r.concMgr.OnLockAcquired(ctx, &lResult.AcquiredLocks[i])
		}
		__antithesis_instrumentation__.Notify(120207)
		lResult.AcquiredLocks = nil
	} else {
		__antithesis_instrumentation__.Notify(120209)
	}
	__antithesis_instrumentation__.Notify(120203)

	if !lResult.IsZero() {
		__antithesis_instrumentation__.Notify(120210)
		log.Fatalf(ctx, "unhandled field in LocalEvalResult: %s", pretty.Diff(lResult, result.LocalResult{}))
	} else {
		__antithesis_instrumentation__.Notify(120211)
	}
	__antithesis_instrumentation__.Notify(120204)
	return nil
}

func (r *Replica) collectSpansRead(
	ba *roachpb.BatchRequest, br *roachpb.BatchResponse,
) (latchSpans, lockSpans *spanset.SpanSet, _ error) {
	__antithesis_instrumentation__.Notify(120212)
	baCopy := *ba
	baCopy.Requests = make([]roachpb.RequestUnion, len(baCopy.Requests))
	j := 0
	for i := 0; i < len(baCopy.Requests); i++ {
		__antithesis_instrumentation__.Notify(120214)
		baReq := ba.Requests[i]
		req := baReq.GetInner()
		header := req.Header()

		resp := br.Responses[i].GetInner()
		if resp.Header().ResumeSpan == nil {
			__antithesis_instrumentation__.Notify(120217)

			baCopy.Requests[j] = baReq
			j++
			continue
		} else {
			__antithesis_instrumentation__.Notify(120218)
		}
		__antithesis_instrumentation__.Notify(120215)

		switch t := resp.(type) {
		case *roachpb.ScanResponse:
			__antithesis_instrumentation__.Notify(120219)
			if header.Key.Equal(t.ResumeSpan.Key) {
				__antithesis_instrumentation__.Notify(120224)

				continue
			} else {
				__antithesis_instrumentation__.Notify(120225)
			}
			__antithesis_instrumentation__.Notify(120220)

			header.EndKey = t.ResumeSpan.Key
		case *roachpb.ReverseScanResponse:
			__antithesis_instrumentation__.Notify(120221)
			if header.EndKey.Equal(t.ResumeSpan.EndKey) {
				__antithesis_instrumentation__.Notify(120226)

				continue
			} else {
				__antithesis_instrumentation__.Notify(120227)
			}
			__antithesis_instrumentation__.Notify(120222)

			header.Key = t.ResumeSpan.EndKey
		default:
			__antithesis_instrumentation__.Notify(120223)

			baCopy.Requests[j] = baReq
			j++
			continue
		}
		__antithesis_instrumentation__.Notify(120216)

		req = req.ShallowCopy()
		req.SetHeader(header)
		baCopy.Requests[j].MustSetInner(req)
		j++
	}
	__antithesis_instrumentation__.Notify(120213)
	baCopy.Requests = baCopy.Requests[:j]

	var err error
	latchSpans, lockSpans, _, err = r.collectSpans(&baCopy)
	return latchSpans, lockSpans, err
}
