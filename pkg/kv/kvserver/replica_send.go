package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"reflect"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/poison"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/txnwait"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/circuit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

var optimisticEvalLimitedScans = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.concurrency.optimistic_eval_limited_scans.enabled",
	"when true, limited scans are optimistically evaluated in the sense of not checking for "+
		"conflicting latches or locks up front for the full key range of the scan, and instead "+
		"subsequently checking for conflicts only over the key range that was read",
	true,
)

func (r *Replica) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120228)
	return r.sendWithoutRangeID(ctx, &ba)
}

func (r *Replica) sendWithoutRangeID(
	ctx context.Context, ba *roachpb.BatchRequest,
) (_ *roachpb.BatchResponse, rErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120229)
	var br *roachpb.BatchResponse

	if r.leaseholderStats != nil && func() bool {
		__antithesis_instrumentation__.Notify(120240)
		return ba.Header.GatewayNodeID != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(120241)
		r.leaseholderStats.recordCount(r.getBatchRequestQPS(ctx, ba), ba.Header.GatewayNodeID)
	} else {
		__antithesis_instrumentation__.Notify(120242)
	}
	__antithesis_instrumentation__.Notify(120230)

	ctx = r.AnnotateCtx(ctx)

	r.maybeInitializeRaftGroup(ctx)

	isReadOnly := ba.IsReadOnly()
	if err := r.checkBatchRequest(ba, isReadOnly); err != nil {
		__antithesis_instrumentation__.Notify(120243)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120244)
	}
	__antithesis_instrumentation__.Notify(120231)

	if err := r.maybeBackpressureBatch(ctx, ba); err != nil {
		__antithesis_instrumentation__.Notify(120245)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120246)
	}
	__antithesis_instrumentation__.Notify(120232)
	if err := r.maybeRateLimitBatch(ctx, ba); err != nil {
		__antithesis_instrumentation__.Notify(120247)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120248)
	}
	__antithesis_instrumentation__.Notify(120233)
	if err := r.maybeCommitWaitBeforeCommitTrigger(ctx, ba); err != nil {
		__antithesis_instrumentation__.Notify(120249)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120250)
	}
	__antithesis_instrumentation__.Notify(120234)

	ba, err := maybeStripInFlightWrites(ba)
	if err != nil {
		__antithesis_instrumentation__.Notify(120251)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120252)
	}
	__antithesis_instrumentation__.Notify(120235)

	if filter := r.store.cfg.TestingKnobs.TestingRequestFilter; filter != nil {
		__antithesis_instrumentation__.Notify(120253)
		if pErr := filter(ctx, *ba); pErr != nil {
			__antithesis_instrumentation__.Notify(120254)
			return nil, pErr
		} else {
			__antithesis_instrumentation__.Notify(120255)
		}
	} else {
		__antithesis_instrumentation__.Notify(120256)
	}
	__antithesis_instrumentation__.Notify(120236)

	var pErr *roachpb.Error
	if isReadOnly {
		__antithesis_instrumentation__.Notify(120257)
		log.Event(ctx, "read-only path")
		fn := (*Replica).executeReadOnlyBatch
		br, pErr = r.executeBatchWithConcurrencyRetries(ctx, ba, fn)
	} else {
		__antithesis_instrumentation__.Notify(120258)
		if ba.IsWrite() {
			__antithesis_instrumentation__.Notify(120259)
			log.Event(ctx, "read-write path")
			fn := (*Replica).executeWriteBatch
			br, pErr = r.executeBatchWithConcurrencyRetries(ctx, ba, fn)
		} else {
			__antithesis_instrumentation__.Notify(120260)
			if ba.IsAdmin() {
				__antithesis_instrumentation__.Notify(120261)
				log.Event(ctx, "admin path")
				br, pErr = r.executeAdminBatch(ctx, ba)
			} else {
				__antithesis_instrumentation__.Notify(120262)
				if len(ba.Requests) == 0 {
					__antithesis_instrumentation__.Notify(120263)

					log.Fatalf(ctx, "empty batch")
				} else {
					__antithesis_instrumentation__.Notify(120264)
					log.Fatalf(ctx, "don't know how to handle command %s", ba)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(120237)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(120265)
		log.Eventf(ctx, "replica.Send got error: %s", pErr)
	} else {
		__antithesis_instrumentation__.Notify(120266)
		if filter := r.store.cfg.TestingKnobs.TestingResponseFilter; filter != nil {
			__antithesis_instrumentation__.Notify(120267)
			pErr = filter(ctx, *ba, br)
		} else {
			__antithesis_instrumentation__.Notify(120268)
		}
	}
	__antithesis_instrumentation__.Notify(120238)

	if pErr == nil {
		__antithesis_instrumentation__.Notify(120269)
		r.maybeAddRangeInfoToResponse(ctx, ba, br)
	} else {
		__antithesis_instrumentation__.Notify(120270)
	}
	__antithesis_instrumentation__.Notify(120239)

	r.recordImpactOnRateLimiter(ctx, br)
	return br, pErr
}

func (r *Replica) maybeCommitWaitBeforeCommitTrigger(
	ctx context.Context, ba *roachpb.BatchRequest,
) error {
	__antithesis_instrumentation__.Notify(120271)
	args, hasET := ba.GetArg(roachpb.EndTxn)
	if !hasET {
		__antithesis_instrumentation__.Notify(120278)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120279)
	}
	__antithesis_instrumentation__.Notify(120272)
	et := args.(*roachpb.EndTxnRequest)
	if !et.Commit || func() bool {
		__antithesis_instrumentation__.Notify(120280)
		return et.InternalCommitTrigger == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(120281)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(120282)
	}
	__antithesis_instrumentation__.Notify(120273)
	txn := ba.Txn
	if txn.ReadTimestamp != txn.WriteTimestamp && func() bool {
		__antithesis_instrumentation__.Notify(120283)
		return !ba.CanForwardReadTimestamp == true
	}() == true {
		__antithesis_instrumentation__.Notify(120284)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(120285)
	}
	__antithesis_instrumentation__.Notify(120274)

	if !txn.WriteTimestamp.Synthetic {
		__antithesis_instrumentation__.Notify(120286)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120287)
	}
	__antithesis_instrumentation__.Notify(120275)
	if !r.Clock().Now().Less(txn.WriteTimestamp) {
		__antithesis_instrumentation__.Notify(120288)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120289)
	}
	__antithesis_instrumentation__.Notify(120276)

	waitUntil := txn.WriteTimestamp
	before := r.Clock().PhysicalTime()
	est := waitUntil.GoTime().Sub(before)
	log.VEventf(ctx, 1, "performing server-side commit-wait sleep for ~%s", est)

	if err := r.Clock().SleepUntil(ctx, waitUntil); err != nil {
		__antithesis_instrumentation__.Notify(120290)
		return err
	} else {
		__antithesis_instrumentation__.Notify(120291)
	}
	__antithesis_instrumentation__.Notify(120277)

	after := r.Clock().PhysicalTime()
	log.VEventf(ctx, 1, "completed server-side commit-wait sleep, took %s", after.Sub(before))
	r.store.metrics.CommitWaitsBeforeCommitTrigger.Inc(1)
	return nil
}

func (r *Replica) maybeAddRangeInfoToResponse(
	ctx context.Context, ba *roachpb.BatchRequest, br *roachpb.BatchResponse,
) {
	__antithesis_instrumentation__.Notify(120292)

	cri := &ba.ClientRangeInfo
	ri := r.GetRangeInfo(ctx)
	needInfo := cri.ExplicitlyRequested || func() bool {
		__antithesis_instrumentation__.Notify(120297)
		return (cri.DescriptorGeneration < ri.Desc.Generation) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(120298)
		return (cri.LeaseSequence < ri.Lease.Sequence) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(120299)
		return (cri.ClosedTimestampPolicy != ri.ClosedTimestampPolicy) == true
	}() == true
	if !needInfo {
		__antithesis_instrumentation__.Notify(120300)
		return
	} else {
		__antithesis_instrumentation__.Notify(120301)
	}
	__antithesis_instrumentation__.Notify(120293)
	log.VEventf(ctx, 3, "client had stale range info; returning an update")
	br.RangeInfos = []roachpb.RangeInfo{ri}

	if cri.DescriptorGeneration >= ri.Desc.Generation {
		__antithesis_instrumentation__.Notify(120302)
		return
	} else {
		__antithesis_instrumentation__.Notify(120303)
	}
	__antithesis_instrumentation__.Notify(120294)

	maybeAddRange := func(repl *Replica) {
		__antithesis_instrumentation__.Notify(120304)
		if repl.Desc().Generation != ri.Desc.Generation {
			__antithesis_instrumentation__.Notify(120306)

			return
		} else {
			__antithesis_instrumentation__.Notify(120307)
		}
		__antithesis_instrumentation__.Notify(120305)

		br.RangeInfos = append(br.RangeInfos, repl.GetRangeInfo(ctx))
	}
	__antithesis_instrumentation__.Notify(120295)

	if repl := r.store.lookupPrecedingReplica(ri.Desc.StartKey); repl != nil {
		__antithesis_instrumentation__.Notify(120308)
		maybeAddRange(repl)
	} else {
		__antithesis_instrumentation__.Notify(120309)
	}
	__antithesis_instrumentation__.Notify(120296)
	if repl := r.store.LookupReplica(ri.Desc.EndKey); repl != nil {
		__antithesis_instrumentation__.Notify(120310)
		maybeAddRange(repl)
	} else {
		__antithesis_instrumentation__.Notify(120311)
	}
}

type batchExecutionFn func(
	*Replica, context.Context, *roachpb.BatchRequest, *concurrency.Guard,
) (*roachpb.BatchResponse, *concurrency.Guard, *roachpb.Error)

var _ batchExecutionFn = (*Replica).executeWriteBatch
var _ batchExecutionFn = (*Replica).executeReadOnlyBatch

func (r *Replica) executeBatchWithConcurrencyRetries(
	ctx context.Context, ba *roachpb.BatchRequest, fn batchExecutionFn,
) (br *roachpb.BatchResponse, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120312)

	var latchSpans, lockSpans *spanset.SpanSet
	var requestEvalKind concurrency.RequestEvalKind
	var g *concurrency.Guard
	defer func() {
		__antithesis_instrumentation__.Notify(120315)

		if g != nil {
			__antithesis_instrumentation__.Notify(120316)
			r.concMgr.FinishReq(g)
		} else {
			__antithesis_instrumentation__.Notify(120317)
		}
	}()
	__antithesis_instrumentation__.Notify(120313)
	pp := poison.Policy_Error
	if r.signallerForBatch(ba).C() == nil {
		__antithesis_instrumentation__.Notify(120318)

		pp = poison.Policy_Wait
	} else {
		__antithesis_instrumentation__.Notify(120319)
	}
	__antithesis_instrumentation__.Notify(120314)
	for first := true; ; first = false {
		__antithesis_instrumentation__.Notify(120320)

		if err := ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(120327)
			return nil, roachpb.NewError(errors.Wrap(err, "aborted during Replica.Send"))
		} else {
			__antithesis_instrumentation__.Notify(120328)
		}
		__antithesis_instrumentation__.Notify(120321)

		if latchSpans == nil && func() bool {
			__antithesis_instrumentation__.Notify(120329)
			return g == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(120330)
			var err error
			latchSpans, lockSpans, requestEvalKind, err = r.collectSpans(ba)
			if err != nil {
				__antithesis_instrumentation__.Notify(120331)
				return nil, roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(120332)
			}
		} else {
			__antithesis_instrumentation__.Notify(120333)
		}
		__antithesis_instrumentation__.Notify(120322)

		if first {
			__antithesis_instrumentation__.Notify(120334)
			r.recordBatchForLoadBasedSplitting(ctx, ba, latchSpans)
		} else {
			__antithesis_instrumentation__.Notify(120335)
		}
		__antithesis_instrumentation__.Notify(120323)

		var resp []roachpb.ResponseUnion
		g, resp, pErr = r.concMgr.SequenceReq(ctx, g, concurrency.Request{
			Txn:             ba.Txn,
			Timestamp:       ba.Timestamp,
			Priority:        ba.UserPriority,
			ReadConsistency: ba.ReadConsistency,
			WaitPolicy:      ba.WaitPolicy,
			LockTimeout:     ba.LockTimeout,
			PoisonPolicy:    pp,
			Requests:        ba.Requests,
			LatchSpans:      latchSpans,
			LockSpans:       lockSpans,
		}, requestEvalKind)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(120336)
			if poisonErr := (*poison.PoisonedError)(nil); errors.As(pErr.GoError(), &poisonErr) {
				__antithesis_instrumentation__.Notify(120338)

				pErr = roachpb.NewError(r.replicaUnavailableError(errors.CombineErrors(
					errors.Mark(poisonErr, circuit.ErrBreakerOpen),
					r.breaker.Signal().Err(),
				)))
			} else {
				__antithesis_instrumentation__.Notify(120339)
			}
			__antithesis_instrumentation__.Notify(120337)
			return nil, pErr
		} else {
			__antithesis_instrumentation__.Notify(120340)
			if resp != nil {
				__antithesis_instrumentation__.Notify(120341)
				br = new(roachpb.BatchResponse)
				br.Responses = resp
				return br, nil
			} else {
				__antithesis_instrumentation__.Notify(120342)
			}
		}
		__antithesis_instrumentation__.Notify(120324)
		latchSpans, lockSpans = nil, nil

		br, g, pErr = fn(r, ctx, ba, g)
		if pErr == nil {
			__antithesis_instrumentation__.Notify(120343)

			return br, nil
		} else {
			__antithesis_instrumentation__.Notify(120344)
			if !isConcurrencyRetryError(pErr) {
				__antithesis_instrumentation__.Notify(120345)

				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120346)
			}
		}
		__antithesis_instrumentation__.Notify(120325)

		g.AssertLatches()
		if filter := r.store.cfg.TestingKnobs.TestingConcurrencyRetryFilter; filter != nil {
			__antithesis_instrumentation__.Notify(120347)
			filter(ctx, *ba, pErr)
		} else {
			__antithesis_instrumentation__.Notify(120348)
		}
		__antithesis_instrumentation__.Notify(120326)

		requestEvalKind = concurrency.PessimisticEval

		switch t := pErr.GetDetail().(type) {
		case *roachpb.WriteIntentError:
			__antithesis_instrumentation__.Notify(120349)

			if g, pErr = r.handleWriteIntentError(ctx, ba, g, pErr, t); pErr != nil {
				__antithesis_instrumentation__.Notify(120357)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120358)
			}
		case *roachpb.TransactionPushError:
			__antithesis_instrumentation__.Notify(120350)

			if g, pErr = r.handleTransactionPushError(ctx, ba, g, pErr, t); pErr != nil {
				__antithesis_instrumentation__.Notify(120359)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120360)
			}
		case *roachpb.IndeterminateCommitError:
			__antithesis_instrumentation__.Notify(120351)

			latchSpans, lockSpans = g.TakeSpanSets()
			r.concMgr.FinishReq(g)
			g = nil

			if pErr = r.handleIndeterminateCommitError(ctx, ba, pErr, t); pErr != nil {
				__antithesis_instrumentation__.Notify(120361)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120362)
			}
		case *roachpb.ReadWithinUncertaintyIntervalError:
			__antithesis_instrumentation__.Notify(120352)

			r.concMgr.FinishReq(g)
			g = nil

			latchSpans, lockSpans = nil, nil

			ba, pErr = r.handleReadWithinUncertaintyIntervalError(ctx, ba, pErr, t)
			if pErr != nil {
				__antithesis_instrumentation__.Notify(120363)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120364)
			}
		case *roachpb.InvalidLeaseError:
			__antithesis_instrumentation__.Notify(120353)

			latchSpans, lockSpans = g.TakeSpanSets()
			r.concMgr.FinishReq(g)
			g = nil

			if pErr = r.handleInvalidLeaseError(ctx, ba); pErr != nil {
				__antithesis_instrumentation__.Notify(120365)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120366)
			}
		case *roachpb.MergeInProgressError:
			__antithesis_instrumentation__.Notify(120354)

			latchSpans, lockSpans = g.TakeSpanSets()
			r.concMgr.FinishReq(g)
			g = nil

			if pErr = r.handleMergeInProgressError(ctx, ba, pErr, t); pErr != nil {
				__antithesis_instrumentation__.Notify(120367)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120368)
			}
		case *roachpb.OptimisticEvalConflictsError:
			__antithesis_instrumentation__.Notify(120355)

			requestEvalKind = concurrency.PessimisticAfterFailedOptimisticEval
		default:
			__antithesis_instrumentation__.Notify(120356)
			log.Fatalf(ctx, "unexpected concurrency retry error %T", t)
		}

	}
}

func isConcurrencyRetryError(pErr *roachpb.Error) bool {
	__antithesis_instrumentation__.Notify(120369)
	switch pErr.GetDetail().(type) {
	case *roachpb.WriteIntentError:
		__antithesis_instrumentation__.Notify(120371)

	case *roachpb.TransactionPushError:
		__antithesis_instrumentation__.Notify(120372)

	case *roachpb.IndeterminateCommitError:
		__antithesis_instrumentation__.Notify(120373)

	case *roachpb.ReadWithinUncertaintyIntervalError:
		__antithesis_instrumentation__.Notify(120374)

	case *roachpb.InvalidLeaseError:
		__antithesis_instrumentation__.Notify(120375)

	case *roachpb.MergeInProgressError:
		__antithesis_instrumentation__.Notify(120376)

	case *roachpb.OptimisticEvalConflictsError:
		__antithesis_instrumentation__.Notify(120377)

	default:
		__antithesis_instrumentation__.Notify(120378)
		return false
	}
	__antithesis_instrumentation__.Notify(120370)
	return true
}

func maybeAttachLease(pErr *roachpb.Error, lease *roachpb.Lease) *roachpb.Error {
	__antithesis_instrumentation__.Notify(120379)
	if wiErr, ok := pErr.GetDetail().(*roachpb.WriteIntentError); ok {
		__antithesis_instrumentation__.Notify(120381)

		if lease.Empty() {
			__antithesis_instrumentation__.Notify(120383)
			return roachpb.NewErrorWithTxn(&roachpb.InvalidLeaseError{}, pErr.GetTxn())
		} else {
			__antithesis_instrumentation__.Notify(120384)
		}
		__antithesis_instrumentation__.Notify(120382)
		wiErr.LeaseSequence = lease.Sequence
		return roachpb.NewErrorWithTxn(wiErr, pErr.GetTxn())
	} else {
		__antithesis_instrumentation__.Notify(120385)
	}
	__antithesis_instrumentation__.Notify(120380)
	return pErr
}

func (r *Replica) handleWriteIntentError(
	ctx context.Context,
	ba *roachpb.BatchRequest,
	g *concurrency.Guard,
	pErr *roachpb.Error,
	t *roachpb.WriteIntentError,
) (*concurrency.Guard, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120386)
	if r.store.cfg.TestingKnobs.DontPushOnWriteIntentError {
		__antithesis_instrumentation__.Notify(120388)
		return g, pErr
	} else {
		__antithesis_instrumentation__.Notify(120389)
	}
	__antithesis_instrumentation__.Notify(120387)

	return r.concMgr.HandleWriterIntentError(ctx, g, t.LeaseSequence, t)
}

func (r *Replica) handleTransactionPushError(
	ctx context.Context,
	ba *roachpb.BatchRequest,
	g *concurrency.Guard,
	pErr *roachpb.Error,
	t *roachpb.TransactionPushError,
) (*concurrency.Guard, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120390)

	dontRetry := r.store.cfg.TestingKnobs.DontRetryPushTxnFailures
	if !dontRetry && func() bool {
		__antithesis_instrumentation__.Notify(120393)
		return ba.IsSinglePushTxnRequest() == true
	}() == true {
		__antithesis_instrumentation__.Notify(120394)
		pushReq := ba.Requests[0].GetInner().(*roachpb.PushTxnRequest)
		dontRetry = txnwait.ShouldPushImmediately(pushReq)
	} else {
		__antithesis_instrumentation__.Notify(120395)
	}
	__antithesis_instrumentation__.Notify(120391)
	if dontRetry {
		__antithesis_instrumentation__.Notify(120396)
		return g, pErr
	} else {
		__antithesis_instrumentation__.Notify(120397)
	}
	__antithesis_instrumentation__.Notify(120392)

	return r.concMgr.HandleTransactionPushError(ctx, g, t), nil
}

func (r *Replica) handleIndeterminateCommitError(
	ctx context.Context,
	ba *roachpb.BatchRequest,
	pErr *roachpb.Error,
	t *roachpb.IndeterminateCommitError,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(120398)
	if r.store.cfg.TestingKnobs.DontRecoverIndeterminateCommits {
		__antithesis_instrumentation__.Notify(120401)
		return pErr
	} else {
		__antithesis_instrumentation__.Notify(120402)
	}
	__antithesis_instrumentation__.Notify(120399)

	if _, err := r.store.recoveryMgr.ResolveIndeterminateCommit(ctx, t); err != nil {
		__antithesis_instrumentation__.Notify(120403)

		if errors.HasType(err, (*roachpb.AmbiguousResultError)(nil)) {
			__antithesis_instrumentation__.Notify(120405)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(120406)
		}
		__antithesis_instrumentation__.Notify(120404)

		newPErr := roachpb.NewError(err)
		newPErr.Index = pErr.Index
		return newPErr
	} else {
		__antithesis_instrumentation__.Notify(120407)
	}
	__antithesis_instrumentation__.Notify(120400)

	return nil
}

func (r *Replica) handleReadWithinUncertaintyIntervalError(
	ctx context.Context,
	ba *roachpb.BatchRequest,
	pErr *roachpb.Error,
	t *roachpb.ReadWithinUncertaintyIntervalError,
) (*roachpb.BatchRequest, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120408)

	if !canDoServersideRetry(ctx, pErr, ba, nil, nil, nil) {
		__antithesis_instrumentation__.Notify(120411)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(120412)
	}
	__antithesis_instrumentation__.Notify(120409)
	if ba.Txn == nil && func() bool {
		__antithesis_instrumentation__.Notify(120413)
		return ba.Timestamp.Synthetic == true
	}() == true {
		__antithesis_instrumentation__.Notify(120414)

		var cancel func()
		ctx, cancel = r.store.Stopper().WithCancelOnQuiesce(ctx)
		defer cancel()
		if err := r.Clock().SleepUntil(ctx, ba.Timestamp); err != nil {
			__antithesis_instrumentation__.Notify(120415)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(120416)
		}
	} else {
		__antithesis_instrumentation__.Notify(120417)
	}
	__antithesis_instrumentation__.Notify(120410)
	return ba, nil
}

func (r *Replica) handleInvalidLeaseError(
	ctx context.Context, ba *roachpb.BatchRequest,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(120418)

	_, pErr := r.redirectOnOrAcquireLeaseForRequest(ctx, ba.Timestamp, r.signallerForBatch(ba))

	return pErr
}

func (r *Replica) handleMergeInProgressError(
	ctx context.Context,
	ba *roachpb.BatchRequest,
	pErr *roachpb.Error,
	t *roachpb.MergeInProgressError,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(120419)

	mergeCompleteCh := r.getMergeCompleteCh()
	if mergeCompleteCh == nil {
		__antithesis_instrumentation__.Notify(120422)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(120423)
	}
	__antithesis_instrumentation__.Notify(120420)

	if ba.IsSingleTransferLeaseRequest() {
		__antithesis_instrumentation__.Notify(120424)
		return roachpb.NewErrorf("cannot transfer lease while merge in progress")
	} else {
		__antithesis_instrumentation__.Notify(120425)
	}
	__antithesis_instrumentation__.Notify(120421)
	log.Event(ctx, "waiting on in-progress range merge")
	select {
	case <-mergeCompleteCh:
		__antithesis_instrumentation__.Notify(120426)

		return nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(120427)
		return roachpb.NewError(errors.Wrap(ctx.Err(), "aborted during merge"))
	case <-r.store.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(120428)
		return roachpb.NewError(&roachpb.NodeUnavailableError{})
	}
}

func (r *Replica) executeAdminBatch(
	ctx context.Context, ba *roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120429)
	if len(ba.Requests) != 1 {
		__antithesis_instrumentation__.Notify(120434)
		return nil, roachpb.NewErrorf("only single-element admin batches allowed")
	} else {
		__antithesis_instrumentation__.Notify(120435)
	}
	__antithesis_instrumentation__.Notify(120430)

	args := ba.Requests[0].GetInner()

	ctx, sp := tracing.EnsureChildSpan(ctx, r.AmbientContext.Tracer, reflect.TypeOf(args).String())
	defer sp.Finish()

	for {
		__antithesis_instrumentation__.Notify(120436)
		if err := ctx.Err(); err != nil {
			__antithesis_instrumentation__.Notify(120440)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(120441)
		}
		__antithesis_instrumentation__.Notify(120437)

		_, err := r.checkExecutionCanProceed(ctx, ba, nil)
		if err == nil {
			__antithesis_instrumentation__.Notify(120442)
			err = r.signallerForBatch(ba).Err()
		} else {
			__antithesis_instrumentation__.Notify(120443)
		}
		__antithesis_instrumentation__.Notify(120438)
		if err == nil {
			__antithesis_instrumentation__.Notify(120444)
			break
		} else {
			__antithesis_instrumentation__.Notify(120445)
		}
		__antithesis_instrumentation__.Notify(120439)
		switch {
		case errors.HasType(err, (*roachpb.InvalidLeaseError)(nil)):
			__antithesis_instrumentation__.Notify(120446)

			_, pErr := r.redirectOnOrAcquireLeaseForRequest(ctx, ba.Timestamp, r.signallerForBatch(ba))
			if pErr != nil {
				__antithesis_instrumentation__.Notify(120448)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(120449)
			}

		default:
			__antithesis_instrumentation__.Notify(120447)
			return nil, roachpb.NewError(err)
		}
	}
	__antithesis_instrumentation__.Notify(120431)

	var resp roachpb.Response
	var pErr *roachpb.Error
	switch tArgs := args.(type) {
	case *roachpb.AdminSplitRequest:
		__antithesis_instrumentation__.Notify(120450)
		var reply roachpb.AdminSplitResponse
		reply, pErr = r.AdminSplit(ctx, *tArgs, "manual")
		resp = &reply

	case *roachpb.AdminUnsplitRequest:
		__antithesis_instrumentation__.Notify(120451)
		var reply roachpb.AdminUnsplitResponse
		reply, pErr = r.AdminUnsplit(ctx, *tArgs, "manual")
		resp = &reply

	case *roachpb.AdminMergeRequest:
		__antithesis_instrumentation__.Notify(120452)
		var reply roachpb.AdminMergeResponse
		reply, pErr = r.AdminMerge(ctx, *tArgs, "manual")
		resp = &reply

	case *roachpb.AdminTransferLeaseRequest:
		__antithesis_instrumentation__.Notify(120453)
		pErr = roachpb.NewError(r.AdminTransferLease(ctx, tArgs.Target))
		resp = &roachpb.AdminTransferLeaseResponse{}

	case *roachpb.AdminChangeReplicasRequest:
		__antithesis_instrumentation__.Notify(120454)
		chgs := tArgs.Changes()
		desc, err := r.ChangeReplicas(ctx, &tArgs.ExpDesc, kvserverpb.SnapshotRequest_REBALANCE, kvserverpb.ReasonAdminRequest, "", chgs)
		pErr = roachpb.NewError(err)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(120460)
			resp = &roachpb.AdminChangeReplicasResponse{}
		} else {
			__antithesis_instrumentation__.Notify(120461)
			resp = &roachpb.AdminChangeReplicasResponse{
				Desc: *desc,
			}
		}

	case *roachpb.AdminRelocateRangeRequest:
		__antithesis_instrumentation__.Notify(120455)

		transferLeaseToFirstVoter := !tArgs.TransferLeaseToFirstVoterAccurate

		transferLeaseToFirstVoter = transferLeaseToFirstVoter || func() bool {
			__antithesis_instrumentation__.Notify(120462)
			return tArgs.TransferLeaseToFirstVoter == true
		}() == true
		err := r.AdminRelocateRange(
			ctx, *r.Desc(), tArgs.VoterTargets, tArgs.NonVoterTargets, transferLeaseToFirstVoter,
		)
		pErr = roachpb.NewError(err)
		resp = &roachpb.AdminRelocateRangeResponse{}

	case *roachpb.CheckConsistencyRequest:
		__antithesis_instrumentation__.Notify(120456)
		var reply roachpb.CheckConsistencyResponse
		reply, pErr = r.CheckConsistency(ctx, *tArgs)
		resp = &reply

	case *roachpb.AdminScatterRequest:
		__antithesis_instrumentation__.Notify(120457)
		reply, err := r.adminScatter(ctx, *tArgs)
		pErr = roachpb.NewError(err)
		resp = &reply

	case *roachpb.AdminVerifyProtectedTimestampRequest:
		__antithesis_instrumentation__.Notify(120458)
		reply, err := r.adminVerifyProtectedTimestamp(ctx, *tArgs)
		pErr = roachpb.NewError(err)
		resp = &reply

	default:
		__antithesis_instrumentation__.Notify(120459)
		return nil, roachpb.NewErrorf("unrecognized admin command: %T", args)
	}
	__antithesis_instrumentation__.Notify(120432)

	if pErr != nil {
		__antithesis_instrumentation__.Notify(120463)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(120464)
	}
	__antithesis_instrumentation__.Notify(120433)

	br := &roachpb.BatchResponse{}
	br.Add(resp)
	br.Txn = resp.Header().Txn
	return br, nil
}

func (r *Replica) getBatchRequestQPS(ctx context.Context, ba *roachpb.BatchRequest) float64 {
	__antithesis_instrumentation__.Notify(120465)
	var count float64 = 1

	requestFact := AddSSTableRequestSizeFactor.Get(&r.store.cfg.Settings.SV)
	if requestFact < 1 {
		__antithesis_instrumentation__.Notify(120468)
		return count
	} else {
		__antithesis_instrumentation__.Notify(120469)
	}
	__antithesis_instrumentation__.Notify(120466)

	var addSSTSize float64 = 0
	for _, req := range ba.Requests {
		__antithesis_instrumentation__.Notify(120470)
		switch t := req.GetInner().(type) {
		case *roachpb.AddSSTableRequest:
			__antithesis_instrumentation__.Notify(120471)
			addSSTSize += float64(len(t.Data))
		default:
			__antithesis_instrumentation__.Notify(120472)
			continue
		}
	}
	__antithesis_instrumentation__.Notify(120467)

	count += addSSTSize / float64(requestFact)
	return count
}

func (r *Replica) checkBatchRequest(ba *roachpb.BatchRequest, isReadOnly bool) error {
	__antithesis_instrumentation__.Notify(120473)
	if ba.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(120476)

		return errors.New("Replica.checkBatchRequest: batch does not have timestamp assigned")
	} else {
		__antithesis_instrumentation__.Notify(120477)
	}
	__antithesis_instrumentation__.Notify(120474)
	consistent := ba.ReadConsistency == roachpb.CONSISTENT
	if isReadOnly {
		__antithesis_instrumentation__.Notify(120478)
		if !consistent && func() bool {
			__antithesis_instrumentation__.Notify(120479)
			return ba.Txn != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(120480)

			return errors.Errorf("cannot allow %v reads within a transaction", ba.ReadConsistency)
		} else {
			__antithesis_instrumentation__.Notify(120481)
		}
	} else {
		__antithesis_instrumentation__.Notify(120482)
		if !consistent {
			__antithesis_instrumentation__.Notify(120483)
			return errors.Errorf("%v mode is only available to reads", ba.ReadConsistency)
		} else {
			__antithesis_instrumentation__.Notify(120484)
		}
	}
	__antithesis_instrumentation__.Notify(120475)

	return nil
}

func (r *Replica) collectSpans(
	ba *roachpb.BatchRequest,
) (latchSpans, lockSpans *spanset.SpanSet, requestEvalKind concurrency.RequestEvalKind, _ error) {
	__antithesis_instrumentation__.Notify(120485)
	latchSpans, lockSpans = spanset.New(), spanset.New()
	r.mu.RLock()
	desc := r.descRLocked()
	liveCount := r.mu.state.Stats.LiveCount
	r.mu.RUnlock()

	if ba.IsLocking() {
		__antithesis_instrumentation__.Notify(120490)
		latchGuess := len(ba.Requests)
		if et, ok := ba.GetArg(roachpb.EndTxn); ok {
			__antithesis_instrumentation__.Notify(120492)

			latchGuess += len(et.(*roachpb.EndTxnRequest).LockSpans) - 1
		} else {
			__antithesis_instrumentation__.Notify(120493)
		}
		__antithesis_instrumentation__.Notify(120491)
		latchSpans.Reserve(spanset.SpanReadWrite, spanset.SpanGlobal, latchGuess)
		lockSpans.Reserve(spanset.SpanReadWrite, spanset.SpanGlobal, len(ba.Requests))
	} else {
		__antithesis_instrumentation__.Notify(120494)
		latchSpans.Reserve(spanset.SpanReadOnly, spanset.SpanGlobal, len(ba.Requests))
		lockSpans.Reserve(spanset.SpanReadOnly, spanset.SpanGlobal, len(ba.Requests))
	}
	__antithesis_instrumentation__.Notify(120486)

	considerOptEval := ba.IsReadOnly() && func() bool {
		__antithesis_instrumentation__.Notify(120495)
		return ba.IsAllTransactional() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(120496)
		return ba.Header.MaxSpanRequestKeys > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(120497)
		return optimisticEvalLimitedScans.Get(&r.ClusterSettings().SV) == true
	}() == true

	hasScans := false
	numGets := 0

	batcheval.DeclareKeysForBatch(desc, &ba.Header, latchSpans)
	for _, union := range ba.Requests {
		__antithesis_instrumentation__.Notify(120498)
		inner := union.GetInner()
		if cmd, ok := batcheval.LookupCommand(inner.Method()); ok {
			__antithesis_instrumentation__.Notify(120499)
			cmd.DeclareKeys(desc, &ba.Header, inner, latchSpans, lockSpans, r.Clock().MaxOffset())
			if considerOptEval {
				__antithesis_instrumentation__.Notify(120500)
				switch inner.(type) {
				case *roachpb.ScanRequest, *roachpb.ReverseScanRequest:
					__antithesis_instrumentation__.Notify(120501)
					hasScans = true
				case *roachpb.GetRequest:
					__antithesis_instrumentation__.Notify(120502)
					numGets++
				}
			} else {
				__antithesis_instrumentation__.Notify(120503)
			}
		} else {
			__antithesis_instrumentation__.Notify(120504)
			return nil, nil, concurrency.PessimisticEval, errors.Errorf("unrecognized command %s", inner.Method())
		}
	}
	__antithesis_instrumentation__.Notify(120487)

	for _, s := range [...]*spanset.SpanSet{latchSpans, lockSpans} {
		__antithesis_instrumentation__.Notify(120505)
		s.SortAndDedup()

		if err := s.Validate(); err != nil {
			__antithesis_instrumentation__.Notify(120506)
			return nil, nil, concurrency.PessimisticEval, err
		} else {
			__antithesis_instrumentation__.Notify(120507)
		}
	}
	__antithesis_instrumentation__.Notify(120488)

	requestEvalKind = concurrency.PessimisticEval
	if considerOptEval {
		__antithesis_instrumentation__.Notify(120508)

		const k = 1
		upperBoundKeys := k * liveCount
		if !hasScans && func() bool {
			__antithesis_instrumentation__.Notify(120510)
			return int64(numGets) < upperBoundKeys == true
		}() == true {
			__antithesis_instrumentation__.Notify(120511)
			upperBoundKeys = int64(numGets)
		} else {
			__antithesis_instrumentation__.Notify(120512)
		}
		__antithesis_instrumentation__.Notify(120509)
		if ba.Header.MaxSpanRequestKeys < upperBoundKeys {
			__antithesis_instrumentation__.Notify(120513)
			requestEvalKind = concurrency.OptimisticEval
		} else {
			__antithesis_instrumentation__.Notify(120514)
		}
	} else {
		__antithesis_instrumentation__.Notify(120515)
	}
	__antithesis_instrumentation__.Notify(120489)

	return latchSpans, lockSpans, requestEvalKind, nil
}

type endCmds struct {
	repl *Replica
	g    *concurrency.Guard
	st   kvserverpb.LeaseStatus
}

func (ec *endCmds) move() endCmds {
	__antithesis_instrumentation__.Notify(120516)
	res := *ec
	*ec = endCmds{}
	return res
}

func (ec *endCmds) poison() {
	__antithesis_instrumentation__.Notify(120517)
	if ec.repl == nil {
		__antithesis_instrumentation__.Notify(120519)

		return
	} else {
		__antithesis_instrumentation__.Notify(120520)
	}
	__antithesis_instrumentation__.Notify(120518)
	ec.repl.concMgr.PoisonReq(ec.g)
}

func (ec *endCmds) done(
	ctx context.Context, ba *roachpb.BatchRequest, br *roachpb.BatchResponse, pErr *roachpb.Error,
) {
	__antithesis_instrumentation__.Notify(120521)
	if ec.repl == nil {
		__antithesis_instrumentation__.Notify(120524)

		return
	} else {
		__antithesis_instrumentation__.Notify(120525)
	}
	__antithesis_instrumentation__.Notify(120522)
	defer ec.move()

	if ba.ReadConsistency == roachpb.CONSISTENT && func() bool {
		__antithesis_instrumentation__.Notify(120526)
		return ec.st.State == kvserverpb.LeaseState_VALID == true
	}() == true {
		__antithesis_instrumentation__.Notify(120527)
		ec.repl.updateTimestampCache(ctx, &ec.st, ba, br, pErr)
	} else {
		__antithesis_instrumentation__.Notify(120528)
	}
	__antithesis_instrumentation__.Notify(120523)

	if ec.g != nil {
		__antithesis_instrumentation__.Notify(120529)
		ec.repl.concMgr.FinishReq(ec.g)
	} else {
		__antithesis_instrumentation__.Notify(120530)
	}
}
