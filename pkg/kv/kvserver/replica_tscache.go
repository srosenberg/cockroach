package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/tscache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/redact"
)

func (r *Replica) addToTSCacheChecked(
	ctx context.Context,
	st *kvserverpb.LeaseStatus,
	ba *roachpb.BatchRequest,
	br *roachpb.BatchResponse,
	pErr *roachpb.Error,
	start, end roachpb.Key,
	ts hlc.Timestamp,
	txnID uuid.UUID,
) {
	__antithesis_instrumentation__.Notify(120827)

	if exp := st.Expiration(); exp.LessEq(ts) {
		__antithesis_instrumentation__.Notify(120830)
		log.Fatalf(ctx, "Unsafe timestamp cache update! Cannot add timestamp %s to timestamp "+
			"cache after evaluating %v (resp=%v; err=%v) with lease expiration %v. The timestamp "+
			"cache update could be lost of a non-cooperative lease change.", ts, ba, br, pErr, exp)
	} else {
		__antithesis_instrumentation__.Notify(120831)
	}
	__antithesis_instrumentation__.Notify(120828)

	if !ts.Synthetic && func() bool {
		__antithesis_instrumentation__.Notify(120832)
		return st.Now.ToTimestamp().Less(ts) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(120833)
		return false == true
	}() == true {
		__antithesis_instrumentation__.Notify(120834)
		log.Fatalf(ctx, "Unsafe timestamp cache update! Cannot add timestamp %s to timestamp "+
			"cache after evaluating %v (resp=%v; err=%v) with local hlc clock at timestamp %s. "+
			"Non-synthetic timestamps should always lag the local hlc clock.", ts, ba, br, pErr, st.Now)
	} else {
		__antithesis_instrumentation__.Notify(120835)
	}
	__antithesis_instrumentation__.Notify(120829)
	r.store.tsCache.Add(start, end, ts, txnID)
}

func (r *Replica) updateTimestampCache(
	ctx context.Context,
	st *kvserverpb.LeaseStatus,
	ba *roachpb.BatchRequest,
	br *roachpb.BatchResponse,
	pErr *roachpb.Error,
) {
	__antithesis_instrumentation__.Notify(120836)
	addToTSCache := func(start, end roachpb.Key, ts hlc.Timestamp, txnID uuid.UUID) {
		__antithesis_instrumentation__.Notify(120840)
		r.addToTSCacheChecked(ctx, st, ba, br, pErr, start, end, ts, txnID)
	}
	__antithesis_instrumentation__.Notify(120837)

	ts := ba.Timestamp
	if br != nil {
		__antithesis_instrumentation__.Notify(120841)
		ts = br.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(120842)
	}
	__antithesis_instrumentation__.Notify(120838)
	var txnID uuid.UUID
	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(120843)
		txnID = ba.Txn.ID
	} else {
		__antithesis_instrumentation__.Notify(120844)
	}
	__antithesis_instrumentation__.Notify(120839)
	for i, union := range ba.Requests {
		__antithesis_instrumentation__.Notify(120845)
		args := union.GetInner()
		if !roachpb.UpdatesTimestampCache(args) {
			__antithesis_instrumentation__.Notify(120848)
			continue
		} else {
			__antithesis_instrumentation__.Notify(120849)
		}
		__antithesis_instrumentation__.Notify(120846)

		if pErr != nil {
			__antithesis_instrumentation__.Notify(120850)
			if index := pErr.Index; !roachpb.UpdatesTimestampCacheOnError(args) || func() bool {
				__antithesis_instrumentation__.Notify(120851)
				return index == nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(120852)
				return int32(i) != index.Index == true
			}() == true {
				__antithesis_instrumentation__.Notify(120853)
				continue
			} else {
				__antithesis_instrumentation__.Notify(120854)
			}
		} else {
			__antithesis_instrumentation__.Notify(120855)
		}
		__antithesis_instrumentation__.Notify(120847)
		header := args.Header()
		start, end := header.Key, header.EndKey
		switch t := args.(type) {
		case *roachpb.EndTxnRequest:
			__antithesis_instrumentation__.Notify(120856)

			key := transactionTombstoneMarker(start, txnID)
			addToTSCache(key, nil, ts, txnID)
		case *roachpb.HeartbeatTxnRequest:
			__antithesis_instrumentation__.Notify(120857)

			key := transactionTombstoneMarker(start, txnID)
			addToTSCache(key, nil, ts, txnID)
		case *roachpb.RecoverTxnRequest:
			__antithesis_instrumentation__.Notify(120858)

			recovered := br.Responses[i].GetInner().(*roachpb.RecoverTxnResponse).RecoveredTxn
			if recovered.Status.IsFinalized() {
				__antithesis_instrumentation__.Notify(120871)
				key := transactionTombstoneMarker(start, recovered.ID)
				addToTSCache(key, nil, ts, recovered.ID)
			} else {
				__antithesis_instrumentation__.Notify(120872)
			}
		case *roachpb.PushTxnRequest:
			__antithesis_instrumentation__.Notify(120859)

			pushee := br.Responses[i].GetInner().(*roachpb.PushTxnResponse).PusheeTxn

			var tombstone bool
			switch pushee.Status {
			case roachpb.PENDING:
				__antithesis_instrumentation__.Notify(120873)
				tombstone = false
			case roachpb.ABORTED:
				__antithesis_instrumentation__.Notify(120874)
				tombstone = true
			case roachpb.STAGING:
				__antithesis_instrumentation__.Notify(120875)

				continue
			case roachpb.COMMITTED:
				__antithesis_instrumentation__.Notify(120876)

				continue
			default:
				__antithesis_instrumentation__.Notify(120877)
			}
			__antithesis_instrumentation__.Notify(120860)

			var key roachpb.Key
			var pushTS hlc.Timestamp
			if tombstone {
				__antithesis_instrumentation__.Notify(120878)

				key = transactionTombstoneMarker(start, pushee.ID)
				pushTS = pushee.MinTimestamp
			} else {
				__antithesis_instrumentation__.Notify(120879)
				key = transactionPushMarker(start, pushee.ID)
				pushTS = pushee.WriteTimestamp
			}
			__antithesis_instrumentation__.Notify(120861)
			addToTSCache(key, nil, pushTS, t.PusherTxn.ID)
		case *roachpb.ConditionalPutRequest:
			__antithesis_instrumentation__.Notify(120862)

			if _, ok := pErr.GetDetail().(*roachpb.ConditionFailedError); ok {
				__antithesis_instrumentation__.Notify(120880)
				addToTSCache(start, end, ts, txnID)
			} else {
				__antithesis_instrumentation__.Notify(120881)
			}
		case *roachpb.InitPutRequest:
			__antithesis_instrumentation__.Notify(120863)

			if _, ok := pErr.GetDetail().(*roachpb.ConditionFailedError); ok {
				__antithesis_instrumentation__.Notify(120882)
				addToTSCache(start, end, ts, txnID)
			} else {
				__antithesis_instrumentation__.Notify(120883)
			}
		case *roachpb.ScanRequest:
			__antithesis_instrumentation__.Notify(120864)
			resp := br.Responses[i].GetInner().(*roachpb.ScanResponse)
			if resp.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(120884)

				end = resp.ResumeSpan.Key
			} else {
				__antithesis_instrumentation__.Notify(120885)
			}
			__antithesis_instrumentation__.Notify(120865)
			addToTSCache(start, end, ts, txnID)
		case *roachpb.ReverseScanRequest:
			__antithesis_instrumentation__.Notify(120866)
			resp := br.Responses[i].GetInner().(*roachpb.ReverseScanResponse)
			if resp.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(120886)

				start = resp.ResumeSpan.EndKey
			} else {
				__antithesis_instrumentation__.Notify(120887)
			}
			__antithesis_instrumentation__.Notify(120867)
			addToTSCache(start, end, ts, txnID)
		case *roachpb.QueryIntentRequest:
			__antithesis_instrumentation__.Notify(120868)
			missing := false
			if pErr != nil {
				__antithesis_instrumentation__.Notify(120888)
				switch t := pErr.GetDetail().(type) {
				case *roachpb.IntentMissingError:
					__antithesis_instrumentation__.Notify(120889)
					missing = true
				case *roachpb.TransactionRetryError:
					__antithesis_instrumentation__.Notify(120890)

					missing = t.Reason == roachpb.RETRY_SERIALIZABLE
				}
			} else {
				__antithesis_instrumentation__.Notify(120891)
				missing = !br.Responses[i].GetInner().(*roachpb.QueryIntentResponse).FoundIntent
			}
			__antithesis_instrumentation__.Notify(120869)
			if missing {
				__antithesis_instrumentation__.Notify(120892)

				addToTSCache(start, end, t.Txn.WriteTimestamp, uuid.UUID{})
			} else {
				__antithesis_instrumentation__.Notify(120893)
			}
		default:
			__antithesis_instrumentation__.Notify(120870)
			addToTSCache(start, end, ts, txnID)
		}
	}
}

var batchesPushedDueToClosedTimestamp telemetry.Counter

func init() {
	batchesPushedDueToClosedTimestamp = telemetry.GetCounter("kv.closed_timestamp.txns_pushed")
}

func (r *Replica) applyTimestampCache(
	ctx context.Context, ba *roachpb.BatchRequest, minReadTS hlc.Timestamp,
) bool {
	__antithesis_instrumentation__.Notify(120894)

	var bumpedDueToMinReadTS bool
	var bumped bool
	var conflictingTxn uuid.UUID

	for _, union := range ba.Requests {
		__antithesis_instrumentation__.Notify(120897)
		args := union.GetInner()
		if roachpb.AppliesTimestampCache(args) {
			__antithesis_instrumentation__.Notify(120898)
			header := args.Header()

			rTS, rTxnID := r.store.tsCache.GetMax(header.Key, header.EndKey)
			var forwardedToMinReadTS bool
			if rTS.Forward(minReadTS) {
				__antithesis_instrumentation__.Notify(120902)
				forwardedToMinReadTS = true
				rTxnID = uuid.Nil
			} else {
				__antithesis_instrumentation__.Notify(120903)
			}
			__antithesis_instrumentation__.Notify(120899)
			nextRTS := rTS.Next()
			var bumpedCurReq bool
			if ba.Txn != nil {
				__antithesis_instrumentation__.Notify(120904)
				if ba.Txn.ID != rTxnID {
					__antithesis_instrumentation__.Notify(120905)
					if ba.Txn.WriteTimestamp.Less(nextRTS) {
						__antithesis_instrumentation__.Notify(120906)
						txn := ba.Txn.Clone()
						bumpedCurReq = txn.WriteTimestamp.Forward(nextRTS)
						ba.Txn = txn
					} else {
						__antithesis_instrumentation__.Notify(120907)
					}
				} else {
					__antithesis_instrumentation__.Notify(120908)
				}
			} else {
				__antithesis_instrumentation__.Notify(120909)
				bumpedCurReq = ba.Timestamp.Forward(nextRTS)
			}
			__antithesis_instrumentation__.Notify(120900)
			if bumpedCurReq && func() bool {
				__antithesis_instrumentation__.Notify(120910)
				return (rTxnID != uuid.Nil) == true
			}() == true {
				__antithesis_instrumentation__.Notify(120911)
				conflictingTxn = rTxnID
			} else {
				__antithesis_instrumentation__.Notify(120912)
			}
			__antithesis_instrumentation__.Notify(120901)

			bumpedDueToMinReadTS = (!bumpedCurReq && func() bool {
				__antithesis_instrumentation__.Notify(120913)
				return bumpedDueToMinReadTS == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(120914)
				return (bumpedCurReq && func() bool {
					__antithesis_instrumentation__.Notify(120915)
					return forwardedToMinReadTS == true
				}() == true) == true
			}() == true
			bumped, bumpedCurReq = bumped || func() bool {
				__antithesis_instrumentation__.Notify(120916)
				return bumpedCurReq == true
			}() == true, false
		} else {
			__antithesis_instrumentation__.Notify(120917)
		}
	}
	__antithesis_instrumentation__.Notify(120895)
	if bumped {
		__antithesis_instrumentation__.Notify(120918)
		bumpedTS := ba.Timestamp
		if ba.Txn != nil {
			__antithesis_instrumentation__.Notify(120920)
			bumpedTS = ba.Txn.WriteTimestamp
		} else {
			__antithesis_instrumentation__.Notify(120921)
		}
		__antithesis_instrumentation__.Notify(120919)

		if bumpedDueToMinReadTS {
			__antithesis_instrumentation__.Notify(120922)
			telemetry.Inc(batchesPushedDueToClosedTimestamp)
			log.VEventf(ctx, 2, "bumped write timestamp due to closed ts: %s", minReadTS)
		} else {
			__antithesis_instrumentation__.Notify(120923)
			conflictMsg := "conflicting txn unknown"
			if conflictingTxn != uuid.Nil {
				__antithesis_instrumentation__.Notify(120925)
				conflictMsg = "conflicting txn: " + conflictingTxn.Short()
			} else {
				__antithesis_instrumentation__.Notify(120926)
			}
			__antithesis_instrumentation__.Notify(120924)
			log.VEventf(ctx, 2, "bumped write timestamp to %s; %s", bumpedTS, redact.Safe(conflictMsg))
		}
	} else {
		__antithesis_instrumentation__.Notify(120927)
	}
	__antithesis_instrumentation__.Notify(120896)
	return bumped
}

func (r *Replica) CanCreateTxnRecord(
	ctx context.Context, txnID uuid.UUID, txnKey []byte, txnMinTS hlc.Timestamp,
) (ok bool, minCommitTS hlc.Timestamp, reason roachpb.TransactionAbortedReason) {
	__antithesis_instrumentation__.Notify(120928)

	tombstoneKey := transactionTombstoneMarker(txnKey, txnID)
	pushKey := transactionPushMarker(txnKey, txnID)

	minCommitTS, _ = r.store.tsCache.GetMax(pushKey, nil)

	tombstoneTimestamp, tombstoneTxnID := r.store.tsCache.GetMax(tombstoneKey, nil)

	if txnMinTS.LessEq(tombstoneTimestamp) {
		__antithesis_instrumentation__.Notify(120930)
		switch tombstoneTxnID {
		case txnID:
			__antithesis_instrumentation__.Notify(120931)

			return false, hlc.Timestamp{},
				roachpb.ABORT_REASON_RECORD_ALREADY_WRITTEN_POSSIBLE_REPLAY
		case uuid.Nil:
			__antithesis_instrumentation__.Notify(120932)
			lease, _ := r.GetLease()

			if tombstoneTimestamp == lease.Start.ToTimestamp() {
				__antithesis_instrumentation__.Notify(120935)
				return false, hlc.Timestamp{}, roachpb.ABORT_REASON_NEW_LEASE_PREVENTS_TXN
			} else {
				__antithesis_instrumentation__.Notify(120936)
			}
			__antithesis_instrumentation__.Notify(120933)
			return false, hlc.Timestamp{}, roachpb.ABORT_REASON_TIMESTAMP_CACHE_REJECTED
		default:
			__antithesis_instrumentation__.Notify(120934)

			return false, hlc.Timestamp{}, roachpb.ABORT_REASON_ABORTED_RECORD_FOUND
		}
	} else {
		__antithesis_instrumentation__.Notify(120937)
	}
	__antithesis_instrumentation__.Notify(120929)
	return true, minCommitTS, 0
}

func transactionTombstoneMarker(key roachpb.Key, txnID uuid.UUID) roachpb.Key {
	__antithesis_instrumentation__.Notify(120938)
	return append(keys.TransactionKey(key, txnID), []byte("-tmbs")...)
}

func transactionPushMarker(key roachpb.Key, txnID uuid.UUID) roachpb.Key {
	__antithesis_instrumentation__.Notify(120939)
	return append(keys.TransactionKey(key, txnID), []byte("-push")...)
}

func (r *Replica) GetCurrentReadSummary(ctx context.Context) rspb.ReadSummary {
	__antithesis_instrumentation__.Notify(120940)
	sum := collectReadSummaryFromTimestampCache(r.store.tsCache, r.Desc())

	closedTS := r.GetClosedTimestamp(ctx)
	sum.Merge(rspb.FromTimestamp(closedTS))
	return sum
}

func collectReadSummaryFromTimestampCache(
	tc tscache.Cache, desc *roachpb.RangeDescriptor,
) rspb.ReadSummary {
	__antithesis_instrumentation__.Notify(120941)
	var s rspb.ReadSummary
	s.Local.LowWater, _ = tc.GetMax(
		keys.MakeRangeKeyPrefix(desc.StartKey),
		keys.MakeRangeKeyPrefix(desc.EndKey),
	)
	userKeys := desc.KeySpan()
	s.Global.LowWater, _ = tc.GetMax(
		userKeys.Key.AsRawKey(),
		userKeys.EndKey.AsRawKey(),
	)

	return s
}

func applyReadSummaryToTimestampCache(
	tc tscache.Cache, desc *roachpb.RangeDescriptor, s rspb.ReadSummary,
) {
	__antithesis_instrumentation__.Notify(120942)
	tc.Add(
		keys.MakeRangeKeyPrefix(desc.StartKey),
		keys.MakeRangeKeyPrefix(desc.EndKey),
		s.Local.LowWater,
		uuid.Nil,
	)
	userKeys := desc.KeySpan()
	tc.Add(
		userKeys.Key.AsRawKey(),
		userKeys.EndKey.AsRawKey(),
		s.Global.LowWater,
		uuid.Nil,
	)
}
