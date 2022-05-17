package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/google/btree"
)

const txnPipelinerBtreeDegree = 32

var pipelinedWritesEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.transaction.write_pipelining_enabled",
	"if enabled, transactional writes are pipelined through Raft consensus",
	true,
)
var pipelinedWritesMaxBatchSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.transaction.write_pipelining_max_batch_size",
	"if non-zero, defines that maximum size batch that will be pipelined through Raft consensus",

	128,
	settings.NonNegativeInt,
)

var TrackedWritesMaxSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.transaction.max_intents_bytes",
	"maximum number of bytes used to track locks in transactions",
	1<<22,
).WithPublic()

var rejectTxnOverTrackedWritesBudget = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.transaction.reject_over_max_intents_budget.enabled",
	"if set, transactions that exceed their lock tracking budget (kv.transaction.max_intents_bytes) "+
		"are rejected instead of having their lock spans imprecisely compressed",
	false,
).WithPublic()

type txnPipeliner struct {
	st                     *cluster.Settings
	riGen                  rangeIteratorFactory
	wrapped                lockedSender
	disabled               bool
	txnMetrics             *TxnMetrics
	condensedIntentsEveryN *log.EveryN

	ifWrites inFlightWriteSet

	lockFootprint condensableSpanSet
}

type condensableSpanSetRangeIterator interface {
	Valid() bool
	Seek(ctx context.Context, key roachpb.RKey, scanDir ScanDirection)
	Error() error
	Desc() *roachpb.RangeDescriptor
}

type rangeIteratorFactory struct {
	factory func() condensableSpanSetRangeIterator
	ds      *DistSender
}

func (f rangeIteratorFactory) newRangeIterator() condensableSpanSetRangeIterator {
	__antithesis_instrumentation__.Notify(88765)
	if f.factory != nil {
		__antithesis_instrumentation__.Notify(88768)
		return f.factory()
	} else {
		__antithesis_instrumentation__.Notify(88769)
	}
	__antithesis_instrumentation__.Notify(88766)
	if f.ds != nil {
		__antithesis_instrumentation__.Notify(88770)
		ri := MakeRangeIterator(f.ds)
		return &ri
	} else {
		__antithesis_instrumentation__.Notify(88771)
	}
	__antithesis_instrumentation__.Notify(88767)
	panic("no iterator factory configured")
}

func (tp *txnPipeliner) SendLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88772)

	ba, pErr := tp.attachLocksToEndTxn(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88776)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(88777)
	}
	__antithesis_instrumentation__.Notify(88773)

	rejectOverBudget := rejectTxnOverTrackedWritesBudget.Get(&tp.st.SV)
	maxBytes := TrackedWritesMaxSize.Get(&tp.st.SV)
	if rejectOverBudget {
		__antithesis_instrumentation__.Notify(88778)
		if err := tp.maybeRejectOverBudget(ba, maxBytes); err != nil {
			__antithesis_instrumentation__.Notify(88779)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(88780)
		}
	} else {
		__antithesis_instrumentation__.Notify(88781)
	}
	__antithesis_instrumentation__.Notify(88774)

	ba.AsyncConsensus = tp.canUseAsyncConsensus(ctx, ba)

	ba = tp.chainToInFlightWrites(ba)

	br, pErr := tp.wrapped.SendLocked(ctx, ba)

	tp.updateLockTracking(ctx, ba, br, pErr, maxBytes, !rejectOverBudget)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88782)
		return nil, tp.adjustError(ctx, ba, pErr)
	} else {
		__antithesis_instrumentation__.Notify(88783)
	}
	__antithesis_instrumentation__.Notify(88775)
	return tp.stripQueryIntents(br), nil
}

func (tp *txnPipeliner) maybeRejectOverBudget(ba roachpb.BatchRequest, maxBytes int64) error {
	__antithesis_instrumentation__.Notify(88784)

	if !ba.IsLocking() {
		__antithesis_instrumentation__.Notify(88788)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(88789)
	}
	__antithesis_instrumentation__.Notify(88785)

	var spans []roachpb.Span
	ba.LockSpanIterate(nil, func(sp roachpb.Span, _ lock.Durability) {
		__antithesis_instrumentation__.Notify(88790)
		spans = append(spans, sp)
	})
	__antithesis_instrumentation__.Notify(88786)

	locksBudget := maxBytes - tp.ifWrites.byteSize()

	estimate := tp.lockFootprint.estimateSize(spans, locksBudget)
	if estimate > locksBudget {
		__antithesis_instrumentation__.Notify(88791)
		tp.txnMetrics.TxnsRejectedByLockSpanBudget.Inc(1)
		bErr := newLockSpansOverBudgetError(estimate+tp.ifWrites.byteSize(), maxBytes, ba)
		return pgerror.WithCandidateCode(bErr, pgcode.ConfigurationLimitExceeded)
	} else {
		__antithesis_instrumentation__.Notify(88792)
	}
	__antithesis_instrumentation__.Notify(88787)
	return nil
}

func (tp *txnPipeliner) attachLocksToEndTxn(
	ctx context.Context, ba roachpb.BatchRequest,
) (roachpb.BatchRequest, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88793)
	args, hasET := ba.GetArg(roachpb.EndTxn)
	if !hasET {
		__antithesis_instrumentation__.Notify(88801)
		return ba, nil
	} else {
		__antithesis_instrumentation__.Notify(88802)
	}
	__antithesis_instrumentation__.Notify(88794)
	et := args.(*roachpb.EndTxnRequest)
	if len(et.LockSpans) > 0 {
		__antithesis_instrumentation__.Notify(88803)
		return ba, roachpb.NewErrorf("client must not pass intents to EndTxn")
	} else {
		__antithesis_instrumentation__.Notify(88804)
	}
	__antithesis_instrumentation__.Notify(88795)
	if len(et.InFlightWrites) > 0 {
		__antithesis_instrumentation__.Notify(88805)
		return ba, roachpb.NewErrorf("client must not pass in-flight writes to EndTxn")
	} else {
		__antithesis_instrumentation__.Notify(88806)
	}
	__antithesis_instrumentation__.Notify(88796)

	if !tp.lockFootprint.empty() {
		__antithesis_instrumentation__.Notify(88807)
		et.LockSpans = append([]roachpb.Span(nil), tp.lockFootprint.asSlice()...)
	} else {
		__antithesis_instrumentation__.Notify(88808)
	}
	__antithesis_instrumentation__.Notify(88797)
	if inFlight := tp.ifWrites.len(); inFlight != 0 {
		__antithesis_instrumentation__.Notify(88809)
		et.InFlightWrites = make([]roachpb.SequencedWrite, 0, inFlight)
		tp.ifWrites.ascend(func(w *inFlightWrite) {
			__antithesis_instrumentation__.Notify(88810)
			et.InFlightWrites = append(et.InFlightWrites, w.SequencedWrite)
		})
	} else {
		__antithesis_instrumentation__.Notify(88811)
	}
	__antithesis_instrumentation__.Notify(88798)

	for _, ru := range ba.Requests[:len(ba.Requests)-1] {
		__antithesis_instrumentation__.Notify(88812)
		req := ru.GetInner()
		h := req.Header()
		if roachpb.IsLocking(req) {
			__antithesis_instrumentation__.Notify(88813)

			if roachpb.IsIntentWrite(req) && func() bool {
				__antithesis_instrumentation__.Notify(88814)
				return !roachpb.IsRange(req) == true
			}() == true {
				__antithesis_instrumentation__.Notify(88815)
				w := roachpb.SequencedWrite{Key: h.Key, Sequence: h.Sequence}
				et.InFlightWrites = append(et.InFlightWrites, w)
			} else {
				__antithesis_instrumentation__.Notify(88816)
				et.LockSpans = append(et.LockSpans, h.Span())
			}
		} else {
			__antithesis_instrumentation__.Notify(88817)
		}
	}
	__antithesis_instrumentation__.Notify(88799)

	et.LockSpans, _ = roachpb.MergeSpans(&et.LockSpans)
	sort.Sort(roachpb.SequencedWriteBySeq(et.InFlightWrites))

	if log.V(3) {
		__antithesis_instrumentation__.Notify(88818)
		for _, intent := range et.LockSpans {
			__antithesis_instrumentation__.Notify(88820)
			log.Infof(ctx, "intent: [%s,%s)", intent.Key, intent.EndKey)
		}
		__antithesis_instrumentation__.Notify(88819)
		for _, write := range et.InFlightWrites {
			__antithesis_instrumentation__.Notify(88821)
			log.Infof(ctx, "in-flight: %d:%s", write.Sequence, write.Key)
		}
	} else {
		__antithesis_instrumentation__.Notify(88822)
	}
	__antithesis_instrumentation__.Notify(88800)
	return ba, nil
}

func (tp *txnPipeliner) canUseAsyncConsensus(ctx context.Context, ba roachpb.BatchRequest) bool {
	__antithesis_instrumentation__.Notify(88823)

	if _, hasET := ba.GetArg(roachpb.EndTxn); hasET {
		__antithesis_instrumentation__.Notify(88828)
		return false
	} else {
		__antithesis_instrumentation__.Notify(88829)
	}
	__antithesis_instrumentation__.Notify(88824)

	if !pipelinedWritesEnabled.Get(&tp.st.SV) || func() bool {
		__antithesis_instrumentation__.Notify(88830)
		return tp.disabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(88831)
		return false
	} else {
		__antithesis_instrumentation__.Notify(88832)
	}
	__antithesis_instrumentation__.Notify(88825)

	addedIFBytes := int64(0)
	maxTrackingBytes := TrackedWritesMaxSize.Get(&tp.st.SV)

	if maxBatch := pipelinedWritesMaxBatchSize.Get(&tp.st.SV); maxBatch > 0 {
		__antithesis_instrumentation__.Notify(88833)
		batchSize := int64(len(ba.Requests))
		if batchSize > maxBatch {
			__antithesis_instrumentation__.Notify(88834)
			return false
		} else {
			__antithesis_instrumentation__.Notify(88835)
		}
	} else {
		__antithesis_instrumentation__.Notify(88836)
	}
	__antithesis_instrumentation__.Notify(88826)

	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(88837)
		req := ru.GetInner()

		if !roachpb.IsIntentWrite(req) || func() bool {
			__antithesis_instrumentation__.Notify(88839)
			return roachpb.IsRange(req) == true
		}() == true {
			__antithesis_instrumentation__.Notify(88840)

			return false
		} else {
			__antithesis_instrumentation__.Notify(88841)
		}
		__antithesis_instrumentation__.Notify(88838)

		addedIFBytes += keySize(req.Header().Key)
		if (tp.ifWrites.byteSize() + addedIFBytes + tp.lockFootprint.bytes) > maxTrackingBytes {
			__antithesis_instrumentation__.Notify(88842)
			log.VEventf(ctx, 2, "cannot perform async consensus because memory budget exceeded")
			return false
		} else {
			__antithesis_instrumentation__.Notify(88843)
		}
	}
	__antithesis_instrumentation__.Notify(88827)
	return true
}

func (tp *txnPipeliner) chainToInFlightWrites(ba roachpb.BatchRequest) roachpb.BatchRequest {
	__antithesis_instrumentation__.Notify(88844)

	if tp.ifWrites.len() == 0 {
		__antithesis_instrumentation__.Notify(88847)
		return ba
	} else {
		__antithesis_instrumentation__.Notify(88848)
	}
	__antithesis_instrumentation__.Notify(88845)

	forked := false
	oldReqs := ba.Requests

	var chainedKeys map[string]struct{}
	for i, ru := range oldReqs {
		__antithesis_instrumentation__.Notify(88849)
		req := ru.GetInner()

		if tp.ifWrites.len() > len(chainedKeys) {
			__antithesis_instrumentation__.Notify(88851)

			writeIter := func(w *inFlightWrite) {
				__antithesis_instrumentation__.Notify(88853)

				if !forked {
					__antithesis_instrumentation__.Notify(88855)
					ba.Requests = append([]roachpb.RequestUnion(nil), ba.Requests[:i]...)
					forked = true
				} else {
					__antithesis_instrumentation__.Notify(88856)
				}
				__antithesis_instrumentation__.Notify(88854)

				if _, ok := chainedKeys[string(w.Key)]; !ok {
					__antithesis_instrumentation__.Notify(88857)

					meta := ba.Txn.TxnMeta
					meta.Sequence = w.Sequence
					ba.Add(&roachpb.QueryIntentRequest{
						RequestHeader: roachpb.RequestHeader{
							Key: w.Key,
						},
						Txn:            meta,
						ErrorIfMissing: true,
					})

					if chainedKeys == nil {
						__antithesis_instrumentation__.Notify(88859)
						chainedKeys = make(map[string]struct{})
					} else {
						__antithesis_instrumentation__.Notify(88860)
					}
					__antithesis_instrumentation__.Notify(88858)
					chainedKeys[string(w.Key)] = struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(88861)
				}
			}
			__antithesis_instrumentation__.Notify(88852)

			if !roachpb.IsTransactional(req) {
				__antithesis_instrumentation__.Notify(88862)

				tp.ifWrites.ascend(writeIter)
			} else {
				__antithesis_instrumentation__.Notify(88863)
				if et, ok := req.(*roachpb.EndTxnRequest); ok {
					__antithesis_instrumentation__.Notify(88864)
					if et.Commit {
						__antithesis_instrumentation__.Notify(88865)

						tp.ifWrites.ascend(writeIter)
					} else {
						__antithesis_instrumentation__.Notify(88866)
					}
				} else {
					__antithesis_instrumentation__.Notify(88867)

					s := req.Header().Span()
					tp.ifWrites.ascendRange(s.Key, s.EndKey, writeIter)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(88868)
		}
		__antithesis_instrumentation__.Notify(88850)

		if forked {
			__antithesis_instrumentation__.Notify(88869)
			ba.Add(req)
		} else {
			__antithesis_instrumentation__.Notify(88870)
		}
	}
	__antithesis_instrumentation__.Notify(88846)

	return ba
}

func (tp *txnPipeliner) updateLockTracking(
	ctx context.Context,
	ba roachpb.BatchRequest,
	br *roachpb.BatchResponse,
	pErr *roachpb.Error,
	maxBytes int64,
	condenseLocksIfOverBudget bool,
) {
	__antithesis_instrumentation__.Notify(88871)
	tp.updateLockTrackingInner(ctx, ba, br, pErr)

	locksBudget := maxBytes - tp.ifWrites.byteSize()

	if tp.lockFootprint.bytesSize() <= locksBudget {
		__antithesis_instrumentation__.Notify(88874)
		return
	} else {
		__antithesis_instrumentation__.Notify(88875)
	}
	__antithesis_instrumentation__.Notify(88872)

	if !condenseLocksIfOverBudget {
		__antithesis_instrumentation__.Notify(88876)
		return
	} else {
		__antithesis_instrumentation__.Notify(88877)
	}
	__antithesis_instrumentation__.Notify(88873)

	alreadyCondensed := tp.lockFootprint.condensed
	condensed := tp.lockFootprint.maybeCondense(ctx, tp.riGen, locksBudget)
	if condensed && func() bool {
		__antithesis_instrumentation__.Notify(88878)
		return !alreadyCondensed == true
	}() == true {
		__antithesis_instrumentation__.Notify(88879)
		if tp.condensedIntentsEveryN.ShouldLog() || func() bool {
			__antithesis_instrumentation__.Notify(88881)
			return log.ExpensiveLogEnabled(ctx, 2) == true
		}() == true {
			__antithesis_instrumentation__.Notify(88882)
			log.Warningf(ctx,
				"a transaction has hit the intent tracking limit (kv.transaction.max_intents_bytes); "+
					"is it a bulk operation? Intent cleanup will be slower. txn: %s ba: %s",
				ba.Txn, ba.Summary())
		} else {
			__antithesis_instrumentation__.Notify(88883)
		}
		__antithesis_instrumentation__.Notify(88880)
		tp.txnMetrics.TxnsWithCondensedIntents.Inc(1)
		tp.txnMetrics.TxnsWithCondensedIntentsGauge.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(88884)
	}
}

func (tp *txnPipeliner) updateLockTrackingInner(
	ctx context.Context, ba roachpb.BatchRequest, br *roachpb.BatchResponse, pErr *roachpb.Error,
) {
	__antithesis_instrumentation__.Notify(88885)

	if pErr != nil {
		__antithesis_instrumentation__.Notify(88888)

		baStripped := ba
		if roachpb.ErrPriority(pErr.GoError()) <= roachpb.ErrorScoreUnambiguousError && func() bool {
			__antithesis_instrumentation__.Notify(88890)
			return pErr.Index != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(88891)
			baStripped.Requests = make([]roachpb.RequestUnion, len(ba.Requests)-1)
			copy(baStripped.Requests, ba.Requests[:pErr.Index.Index])
			copy(baStripped.Requests[pErr.Index.Index:], ba.Requests[pErr.Index.Index+1:])
		} else {
			__antithesis_instrumentation__.Notify(88892)
		}
		__antithesis_instrumentation__.Notify(88889)
		baStripped.LockSpanIterate(nil, tp.trackLocks)
		return
	} else {
		__antithesis_instrumentation__.Notify(88893)
	}
	__antithesis_instrumentation__.Notify(88886)

	if br.Txn.Status.IsFinalized() {
		__antithesis_instrumentation__.Notify(88894)
		switch br.Txn.Status {
		case roachpb.ABORTED:
			__antithesis_instrumentation__.Notify(88896)

			ba.LockSpanIterate(nil, tp.trackLocks)
		case roachpb.COMMITTED:
			__antithesis_instrumentation__.Notify(88897)

			tp.ifWrites.clear(

				false)
		default:
			__antithesis_instrumentation__.Notify(88898)
			panic("unexpected")
		}
		__antithesis_instrumentation__.Notify(88895)
		return
	} else {
		__antithesis_instrumentation__.Notify(88899)
	}
	__antithesis_instrumentation__.Notify(88887)

	for i, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(88900)
		req := ru.GetInner()
		resp := br.Responses[i].GetInner()

		if qiReq, ok := req.(*roachpb.QueryIntentRequest); ok {
			__antithesis_instrumentation__.Notify(88901)

			if resp.(*roachpb.QueryIntentResponse).FoundIntent {
				__antithesis_instrumentation__.Notify(88902)
				tp.ifWrites.remove(qiReq.Key, qiReq.Txn.Sequence)

				tp.lockFootprint.insert(roachpb.Span{Key: qiReq.Key})
			} else {
				__antithesis_instrumentation__.Notify(88903)
			}
		} else {
			__antithesis_instrumentation__.Notify(88904)
			if roachpb.IsLocking(req) {
				__antithesis_instrumentation__.Notify(88905)

				if ba.AsyncConsensus {
					__antithesis_instrumentation__.Notify(88906)

					header := req.Header()
					tp.ifWrites.insert(header.Key, header.Sequence)

					if roachpb.IsRange(req) {
						__antithesis_instrumentation__.Notify(88907)
						log.Fatalf(ctx, "unexpected range request with AsyncConsensus: %s", req)
					} else {
						__antithesis_instrumentation__.Notify(88908)
					}
				} else {
					__antithesis_instrumentation__.Notify(88909)

					if sp, ok := roachpb.ActualSpan(req, resp); ok {
						__antithesis_instrumentation__.Notify(88910)
						tp.lockFootprint.insert(sp)
					} else {
						__antithesis_instrumentation__.Notify(88911)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(88912)
			}
		}
	}
}

func (tp *txnPipeliner) trackLocks(s roachpb.Span, _ lock.Durability) {
	__antithesis_instrumentation__.Notify(88913)
	tp.lockFootprint.insert(s)
}

func (tp *txnPipeliner) stripQueryIntents(br *roachpb.BatchResponse) *roachpb.BatchResponse {
	__antithesis_instrumentation__.Notify(88914)
	j := 0
	for i, ru := range br.Responses {
		__antithesis_instrumentation__.Notify(88916)
		if ru.GetQueryIntent() != nil {
			__antithesis_instrumentation__.Notify(88919)
			continue
		} else {
			__antithesis_instrumentation__.Notify(88920)
		}
		__antithesis_instrumentation__.Notify(88917)
		if i != j {
			__antithesis_instrumentation__.Notify(88921)
			br.Responses[j] = br.Responses[i]
		} else {
			__antithesis_instrumentation__.Notify(88922)
		}
		__antithesis_instrumentation__.Notify(88918)
		j++
	}
	__antithesis_instrumentation__.Notify(88915)
	br.Responses = br.Responses[:j]
	return br
}

func (tp *txnPipeliner) adjustError(
	ctx context.Context, ba roachpb.BatchRequest, pErr *roachpb.Error,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(88923)

	if pErr.Index != nil {
		__antithesis_instrumentation__.Notify(88926)
		before := int32(0)
		for _, ru := range ba.Requests[:int(pErr.Index.Index)] {
			__antithesis_instrumentation__.Notify(88928)
			req := ru.GetInner()
			if req.Method() == roachpb.QueryIntent {
				__antithesis_instrumentation__.Notify(88929)
				before++
			} else {
				__antithesis_instrumentation__.Notify(88930)
			}
		}
		__antithesis_instrumentation__.Notify(88927)
		pErr.Index.Index -= before
	} else {
		__antithesis_instrumentation__.Notify(88931)
	}
	__antithesis_instrumentation__.Notify(88924)

	if ime, ok := pErr.GetDetail().(*roachpb.IntentMissingError); ok {
		__antithesis_instrumentation__.Notify(88932)
		log.VEventf(ctx, 2, "transforming intent missing error into retry: %v", ime)
		err := roachpb.NewTransactionRetryError(
			roachpb.RETRY_ASYNC_WRITE_FAILURE, fmt.Sprintf("missing intent on: %s", ime.Key))
		retryErr := roachpb.NewErrorWithTxn(err, pErr.GetTxn())
		retryErr.Index = pErr.Index
		return retryErr
	} else {
		__antithesis_instrumentation__.Notify(88933)
	}
	__antithesis_instrumentation__.Notify(88925)
	return pErr
}

func (tp *txnPipeliner) setWrapped(wrapped lockedSender) {
	__antithesis_instrumentation__.Notify(88934)
	tp.wrapped = wrapped
}

func (tp *txnPipeliner) populateLeafInputState(tis *roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(88935)
	tis.InFlightWrites = tp.ifWrites.asSlice()
}

func (tp *txnPipeliner) initializeLeaf(tis *roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(88936)

	for _, w := range tis.InFlightWrites {
		__antithesis_instrumentation__.Notify(88937)
		tp.ifWrites.insert(w.Key, w.Sequence)
	}
}

func (tp *txnPipeliner) populateLeafFinalState(*roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88938)
}

func (tp *txnPipeliner) importLeafFinalState(context.Context, *roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88939)
}

func (tp *txnPipeliner) epochBumpedLocked() {
	__antithesis_instrumentation__.Notify(88940)

	if tp.ifWrites.len() > 0 {
		__antithesis_instrumentation__.Notify(88941)
		tp.ifWrites.ascend(func(w *inFlightWrite) {
			__antithesis_instrumentation__.Notify(88943)
			tp.lockFootprint.insert(roachpb.Span{Key: w.Key})
		})
		__antithesis_instrumentation__.Notify(88942)
		tp.lockFootprint.mergeAndSort()
		tp.ifWrites.clear(true)
	} else {
		__antithesis_instrumentation__.Notify(88944)
	}
}

func (tp *txnPipeliner) createSavepointLocked(context.Context, *savepoint) {
	__antithesis_instrumentation__.Notify(88945)
}

func (tp *txnPipeliner) rollbackToSavepointLocked(ctx context.Context, s savepoint) {
	__antithesis_instrumentation__.Notify(88946)

	var writesToDelete []*inFlightWrite
	needCollecting := !s.Initial()
	tp.ifWrites.ascend(func(w *inFlightWrite) {
		__antithesis_instrumentation__.Notify(88948)
		if w.Sequence > s.seqNum {
			__antithesis_instrumentation__.Notify(88949)
			tp.lockFootprint.insert(roachpb.Span{Key: w.Key})
			if needCollecting {
				__antithesis_instrumentation__.Notify(88950)
				writesToDelete = append(writesToDelete, w)
			} else {
				__antithesis_instrumentation__.Notify(88951)
			}
		} else {
			__antithesis_instrumentation__.Notify(88952)
		}
	})
	__antithesis_instrumentation__.Notify(88947)
	tp.lockFootprint.mergeAndSort()

	if needCollecting {
		__antithesis_instrumentation__.Notify(88953)
		for _, ifw := range writesToDelete {
			__antithesis_instrumentation__.Notify(88954)
			tp.ifWrites.remove(ifw.Key, ifw.Sequence)
		}
	} else {
		__antithesis_instrumentation__.Notify(88955)
		tp.ifWrites.clear(true)
	}
}

func (tp *txnPipeliner) closeLocked() {
	__antithesis_instrumentation__.Notify(88956)
	if tp.lockFootprint.condensed {
		__antithesis_instrumentation__.Notify(88957)
		tp.txnMetrics.TxnsWithCondensedIntentsGauge.Dec(1)
	} else {
		__antithesis_instrumentation__.Notify(88958)
	}
}

func (tp *txnPipeliner) hasAcquiredLocks() bool {
	__antithesis_instrumentation__.Notify(88959)
	return tp.ifWrites.len() > 0 || func() bool {
		__antithesis_instrumentation__.Notify(88960)
		return !tp.lockFootprint.empty() == true
	}() == true
}

type inFlightWrite struct {
	roachpb.SequencedWrite
}

func (a *inFlightWrite) Less(b btree.Item) bool {
	__antithesis_instrumentation__.Notify(88961)
	return a.Key.Compare(b.(*inFlightWrite).Key) < 0
}

type inFlightWriteSet struct {
	t     *btree.BTree
	bytes int64

	tmp1, tmp2 inFlightWrite
	alloc      inFlightWriteAlloc
}

func (s *inFlightWriteSet) insert(key roachpb.Key, seq enginepb.TxnSeq) {
	__antithesis_instrumentation__.Notify(88962)
	if s.t == nil {
		__antithesis_instrumentation__.Notify(88965)

		s.t = btree.New(txnPipelinerBtreeDegree)
	} else {
		__antithesis_instrumentation__.Notify(88966)
	}
	__antithesis_instrumentation__.Notify(88963)

	s.tmp1.Key = key
	item := s.t.Get(&s.tmp1)
	if item != nil {
		__antithesis_instrumentation__.Notify(88967)
		otherW := item.(*inFlightWrite)
		if seq > otherW.Sequence {
			__antithesis_instrumentation__.Notify(88969)

			otherW.Sequence = seq
		} else {
			__antithesis_instrumentation__.Notify(88970)
		}
		__antithesis_instrumentation__.Notify(88968)
		return
	} else {
		__antithesis_instrumentation__.Notify(88971)
	}
	__antithesis_instrumentation__.Notify(88964)

	w := s.alloc.alloc(key, seq)
	s.t.ReplaceOrInsert(w)
	s.bytes += keySize(key)
}

func (s *inFlightWriteSet) remove(key roachpb.Key, seq enginepb.TxnSeq) {
	__antithesis_instrumentation__.Notify(88972)
	if s.len() == 0 {
		__antithesis_instrumentation__.Notify(88977)

		return
	} else {
		__antithesis_instrumentation__.Notify(88978)
	}
	__antithesis_instrumentation__.Notify(88973)

	s.tmp1.Key = key
	item := s.t.Get(&s.tmp1)
	if item == nil {
		__antithesis_instrumentation__.Notify(88979)

		return
	} else {
		__antithesis_instrumentation__.Notify(88980)
	}
	__antithesis_instrumentation__.Notify(88974)

	w := item.(*inFlightWrite)
	if seq < w.Sequence {
		__antithesis_instrumentation__.Notify(88981)

		return
	} else {
		__antithesis_instrumentation__.Notify(88982)
	}
	__antithesis_instrumentation__.Notify(88975)

	delItem := s.t.Delete(item)
	if delItem != nil {
		__antithesis_instrumentation__.Notify(88983)
		*delItem.(*inFlightWrite) = inFlightWrite{}
	} else {
		__antithesis_instrumentation__.Notify(88984)
	}
	__antithesis_instrumentation__.Notify(88976)
	s.bytes -= keySize(key)

	if s.bytes < 0 {
		__antithesis_instrumentation__.Notify(88985)
		panic("negative in-flight write size")
	} else {
		__antithesis_instrumentation__.Notify(88986)
		if s.t.Len() == 0 && func() bool {
			__antithesis_instrumentation__.Notify(88987)
			return s.bytes != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(88988)
			panic("non-zero in-flight write size with 0 in-flight writes")
		} else {
			__antithesis_instrumentation__.Notify(88989)
		}
	}
}

func (s *inFlightWriteSet) ascend(f func(w *inFlightWrite)) {
	__antithesis_instrumentation__.Notify(88990)
	if s.len() == 0 {
		__antithesis_instrumentation__.Notify(88992)

		return
	} else {
		__antithesis_instrumentation__.Notify(88993)
	}
	__antithesis_instrumentation__.Notify(88991)
	s.t.Ascend(func(i btree.Item) bool {
		__antithesis_instrumentation__.Notify(88994)
		f(i.(*inFlightWrite))
		return true
	})
}

func (s *inFlightWriteSet) ascendRange(start, end roachpb.Key, f func(w *inFlightWrite)) {
	__antithesis_instrumentation__.Notify(88995)
	if s.len() == 0 {
		__antithesis_instrumentation__.Notify(88997)

		return
	} else {
		__antithesis_instrumentation__.Notify(88998)
	}
	__antithesis_instrumentation__.Notify(88996)
	if end == nil {
		__antithesis_instrumentation__.Notify(88999)

		s.tmp1.Key = start
		if i := s.t.Get(&s.tmp1); i != nil {
			__antithesis_instrumentation__.Notify(89000)
			f(i.(*inFlightWrite))
		} else {
			__antithesis_instrumentation__.Notify(89001)
		}
	} else {
		__antithesis_instrumentation__.Notify(89002)

		s.tmp1.Key, s.tmp2.Key = start, end
		s.t.AscendRange(&s.tmp1, &s.tmp2, func(i btree.Item) bool {
			__antithesis_instrumentation__.Notify(89003)
			f(i.(*inFlightWrite))
			return true
		})
	}
}

func (s *inFlightWriteSet) len() int {
	__antithesis_instrumentation__.Notify(89004)
	if s.t == nil {
		__antithesis_instrumentation__.Notify(89006)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(89007)
	}
	__antithesis_instrumentation__.Notify(89005)
	return s.t.Len()
}

func (s *inFlightWriteSet) byteSize() int64 {
	__antithesis_instrumentation__.Notify(89008)
	return s.bytes
}

func (s *inFlightWriteSet) clear(reuse bool) {
	__antithesis_instrumentation__.Notify(89009)
	if s.t == nil {
		__antithesis_instrumentation__.Notify(89011)
		return
	} else {
		__antithesis_instrumentation__.Notify(89012)
	}
	__antithesis_instrumentation__.Notify(89010)
	s.t.Clear(reuse)
	s.bytes = 0
	s.alloc.clear()
}

func (s *inFlightWriteSet) asSlice() []roachpb.SequencedWrite {
	__antithesis_instrumentation__.Notify(89013)
	l := s.len()
	if l == 0 {
		__antithesis_instrumentation__.Notify(89016)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89017)
	}
	__antithesis_instrumentation__.Notify(89014)
	writes := make([]roachpb.SequencedWrite, 0, l)
	s.ascend(func(w *inFlightWrite) {
		__antithesis_instrumentation__.Notify(89018)
		writes = append(writes, w.SequencedWrite)
	})
	__antithesis_instrumentation__.Notify(89015)
	return writes
}

type inFlightWriteAlloc []inFlightWrite

func (a *inFlightWriteAlloc) alloc(key roachpb.Key, seq enginepb.TxnSeq) *inFlightWrite {
	__antithesis_instrumentation__.Notify(89019)

	if cap(*a)-len(*a) == 0 {
		__antithesis_instrumentation__.Notify(89021)
		const chunkAllocMinSize = 4
		const chunkAllocMaxSize = 1024

		allocSize := cap(*a) * 2
		if allocSize < chunkAllocMinSize {
			__antithesis_instrumentation__.Notify(89023)
			allocSize = chunkAllocMinSize
		} else {
			__antithesis_instrumentation__.Notify(89024)
			if allocSize > chunkAllocMaxSize {
				__antithesis_instrumentation__.Notify(89025)
				allocSize = chunkAllocMaxSize
			} else {
				__antithesis_instrumentation__.Notify(89026)
			}
		}
		__antithesis_instrumentation__.Notify(89022)
		*a = make([]inFlightWrite, 0, allocSize)
	} else {
		__antithesis_instrumentation__.Notify(89027)
	}
	__antithesis_instrumentation__.Notify(89020)

	*a = (*a)[:len(*a)+1]
	w := &(*a)[len(*a)-1]
	*w = inFlightWrite{
		SequencedWrite: roachpb.SequencedWrite{Key: key, Sequence: seq},
	}
	return w
}

func (a *inFlightWriteAlloc) clear() {
	__antithesis_instrumentation__.Notify(89028)
	for i := range *a {
		__antithesis_instrumentation__.Notify(89030)
		(*a)[i] = inFlightWrite{}
	}
	__antithesis_instrumentation__.Notify(89029)
	*a = (*a)[:0]
}
