package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/limit"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func (s *Store) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (br *roachpb.BatchResponse, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(125843)

	ctx = s.AnnotateCtx(ctx)
	for _, union := range ba.Requests {
		__antithesis_instrumentation__.Notify(125853)
		arg := union.GetInner()
		header := arg.Header()
		if err := verifyKeys(header.Key, header.EndKey, roachpb.IsRange(arg)); err != nil {
			__antithesis_instrumentation__.Notify(125854)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(125855)
		}
	}
	__antithesis_instrumentation__.Notify(125844)

	if res, err := s.maybeThrottleBatch(ctx, ba); err != nil {
		__antithesis_instrumentation__.Notify(125856)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(125857)
		if res != nil {
			__antithesis_instrumentation__.Notify(125858)
			defer res.Release()
		} else {
			__antithesis_instrumentation__.Notify(125859)
		}
	}
	__antithesis_instrumentation__.Notify(125845)

	if ba.BoundedStaleness != nil {
		__antithesis_instrumentation__.Notify(125860)
		newBa, pErr := s.executeServerSideBoundedStalenessNegotiation(ctx, ba)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(125862)
			return nil, pErr
		} else {
			__antithesis_instrumentation__.Notify(125863)
		}
		__antithesis_instrumentation__.Notify(125861)
		ba = newBa
	} else {
		__antithesis_instrumentation__.Notify(125864)
	}
	__antithesis_instrumentation__.Notify(125846)

	if err := ba.SetActiveTimestamp(s.Clock()); err != nil {
		__antithesis_instrumentation__.Notify(125865)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(125866)
	}
	__antithesis_instrumentation__.Notify(125847)

	if baClockTS, ok := ba.Timestamp.TryToClockTimestamp(); ok {
		__antithesis_instrumentation__.Notify(125867)
		if s.cfg.TestingKnobs.DisableMaxOffsetCheck {
			__antithesis_instrumentation__.Notify(125868)
			s.cfg.Clock.Update(baClockTS)
		} else {
			__antithesis_instrumentation__.Notify(125869)

			if err := s.cfg.Clock.UpdateAndCheckMaxOffset(ctx, baClockTS); err != nil {
				__antithesis_instrumentation__.Notify(125870)
				return nil, roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(125871)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(125872)
	}
	__antithesis_instrumentation__.Notify(125848)

	defer func() {
		__antithesis_instrumentation__.Notify(125873)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(125876)

			panic(r)
		} else {
			__antithesis_instrumentation__.Notify(125877)
		}
		__antithesis_instrumentation__.Notify(125874)
		if ba.Txn != nil {
			__antithesis_instrumentation__.Notify(125878)

			if pErr != nil {
				__antithesis_instrumentation__.Notify(125879)
				pErr.OriginNode = s.NodeID()
				if txn := pErr.GetTxn(); txn == nil {
					__antithesis_instrumentation__.Notify(125880)
					pErr.SetTxn(ba.Txn)
				} else {
					__antithesis_instrumentation__.Notify(125881)
				}
			} else {
				__antithesis_instrumentation__.Notify(125882)
				if br.Txn == nil {
					__antithesis_instrumentation__.Notify(125884)
					br.Txn = ba.Txn
				} else {
					__antithesis_instrumentation__.Notify(125885)
				}
				__antithesis_instrumentation__.Notify(125883)

				if ba.Timestamp.Less(br.Txn.WriteTimestamp) {
					__antithesis_instrumentation__.Notify(125886)
					if clockTS, ok := br.Txn.WriteTimestamp.TryToClockTimestamp(); ok {
						__antithesis_instrumentation__.Notify(125887)
						s.cfg.Clock.Update(clockTS)
					} else {
						__antithesis_instrumentation__.Notify(125888)
					}
				} else {
					__antithesis_instrumentation__.Notify(125889)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(125890)
			if pErr == nil {
				__antithesis_instrumentation__.Notify(125891)

				if ba.Timestamp.Less(br.Timestamp) {
					__antithesis_instrumentation__.Notify(125892)
					if clockTS, ok := br.Timestamp.TryToClockTimestamp(); ok {
						__antithesis_instrumentation__.Notify(125893)
						s.cfg.Clock.Update(clockTS)
					} else {
						__antithesis_instrumentation__.Notify(125894)
					}
				} else {
					__antithesis_instrumentation__.Notify(125895)
				}
			} else {
				__antithesis_instrumentation__.Notify(125896)
			}
		}
		__antithesis_instrumentation__.Notify(125875)

		now := s.cfg.Clock.NowAsClockTimestamp()
		if pErr != nil {
			__antithesis_instrumentation__.Notify(125897)
			pErr.Now = now
		} else {
			__antithesis_instrumentation__.Notify(125898)
			br.Now = now
		}
	}()
	__antithesis_instrumentation__.Notify(125849)

	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(125899)

		if _, ok := ba.Txn.GetObservedTimestamp(s.NodeID()); !ok {
			__antithesis_instrumentation__.Notify(125900)
			txnClone := ba.Txn.Clone()
			txnClone.UpdateObservedTimestamp(s.NodeID(), s.Clock().NowAsClockTimestamp())
			ba.Txn = txnClone
		} else {
			__antithesis_instrumentation__.Notify(125901)
		}
	} else {
		__antithesis_instrumentation__.Notify(125902)
	}
	__antithesis_instrumentation__.Notify(125850)

	if log.ExpensiveLogEnabled(ctx, 1) {
		__antithesis_instrumentation__.Notify(125903)
		log.Eventf(ctx, "executing %s", ba)
	} else {
		__antithesis_instrumentation__.Notify(125904)
	}
	__antithesis_instrumentation__.Notify(125851)

	var rangeInfos []roachpb.RangeInfo

	for {
		__antithesis_instrumentation__.Notify(125905)

		repl, err := s.GetReplica(ba.RangeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(125910)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(125911)
		}
		__antithesis_instrumentation__.Notify(125906)
		if !repl.IsInitialized() {
			__antithesis_instrumentation__.Notify(125912)

			return nil, roachpb.NewError(&roachpb.NotLeaseHolderError{
				RangeID:     ba.RangeID,
				LeaseHolder: repl.creatingReplica,

				Replica: roachpb.ReplicaDescriptor{
					NodeID:    repl.store.nodeDesc.NodeID,
					StoreID:   repl.store.StoreID(),
					ReplicaID: repl.replicaID,
				},
			})
		} else {
			__antithesis_instrumentation__.Notify(125913)
		}
		__antithesis_instrumentation__.Notify(125907)

		br, pErr = repl.Send(ctx, ba)
		if pErr == nil {
			__antithesis_instrumentation__.Notify(125914)

			if len(rangeInfos) > 0 {
				__antithesis_instrumentation__.Notify(125916)
				br.RangeInfos = append(rangeInfos, br.RangeInfos...)
			} else {
				__antithesis_instrumentation__.Notify(125917)
			}
			__antithesis_instrumentation__.Notify(125915)

			return br, nil
		} else {
			__antithesis_instrumentation__.Notify(125918)
		}
		__antithesis_instrumentation__.Notify(125908)

		switch t := pErr.GetDetail().(type) {
		case *roachpb.RangeKeyMismatchError:
			__antithesis_instrumentation__.Notify(125919)

			rSpan, err := keys.Range(ba.Requests)
			if err != nil {
				__antithesis_instrumentation__.Notify(125928)
				return nil, roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(125929)
			}
			__antithesis_instrumentation__.Notify(125920)

			ri, err := t.MismatchedRange()
			if err != nil {
				__antithesis_instrumentation__.Notify(125930)
				return nil, roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(125931)
			}
			__antithesis_instrumentation__.Notify(125921)
			skipRID := ri.Desc.RangeID
			startKey := ri.Desc.StartKey
			if rSpan.Key.Less(startKey) {
				__antithesis_instrumentation__.Notify(125932)
				startKey = rSpan.Key
			} else {
				__antithesis_instrumentation__.Notify(125933)
			}
			__antithesis_instrumentation__.Notify(125922)
			endKey := ri.Desc.EndKey
			if endKey.Less(rSpan.EndKey) {
				__antithesis_instrumentation__.Notify(125934)
				endKey = rSpan.EndKey
			} else {
				__antithesis_instrumentation__.Notify(125935)
			}
			__antithesis_instrumentation__.Notify(125923)
			var ris []roachpb.RangeInfo
			if err := s.visitReplicasByKey(ctx, startKey, endKey, AscendingKeyOrder, func(ctx context.Context, repl *Replica) error {
				__antithesis_instrumentation__.Notify(125936)

				ri := repl.GetRangeInfo(ctx)
				if ri.Desc.RangeID == skipRID {
					__antithesis_instrumentation__.Notify(125938)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(125939)
				}
				__antithesis_instrumentation__.Notify(125937)
				ris = append(ris, ri)
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(125940)

				log.Warningf(ctx, "unexpected error visiting replicas: %s", err)
				ris = nil
			} else {
				__antithesis_instrumentation__.Notify(125941)
			}
			__antithesis_instrumentation__.Notify(125924)

			t.AppendRangeInfo(ctx, ris...)

			isRetriableMismatch := false
			for _, ri := range ris {
				__antithesis_instrumentation__.Notify(125942)

				if ri.Desc.RSpan().ContainsKeyRange(rSpan.Key, rSpan.EndKey) {
					__antithesis_instrumentation__.Notify(125943)

					ba.RangeID = ri.Desc.RangeID
					ba.ClientRangeInfo = roachpb.ClientRangeInfo{
						ClosedTimestampPolicy: ri.ClosedTimestampPolicy,
						DescriptorGeneration:  ri.Desc.Generation,
						LeaseSequence:         ri.Lease.Sequence,
					}

					rangeInfos = append(rangeInfos, t.Ranges...)
					isRetriableMismatch = true
					break
				} else {
					__antithesis_instrumentation__.Notify(125944)
				}
			}
			__antithesis_instrumentation__.Notify(125925)

			if isRetriableMismatch {
				__antithesis_instrumentation__.Notify(125945)
				continue
			} else {
				__antithesis_instrumentation__.Notify(125946)
			}
			__antithesis_instrumentation__.Notify(125926)

			t.Ranges = append(rangeInfos, t.Ranges...)

			pErr = roachpb.NewError(t)
		case *roachpb.RaftGroupDeletedError:
			__antithesis_instrumentation__.Notify(125927)

			err := roachpb.NewRangeNotFoundError(repl.RangeID, repl.store.StoreID())
			pErr = roachpb.NewError(err)
		}
		__antithesis_instrumentation__.Notify(125909)

		break
	}
	__antithesis_instrumentation__.Notify(125852)
	return nil, pErr
}

func (s *Store) maybeThrottleBatch(
	ctx context.Context, ba roachpb.BatchRequest,
) (limit.Reservation, error) {
	__antithesis_instrumentation__.Notify(125947)
	if !ba.IsSingleRequest() {
		__antithesis_instrumentation__.Notify(125949)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(125950)
	}
	__antithesis_instrumentation__.Notify(125948)

	switch t := ba.Requests[0].GetInner().(type) {
	case *roachpb.AddSSTableRequest:
		__antithesis_instrumentation__.Notify(125951)
		limiter := s.limiters.ConcurrentAddSSTableRequests
		if t.IngestAsWrites {
			__antithesis_instrumentation__.Notify(125962)
			limiter = s.limiters.ConcurrentAddSSTableAsWritesRequests
		} else {
			__antithesis_instrumentation__.Notify(125963)
		}
		__antithesis_instrumentation__.Notify(125952)
		before := timeutil.Now()
		res, err := limiter.Begin(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(125964)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(125965)
		}
		__antithesis_instrumentation__.Notify(125953)

		beforeEngineDelay := timeutil.Now()
		s.engine.PreIngestDelay(ctx)
		after := timeutil.Now()

		waited, waitedEngine := after.Sub(before), after.Sub(beforeEngineDelay)
		s.metrics.AddSSTableProposalTotalDelay.Inc(waited.Nanoseconds())
		s.metrics.AddSSTableProposalEngineDelay.Inc(waitedEngine.Nanoseconds())
		if waited > time.Second {
			__antithesis_instrumentation__.Notify(125966)
			log.Infof(ctx, "SST ingestion was delayed by %v (%v for storage engine back-pressure)",
				waited, waitedEngine)
		} else {
			__antithesis_instrumentation__.Notify(125967)
		}
		__antithesis_instrumentation__.Notify(125954)
		return res, nil

	case *roachpb.ExportRequest:
		__antithesis_instrumentation__.Notify(125955)

		before := timeutil.Now()
		res, err := s.limiters.ConcurrentExportRequests.Begin(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(125968)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(125969)
		}
		__antithesis_instrumentation__.Notify(125956)

		waited := timeutil.Since(before)
		s.metrics.ExportRequestProposalTotalDelay.Inc(waited.Nanoseconds())
		if waited > time.Second {
			__antithesis_instrumentation__.Notify(125970)
			log.Infof(ctx, "Export request was delayed by %v", waited)
		} else {
			__antithesis_instrumentation__.Notify(125971)
		}
		__antithesis_instrumentation__.Notify(125957)
		return res, nil

	case *roachpb.ScanInterleavedIntentsRequest:
		__antithesis_instrumentation__.Notify(125958)
		before := timeutil.Now()
		res, err := s.limiters.ConcurrentScanInterleavedIntents.Begin(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(125972)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(125973)
		}
		__antithesis_instrumentation__.Notify(125959)

		waited := timeutil.Since(before)
		if waited > time.Second {
			__antithesis_instrumentation__.Notify(125974)
			log.Infof(ctx, "ScanInterleavedIntents request was delayed by %v", waited)
		} else {
			__antithesis_instrumentation__.Notify(125975)
		}
		__antithesis_instrumentation__.Notify(125960)
		return res, nil

	default:
		__antithesis_instrumentation__.Notify(125961)
		return nil, nil
	}
}

func (s *Store) executeServerSideBoundedStalenessNegotiation(
	ctx context.Context, ba roachpb.BatchRequest,
) (roachpb.BatchRequest, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(125976)
	if ba.BoundedStaleness == nil {
		__antithesis_instrumentation__.Notify(125987)
		log.Fatal(ctx, "BoundedStaleness header required for server-side negotiation fast-path")
	} else {
		__antithesis_instrumentation__.Notify(125988)
	}
	__antithesis_instrumentation__.Notify(125977)
	cfg := ba.BoundedStaleness
	if cfg.MinTimestampBound.IsEmpty() {
		__antithesis_instrumentation__.Notify(125989)
		return ba, roachpb.NewError(errors.AssertionFailedf(
			"MinTimestampBound must be set in batch"))
	} else {
		__antithesis_instrumentation__.Notify(125990)
	}
	__antithesis_instrumentation__.Notify(125978)
	if !cfg.MaxTimestampBound.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(125991)
		return cfg.MaxTimestampBound.LessEq(cfg.MinTimestampBound) == true
	}() == true {
		__antithesis_instrumentation__.Notify(125992)
		return ba, roachpb.NewError(errors.AssertionFailedf(
			"MaxTimestampBound, if set in batch, must be greater than MinTimestampBound"))
	} else {
		__antithesis_instrumentation__.Notify(125993)
	}
	__antithesis_instrumentation__.Notify(125979)
	if !ba.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(125994)
		return ba, roachpb.NewError(errors.AssertionFailedf(
			"MinTimestampBound and Timestamp cannot both be set in batch"))
	} else {
		__antithesis_instrumentation__.Notify(125995)
	}
	__antithesis_instrumentation__.Notify(125980)
	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(125996)
		return ba, roachpb.NewError(errors.AssertionFailedf(
			"MinTimestampBound and Txn cannot both be set in batch"))
	} else {
		__antithesis_instrumentation__.Notify(125997)
	}
	__antithesis_instrumentation__.Notify(125981)

	var queryResBa roachpb.BatchRequest
	queryResBa.RangeID = ba.RangeID
	queryResBa.Replica = ba.Replica
	queryResBa.ClientRangeInfo = ba.ClientRangeInfo
	queryResBa.ReadConsistency = roachpb.INCONSISTENT
	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(125998)
		span := ru.GetInner().Header().Span()
		if len(span.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(126000)

			span.EndKey = span.Key.Next()
		} else {
			__antithesis_instrumentation__.Notify(126001)
		}
		__antithesis_instrumentation__.Notify(125999)
		queryResBa.Add(&roachpb.QueryResolvedTimestampRequest{
			RequestHeader: roachpb.RequestHeaderFromSpan(span),
		})
	}
	__antithesis_instrumentation__.Notify(125982)

	br, pErr := s.Send(ctx, queryResBa)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(126002)
		return ba, pErr
	} else {
		__antithesis_instrumentation__.Notify(126003)
	}
	__antithesis_instrumentation__.Notify(125983)

	var resTS hlc.Timestamp
	for _, ru := range br.Responses {
		__antithesis_instrumentation__.Notify(126004)
		ts := ru.GetQueryResolvedTimestamp().ResolvedTS
		if resTS.IsEmpty() {
			__antithesis_instrumentation__.Notify(126005)
			resTS = ts
		} else {
			__antithesis_instrumentation__.Notify(126006)
			resTS.Backward(ts)
		}
	}
	__antithesis_instrumentation__.Notify(125984)
	if resTS.Less(cfg.MinTimestampBound) {
		__antithesis_instrumentation__.Notify(126007)

		if cfg.MinTimestampBoundStrict {
			__antithesis_instrumentation__.Notify(126009)
			return ba, roachpb.NewError(roachpb.NewMinTimestampBoundUnsatisfiableError(
				cfg.MinTimestampBound, resTS,
			))
		} else {
			__antithesis_instrumentation__.Notify(126010)
		}
		__antithesis_instrumentation__.Notify(126008)
		resTS = cfg.MinTimestampBound
	} else {
		__antithesis_instrumentation__.Notify(126011)
	}
	__antithesis_instrumentation__.Notify(125985)
	if !cfg.MaxTimestampBound.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(126012)
		return cfg.MaxTimestampBound.LessEq(resTS) == true
	}() == true {
		__antithesis_instrumentation__.Notify(126013)

		resTS = cfg.MaxTimestampBound.Prev()
	} else {
		__antithesis_instrumentation__.Notify(126014)
	}
	__antithesis_instrumentation__.Notify(125986)

	ba.Timestamp = resTS
	ba.BoundedStaleness = nil
	return ba, nil
}
