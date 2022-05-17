package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/abortspan"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadWriteCommand(roachpb.EndTxn, declareKeysEndTxn, EndTxn)
}

func declareKeysWriteTransaction(
	_ ImmutableRangeState, header *roachpb.Header, req roachpb.Request, latchSpans *spanset.SpanSet,
) {
	__antithesis_instrumentation__.Notify(96512)
	if header.Txn != nil {
		__antithesis_instrumentation__.Notify(96513)
		header.Txn.AssertInitialized(context.TODO())
		latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
			Key: keys.TransactionKey(req.Header().Key, header.Txn.ID),
		})
	} else {
		__antithesis_instrumentation__.Notify(96514)
	}
}

func declareKeysEndTxn(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(96515)
	et := req.(*roachpb.EndTxnRequest)
	declareKeysWriteTransaction(rs, header, req, latchSpans)
	var minTxnTS hlc.Timestamp
	if header.Txn != nil {
		__antithesis_instrumentation__.Notify(96517)
		header.Txn.AssertInitialized(context.TODO())
		minTxnTS = header.Txn.MinTimestamp
		abortSpanAccess := spanset.SpanReadOnly
		if !et.Commit {
			__antithesis_instrumentation__.Notify(96519)

			abortSpanAccess = spanset.SpanReadWrite
		} else {
			__antithesis_instrumentation__.Notify(96520)
		}
		__antithesis_instrumentation__.Notify(96518)
		latchSpans.AddNonMVCC(abortSpanAccess, roachpb.Span{
			Key: keys.AbortSpanKey(rs.GetRangeID(), header.Txn.ID),
		})
	} else {
		__antithesis_instrumentation__.Notify(96521)
	}
	__antithesis_instrumentation__.Notify(96516)

	if !et.IsParallelCommit() {
		__antithesis_instrumentation__.Notify(96522)

		latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{
			Key: keys.RangeDescriptorKey(rs.GetStartKey()),
		})

		for _, span := range et.LockSpans {
			__antithesis_instrumentation__.Notify(96524)
			latchSpans.AddMVCC(spanset.SpanReadWrite, span, minTxnTS)
		}
		__antithesis_instrumentation__.Notify(96523)

		if et.InternalCommitTrigger != nil {
			__antithesis_instrumentation__.Notify(96525)
			if st := et.InternalCommitTrigger.SplitTrigger; st != nil {
				__antithesis_instrumentation__.Notify(96527)

				latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{
					Key:    st.LeftDesc.StartKey.AsRawKey(),
					EndKey: st.LeftDesc.EndKey.AsRawKey(),
				})
				latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
					Key:    st.RightDesc.StartKey.AsRawKey(),
					EndKey: st.RightDesc.EndKey.AsRawKey(),
				})
				latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
					Key:    keys.MakeRangeKeyPrefix(st.LeftDesc.StartKey),
					EndKey: keys.MakeRangeKeyPrefix(st.RightDesc.EndKey).PrefixEnd(),
				})

				leftRangeIDPrefix := keys.MakeRangeIDReplicatedPrefix(rs.GetRangeID())
				latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{
					Key:    leftRangeIDPrefix,
					EndKey: leftRangeIDPrefix.PrefixEnd(),
				})
				rightRangeIDPrefix := keys.MakeRangeIDReplicatedPrefix(st.RightDesc.RangeID)
				latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
					Key:    rightRangeIDPrefix,
					EndKey: rightRangeIDPrefix.PrefixEnd(),
				})

				rightRangeIDUnreplicatedPrefix := keys.MakeRangeIDUnreplicatedPrefix(st.RightDesc.RangeID)
				latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
					Key:    rightRangeIDUnreplicatedPrefix,
					EndKey: rightRangeIDUnreplicatedPrefix.PrefixEnd(),
				})

				latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{
					Key: keys.RangeLastReplicaGCTimestampKey(st.LeftDesc.RangeID),
				})
				latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
					Key: keys.RangeLastReplicaGCTimestampKey(st.RightDesc.RangeID),
				})

				latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{
					Key:    abortspan.MinKey(rs.GetRangeID()),
					EndKey: abortspan.MaxKey(rs.GetRangeID()),
				})
			} else {
				__antithesis_instrumentation__.Notify(96528)
			}
			__antithesis_instrumentation__.Notify(96526)
			if mt := et.InternalCommitTrigger.MergeTrigger; mt != nil {
				__antithesis_instrumentation__.Notify(96529)

				latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
					Key:    abortspan.MinKey(mt.LeftDesc.RangeID),
					EndKey: abortspan.MaxKey(mt.LeftDesc.RangeID).PrefixEnd(),
				})
				latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{
					Key:    keys.MakeRangeIDReplicatedPrefix(mt.RightDesc.RangeID),
					EndKey: keys.MakeRangeIDReplicatedPrefix(mt.RightDesc.RangeID).PrefixEnd(),
				})

				latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{
					Key: keys.RangePriorReadSummaryKey(mt.LeftDesc.RangeID),
				})
			} else {
				__antithesis_instrumentation__.Notify(96530)
			}
		} else {
			__antithesis_instrumentation__.Notify(96531)
		}
	} else {
		__antithesis_instrumentation__.Notify(96532)
	}
}

func EndTxn(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96533)
	args := cArgs.Args.(*roachpb.EndTxnRequest)
	h := cArgs.Header
	ms := cArgs.Stats
	reply := resp.(*roachpb.EndTxnResponse)

	if err := VerifyTransaction(h, args, roachpb.PENDING, roachpb.STAGING, roachpb.ABORTED); err != nil {
		__antithesis_instrumentation__.Notify(96542)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96543)
	}
	__antithesis_instrumentation__.Notify(96534)
	if args.Require1PC {
		__antithesis_instrumentation__.Notify(96544)

		return result.Result{}, errors.AssertionFailedf("unexpectedly trying to evaluate EndTxn with the Require1PC flag set")
	} else {
		__antithesis_instrumentation__.Notify(96545)
	}
	__antithesis_instrumentation__.Notify(96535)
	if args.Commit && func() bool {
		__antithesis_instrumentation__.Notify(96546)
		return args.Poison == true
	}() == true {
		__antithesis_instrumentation__.Notify(96547)
		return result.Result{}, errors.AssertionFailedf("cannot poison during a committing EndTxn request")
	} else {
		__antithesis_instrumentation__.Notify(96548)
	}
	__antithesis_instrumentation__.Notify(96536)

	key := keys.TransactionKey(h.Txn.Key, h.Txn.ID)

	var existingTxn roachpb.Transaction
	recordAlreadyExisted, err := storage.MVCCGetProto(
		ctx, readWriter, key, hlc.Timestamp{}, &existingTxn, storage.MVCCGetOptions{},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(96549)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96550)
		if !recordAlreadyExisted {
			__antithesis_instrumentation__.Notify(96551)

			reply.Txn = h.Txn.Clone()

			if args.Commit {
				__antithesis_instrumentation__.Notify(96552)
				if err := CanCreateTxnRecord(ctx, cArgs.EvalCtx, reply.Txn); err != nil {
					__antithesis_instrumentation__.Notify(96553)
					return result.Result{}, err
				} else {
					__antithesis_instrumentation__.Notify(96554)
				}
			} else {
				__antithesis_instrumentation__.Notify(96555)
			}
		} else {
			__antithesis_instrumentation__.Notify(96556)

			reply.Txn = &existingTxn

			switch reply.Txn.Status {
			case roachpb.COMMITTED:
				__antithesis_instrumentation__.Notify(96558)

				log.VEventf(ctx, 2, "transaction found to be already committed")
				return result.Result{}, roachpb.NewTransactionStatusError(
					roachpb.TransactionStatusError_REASON_TXN_COMMITTED,
					"already committed")

			case roachpb.ABORTED:
				__antithesis_instrumentation__.Notify(96559)
				if !args.Commit {
					__antithesis_instrumentation__.Notify(96565)

					desc := cArgs.EvalCtx.Desc()
					resolvedLocks, externalLocks, err := resolveLocalLocks(ctx, desc, readWriter, ms, args, reply.Txn, cArgs.EvalCtx)
					if err != nil {
						__antithesis_instrumentation__.Notify(96568)
						return result.Result{}, err
					} else {
						__antithesis_instrumentation__.Notify(96569)
					}
					__antithesis_instrumentation__.Notify(96566)
					if err := updateFinalizedTxn(
						ctx, readWriter, ms, key, args, reply.Txn, recordAlreadyExisted, externalLocks,
					); err != nil {
						__antithesis_instrumentation__.Notify(96570)
						return result.Result{}, err
					} else {
						__antithesis_instrumentation__.Notify(96571)
					}
					__antithesis_instrumentation__.Notify(96567)

					res := result.FromEndTxn(reply.Txn, true, args.Poison)
					res.Local.ResolvedLocks = resolvedLocks
					return res, nil
				} else {
					__antithesis_instrumentation__.Notify(96572)
				}
				__antithesis_instrumentation__.Notify(96560)

				reply.Txn.LockSpans = args.LockSpans
				return result.FromEndTxn(reply.Txn, true, args.Poison),
					roachpb.NewTransactionAbortedError(roachpb.ABORT_REASON_ABORTED_RECORD_FOUND)

			case roachpb.PENDING:
				__antithesis_instrumentation__.Notify(96561)
				if h.Txn.Epoch < reply.Txn.Epoch {
					__antithesis_instrumentation__.Notify(96573)
					return result.Result{}, errors.AssertionFailedf(
						"programming error: epoch regression: %d", h.Txn.Epoch)
				} else {
					__antithesis_instrumentation__.Notify(96574)
				}

			case roachpb.STAGING:
				__antithesis_instrumentation__.Notify(96562)
				if h.Txn.Epoch < reply.Txn.Epoch {
					__antithesis_instrumentation__.Notify(96575)
					return result.Result{}, errors.AssertionFailedf(
						"programming error: epoch regression: %d", h.Txn.Epoch)
				} else {
					__antithesis_instrumentation__.Notify(96576)
				}
				__antithesis_instrumentation__.Notify(96563)
				if h.Txn.Epoch > reply.Txn.Epoch {
					__antithesis_instrumentation__.Notify(96577)

					reply.Txn.Status = roachpb.PENDING
				} else {
					__antithesis_instrumentation__.Notify(96578)
				}

			default:
				__antithesis_instrumentation__.Notify(96564)
				return result.Result{}, errors.AssertionFailedf("bad txn status: %s", reply.Txn)
			}
			__antithesis_instrumentation__.Notify(96557)

			reply.Txn.Update(h.Txn)
		}
	}
	__antithesis_instrumentation__.Notify(96537)

	if args.Commit {
		__antithesis_instrumentation__.Notify(96579)
		if retry, reason, extraMsg := IsEndTxnTriggeringRetryError(reply.Txn, args); retry {
			__antithesis_instrumentation__.Notify(96582)
			return result.Result{}, roachpb.NewTransactionRetryError(reason, extraMsg)
		} else {
			__antithesis_instrumentation__.Notify(96583)
		}
		__antithesis_instrumentation__.Notify(96580)

		if args.IsParallelCommit() {
			__antithesis_instrumentation__.Notify(96584)

			if ct := args.InternalCommitTrigger; ct != nil {
				__antithesis_instrumentation__.Notify(96587)
				err := errors.Errorf("cannot stage transaction with a commit trigger: %+v", ct)
				return result.Result{}, err
			} else {
				__antithesis_instrumentation__.Notify(96588)
			}
			__antithesis_instrumentation__.Notify(96585)

			reply.Txn.Status = roachpb.STAGING
			reply.StagingTimestamp = reply.Txn.WriteTimestamp
			if err := updateStagingTxn(ctx, readWriter, ms, key, args, reply.Txn); err != nil {
				__antithesis_instrumentation__.Notify(96589)
				return result.Result{}, err
			} else {
				__antithesis_instrumentation__.Notify(96590)
			}
			__antithesis_instrumentation__.Notify(96586)
			return result.Result{}, nil
		} else {
			__antithesis_instrumentation__.Notify(96591)
		}
		__antithesis_instrumentation__.Notify(96581)

		reply.Txn.Status = roachpb.COMMITTED
	} else {
		__antithesis_instrumentation__.Notify(96592)

		if reply.Txn.Status == roachpb.STAGING {
			__antithesis_instrumentation__.Notify(96594)
			err := roachpb.NewIndeterminateCommitError(*reply.Txn)
			log.VEventf(ctx, 1, "%v", err)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96595)
		}
		__antithesis_instrumentation__.Notify(96593)

		reply.Txn.Status = roachpb.ABORTED
	}
	__antithesis_instrumentation__.Notify(96538)

	desc := cArgs.EvalCtx.Desc()
	resolvedLocks, externalLocks, err := resolveLocalLocks(ctx, desc, readWriter, ms, args, reply.Txn, cArgs.EvalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(96596)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96597)
	}
	__antithesis_instrumentation__.Notify(96539)
	if err := updateFinalizedTxn(
		ctx, readWriter, ms, key, args, reply.Txn, recordAlreadyExisted, externalLocks,
	); err != nil {
		__antithesis_instrumentation__.Notify(96598)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96599)
	}
	__antithesis_instrumentation__.Notify(96540)

	txnResult := result.FromEndTxn(reply.Txn, false, args.Poison)
	txnResult.Local.UpdatedTxns = []*roachpb.Transaction{reply.Txn}
	txnResult.Local.ResolvedLocks = resolvedLocks

	if reply.Txn.Status == roachpb.COMMITTED {
		__antithesis_instrumentation__.Notify(96600)
		triggerResult, err := RunCommitTrigger(
			ctx, cArgs.EvalCtx, readWriter.(storage.Batch), ms, args, reply.Txn,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(96602)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96603)
		}
		__antithesis_instrumentation__.Notify(96601)
		if err := txnResult.MergeAndDestroy(triggerResult); err != nil {
			__antithesis_instrumentation__.Notify(96604)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96605)
		}
	} else {
		__antithesis_instrumentation__.Notify(96606)
		if reply.Txn.Status == roachpb.ABORTED {
			__antithesis_instrumentation__.Notify(96607)

			if cArgs.EvalCtx.ContainsKey(keys.SystemConfigSpan.Key) && func() bool {
				__antithesis_instrumentation__.Notify(96608)
				return !cArgs.EvalCtx.ClusterSettings().Version.IsActive(
					ctx, clusterversion.DisableSystemConfigGossipTrigger,
				) == true
			}() == true {
				__antithesis_instrumentation__.Notify(96609)
				txnResult.Local.MaybeGossipSystemConfigIfHaveFailure = true
			} else {
				__antithesis_instrumentation__.Notify(96610)
			}
		} else {
			__antithesis_instrumentation__.Notify(96611)
		}
	}
	__antithesis_instrumentation__.Notify(96541)

	return txnResult, nil
}

func IsEndTxnExceedingDeadline(commitTS hlc.Timestamp, deadline *hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(96612)
	return deadline != nil && func() bool {
		__antithesis_instrumentation__.Notify(96613)
		return !deadline.IsEmpty() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(96614)
		return deadline.LessEq(commitTS) == true
	}() == true
}

func IsEndTxnTriggeringRetryError(
	txn *roachpb.Transaction, args *roachpb.EndTxnRequest,
) (retry bool, reason roachpb.TransactionRetryReason, extraMsg string) {
	__antithesis_instrumentation__.Notify(96615)

	if txn.WriteTooOld {
		__antithesis_instrumentation__.Notify(96618)
		retry, reason = true, roachpb.RETRY_WRITE_TOO_OLD
	} else {
		__antithesis_instrumentation__.Notify(96619)
		readTimestamp := txn.ReadTimestamp
		isTxnPushed := txn.WriteTimestamp != readTimestamp

		if isTxnPushed {
			__antithesis_instrumentation__.Notify(96620)
			retry, reason = true, roachpb.RETRY_SERIALIZABLE
		} else {
			__antithesis_instrumentation__.Notify(96621)
		}
	}
	__antithesis_instrumentation__.Notify(96616)

	if !retry && func() bool {
		__antithesis_instrumentation__.Notify(96622)
		return IsEndTxnExceedingDeadline(txn.WriteTimestamp, args.Deadline) == true
	}() == true {
		__antithesis_instrumentation__.Notify(96623)
		exceededBy := txn.WriteTimestamp.GoTime().Sub(args.Deadline.GoTime())
		extraMsg = fmt.Sprintf(
			"txn timestamp pushed too much; deadline exceeded by %s (%s > %s)",
			exceededBy, txn.WriteTimestamp, args.Deadline)
		retry, reason = true, roachpb.RETRY_COMMIT_DEADLINE_EXCEEDED
	} else {
		__antithesis_instrumentation__.Notify(96624)
	}
	__antithesis_instrumentation__.Notify(96617)
	return retry, reason, extraMsg
}

const lockResolutionBatchSize = 500

func resolveLocalLocks(
	ctx context.Context,
	desc *roachpb.RangeDescriptor,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	args *roachpb.EndTxnRequest,
	txn *roachpb.Transaction,
	evalCtx EvalContext,
) (resolvedLocks []roachpb.LockUpdate, externalLocks []roachpb.Span, _ error) {
	__antithesis_instrumentation__.Notify(96625)
	if mergeTrigger := args.InternalCommitTrigger.GetMergeTrigger(); mergeTrigger != nil {
		__antithesis_instrumentation__.Notify(96630)

		desc = &mergeTrigger.LeftDesc
	} else {
		__antithesis_instrumentation__.Notify(96631)
	}
	__antithesis_instrumentation__.Notify(96626)

	var resolveAllowance int64 = lockResolutionBatchSize
	if args.InternalCommitTrigger != nil {
		__antithesis_instrumentation__.Notify(96632)

		resolveAllowance = math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(96633)
	}
	__antithesis_instrumentation__.Notify(96627)

	for _, span := range args.LockSpans {
		__antithesis_instrumentation__.Notify(96634)
		if err := func() error {
			__antithesis_instrumentation__.Notify(96635)
			if resolveAllowance == 0 {
				__antithesis_instrumentation__.Notify(96639)
				externalLocks = append(externalLocks, span)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(96640)
			}
			__antithesis_instrumentation__.Notify(96636)
			update := roachpb.MakeLockUpdate(txn, span)
			if len(span.EndKey) == 0 {
				__antithesis_instrumentation__.Notify(96641)

				if !kvserverbase.ContainsKey(desc, span.Key) {
					__antithesis_instrumentation__.Notify(96645)
					externalLocks = append(externalLocks, span)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(96646)
				}
				__antithesis_instrumentation__.Notify(96642)

				ok, err := storage.MVCCResolveWriteIntent(ctx, readWriter, ms, update)
				if err != nil {
					__antithesis_instrumentation__.Notify(96647)
					return err
				} else {
					__antithesis_instrumentation__.Notify(96648)
				}
				__antithesis_instrumentation__.Notify(96643)
				if ok {
					__antithesis_instrumentation__.Notify(96649)
					resolveAllowance--
				} else {
					__antithesis_instrumentation__.Notify(96650)
				}
				__antithesis_instrumentation__.Notify(96644)
				resolvedLocks = append(resolvedLocks, update)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(96651)
			}
			__antithesis_instrumentation__.Notify(96637)

			inSpan, outSpans := kvserverbase.IntersectSpan(span, desc)
			externalLocks = append(externalLocks, outSpans...)
			if inSpan != nil {
				__antithesis_instrumentation__.Notify(96652)
				update.Span = *inSpan
				num, resumeSpan, err := storage.MVCCResolveWriteIntentRange(
					ctx, readWriter, ms, update, resolveAllowance)
				if err != nil {
					__antithesis_instrumentation__.Notify(96656)
					return err
				} else {
					__antithesis_instrumentation__.Notify(96657)
				}
				__antithesis_instrumentation__.Notify(96653)
				if evalCtx.EvalKnobs().NumKeysEvaluatedForRangeIntentResolution != nil {
					__antithesis_instrumentation__.Notify(96658)
					atomic.AddInt64(evalCtx.EvalKnobs().NumKeysEvaluatedForRangeIntentResolution, num)
				} else {
					__antithesis_instrumentation__.Notify(96659)
				}
				__antithesis_instrumentation__.Notify(96654)
				resolveAllowance -= num
				if resumeSpan != nil {
					__antithesis_instrumentation__.Notify(96660)
					if resolveAllowance != 0 {
						__antithesis_instrumentation__.Notify(96662)
						log.Fatalf(ctx, "expected resolve allowance to be exactly 0 resolving %s; got %d", update.Span, resolveAllowance)
					} else {
						__antithesis_instrumentation__.Notify(96663)
					}
					__antithesis_instrumentation__.Notify(96661)
					update.EndKey = resumeSpan.Key
					externalLocks = append(externalLocks, *resumeSpan)
				} else {
					__antithesis_instrumentation__.Notify(96664)
				}
				__antithesis_instrumentation__.Notify(96655)
				resolvedLocks = append(resolvedLocks, update)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(96665)
			}
			__antithesis_instrumentation__.Notify(96638)
			return nil
		}(); err != nil {
			__antithesis_instrumentation__.Notify(96666)
			return nil, nil, errors.Wrapf(err, "resolving lock at %s on end transaction [%s]", span, txn.Status)
		} else {
			__antithesis_instrumentation__.Notify(96667)
		}
	}
	__antithesis_instrumentation__.Notify(96628)

	removedAny := resolveAllowance != lockResolutionBatchSize
	if WriteAbortSpanOnResolve(txn.Status, args.Poison, removedAny) {
		__antithesis_instrumentation__.Notify(96668)
		if err := UpdateAbortSpan(ctx, evalCtx, readWriter, ms, txn.TxnMeta, args.Poison); err != nil {
			__antithesis_instrumentation__.Notify(96669)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(96670)
		}
	} else {
		__antithesis_instrumentation__.Notify(96671)
	}
	__antithesis_instrumentation__.Notify(96629)
	return resolvedLocks, externalLocks, nil
}

func updateStagingTxn(
	ctx context.Context,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	key []byte,
	args *roachpb.EndTxnRequest,
	txn *roachpb.Transaction,
) error {
	__antithesis_instrumentation__.Notify(96672)
	txn.LockSpans = args.LockSpans
	txn.InFlightWrites = args.InFlightWrites
	txnRecord := txn.AsRecord()
	return storage.MVCCPutProto(ctx, readWriter, ms, key, hlc.Timestamp{}, nil, &txnRecord)
}

func updateFinalizedTxn(
	ctx context.Context,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	key []byte,
	args *roachpb.EndTxnRequest,
	txn *roachpb.Transaction,
	recordAlreadyExisted bool,
	externalLocks []roachpb.Span,
) error {
	__antithesis_instrumentation__.Notify(96673)
	if txnAutoGC && func() bool {
		__antithesis_instrumentation__.Notify(96675)
		return len(externalLocks) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(96676)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(96679)
			log.Infof(ctx, "auto-gc'ed %s (%d locks)", txn.Short(), len(args.LockSpans))
		} else {
			__antithesis_instrumentation__.Notify(96680)
		}
		__antithesis_instrumentation__.Notify(96677)
		if !recordAlreadyExisted {
			__antithesis_instrumentation__.Notify(96681)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(96682)
		}
		__antithesis_instrumentation__.Notify(96678)
		return storage.MVCCDelete(ctx, readWriter, ms, key, hlc.Timestamp{}, nil)
	} else {
		__antithesis_instrumentation__.Notify(96683)
	}
	__antithesis_instrumentation__.Notify(96674)
	txn.LockSpans = externalLocks
	txn.InFlightWrites = nil
	txnRecord := txn.AsRecord()
	return storage.MVCCPutProto(ctx, readWriter, ms, key, hlc.Timestamp{}, nil, &txnRecord)
}

func RunCommitTrigger(
	ctx context.Context,
	rec EvalContext,
	batch storage.Batch,
	ms *enginepb.MVCCStats,
	args *roachpb.EndTxnRequest,
	txn *roachpb.Transaction,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96684)
	ct := args.InternalCommitTrigger
	if ct == nil {
		__antithesis_instrumentation__.Notify(96692)
		return result.Result{}, nil
	} else {
		__antithesis_instrumentation__.Notify(96693)
	}
	__antithesis_instrumentation__.Notify(96685)

	if txn.WriteTimestamp.Synthetic && func() bool {
		__antithesis_instrumentation__.Notify(96694)
		return rec.Clock().Now().Less(txn.WriteTimestamp) == true
	}() == true {
		__antithesis_instrumentation__.Notify(96695)
		return result.Result{}, errors.AssertionFailedf("txn %s with %s commit trigger needs "+
			"commit wait. Was its timestamp bumped after acquiring latches?", txn, ct.Kind())
	} else {
		__antithesis_instrumentation__.Notify(96696)
	}
	__antithesis_instrumentation__.Notify(96686)

	if ct.GetSplitTrigger() != nil {
		__antithesis_instrumentation__.Notify(96697)
		newMS, res, err := splitTrigger(
			ctx, rec, batch, *ms, ct.SplitTrigger, txn.WriteTimestamp,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(96699)
			return result.Result{}, roachpb.NewReplicaCorruptionError(err)
		} else {
			__antithesis_instrumentation__.Notify(96700)
		}
		__antithesis_instrumentation__.Notify(96698)
		*ms = newMS
		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(96701)
	}
	__antithesis_instrumentation__.Notify(96687)
	if mt := ct.GetMergeTrigger(); mt != nil {
		__antithesis_instrumentation__.Notify(96702)
		res, err := mergeTrigger(ctx, rec, batch, ms, mt, txn.WriteTimestamp)
		if err != nil {
			__antithesis_instrumentation__.Notify(96704)
			return result.Result{}, roachpb.NewReplicaCorruptionError(err)
		} else {
			__antithesis_instrumentation__.Notify(96705)
		}
		__antithesis_instrumentation__.Notify(96703)
		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(96706)
	}
	__antithesis_instrumentation__.Notify(96688)
	if crt := ct.GetChangeReplicasTrigger(); crt != nil {
		__antithesis_instrumentation__.Notify(96707)

		return changeReplicasTrigger(ctx, rec, batch, crt), nil
	} else {
		__antithesis_instrumentation__.Notify(96708)
	}
	__antithesis_instrumentation__.Notify(96689)
	if ct.GetModifiedSpanTrigger() != nil {
		__antithesis_instrumentation__.Notify(96709)
		var pd result.Result
		if ct.ModifiedSpanTrigger.SystemConfigSpan {
			__antithesis_instrumentation__.Notify(96712)

			if rec.ContainsKey(keys.SystemConfigSpan.Key) {
				__antithesis_instrumentation__.Notify(96713)
				if err := pd.MergeAndDestroy(
					result.Result{
						Local: result.LocalResult{
							MaybeGossipSystemConfig: true,
						},
					},
				); err != nil {
					__antithesis_instrumentation__.Notify(96714)
					return result.Result{}, err
				} else {
					__antithesis_instrumentation__.Notify(96715)
				}
			} else {
				__antithesis_instrumentation__.Notify(96716)
				log.Errorf(ctx, "System configuration span was modified, but the "+
					"modification trigger is executing on a non-system range. "+
					"Configuration changes will not be gossiped.")
			}
		} else {
			__antithesis_instrumentation__.Notify(96717)
		}
		__antithesis_instrumentation__.Notify(96710)
		if nlSpan := ct.ModifiedSpanTrigger.NodeLivenessSpan; nlSpan != nil {
			__antithesis_instrumentation__.Notify(96718)
			if err := pd.MergeAndDestroy(
				result.Result{
					Local: result.LocalResult{
						MaybeGossipNodeLiveness: nlSpan,
					},
				},
			); err != nil {
				__antithesis_instrumentation__.Notify(96719)
				return result.Result{}, err
			} else {
				__antithesis_instrumentation__.Notify(96720)
			}
		} else {
			__antithesis_instrumentation__.Notify(96721)
		}
		__antithesis_instrumentation__.Notify(96711)
		return pd, nil
	} else {
		__antithesis_instrumentation__.Notify(96722)
	}
	__antithesis_instrumentation__.Notify(96690)
	if sbt := ct.GetStickyBitTrigger(); sbt != nil {
		__antithesis_instrumentation__.Notify(96723)
		newDesc := *rec.Desc()
		if !sbt.StickyBit.IsEmpty() {
			__antithesis_instrumentation__.Notify(96725)
			newDesc.StickyBit = &sbt.StickyBit
		} else {
			__antithesis_instrumentation__.Notify(96726)
			newDesc.StickyBit = nil
		}
		__antithesis_instrumentation__.Notify(96724)
		var res result.Result
		res.Replicated.State = &kvserverpb.ReplicaState{
			Desc: &newDesc,
		}
		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(96727)
	}
	__antithesis_instrumentation__.Notify(96691)

	log.Fatalf(ctx, "unknown commit trigger: %+v", ct)
	return result.Result{}, nil
}

func splitTrigger(
	ctx context.Context,
	rec EvalContext,
	batch storage.Batch,
	bothDeltaMS enginepb.MVCCStats,
	split *roachpb.SplitTrigger,
	ts hlc.Timestamp,
) (enginepb.MVCCStats, result.Result, error) {
	__antithesis_instrumentation__.Notify(96728)
	desc := rec.Desc()
	if !bytes.Equal(desc.StartKey, split.LeftDesc.StartKey) || func() bool {
		__antithesis_instrumentation__.Notify(96731)
		return !bytes.Equal(desc.EndKey, split.RightDesc.EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(96732)
		return enginepb.MVCCStats{}, result.Result{}, errors.Errorf("range does not match splits: (%s-%s) + (%s-%s) != %s",
			split.LeftDesc.StartKey, split.LeftDesc.EndKey,
			split.RightDesc.StartKey, split.RightDesc.EndKey, desc)
	} else {
		__antithesis_instrumentation__.Notify(96733)
	}
	__antithesis_instrumentation__.Notify(96729)

	emptyRHS, err := isGlobalKeyspaceEmpty(batch, &split.RightDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(96734)
		return enginepb.MVCCStats{}, result.Result{}, errors.Wrapf(err,
			"unable to determine whether right hand side of split is empty")
	} else {
		__antithesis_instrumentation__.Notify(96735)
	}
	__antithesis_instrumentation__.Notify(96730)

	h := splitStatsHelperInput{
		AbsPreSplitBothEstimated: rec.GetMVCCStats(),
		DeltaBatchEstimated:      bothDeltaMS,
		AbsPostSplitLeftFn:       makeScanStatsFn(ctx, batch, ts, &split.LeftDesc, "left hand side"),
		AbsPostSplitRightFn:      makeScanStatsFn(ctx, batch, ts, &split.RightDesc, "right hand side"),
		ScanRightFirst: splitScansRightForStatsFirst || func() bool {
			__antithesis_instrumentation__.Notify(96736)
			return emptyRHS == true
		}() == true,
	}
	return splitTriggerHelper(ctx, rec, batch, h, split, ts)
}

var splitScansRightForStatsFirst = util.ConstantWithMetamorphicTestBool(
	"split-scans-right-for-stats-first", false)

func isGlobalKeyspaceEmpty(reader storage.Reader, d *roachpb.RangeDescriptor) (bool, error) {
	__antithesis_instrumentation__.Notify(96737)
	span := d.KeySpan().AsRawSpanWithNoLocals()
	iter := reader.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{UpperBound: span.EndKey})
	defer iter.Close()
	iter.SeekGE(storage.MakeMVCCMetadataKey(span.Key))
	ok, err := iter.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(96739)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(96740)
	}
	__antithesis_instrumentation__.Notify(96738)
	return !ok, nil
}

func makeScanStatsFn(
	ctx context.Context,
	reader storage.Reader,
	ts hlc.Timestamp,
	sideDesc *roachpb.RangeDescriptor,
	sideName string,
) splitStatsScanFn {
	__antithesis_instrumentation__.Notify(96741)
	return func() (enginepb.MVCCStats, error) {
		__antithesis_instrumentation__.Notify(96742)
		sideMS, err := rditer.ComputeStatsForRange(sideDesc, reader, ts.WallTime)
		if err != nil {
			__antithesis_instrumentation__.Notify(96744)
			return enginepb.MVCCStats{}, errors.Wrapf(err,
				"unable to compute stats for %s range after split", sideName)
		} else {
			__antithesis_instrumentation__.Notify(96745)
		}
		__antithesis_instrumentation__.Notify(96743)
		log.Eventf(ctx, "computed stats for %s range", sideName)
		return sideMS, nil
	}
}

func splitTriggerHelper(
	ctx context.Context,
	rec EvalContext,
	batch storage.Batch,
	statsInput splitStatsHelperInput,
	split *roachpb.SplitTrigger,
	ts hlc.Timestamp,
) (enginepb.MVCCStats, result.Result, error) {
	__antithesis_instrumentation__.Notify(96746)

	replicaGCTS, err := rec.GetLastReplicaGCTimestamp(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(96751)
		return enginepb.MVCCStats{}, result.Result{}, errors.Wrap(err, "unable to fetch last replica GC timestamp")
	} else {
		__antithesis_instrumentation__.Notify(96752)
	}
	__antithesis_instrumentation__.Notify(96747)
	if err := storage.MVCCPutProto(ctx, batch, nil, keys.RangeLastReplicaGCTimestampKey(split.RightDesc.RangeID), hlc.Timestamp{}, nil, &replicaGCTS); err != nil {
		__antithesis_instrumentation__.Notify(96753)
		return enginepb.MVCCStats{}, result.Result{}, errors.Wrap(err, "unable to copy last replica GC timestamp")
	} else {
		__antithesis_instrumentation__.Notify(96754)
	}
	__antithesis_instrumentation__.Notify(96748)

	h, err := makeSplitStatsHelper(statsInput)
	if err != nil {
		__antithesis_instrumentation__.Notify(96755)
		return enginepb.MVCCStats{}, result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96756)
	}
	__antithesis_instrumentation__.Notify(96749)

	if err := rec.AbortSpan().CopyTo(
		ctx, batch, batch, h.AbsPostSplitRight(), ts, split.RightDesc.RangeID,
	); err != nil {
		__antithesis_instrumentation__.Notify(96757)
		return enginepb.MVCCStats{}, result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96758)
	}

	{
		__antithesis_instrumentation__.Notify(96759)

		sl := MakeStateLoader(rec)
		leftLease, err := sl.LoadLease(ctx, batch)
		if err != nil {
			__antithesis_instrumentation__.Notify(96768)
			return enginepb.MVCCStats{}, result.Result{}, errors.Wrap(err, "unable to load lease")
		} else {
			__antithesis_instrumentation__.Notify(96769)
		}
		__antithesis_instrumentation__.Notify(96760)
		if leftLease.Empty() {
			__antithesis_instrumentation__.Notify(96770)
			log.Fatalf(ctx, "LHS of split has no lease")
		} else {
			__antithesis_instrumentation__.Notify(96771)
		}
		__antithesis_instrumentation__.Notify(96761)

		replica, found := split.RightDesc.GetReplicaDescriptor(leftLease.Replica.StoreID)
		if !found {
			__antithesis_instrumentation__.Notify(96772)
			return enginepb.MVCCStats{}, result.Result{}, errors.Errorf(
				"pre-split lease holder %+v not found in post-split descriptor %+v",
				leftLease.Replica, split.RightDesc,
			)
		} else {
			__antithesis_instrumentation__.Notify(96773)
		}
		__antithesis_instrumentation__.Notify(96762)
		rightLease := leftLease
		rightLease.Replica = replica
		gcThreshold, err := sl.LoadGCThreshold(ctx, batch)
		if err != nil {
			__antithesis_instrumentation__.Notify(96774)
			return enginepb.MVCCStats{}, result.Result{}, errors.Wrap(err, "unable to load GCThreshold")
		} else {
			__antithesis_instrumentation__.Notify(96775)
		}
		__antithesis_instrumentation__.Notify(96763)
		if gcThreshold.IsEmpty() {
			__antithesis_instrumentation__.Notify(96776)
			log.VEventf(ctx, 1, "LHS's GCThreshold of split is not set")
		} else {
			__antithesis_instrumentation__.Notify(96777)
		}
		__antithesis_instrumentation__.Notify(96764)

		replicaVersion, err := sl.LoadVersion(ctx, batch)
		if err != nil {
			__antithesis_instrumentation__.Notify(96778)
			return enginepb.MVCCStats{}, result.Result{}, errors.Wrap(err, "unable to load replica version")
		} else {
			__antithesis_instrumentation__.Notify(96779)
		}
		__antithesis_instrumentation__.Notify(96765)

		rangeAppliedState, err := sl.LoadRangeAppliedState(ctx, batch)
		if err != nil {
			__antithesis_instrumentation__.Notify(96780)
			return enginepb.MVCCStats{}, result.Result{},
				errors.Wrap(err, "unable to load range applied state")
		} else {
			__antithesis_instrumentation__.Notify(96781)
		}
		__antithesis_instrumentation__.Notify(96766)
		writeRaftAppliedIndexTerm := false
		if rangeAppliedState.RaftAppliedIndexTerm > 0 {
			__antithesis_instrumentation__.Notify(96782)
			writeRaftAppliedIndexTerm = true
		} else {
			__antithesis_instrumentation__.Notify(96783)
		}
		__antithesis_instrumentation__.Notify(96767)
		*h.AbsPostSplitRight(), err = stateloader.WriteInitialReplicaState(
			ctx, batch, *h.AbsPostSplitRight(), split.RightDesc, rightLease,
			*gcThreshold, replicaVersion, writeRaftAppliedIndexTerm,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(96784)
			return enginepb.MVCCStats{}, result.Result{}, errors.Wrap(err, "unable to write initial Replica state")
		} else {
			__antithesis_instrumentation__.Notify(96785)
		}
	}
	__antithesis_instrumentation__.Notify(96750)

	var pd result.Result
	pd.Replicated.Split = &kvserverpb.Split{
		SplitTrigger: *split,

		RHSDelta: *h.AbsPostSplitRight(),
	}

	deltaPostSplitLeft := h.DeltaPostSplitLeft()
	return deltaPostSplitLeft, pd, nil
}

func mergeTrigger(
	ctx context.Context,
	rec EvalContext,
	batch storage.Batch,
	ms *enginepb.MVCCStats,
	merge *roachpb.MergeTrigger,
	ts hlc.Timestamp,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96786)
	desc := rec.Desc()
	if !bytes.Equal(desc.StartKey, merge.LeftDesc.StartKey) {
		__antithesis_instrumentation__.Notify(96792)
		return result.Result{}, errors.AssertionFailedf("LHS range start keys do not match: %s != %s",
			desc.StartKey, merge.LeftDesc.StartKey)
	} else {
		__antithesis_instrumentation__.Notify(96793)
	}
	__antithesis_instrumentation__.Notify(96787)
	if !desc.EndKey.Less(merge.LeftDesc.EndKey) {
		__antithesis_instrumentation__.Notify(96794)
		return result.Result{}, errors.AssertionFailedf("original LHS end key is not less than the post merge end key: %s >= %s",
			desc.EndKey, merge.LeftDesc.EndKey)
	} else {
		__antithesis_instrumentation__.Notify(96795)
	}
	__antithesis_instrumentation__.Notify(96788)

	if err := abortspan.New(merge.RightDesc.RangeID).CopyTo(
		ctx, batch, batch, ms, ts, merge.LeftDesc.RangeID,
	); err != nil {
		__antithesis_instrumentation__.Notify(96796)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96797)
	}
	__antithesis_instrumentation__.Notify(96789)

	if merge.RightReadSummary != nil {
		__antithesis_instrumentation__.Notify(96798)
		mergedSum := merge.RightReadSummary.Clone()
		if priorSum, err := readsummary.Load(ctx, batch, rec.GetRangeID()); err != nil {
			__antithesis_instrumentation__.Notify(96800)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96801)
			if priorSum != nil {
				__antithesis_instrumentation__.Notify(96802)
				mergedSum.Merge(*priorSum)
			} else {
				__antithesis_instrumentation__.Notify(96803)
			}
		}
		__antithesis_instrumentation__.Notify(96799)
		if err := readsummary.Set(ctx, batch, rec.GetRangeID(), ms, mergedSum); err != nil {
			__antithesis_instrumentation__.Notify(96804)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96805)
		}
	} else {
		__antithesis_instrumentation__.Notify(96806)
	}
	__antithesis_instrumentation__.Notify(96790)

	ms.Add(merge.RightMVCCStats)
	{
		__antithesis_instrumentation__.Notify(96807)
		ridPrefix := keys.MakeRangeIDReplicatedPrefix(merge.RightDesc.RangeID)

		iter := batch.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{UpperBound: ridPrefix.PrefixEnd()})
		defer iter.Close()
		sysMS, err := iter.ComputeStats(ridPrefix, ridPrefix.PrefixEnd(), 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(96809)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96810)
		}
		__antithesis_instrumentation__.Notify(96808)
		ms.Subtract(sysMS)
	}
	__antithesis_instrumentation__.Notify(96791)

	var pd result.Result
	pd.Replicated.Merge = &kvserverpb.Merge{
		MergeTrigger: *merge,
	}
	return pd, nil
}

func changeReplicasTrigger(
	_ context.Context, rec EvalContext, _ storage.Batch, change *roachpb.ChangeReplicasTrigger,
) result.Result {
	__antithesis_instrumentation__.Notify(96811)
	var pd result.Result

	pd.Local.MaybeAddToSplitQueue = true

	pd.Local.GossipFirstRange = rec.IsFirstRange()

	pd.Replicated.State = &kvserverpb.ReplicaState{
		Desc: change.Desc,
	}
	pd.Replicated.ChangeReplicas = &kvserverpb.ChangeReplicas{
		ChangeReplicasTrigger: *change,
	}

	return pd
}

var txnAutoGC = true

func TestingSetTxnAutoGC(to bool) func() {
	__antithesis_instrumentation__.Notify(96812)
	prev := txnAutoGC
	txnAutoGC = to
	return func() { __antithesis_instrumentation__.Notify(96813); txnAutoGC = prev }
}
