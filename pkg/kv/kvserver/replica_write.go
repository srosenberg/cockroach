package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var migrateApplicationTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.migration.migrate_application.timeout",
	"timeout for a Migrate request to be applied across all replicas of a range",
	1*time.Minute,
	settings.PositiveDuration,
)

func (r *Replica) executeWriteBatch(
	ctx context.Context, ba *roachpb.BatchRequest, g *concurrency.Guard,
) (br *roachpb.BatchResponse, _ *concurrency.Guard, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(120943)
	startTime := timeutil.Now()

	readOnlyCmdMu := &r.readOnlyCmdMu
	readOnlyCmdMu.RLock()
	defer func() {
		__antithesis_instrumentation__.Notify(120951)
		if readOnlyCmdMu != nil {
			__antithesis_instrumentation__.Notify(120952)
			readOnlyCmdMu.RUnlock()
		} else {
			__antithesis_instrumentation__.Notify(120953)
		}
	}()
	__antithesis_instrumentation__.Notify(120944)

	st, err := r.checkExecutionCanProceed(ctx, ba, g)
	if err != nil {
		__antithesis_instrumentation__.Notify(120954)
		return nil, g, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120955)
	}
	__antithesis_instrumentation__.Notify(120945)

	if err := r.signallerForBatch(ba).Err(); err != nil {
		__antithesis_instrumentation__.Notify(120956)
		return nil, g, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(120957)
	}
	__antithesis_instrumentation__.Notify(120946)

	ui := uncertainty.ComputeInterval(&ba.Header, st, r.Clock().MaxOffset())

	var minTS hlc.Timestamp
	var tok TrackedRequestToken
	if ba.AppliesTimestampCache() {
		__antithesis_instrumentation__.Notify(120958)
		minTS, tok = r.mu.proposalBuf.TrackEvaluatingRequest(ctx, ba.WriteTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(120959)
	}
	__antithesis_instrumentation__.Notify(120947)
	defer tok.DoneIfNotMoved(ctx)

	if bumped := r.applyTimestampCache(ctx, ba, minTS); bumped {
		__antithesis_instrumentation__.Notify(120960)

		defer func() {
			__antithesis_instrumentation__.Notify(120961)
			if br != nil && func() bool {
				__antithesis_instrumentation__.Notify(120962)
				return ba.Txn != nil == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(120963)
				return br.Txn == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(120964)
				log.Fatalf(ctx, "assertion failed: transaction updated by "+
					"timestamp cache, but transaction returned in response; "+
					"updated timestamp would have been lost (recovered): "+
					"%s in batch %s", ba.Txn, ba,
				)
			} else {
				__antithesis_instrumentation__.Notify(120965)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(120966)
	}
	__antithesis_instrumentation__.Notify(120948)
	log.Event(ctx, "applied timestamp cache")

	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(120967)
		log.VEventf(ctx, 2, "%s before proposing: %s", err, ba.Summary())
		return nil, g, roachpb.NewError(errors.Wrapf(err, "aborted before proposing"))
	} else {
		__antithesis_instrumentation__.Notify(120968)
	}
	__antithesis_instrumentation__.Notify(120949)

	ch, abandon, _, pErr := r.evalAndPropose(ctx, ba, g, st, ui, tok.Move(ctx))
	if pErr != nil {
		__antithesis_instrumentation__.Notify(120969)
		if cErr, ok := pErr.GetDetail().(*roachpb.ReplicaCorruptionError); ok {
			__antithesis_instrumentation__.Notify(120971)

			readOnlyCmdMu.RUnlock()
			readOnlyCmdMu = nil

			r.raftMu.Lock()
			defer r.raftMu.Unlock()

			return nil, g, r.setCorruptRaftMuLocked(ctx, cErr)
		} else {
			__antithesis_instrumentation__.Notify(120972)
		}
		__antithesis_instrumentation__.Notify(120970)
		return nil, g, pErr
	} else {
		__antithesis_instrumentation__.Notify(120973)
	}
	__antithesis_instrumentation__.Notify(120950)
	g = nil

	readOnlyCmdMu.RUnlock()
	readOnlyCmdMu = nil

	ctxDone := ctx.Done()
	shouldQuiesce := r.store.stopper.ShouldQuiesce()

	for {
		__antithesis_instrumentation__.Notify(120974)
		select {
		case propResult := <-ch:
			__antithesis_instrumentation__.Notify(120975)

			if len(propResult.EndTxns) > 0 {
				__antithesis_instrumentation__.Notify(120983)
				if err := r.store.intentResolver.CleanupTxnIntentsAsync(
					ctx, r.RangeID, propResult.EndTxns, true,
				); err != nil {
					__antithesis_instrumentation__.Notify(120984)
					log.Warningf(ctx, "transaction cleanup failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(120985)
				}
			} else {
				__antithesis_instrumentation__.Notify(120986)
			}
			__antithesis_instrumentation__.Notify(120976)
			if len(propResult.EncounteredIntents) > 0 {
				__antithesis_instrumentation__.Notify(120987)
				if err := r.store.intentResolver.CleanupIntentsAsync(
					ctx, propResult.EncounteredIntents, true,
				); err != nil {
					__antithesis_instrumentation__.Notify(120988)
					log.Warningf(ctx, "intent cleanup failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(120989)
				}
			} else {
				__antithesis_instrumentation__.Notify(120990)
			}
			__antithesis_instrumentation__.Notify(120977)
			if ba.Requests[0].GetMigrate() != nil && func() bool {
				__antithesis_instrumentation__.Notify(120991)
				return propResult.Err == nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(120992)

				applicationErr := contextutil.RunWithTimeout(ctx, "wait for Migrate application",
					migrateApplicationTimeout.Get(&r.ClusterSettings().SV),
					func(ctx context.Context) error {
						__antithesis_instrumentation__.Notify(120994)
						desc := r.Desc()
						return waitForApplication(
							ctx, r.store.cfg.NodeDialer, desc.RangeID, desc.Replicas().Descriptors(),

							r.GetLeaseAppliedIndex())
					})
				__antithesis_instrumentation__.Notify(120993)
				propResult.Err = roachpb.NewError(applicationErr)
			} else {
				__antithesis_instrumentation__.Notify(120995)
			}
			__antithesis_instrumentation__.Notify(120978)
			if propResult.Err != nil && func() bool {
				__antithesis_instrumentation__.Notify(120996)
				return ba.IsSingleProbeRequest() == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(120997)
				return errors.Is(
					propResult.Err.GoError(), noopOnProbeCommandErr.GoError(),
				) == true
			}() == true {
				__antithesis_instrumentation__.Notify(120998)

				propResult.Reply, propResult.Err = ba.CreateReply(), nil
			} else {
				__antithesis_instrumentation__.Notify(120999)
			}
			__antithesis_instrumentation__.Notify(120979)

			return propResult.Reply, nil, propResult.Err

		case <-ctxDone:
			__antithesis_instrumentation__.Notify(120980)

			if _, ok := ba.GetArg(roachpb.EndTxn); ok {
				__antithesis_instrumentation__.Notify(121000)
				const taskName = "async txn cleanup"
				_ = r.store.stopper.RunAsyncTask(
					r.AnnotateCtx(context.Background()),
					taskName,
					func(ctx context.Context) {
						__antithesis_instrumentation__.Notify(121001)
						err := contextutil.RunWithTimeout(ctx, taskName, 20*time.Second,
							func(ctx context.Context) error {
								__antithesis_instrumentation__.Notify(121003)
								select {
								case propResult := <-ch:
									__antithesis_instrumentation__.Notify(121005)
									if len(propResult.EndTxns) > 0 {
										__antithesis_instrumentation__.Notify(121008)
										return r.store.intentResolver.CleanupTxnIntentsAsync(ctx,
											r.RangeID, propResult.EndTxns, false)
									} else {
										__antithesis_instrumentation__.Notify(121009)
									}
								case <-shouldQuiesce:
									__antithesis_instrumentation__.Notify(121006)
								case <-ctx.Done():
									__antithesis_instrumentation__.Notify(121007)
								}
								__antithesis_instrumentation__.Notify(121004)
								return ctx.Err()
							})
						__antithesis_instrumentation__.Notify(121002)
						if err != nil {
							__antithesis_instrumentation__.Notify(121010)
							log.Warningf(ctx, "transaction cleanup failed: %v", err)
							r.store.intentResolver.Metrics.FinalizedTxnCleanupFailed.Inc(1)
						} else {
							__antithesis_instrumentation__.Notify(121011)
						}
					})
			} else {
				__antithesis_instrumentation__.Notify(121012)
			}
			__antithesis_instrumentation__.Notify(120981)
			abandon()
			dur := timeutil.Since(startTime)
			log.VEventf(ctx, 2, "context cancellation after %.2fs of attempting command %s",
				dur.Seconds(), ba)
			return nil, nil, roachpb.NewError(roachpb.NewAmbiguousResultError(
				errors.Wrapf(ctx.Err(), "after %.2fs of attempting command", dur.Seconds()),
			))

		case <-shouldQuiesce:
			__antithesis_instrumentation__.Notify(120982)

			abandon()
			log.VEventf(ctx, 2, "shutdown cancellation after %0.1fs of attempting command %s",
				timeutil.Since(startTime).Seconds(), ba)
			return nil, nil, roachpb.NewError(roachpb.NewAmbiguousResultErrorf("server shutdown"))
		}
	}
}

func (r *Replica) canAttempt1PCEvaluation(
	ctx context.Context, ba *roachpb.BatchRequest, g *concurrency.Guard,
) bool {
	__antithesis_instrumentation__.Notify(121013)
	if !isOnePhaseCommit(ba) {
		__antithesis_instrumentation__.Notify(121018)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121019)
	}
	__antithesis_instrumentation__.Notify(121014)

	if ba.Timestamp != ba.Txn.WriteTimestamp {
		__antithesis_instrumentation__.Notify(121020)
		log.Fatalf(ctx, "unexpected 1PC execution with diverged timestamp. %s != %s",
			ba.Timestamp, ba.Txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(121021)
	}
	__antithesis_instrumentation__.Notify(121015)

	ok, minCommitTS, _ := r.CanCreateTxnRecord(ctx, ba.Txn.ID, ba.Txn.Key, ba.Txn.MinTimestamp)
	if !ok {
		__antithesis_instrumentation__.Notify(121022)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121023)
	}
	__antithesis_instrumentation__.Notify(121016)
	if ba.Timestamp.Less(minCommitTS) {
		__antithesis_instrumentation__.Notify(121024)
		ba.Txn.WriteTimestamp = minCommitTS

		return maybeBumpReadTimestampToWriteTimestamp(ctx, ba, g)
	} else {
		__antithesis_instrumentation__.Notify(121025)
	}
	__antithesis_instrumentation__.Notify(121017)
	return true
}

func (r *Replica) evaluateWriteBatch(
	ctx context.Context,
	idKey kvserverbase.CmdIDKey,
	ba *roachpb.BatchRequest,
	ui uncertainty.Interval,
	g *concurrency.Guard,
) (storage.Batch, enginepb.MVCCStats, *roachpb.BatchResponse, result.Result, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(121026)
	log.Event(ctx, "executing read-write batch")

	maybeBumpReadTimestampToWriteTimestamp(ctx, ba, g)

	if r.canAttempt1PCEvaluation(ctx, ba, g) {
		__antithesis_instrumentation__.Notify(121029)
		res := r.evaluate1PC(ctx, idKey, ba, g)
		switch res.success {
		case onePCSucceeded:
			__antithesis_instrumentation__.Notify(121030)
			return res.batch, res.stats, res.br, res.res, nil
		case onePCFailed:
			__antithesis_instrumentation__.Notify(121031)
			if res.pErr == nil {
				__antithesis_instrumentation__.Notify(121035)
				log.Fatalf(ctx, "1PC failed but no err. ba: %s", ba.String())
			} else {
				__antithesis_instrumentation__.Notify(121036)
			}
			__antithesis_instrumentation__.Notify(121032)
			return nil, enginepb.MVCCStats{}, nil, result.Result{}, res.pErr
		case onePCFallbackToTransactionalEvaluation:
			__antithesis_instrumentation__.Notify(121033)
		default:
			__antithesis_instrumentation__.Notify(121034)

		}
	} else {
		__antithesis_instrumentation__.Notify(121037)

		arg, ok := ba.GetArg(roachpb.EndTxn)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(121038)
			return arg.(*roachpb.EndTxnRequest).Require1PC == true
		}() == true {
			__antithesis_instrumentation__.Notify(121039)
			return nil, enginepb.MVCCStats{}, nil, result.Result{}, roachpb.NewError(kv.OnePCNotAllowedError{})
		} else {
			__antithesis_instrumentation__.Notify(121040)
		}
	}
	__antithesis_instrumentation__.Notify(121027)

	if ba.Require1PC() {
		__antithesis_instrumentation__.Notify(121041)
		log.Fatalf(ctx,
			"Require1PC should not have gotten to transactional evaluation. ba: %s", ba.String())
	} else {
		__antithesis_instrumentation__.Notify(121042)
	}
	__antithesis_instrumentation__.Notify(121028)

	ms := new(enginepb.MVCCStats)
	rec := NewReplicaEvalContext(r, g.LatchSpans())
	batch, br, res, pErr := r.evaluateWriteBatchWithServersideRefreshes(
		ctx, idKey, rec, ms, ba, ui, g, nil)
	return batch, *ms, br, res, pErr
}

type onePCSuccess int

const (
	onePCSucceeded onePCSuccess = iota

	onePCFailed

	onePCFallbackToTransactionalEvaluation
)

type onePCResult struct {
	success onePCSuccess

	pErr *roachpb.Error

	stats enginepb.MVCCStats
	br    *roachpb.BatchResponse
	res   result.Result
	batch storage.Batch
}

func (r *Replica) evaluate1PC(
	ctx context.Context, idKey kvserverbase.CmdIDKey, ba *roachpb.BatchRequest, g *concurrency.Guard,
) (onePCRes onePCResult) {
	__antithesis_instrumentation__.Notify(121043)
	log.VEventf(ctx, 2, "attempting 1PC execution")

	var batch storage.Batch
	defer func() {
		__antithesis_instrumentation__.Notify(121049)

		if onePCRes.success != onePCSucceeded {
			__antithesis_instrumentation__.Notify(121050)
			batch.Close()
		} else {
			__antithesis_instrumentation__.Notify(121051)
		}
	}()
	__antithesis_instrumentation__.Notify(121044)

	strippedBa := *ba
	strippedBa.Txn = nil
	strippedBa.Requests = ba.Requests[:len(ba.Requests)-1]

	ui := uncertainty.Interval{}

	rec := NewReplicaEvalContext(r, g.LatchSpans())
	var br *roachpb.BatchResponse
	var res result.Result
	var pErr *roachpb.Error

	arg, _ := ba.GetArg(roachpb.EndTxn)
	etArg := arg.(*roachpb.EndTxnRequest)

	ms := new(enginepb.MVCCStats)
	if ba.CanForwardReadTimestamp {
		__antithesis_instrumentation__.Notify(121052)
		batch, br, res, pErr = r.evaluateWriteBatchWithServersideRefreshes(
			ctx, idKey, rec, ms, &strippedBa, ui, g, etArg.Deadline)
	} else {
		__antithesis_instrumentation__.Notify(121053)
		batch, br, res, pErr = r.evaluateWriteBatchWrapper(
			ctx, idKey, rec, ms, &strippedBa, ui, g)
	}
	__antithesis_instrumentation__.Notify(121045)

	if pErr != nil || func() bool {
		__antithesis_instrumentation__.Notify(121054)
		return (!ba.CanForwardReadTimestamp && func() bool {
			__antithesis_instrumentation__.Notify(121055)
			return ba.Timestamp != br.Timestamp == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(121056)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(121059)
			log.VEventf(ctx, 2,
				"1PC execution failed, falling back to transactional execution. pErr: %v", pErr.String())
		} else {
			__antithesis_instrumentation__.Notify(121060)
			log.VEventf(ctx, 2,
				"1PC execution failed, falling back to transactional execution; the batch was pushed")
		}
		__antithesis_instrumentation__.Notify(121057)
		if etArg.Require1PC {
			__antithesis_instrumentation__.Notify(121061)
			return onePCResult{success: onePCFailed, pErr: pErr}
		} else {
			__antithesis_instrumentation__.Notify(121062)
		}
		__antithesis_instrumentation__.Notify(121058)
		return onePCResult{success: onePCFallbackToTransactionalEvaluation}
	} else {
		__antithesis_instrumentation__.Notify(121063)
	}
	__antithesis_instrumentation__.Notify(121046)

	clonedTxn := ba.Txn.Clone()
	clonedTxn.Status = roachpb.COMMITTED

	clonedTxn.ReadTimestamp = br.Timestamp
	clonedTxn.WriteTimestamp = br.Timestamp

	if !etArg.Commit {
		__antithesis_instrumentation__.Notify(121064)
		clonedTxn.Status = roachpb.ABORTED
		batch.Close()
		batch = r.store.Engine().NewBatch()
		ms = new(enginepb.MVCCStats)
	} else {
		__antithesis_instrumentation__.Notify(121065)

		innerResult, err := batcheval.RunCommitTrigger(ctx, rec, batch, ms, etArg, clonedTxn)
		if err != nil {
			__antithesis_instrumentation__.Notify(121067)
			return onePCResult{
				success: onePCFailed,
				pErr:    roachpb.NewError(errors.Wrap(err, "failed to run commit trigger")),
			}
		} else {
			__antithesis_instrumentation__.Notify(121068)
		}
		__antithesis_instrumentation__.Notify(121066)
		if err := res.MergeAndDestroy(innerResult); err != nil {
			__antithesis_instrumentation__.Notify(121069)
			return onePCResult{
				success: onePCFailed,
				pErr:    roachpb.NewError(err),
			}
		} else {
			__antithesis_instrumentation__.Notify(121070)
		}
	}
	__antithesis_instrumentation__.Notify(121047)

	res.Local.UpdatedTxns = []*roachpb.Transaction{clonedTxn}
	res.Local.ResolvedLocks = make([]roachpb.LockUpdate, len(etArg.LockSpans))
	for i, sp := range etArg.LockSpans {
		__antithesis_instrumentation__.Notify(121071)
		res.Local.ResolvedLocks[i] = roachpb.LockUpdate{
			Span:           sp,
			Txn:            clonedTxn.TxnMeta,
			Status:         clonedTxn.Status,
			IgnoredSeqNums: clonedTxn.IgnoredSeqNums,
		}
	}
	__antithesis_instrumentation__.Notify(121048)

	br.Add(&roachpb.EndTxnResponse{OnePhaseCommit: true})
	br.Txn = clonedTxn
	return onePCResult{
		success: onePCSucceeded,
		stats:   *ms,
		br:      br,
		res:     res,
		batch:   batch,
	}
}

func (r *Replica) evaluateWriteBatchWithServersideRefreshes(
	ctx context.Context,
	idKey kvserverbase.CmdIDKey,
	rec batcheval.EvalContext,
	ms *enginepb.MVCCStats,
	ba *roachpb.BatchRequest,
	ui uncertainty.Interval,
	g *concurrency.Guard,
	deadline *hlc.Timestamp,
) (batch storage.Batch, br *roachpb.BatchResponse, res result.Result, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(121072)
	goldenMS := *ms
	for retries := 0; ; retries++ {
		__antithesis_instrumentation__.Notify(121074)
		if retries > 0 {
			__antithesis_instrumentation__.Notify(121078)
			log.VEventf(ctx, 2, "server-side retry of batch")
		} else {
			__antithesis_instrumentation__.Notify(121079)
		}
		__antithesis_instrumentation__.Notify(121075)
		if batch != nil {
			__antithesis_instrumentation__.Notify(121080)

			*ms = goldenMS
			batch.Close()
		} else {
			__antithesis_instrumentation__.Notify(121081)
		}
		__antithesis_instrumentation__.Notify(121076)

		batch, br, res, pErr = r.evaluateWriteBatchWrapper(ctx, idKey, rec, ms, ba, ui, g)

		var success bool
		if pErr == nil {
			__antithesis_instrumentation__.Notify(121082)
			wto := br.Txn != nil && func() bool {
				__antithesis_instrumentation__.Notify(121083)
				return br.Txn.WriteTooOld == true
			}() == true
			success = !wto
		} else {
			__antithesis_instrumentation__.Notify(121084)
			success = false
		}
		__antithesis_instrumentation__.Notify(121077)

		if success || func() bool {
			__antithesis_instrumentation__.Notify(121085)
			return retries > 0 == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(121086)
			return !canDoServersideRetry(ctx, pErr, ba, br, g, deadline) == true
		}() == true {
			__antithesis_instrumentation__.Notify(121087)
			break
		} else {
			__antithesis_instrumentation__.Notify(121088)
		}
	}
	__antithesis_instrumentation__.Notify(121073)
	return batch, br, res, pErr
}

func (r *Replica) evaluateWriteBatchWrapper(
	ctx context.Context,
	idKey kvserverbase.CmdIDKey,
	rec batcheval.EvalContext,
	ms *enginepb.MVCCStats,
	ba *roachpb.BatchRequest,
	ui uncertainty.Interval,
	g *concurrency.Guard,
) (storage.Batch, *roachpb.BatchResponse, result.Result, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(121089)
	batch, opLogger := r.newBatchedEngine(ba, g)
	br, res, pErr := evaluateBatch(ctx, idKey, batch, rec, ms, ba, ui, false)
	if pErr == nil {
		__antithesis_instrumentation__.Notify(121091)
		if opLogger != nil {
			__antithesis_instrumentation__.Notify(121092)
			res.LogicalOpLog = &kvserverpb.LogicalOpLog{
				Ops: opLogger.LogicalOps(),
			}
		} else {
			__antithesis_instrumentation__.Notify(121093)
		}
	} else {
		__antithesis_instrumentation__.Notify(121094)
	}
	__antithesis_instrumentation__.Notify(121090)
	return batch, br, res, pErr
}

func (r *Replica) newBatchedEngine(
	ba *roachpb.BatchRequest, g *concurrency.Guard,
) (storage.Batch, *storage.OpLoggerBatch) {
	__antithesis_instrumentation__.Notify(121095)
	batch := r.store.Engine().NewBatch()
	if !batch.ConsistentIterators() {
		__antithesis_instrumentation__.Notify(121099)

		panic("expected consistent iterators")
	} else {
		__antithesis_instrumentation__.Notify(121100)
	}
	__antithesis_instrumentation__.Notify(121096)
	var opLogger *storage.OpLoggerBatch
	if r.isRangefeedEnabled() || func() bool {
		__antithesis_instrumentation__.Notify(121101)
		return RangefeedEnabled.Get(&r.store.cfg.Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(121102)

		opLogger = storage.NewOpLoggerBatch(batch)
		batch = opLogger
	} else {
		__antithesis_instrumentation__.Notify(121103)
	}
	__antithesis_instrumentation__.Notify(121097)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(121104)

		batch = spanset.NewBatch(batch, g.LatchSpans())
	} else {
		__antithesis_instrumentation__.Notify(121105)
	}
	__antithesis_instrumentation__.Notify(121098)
	return batch, opLogger
}

func isOnePhaseCommit(ba *roachpb.BatchRequest) bool {
	__antithesis_instrumentation__.Notify(121106)
	if ba.Txn == nil {
		__antithesis_instrumentation__.Notify(121110)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121111)
	}
	__antithesis_instrumentation__.Notify(121107)
	if !ba.IsCompleteTransaction() {
		__antithesis_instrumentation__.Notify(121112)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121113)
	}
	__antithesis_instrumentation__.Notify(121108)
	arg, _ := ba.GetArg(roachpb.EndTxn)
	etArg := arg.(*roachpb.EndTxnRequest)
	if retry, _, _ := batcheval.IsEndTxnTriggeringRetryError(ba.Txn, etArg); retry {
		__antithesis_instrumentation__.Notify(121114)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121115)
	}
	__antithesis_instrumentation__.Notify(121109)

	return ba.Txn.Epoch == 0 || func() bool {
		__antithesis_instrumentation__.Notify(121116)
		return etArg.Require1PC == true
	}() == true
}
