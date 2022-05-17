package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

func Validate(steps []Step, kvs *Engine) []error {
	__antithesis_instrumentation__.Notify(93473)
	v, err := makeValidator(kvs)
	if err != nil {
		__antithesis_instrumentation__.Notify(93480)
		return []error{err}
	} else {
		__antithesis_instrumentation__.Notify(93481)
	}
	__antithesis_instrumentation__.Notify(93474)

	sort.Slice(steps, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(93482)
		return steps[i].After.Less(steps[j].After)
	})
	__antithesis_instrumentation__.Notify(93475)
	for _, s := range steps {
		__antithesis_instrumentation__.Notify(93483)
		v.processOp(nil, s.Op)
	}
	__antithesis_instrumentation__.Notify(93476)

	var extraKVs []observedOp
	for _, kv := range v.kvByValue {
		__antithesis_instrumentation__.Notify(93484)
		kv := &observedWrite{
			Key:          kv.Key.Key,
			Value:        roachpb.Value{RawBytes: kv.Value},
			Timestamp:    kv.Key.Timestamp,
			Materialized: true,
		}
		extraKVs = append(extraKVs, kv)
	}
	__antithesis_instrumentation__.Notify(93477)
	for key, tombstones := range v.tombstonesForKey {
		__antithesis_instrumentation__.Notify(93485)
		numExtraWrites := len(tombstones) - v.committedDeletesForKey[key]
		for i := 0; i < numExtraWrites; i++ {
			__antithesis_instrumentation__.Notify(93486)
			kv := &observedWrite{
				Key:   []byte(key),
				Value: roachpb.Value{},

				Materialized: true,
			}
			extraKVs = append(extraKVs, kv)
		}
	}
	__antithesis_instrumentation__.Notify(93478)
	if len(extraKVs) > 0 {
		__antithesis_instrumentation__.Notify(93487)
		err := errors.Errorf(`extra writes: %s`, printObserved(extraKVs...))
		v.failures = append(v.failures, err)
	} else {
		__antithesis_instrumentation__.Notify(93488)
	}
	__antithesis_instrumentation__.Notify(93479)

	return v.failures
}

type timeSpan struct {
	Start, End hlc.Timestamp
}

func (ts timeSpan) Intersect(o timeSpan) timeSpan {
	__antithesis_instrumentation__.Notify(93489)
	i := ts
	if i.Start.Less(o.Start) {
		__antithesis_instrumentation__.Notify(93492)
		i.Start = o.Start
	} else {
		__antithesis_instrumentation__.Notify(93493)
	}
	__antithesis_instrumentation__.Notify(93490)
	if o.End.Less(i.End) {
		__antithesis_instrumentation__.Notify(93494)
		i.End = o.End
	} else {
		__antithesis_instrumentation__.Notify(93495)
	}
	__antithesis_instrumentation__.Notify(93491)
	return i
}

func (ts timeSpan) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(93496)
	return !ts.Start.Less(ts.End)
}

func (ts timeSpan) String() string {
	__antithesis_instrumentation__.Notify(93497)
	var start string
	if ts.Start == hlc.MinTimestamp {
		__antithesis_instrumentation__.Notify(93500)
		start = `<min>`
	} else {
		__antithesis_instrumentation__.Notify(93501)
		start = ts.Start.String()
	}
	__antithesis_instrumentation__.Notify(93498)
	var end string
	if ts.End == hlc.MaxTimestamp {
		__antithesis_instrumentation__.Notify(93502)
		end = `<max>`
	} else {
		__antithesis_instrumentation__.Notify(93503)
		end = ts.End.String()
	}
	__antithesis_instrumentation__.Notify(93499)
	return fmt.Sprintf(`[%s, %s)`, start, end)
}

type disjointTimeSpans []timeSpan

func (ts disjointTimeSpans) validIntersections(
	keyOpValidTimeSpans disjointTimeSpans,
) disjointTimeSpans {
	__antithesis_instrumentation__.Notify(93504)
	var newValidSpans disjointTimeSpans
	for _, existingValidSpan := range ts {
		__antithesis_instrumentation__.Notify(93506)
		for _, opValidSpan := range keyOpValidTimeSpans {
			__antithesis_instrumentation__.Notify(93507)
			intersection := existingValidSpan.Intersect(opValidSpan)
			if !intersection.IsEmpty() {
				__antithesis_instrumentation__.Notify(93508)
				newValidSpans = append(newValidSpans, intersection)
			} else {
				__antithesis_instrumentation__.Notify(93509)
			}
		}
	}
	__antithesis_instrumentation__.Notify(93505)
	return newValidSpans
}

type multiKeyTimeSpan struct {
	Keys []disjointTimeSpans
	Gaps disjointTimeSpans
}

func (mts multiKeyTimeSpan) Combined() disjointTimeSpans {
	__antithesis_instrumentation__.Notify(93510)
	validPossibilities := mts.Gaps
	for _, validKey := range mts.Keys {
		__antithesis_instrumentation__.Notify(93512)
		validPossibilities = validPossibilities.validIntersections(validKey)
	}
	__antithesis_instrumentation__.Notify(93511)
	return validPossibilities
}

func (mts multiKeyTimeSpan) String() string {
	__antithesis_instrumentation__.Notify(93513)
	var buf strings.Builder
	buf.WriteByte('{')
	for i, timeSpans := range mts.Keys {
		__antithesis_instrumentation__.Notify(93516)
		fmt.Fprintf(&buf, "%d:", i)
		for tsIdx, ts := range timeSpans {
			__antithesis_instrumentation__.Notify(93518)
			if tsIdx != 0 {
				__antithesis_instrumentation__.Notify(93520)
				fmt.Fprintf(&buf, ",")
			} else {
				__antithesis_instrumentation__.Notify(93521)
			}
			__antithesis_instrumentation__.Notify(93519)
			fmt.Fprintf(&buf, "%s", ts)
		}
		__antithesis_instrumentation__.Notify(93517)
		fmt.Fprintf(&buf, ", ")
	}
	__antithesis_instrumentation__.Notify(93514)
	fmt.Fprintf(&buf, "gap:")
	for idx, gapSpan := range mts.Gaps {
		__antithesis_instrumentation__.Notify(93522)
		if idx != 0 {
			__antithesis_instrumentation__.Notify(93524)
			fmt.Fprintf(&buf, ",")
		} else {
			__antithesis_instrumentation__.Notify(93525)
		}
		__antithesis_instrumentation__.Notify(93523)
		fmt.Fprintf(&buf, "%s", gapSpan)
	}
	__antithesis_instrumentation__.Notify(93515)
	fmt.Fprintf(&buf, "}")
	return buf.String()
}

type observedOp interface {
	observedMarker()
}

type observedWrite struct {
	Key   roachpb.Key
	Value roachpb.Value

	Timestamp     hlc.Timestamp
	Materialized  bool
	IsDeleteRange bool
}

func (*observedWrite) observedMarker() { __antithesis_instrumentation__.Notify(93526) }

func (o *observedWrite) isDelete() bool {
	__antithesis_instrumentation__.Notify(93527)
	return !o.Value.IsPresent()
}

type observedRead struct {
	Key        roachpb.Key
	Value      roachpb.Value
	ValidTimes disjointTimeSpans
}

func (*observedRead) observedMarker() { __antithesis_instrumentation__.Notify(93528) }

type observedScan struct {
	Span          roachpb.Span
	IsDeleteRange bool
	Reverse       bool
	KVs           []roachpb.KeyValue
	Valid         multiKeyTimeSpan
}

func (*observedScan) observedMarker() { __antithesis_instrumentation__.Notify(93529) }

type validator struct {
	kvs              *Engine
	observedOpsByTxn map[string][]observedOp

	kvByValue map[string]storage.MVCCKeyValue

	tombstonesForKey       map[string]map[hlc.Timestamp]bool
	committedDeletesForKey map[string]int

	failures []error
}

func makeValidator(kvs *Engine) (*validator, error) {
	__antithesis_instrumentation__.Notify(93530)
	kvByValue := make(map[string]storage.MVCCKeyValue)
	tombstonesForKey := make(map[string]map[hlc.Timestamp]bool)
	var err error
	kvs.Iterate(func(key storage.MVCCKey, value []byte, iterErr error) {
		__antithesis_instrumentation__.Notify(93533)
		if iterErr != nil {
			__antithesis_instrumentation__.Notify(93535)
			err = errors.CombineErrors(err, iterErr)
			return
		} else {
			__antithesis_instrumentation__.Notify(93536)
		}
		__antithesis_instrumentation__.Notify(93534)
		v := roachpb.Value{RawBytes: value}
		if v.GetTag() != roachpb.ValueType_UNKNOWN {
			__antithesis_instrumentation__.Notify(93537)
			valueStr := mustGetStringValue(value)
			if existing, ok := kvByValue[valueStr]; ok {
				__antithesis_instrumentation__.Notify(93539)

				panic(errors.AssertionFailedf(
					`invariant violation: value %s was written by two operations %s and %s`,
					valueStr, existing.Key, key))
			} else {
				__antithesis_instrumentation__.Notify(93540)
			}
			__antithesis_instrumentation__.Notify(93538)

			kvByValue[valueStr] = storage.MVCCKeyValue{Key: key, Value: value}
		} else {
			__antithesis_instrumentation__.Notify(93541)
			if len(value) == 0 {
				__antithesis_instrumentation__.Notify(93542)
				rawKey := string(key.Key)
				if _, ok := tombstonesForKey[rawKey]; !ok {
					__antithesis_instrumentation__.Notify(93544)
					tombstonesForKey[rawKey] = make(map[hlc.Timestamp]bool)
				} else {
					__antithesis_instrumentation__.Notify(93545)
				}
				__antithesis_instrumentation__.Notify(93543)
				tombstonesForKey[rawKey][key.Timestamp] = false
			} else {
				__antithesis_instrumentation__.Notify(93546)
			}
		}
	})
	__antithesis_instrumentation__.Notify(93531)
	if err != nil {
		__antithesis_instrumentation__.Notify(93547)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(93548)
	}
	__antithesis_instrumentation__.Notify(93532)

	return &validator{
		kvs:                    kvs,
		kvByValue:              kvByValue,
		tombstonesForKey:       tombstonesForKey,
		committedDeletesForKey: make(map[string]int),
		observedOpsByTxn:       make(map[string][]observedOp),
	}, nil
}

func (v *validator) getDeleteForKey(key string, txn *roachpb.Transaction) (storage.MVCCKey, bool) {
	__antithesis_instrumentation__.Notify(93549)
	if txn == nil {
		__antithesis_instrumentation__.Notify(93552)
		panic(errors.AssertionFailedf(`transaction required to look up delete for key: %v`, key))
	} else {
		__antithesis_instrumentation__.Notify(93553)
	}
	__antithesis_instrumentation__.Notify(93550)

	if used, ok := v.tombstonesForKey[key][txn.TxnMeta.WriteTimestamp]; !used && func() bool {
		__antithesis_instrumentation__.Notify(93554)
		return ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(93555)
		v.tombstonesForKey[key][txn.TxnMeta.WriteTimestamp] = true
		return storage.MVCCKey{Key: []byte(key), Timestamp: txn.TxnMeta.WriteTimestamp}, true
	} else {
		__antithesis_instrumentation__.Notify(93556)
	}
	__antithesis_instrumentation__.Notify(93551)

	return storage.MVCCKey{}, false
}

func (v *validator) processOp(txnID *string, op Operation) {
	__antithesis_instrumentation__.Notify(93557)
	switch t := op.GetValue().(type) {
	case *GetOperation:
		__antithesis_instrumentation__.Notify(93558)
		v.failIfError(op, t.Result)
		if txnID == nil {
			__antithesis_instrumentation__.Notify(93573)
			v.checkAtomic(`get`, t.Result, nil, op)
		} else {
			__antithesis_instrumentation__.Notify(93574)
			read := &observedRead{
				Key:   t.Key,
				Value: roachpb.Value{RawBytes: t.Result.Value},
			}
			v.observedOpsByTxn[*txnID] = append(v.observedOpsByTxn[*txnID], read)
		}
	case *PutOperation:
		__antithesis_instrumentation__.Notify(93559)
		if txnID == nil {
			__antithesis_instrumentation__.Notify(93575)
			v.checkAtomic(`put`, t.Result, nil, op)
		} else {
			__antithesis_instrumentation__.Notify(93576)

			kv, ok := v.kvByValue[string(t.Value)]
			delete(v.kvByValue, string(t.Value))
			write := &observedWrite{
				Key:          t.Key,
				Value:        roachpb.MakeValueFromBytes(t.Value),
				Materialized: ok,
			}
			if write.Materialized {
				__antithesis_instrumentation__.Notify(93578)
				write.Timestamp = kv.Key.Timestamp
			} else {
				__antithesis_instrumentation__.Notify(93579)
			}
			__antithesis_instrumentation__.Notify(93577)
			v.observedOpsByTxn[*txnID] = append(v.observedOpsByTxn[*txnID], write)
		}
	case *DeleteOperation:
		__antithesis_instrumentation__.Notify(93560)
		if txnID == nil {
			__antithesis_instrumentation__.Notify(93580)
			v.checkAtomic(`delete`, t.Result, nil, op)
		} else {
			__antithesis_instrumentation__.Notify(93581)

			write := &observedWrite{
				Key:   t.Key,
				Value: roachpb.Value{},
			}
			v.observedOpsByTxn[*txnID] = append(v.observedOpsByTxn[*txnID], write)
		}
	case *DeleteRangeOperation:
		__antithesis_instrumentation__.Notify(93561)
		if txnID == nil {
			__antithesis_instrumentation__.Notify(93582)
			v.checkAtomic(`deleteRange`, t.Result, nil, op)
		} else {
			__antithesis_instrumentation__.Notify(93583)

			scan := &observedScan{
				Span: roachpb.Span{
					Key:    t.Key,
					EndKey: t.EndKey,
				},
				IsDeleteRange: true,
				KVs:           make([]roachpb.KeyValue, len(t.Result.Keys)),
			}
			deleteOps := make([]observedOp, len(t.Result.Keys))
			for i, key := range t.Result.Keys {
				__antithesis_instrumentation__.Notify(93585)
				scan.KVs[i] = roachpb.KeyValue{
					Key:   key,
					Value: roachpb.Value{},
				}
				write := &observedWrite{
					Key:           key,
					Value:         roachpb.Value{},
					IsDeleteRange: true,
				}
				deleteOps[i] = write
			}
			__antithesis_instrumentation__.Notify(93584)
			v.observedOpsByTxn[*txnID] = append(v.observedOpsByTxn[*txnID], scan)
			v.observedOpsByTxn[*txnID] = append(v.observedOpsByTxn[*txnID], deleteOps...)
		}
	case *ScanOperation:
		__antithesis_instrumentation__.Notify(93562)
		v.failIfError(op, t.Result)
		if txnID == nil {
			__antithesis_instrumentation__.Notify(93586)
			atomicScanType := `scan`
			if t.Reverse {
				__antithesis_instrumentation__.Notify(93588)
				atomicScanType = `reverse scan`
			} else {
				__antithesis_instrumentation__.Notify(93589)
			}
			__antithesis_instrumentation__.Notify(93587)
			v.checkAtomic(atomicScanType, t.Result, nil, op)
		} else {
			__antithesis_instrumentation__.Notify(93590)
			scan := &observedScan{
				Span: roachpb.Span{
					Key:    t.Key,
					EndKey: t.EndKey,
				},
				KVs:     make([]roachpb.KeyValue, len(t.Result.Values)),
				Reverse: t.Reverse,
			}
			for i, kv := range t.Result.Values {
				__antithesis_instrumentation__.Notify(93592)
				scan.KVs[i] = roachpb.KeyValue{
					Key:   kv.Key,
					Value: roachpb.Value{RawBytes: kv.Value},
				}
			}
			__antithesis_instrumentation__.Notify(93591)
			v.observedOpsByTxn[*txnID] = append(v.observedOpsByTxn[*txnID], scan)
		}
	case *SplitOperation:
		__antithesis_instrumentation__.Notify(93563)
		v.failIfError(op, t.Result)
	case *MergeOperation:
		__antithesis_instrumentation__.Notify(93564)
		if resultIsErrorStr(t.Result, `cannot merge final range`) {
			__antithesis_instrumentation__.Notify(93593)

		} else {
			__antithesis_instrumentation__.Notify(93594)
			if resultIsErrorStr(t.Result, `merge failed: unexpected value`) {
				__antithesis_instrumentation__.Notify(93595)

			} else {
				__antithesis_instrumentation__.Notify(93596)
				if resultIsErrorStr(t.Result, `merge failed: cannot merge ranges when (rhs)|(lhs) is in a joint state or has learners`) {
					__antithesis_instrumentation__.Notify(93597)

				} else {
					__antithesis_instrumentation__.Notify(93598)
					if resultIsErrorStr(t.Result, `merge failed: ranges not collocated`) {
						__antithesis_instrumentation__.Notify(93599)

					} else {
						__antithesis_instrumentation__.Notify(93600)
						if resultIsErrorStr(t.Result, `merge failed: waiting for all left-hand replicas to initialize`) {
							__antithesis_instrumentation__.Notify(93601)

						} else {
							__antithesis_instrumentation__.Notify(93602)
							if resultIsErrorStr(t.Result, `merge failed: waiting for all right-hand replicas to catch up`) {
								__antithesis_instrumentation__.Notify(93603)

							} else {
								__antithesis_instrumentation__.Notify(93604)
								if resultIsErrorStr(t.Result, `merge failed: non-deletion intent on local range descriptor`) {
									__antithesis_instrumentation__.Notify(93605)

								} else {
									__antithesis_instrumentation__.Notify(93606)
									if resultIsErrorStr(t.Result, `merge failed: range missing intent on its local descriptor`) {
										__antithesis_instrumentation__.Notify(93607)

									} else {
										__antithesis_instrumentation__.Notify(93608)
										if resultIsErrorStr(t.Result, `merge failed: RHS range bounds do not match`) {
											__antithesis_instrumentation__.Notify(93609)

										} else {
											__antithesis_instrumentation__.Notify(93610)
											v.failIfError(op, t.Result)
										}
									}
								}
							}
						}
					}
				}
			}
		}
	case *ChangeReplicasOperation:
		__antithesis_instrumentation__.Notify(93565)
		var ignore bool
		if err := errorFromResult(t.Result); err != nil {
			__antithesis_instrumentation__.Notify(93611)
			ignore = kvserver.IsRetriableReplicationChangeError(err) || func() bool {
				__antithesis_instrumentation__.Notify(93612)
				return kvserver.IsIllegalReplicationChangeError(err) == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(93613)
		}
		__antithesis_instrumentation__.Notify(93566)
		if !ignore {
			__antithesis_instrumentation__.Notify(93614)
			v.failIfError(op, t.Result)
		} else {
			__antithesis_instrumentation__.Notify(93615)
		}
	case *TransferLeaseOperation:
		__antithesis_instrumentation__.Notify(93567)
		if resultIsErrorStr(t.Result, `replica cannot hold lease`) {
			__antithesis_instrumentation__.Notify(93616)

		} else {
			__antithesis_instrumentation__.Notify(93617)
			if resultIsErrorStr(t.Result, `replica not found in RangeDescriptor`) {
				__antithesis_instrumentation__.Notify(93618)

			} else {
				__antithesis_instrumentation__.Notify(93619)
				if resultIsErrorStr(t.Result, `unable to find store \d+ in range`) {
					__antithesis_instrumentation__.Notify(93620)

				} else {
					__antithesis_instrumentation__.Notify(93621)
					if resultIsErrorStr(t.Result, `cannot transfer lease while merge in progress`) {
						__antithesis_instrumentation__.Notify(93622)

					} else {
						__antithesis_instrumentation__.Notify(93623)
						if resultIsError(t.Result, liveness.ErrRecordCacheMiss) {
							__antithesis_instrumentation__.Notify(93624)

						} else {
							__antithesis_instrumentation__.Notify(93625)
							if resultIsErrorStr(t.Result, liveness.ErrRecordCacheMiss.Error()) {
								__antithesis_instrumentation__.Notify(93626)

							} else {
								__antithesis_instrumentation__.Notify(93627)
								v.failIfError(op, t.Result)
							}
						}
					}
				}
			}
		}
	case *ChangeZoneOperation:
		__antithesis_instrumentation__.Notify(93568)
		v.failIfError(op, t.Result)
	case *BatchOperation:
		__antithesis_instrumentation__.Notify(93569)
		if !resultIsRetryable(t.Result) {
			__antithesis_instrumentation__.Notify(93628)
			v.failIfError(op, t.Result)
			if txnID == nil {
				__antithesis_instrumentation__.Notify(93629)
				v.checkAtomic(`batch`, t.Result, nil, t.Ops...)
			} else {
				__antithesis_instrumentation__.Notify(93630)
				for _, op := range t.Ops {
					__antithesis_instrumentation__.Notify(93631)
					v.processOp(txnID, op)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(93632)
		}
	case *ClosureTxnOperation:
		__antithesis_instrumentation__.Notify(93570)
		ops := t.Ops
		if t.CommitInBatch != nil {
			__antithesis_instrumentation__.Notify(93633)
			ops = append(ops, t.CommitInBatch.Ops...)
		} else {
			__antithesis_instrumentation__.Notify(93634)
		}
		__antithesis_instrumentation__.Notify(93571)
		v.checkAtomic(`txn`, t.Result, t.Txn, ops...)
	default:
		__antithesis_instrumentation__.Notify(93572)
		panic(errors.AssertionFailedf(`unknown operation type: %T %v`, t, t))
	}
}

func (v *validator) checkAtomic(
	atomicType string, result Result, optTxn *roachpb.Transaction, ops ...Operation,
) {
	__antithesis_instrumentation__.Notify(93635)
	fakeTxnID := uuid.MakeV4().String()
	for _, op := range ops {
		__antithesis_instrumentation__.Notify(93637)
		v.processOp(&fakeTxnID, op)
	}
	__antithesis_instrumentation__.Notify(93636)
	txnObservations := v.observedOpsByTxn[fakeTxnID]
	delete(v.observedOpsByTxn, fakeTxnID)

	if result.Type != ResultType_Error {
		__antithesis_instrumentation__.Notify(93638)
		v.checkCommittedTxn(`committed `+atomicType, txnObservations, optTxn)
	} else {
		__antithesis_instrumentation__.Notify(93639)
		if resultIsAmbiguous(result) {
			__antithesis_instrumentation__.Notify(93640)
			v.checkAmbiguousTxn(`ambiguous `+atomicType, txnObservations)
		} else {
			__antithesis_instrumentation__.Notify(93641)
			v.checkUncommittedTxn(`uncommitted `+atomicType, txnObservations)
		}
	}
}

func (v *validator) checkCommittedTxn(
	atomicType string, txnObservations []observedOp, optTxn *roachpb.Transaction,
) {
	__antithesis_instrumentation__.Notify(93642)

	batch := v.kvs.kvs.NewIndexedBatch()
	defer func() { __antithesis_instrumentation__.Notify(93648); _ = batch.Close() }()
	__antithesis_instrumentation__.Notify(93643)

	lastWriteIdxByKey := make(map[string]int, len(txnObservations))
	for idx := len(txnObservations) - 1; idx >= 0; idx-- {
		__antithesis_instrumentation__.Notify(93649)
		observation := txnObservations[idx]
		switch o := observation.(type) {
		case *observedWrite:
			__antithesis_instrumentation__.Notify(93650)
			if _, ok := lastWriteIdxByKey[string(o.Key)]; !ok {
				__antithesis_instrumentation__.Notify(93652)
				lastWriteIdxByKey[string(o.Key)] = idx

				if o.isDelete() {
					__antithesis_instrumentation__.Notify(93653)
					key := string(o.Key)
					v.committedDeletesForKey[key]++
					if optTxn == nil {
						__antithesis_instrumentation__.Notify(93654)

						o.Materialized = v.committedDeletesForKey[key] <= len(v.tombstonesForKey[key])
					} else {
						__antithesis_instrumentation__.Notify(93655)
						if storedDelete, ok := v.getDeleteForKey(key, optTxn); ok {
							__antithesis_instrumentation__.Notify(93656)
							o.Materialized = true
							o.Timestamp = storedDelete.Timestamp
						} else {
							__antithesis_instrumentation__.Notify(93657)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(93658)
				}
			} else {
				__antithesis_instrumentation__.Notify(93659)
			}
			__antithesis_instrumentation__.Notify(93651)
			if !o.Timestamp.IsEmpty() {
				__antithesis_instrumentation__.Notify(93660)
				mvccKey := storage.MVCCKey{Key: o.Key, Timestamp: o.Timestamp}
				if err := batch.Delete(storage.EncodeMVCCKey(mvccKey), nil); err != nil {
					__antithesis_instrumentation__.Notify(93661)
					panic(err)
				} else {
					__antithesis_instrumentation__.Notify(93662)
				}
			} else {
				__antithesis_instrumentation__.Notify(93663)
			}
		}
	}
	__antithesis_instrumentation__.Notify(93644)

	var failure string
	for idx, observation := range txnObservations {
		__antithesis_instrumentation__.Notify(93664)
		if failure != `` {
			__antithesis_instrumentation__.Notify(93666)
			break
		} else {
			__antithesis_instrumentation__.Notify(93667)
		}
		__antithesis_instrumentation__.Notify(93665)
		switch o := observation.(type) {
		case *observedWrite:
			__antithesis_instrumentation__.Notify(93668)
			var mvccKey storage.MVCCKey
			if lastWriteIdx := lastWriteIdxByKey[string(o.Key)]; idx == lastWriteIdx {
				__antithesis_instrumentation__.Notify(93676)

				mvccKey = storage.MVCCKey{Key: o.Key, Timestamp: o.Timestamp}
			} else {
				__antithesis_instrumentation__.Notify(93677)
				if o.Materialized {
					__antithesis_instrumentation__.Notify(93679)
					failure = `committed txn overwritten key had write`
				} else {
					__antithesis_instrumentation__.Notify(93680)
				}
				__antithesis_instrumentation__.Notify(93678)

				mvccKey = storage.MVCCKey{
					Key:       o.Key,
					Timestamp: txnObservations[lastWriteIdx].(*observedWrite).Timestamp,
				}
			}
			__antithesis_instrumentation__.Notify(93669)
			if err := batch.Set(storage.EncodeMVCCKey(mvccKey), o.Value.RawBytes, nil); err != nil {
				__antithesis_instrumentation__.Notify(93681)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(93682)
			}
		case *observedRead:
			__antithesis_instrumentation__.Notify(93670)
			o.ValidTimes = validReadTimes(batch, o.Key, o.Value.RawBytes, false)
		case *observedScan:
			__antithesis_instrumentation__.Notify(93671)

			for _, kv := range o.KVs {
				__antithesis_instrumentation__.Notify(93683)
				if !o.Span.ContainsKey(kv.Key) {
					__antithesis_instrumentation__.Notify(93684)
					opCode := "scan"
					if o.IsDeleteRange {
						__antithesis_instrumentation__.Notify(93686)
						opCode = "delete range"
					} else {
						__antithesis_instrumentation__.Notify(93687)
					}
					__antithesis_instrumentation__.Notify(93685)
					failure = fmt.Sprintf(`key %s outside %s bounds`, kv.Key, opCode)
					break
				} else {
					__antithesis_instrumentation__.Notify(93688)
				}
			}
			__antithesis_instrumentation__.Notify(93672)

			orderedKVs := sort.Interface(roachpb.KeyValueByKey(o.KVs))
			if o.Reverse {
				__antithesis_instrumentation__.Notify(93689)
				orderedKVs = sort.Reverse(orderedKVs)
			} else {
				__antithesis_instrumentation__.Notify(93690)
			}
			__antithesis_instrumentation__.Notify(93673)
			if !sort.IsSorted(orderedKVs) {
				__antithesis_instrumentation__.Notify(93691)
				failure = `scan result not ordered correctly`
			} else {
				__antithesis_instrumentation__.Notify(93692)
			}
			__antithesis_instrumentation__.Notify(93674)
			o.Valid = validScanTime(batch, o.Span, o.KVs, o.IsDeleteRange)
		default:
			__antithesis_instrumentation__.Notify(93675)
			panic(errors.AssertionFailedf(`unknown observedOp: %T %s`, observation, observation))
		}
	}
	__antithesis_instrumentation__.Notify(93645)

	validPossibilities := disjointTimeSpans{{Start: hlc.MinTimestamp, End: hlc.MaxTimestamp}}
	for idx, observation := range txnObservations {
		__antithesis_instrumentation__.Notify(93693)
		if failure != `` {
			__antithesis_instrumentation__.Notify(93697)
			break
		} else {
			__antithesis_instrumentation__.Notify(93698)
		}
		__antithesis_instrumentation__.Notify(93694)
		var opValid disjointTimeSpans
		switch o := observation.(type) {
		case *observedWrite:
			__antithesis_instrumentation__.Notify(93699)
			isLastWriteForKey := idx == lastWriteIdxByKey[string(o.Key)]
			if !isLastWriteForKey {
				__antithesis_instrumentation__.Notify(93706)
				continue
			} else {
				__antithesis_instrumentation__.Notify(93707)
			}
			__antithesis_instrumentation__.Notify(93700)
			if !o.Materialized {
				__antithesis_instrumentation__.Notify(93708)
				failure = atomicType + ` missing write`
				continue
			} else {
				__antithesis_instrumentation__.Notify(93709)
			}
			__antithesis_instrumentation__.Notify(93701)

			if o.isDelete() && func() bool {
				__antithesis_instrumentation__.Notify(93710)
				return len(txnObservations) == 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(93711)

				continue
			} else {
				__antithesis_instrumentation__.Notify(93712)
			}
			__antithesis_instrumentation__.Notify(93702)

			opValid = disjointTimeSpans{{Start: o.Timestamp, End: o.Timestamp.Next()}}
		case *observedRead:
			__antithesis_instrumentation__.Notify(93703)
			opValid = o.ValidTimes
		case *observedScan:
			__antithesis_instrumentation__.Notify(93704)
			opValid = o.Valid.Combined()
		default:
			__antithesis_instrumentation__.Notify(93705)
			panic(errors.AssertionFailedf(`unknown observedOp: %T %s`, observation, observation))
		}
		__antithesis_instrumentation__.Notify(93695)
		newValidSpans := validPossibilities.validIntersections(opValid)
		if len(newValidSpans) == 0 {
			__antithesis_instrumentation__.Notify(93713)
			failure = atomicType + ` non-atomic timestamps`
		} else {
			__antithesis_instrumentation__.Notify(93714)
		}
		__antithesis_instrumentation__.Notify(93696)
		validPossibilities = newValidSpans
	}
	__antithesis_instrumentation__.Notify(93646)

	for _, observation := range txnObservations {
		__antithesis_instrumentation__.Notify(93715)
		if failure != `` {
			__antithesis_instrumentation__.Notify(93717)
			break
		} else {
			__antithesis_instrumentation__.Notify(93718)
		}
		__antithesis_instrumentation__.Notify(93716)
		switch o := observation.(type) {
		case *observedWrite:
			__antithesis_instrumentation__.Notify(93719)
			if optTxn != nil && func() bool {
				__antithesis_instrumentation__.Notify(93720)
				return o.Materialized == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(93721)
				return optTxn.TxnMeta.WriteTimestamp != o.Timestamp == true
			}() == true {
				__antithesis_instrumentation__.Notify(93722)
				failure = fmt.Sprintf(`committed txn mismatched write timestamp %s`, optTxn.TxnMeta.WriteTimestamp)
			} else {
				__antithesis_instrumentation__.Notify(93723)
			}
		}
	}
	__antithesis_instrumentation__.Notify(93647)

	if failure != `` {
		__antithesis_instrumentation__.Notify(93724)
		err := errors.Errorf("%s: %s", failure, printObserved(txnObservations...))
		v.failures = append(v.failures, err)
	} else {
		__antithesis_instrumentation__.Notify(93725)
	}
}

func (v *validator) checkAmbiguousTxn(atomicType string, txnObservations []observedOp) {
	__antithesis_instrumentation__.Notify(93726)
	var somethingCommitted bool
	deletedKeysInTxn := make(map[string]int)
	var hadWrite bool
	for _, observation := range txnObservations {
		__antithesis_instrumentation__.Notify(93728)
		switch o := observation.(type) {
		case *observedWrite:
			__antithesis_instrumentation__.Notify(93729)
			hadWrite = true
			if o.Materialized {
				__antithesis_instrumentation__.Notify(93731)
				somethingCommitted = true
				break
			} else {
				__antithesis_instrumentation__.Notify(93732)
			}
			__antithesis_instrumentation__.Notify(93730)
			if o.isDelete() && func() bool {
				__antithesis_instrumentation__.Notify(93733)
				return len(v.tombstonesForKey[string(o.Key)]) > v.committedDeletesForKey[string(o.Key)] == true
			}() == true {
				__antithesis_instrumentation__.Notify(93734)
				deletedKeysInTxn[string(o.Key)]++
				break
			} else {
				__antithesis_instrumentation__.Notify(93735)
			}
		}
	}
	__antithesis_instrumentation__.Notify(93727)

	if len(deletedKeysInTxn) > 0 {
		__antithesis_instrumentation__.Notify(93736)

		err := errors.Errorf(
			`unable to validate delete operations in ambiguous transactions: %s`,
			printObserved(txnObservations...),
		)
		v.failures = append(v.failures, err)

		for key := range deletedKeysInTxn {
			__antithesis_instrumentation__.Notify(93737)

			v.committedDeletesForKey[key]++
		}
	} else {
		__antithesis_instrumentation__.Notify(93738)
		if !hadWrite {
			__antithesis_instrumentation__.Notify(93739)

			v.checkCommittedTxn(atomicType, txnObservations, nil)
		} else {
			__antithesis_instrumentation__.Notify(93740)
			if somethingCommitted {
				__antithesis_instrumentation__.Notify(93741)
				v.checkCommittedTxn(atomicType, txnObservations, nil)
			} else {
				__antithesis_instrumentation__.Notify(93742)
				v.checkUncommittedTxn(atomicType, txnObservations)
			}
		}
	}
}

func (v *validator) checkUncommittedTxn(atomicType string, txnObservations []observedOp) {
	__antithesis_instrumentation__.Notify(93743)
	var failure string
	for _, observed := range txnObservations {
		__antithesis_instrumentation__.Notify(93745)
		if failure != `` {
			__antithesis_instrumentation__.Notify(93747)
			break
		} else {
			__antithesis_instrumentation__.Notify(93748)
		}
		__antithesis_instrumentation__.Notify(93746)
		switch o := observed.(type) {
		case *observedWrite:
			__antithesis_instrumentation__.Notify(93749)
			if o.Materialized {
				__antithesis_instrumentation__.Notify(93753)
				failure = atomicType + ` had writes`
			} else {
				__antithesis_instrumentation__.Notify(93754)
			}

		case *observedRead:
			__antithesis_instrumentation__.Notify(93750)

		case *observedScan:
			__antithesis_instrumentation__.Notify(93751)

		default:
			__antithesis_instrumentation__.Notify(93752)
			panic(errors.AssertionFailedf(`unknown observedOp: %T %s`, observed, observed))
		}
	}
	__antithesis_instrumentation__.Notify(93744)

	if failure != `` {
		__antithesis_instrumentation__.Notify(93755)
		err := errors.Errorf("%s: %s", failure, printObserved(txnObservations...))
		v.failures = append(v.failures, err)
	} else {
		__antithesis_instrumentation__.Notify(93756)
	}
}

func (v *validator) failIfError(op Operation, r Result) {
	__antithesis_instrumentation__.Notify(93757)
	switch r.Type {
	case ResultType_Unknown:
		__antithesis_instrumentation__.Notify(93758)
		err := errors.AssertionFailedf(`unknown result %s`, op)
		v.failures = append(v.failures, err)
	case ResultType_Error:
		__antithesis_instrumentation__.Notify(93759)
		ctx := context.Background()
		err := errors.DecodeError(ctx, *r.Err)
		err = errors.Wrapf(err, `error applying %s`, op)
		v.failures = append(v.failures, err)
	default:
		__antithesis_instrumentation__.Notify(93760)
	}
}

func errorFromResult(r Result) error {
	__antithesis_instrumentation__.Notify(93761)
	if r.Type != ResultType_Error {
		__antithesis_instrumentation__.Notify(93763)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(93764)
	}
	__antithesis_instrumentation__.Notify(93762)
	ctx := context.Background()
	return errors.DecodeError(ctx, *r.Err)
}

func resultIsError(r Result, reference error) bool {
	__antithesis_instrumentation__.Notify(93765)
	return errors.Is(errorFromResult(r), reference)
}

func resultIsRetryable(r Result) bool {
	__antithesis_instrumentation__.Notify(93766)
	return errors.HasInterface(errorFromResult(r), (*roachpb.ClientVisibleRetryError)(nil))
}

func resultIsAmbiguous(r Result) bool {
	__antithesis_instrumentation__.Notify(93767)
	return errors.HasInterface(errorFromResult(r), (*roachpb.ClientVisibleAmbiguousError)(nil))
}

func resultIsErrorStr(r Result, msgRE string) bool {
	__antithesis_instrumentation__.Notify(93768)
	if err := errorFromResult(r); err != nil {
		__antithesis_instrumentation__.Notify(93770)
		return regexp.MustCompile(msgRE).MatchString(err.Error())
	} else {
		__antithesis_instrumentation__.Notify(93771)
	}
	__antithesis_instrumentation__.Notify(93769)
	return false
}

func mustGetStringValue(value []byte) string {
	__antithesis_instrumentation__.Notify(93772)
	if len(value) == 0 {
		__antithesis_instrumentation__.Notify(93775)
		return `<nil>`
	} else {
		__antithesis_instrumentation__.Notify(93776)
	}
	__antithesis_instrumentation__.Notify(93773)
	v, err := roachpb.Value{RawBytes: value}.GetBytes()
	if err != nil {
		__antithesis_instrumentation__.Notify(93777)
		panic(errors.Wrapf(err, "decoding %x", value))
	} else {
		__antithesis_instrumentation__.Notify(93778)
	}
	__antithesis_instrumentation__.Notify(93774)
	return string(v)
}

func validReadTimes(
	b *pebble.Batch, key roachpb.Key, value []byte, anyValueAccepted bool,
) disjointTimeSpans {
	__antithesis_instrumentation__.Notify(93779)
	var validTimes disjointTimeSpans
	end := hlc.MaxTimestamp

	iter := b.NewIter(nil)
	defer func() { __antithesis_instrumentation__.Notify(93783); _ = iter.Close() }()
	__antithesis_instrumentation__.Notify(93780)
	iter.SeekGE(storage.EncodeMVCCKey(storage.MVCCKey{Key: key}))
	for ; iter.Valid(); iter.Next() {
		__antithesis_instrumentation__.Notify(93784)
		mvccKey, err := storage.DecodeMVCCKey(iter.Key())
		if err != nil {
			__antithesis_instrumentation__.Notify(93788)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(93789)
		}
		__antithesis_instrumentation__.Notify(93785)
		if !mvccKey.Key.Equal(key) {
			__antithesis_instrumentation__.Notify(93790)
			break
		} else {
			__antithesis_instrumentation__.Notify(93791)
		}
		__antithesis_instrumentation__.Notify(93786)
		if (anyValueAccepted && func() bool {
			__antithesis_instrumentation__.Notify(93792)
			return len(iter.Value()) > 0 == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(93793)
			return (!anyValueAccepted && func() bool {
				__antithesis_instrumentation__.Notify(93794)
				return mustGetStringValue(iter.Value()) == mustGetStringValue(value) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(93795)
			validTimes = append(validTimes, timeSpan{Start: mvccKey.Timestamp, End: end})
		} else {
			__antithesis_instrumentation__.Notify(93796)
		}
		__antithesis_instrumentation__.Notify(93787)
		end = mvccKey.Timestamp
	}
	__antithesis_instrumentation__.Notify(93781)
	if !anyValueAccepted && func() bool {
		__antithesis_instrumentation__.Notify(93797)
		return len(value) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(93798)
		validTimes = append(disjointTimeSpans{{Start: hlc.MinTimestamp, End: end}}, validTimes...)
	} else {
		__antithesis_instrumentation__.Notify(93799)
	}
	__antithesis_instrumentation__.Notify(93782)

	return validTimes
}

func validScanTime(
	b *pebble.Batch, span roachpb.Span, kvs []roachpb.KeyValue, isDeleteRange bool,
) multiKeyTimeSpan {
	__antithesis_instrumentation__.Notify(93800)
	valid := multiKeyTimeSpan{
		Gaps: disjointTimeSpans{{Start: hlc.MinTimestamp, End: hlc.MaxTimestamp}},
	}

	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(93806)

		validTimes := validReadTimes(b, kv.Key, kv.Value.RawBytes, isDeleteRange)
		if !isDeleteRange && func() bool {
			__antithesis_instrumentation__.Notify(93809)
			return len(validTimes) > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(93810)
			panic(errors.AssertionFailedf(
				`invalid number of read time spans for a (key,non-nil-value) pair in scan results: %s->%s`,
				kv.Key, mustGetStringValue(kv.Value.RawBytes)))
		} else {
			__antithesis_instrumentation__.Notify(93811)
		}
		__antithesis_instrumentation__.Notify(93807)
		if len(validTimes) == 0 {
			__antithesis_instrumentation__.Notify(93812)
			validTimes = append(validTimes, timeSpan{})
		} else {
			__antithesis_instrumentation__.Notify(93813)
		}
		__antithesis_instrumentation__.Notify(93808)
		valid.Keys = append(valid.Keys, validTimes)
	}
	__antithesis_instrumentation__.Notify(93801)

	keys := make(map[string]struct{}, len(kvs))
	for _, kv := range kvs {
		__antithesis_instrumentation__.Notify(93814)
		keys[string(kv.Key)] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(93802)

	missingKeys := make(map[string]disjointTimeSpans)
	iter := b.NewIter(nil)
	defer func() { __antithesis_instrumentation__.Notify(93815); _ = iter.Close() }()
	__antithesis_instrumentation__.Notify(93803)
	iter.SeekGE(storage.EncodeMVCCKey(storage.MVCCKey{Key: span.Key}))
	for ; iter.Valid(); iter.Next() {
		__antithesis_instrumentation__.Notify(93816)
		mvccKey, err := storage.DecodeMVCCKey(iter.Key())
		if err != nil {
			__antithesis_instrumentation__.Notify(93820)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(93821)
		}
		__antithesis_instrumentation__.Notify(93817)
		if mvccKey.Key.Compare(span.EndKey) >= 0 {
			__antithesis_instrumentation__.Notify(93822)

			break
		} else {
			__antithesis_instrumentation__.Notify(93823)
		}
		__antithesis_instrumentation__.Notify(93818)
		if _, ok := keys[string(mvccKey.Key)]; ok {
			__antithesis_instrumentation__.Notify(93824)

			continue
		} else {
			__antithesis_instrumentation__.Notify(93825)
		}
		__antithesis_instrumentation__.Notify(93819)
		if _, ok := missingKeys[string(mvccKey.Key)]; !ok {
			__antithesis_instrumentation__.Notify(93826)

			missingKeys[string(mvccKey.Key)] = validReadTimes(b, mvccKey.Key, nil, false)
		} else {
			__antithesis_instrumentation__.Notify(93827)
		}
	}
	__antithesis_instrumentation__.Notify(93804)

	for _, nilValueReadTimes := range missingKeys {
		__antithesis_instrumentation__.Notify(93828)
		valid.Gaps = valid.Gaps.validIntersections(nilValueReadTimes)
	}
	__antithesis_instrumentation__.Notify(93805)

	return valid
}

func printObserved(observedOps ...observedOp) string {
	__antithesis_instrumentation__.Notify(93829)
	var buf strings.Builder
	for _, observed := range observedOps {
		__antithesis_instrumentation__.Notify(93831)
		if buf.Len() > 0 {
			__antithesis_instrumentation__.Notify(93833)
			buf.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(93834)
		}
		__antithesis_instrumentation__.Notify(93832)
		switch o := observed.(type) {
		case *observedWrite:
			__antithesis_instrumentation__.Notify(93835)
			opCode := "w"
			if o.isDelete() {
				__antithesis_instrumentation__.Notify(93846)
				if o.IsDeleteRange {
					__antithesis_instrumentation__.Notify(93847)
					opCode = "dr.d"
				} else {
					__antithesis_instrumentation__.Notify(93848)
					opCode = "d"
				}
			} else {
				__antithesis_instrumentation__.Notify(93849)
			}
			__antithesis_instrumentation__.Notify(93836)
			ts := `missing`
			if o.Materialized {
				__antithesis_instrumentation__.Notify(93850)
				if o.isDelete() && func() bool {
					__antithesis_instrumentation__.Notify(93851)
					return o.Timestamp.IsEmpty() == true
				}() == true {
					__antithesis_instrumentation__.Notify(93852)
					ts = `uncertain`
				} else {
					__antithesis_instrumentation__.Notify(93853)
					ts = o.Timestamp.String()
				}
			} else {
				__antithesis_instrumentation__.Notify(93854)
			}
			__antithesis_instrumentation__.Notify(93837)
			fmt.Fprintf(&buf, "[%s]%s:%s->%s",
				opCode, o.Key, ts, mustGetStringValue(o.Value.RawBytes))
		case *observedRead:
			__antithesis_instrumentation__.Notify(93838)
			fmt.Fprintf(&buf, "[r]%s:", o.Key)
			validTimes := o.ValidTimes
			if len(validTimes) == 0 {
				__antithesis_instrumentation__.Notify(93855)
				validTimes = append(validTimes, timeSpan{})
			} else {
				__antithesis_instrumentation__.Notify(93856)
			}
			__antithesis_instrumentation__.Notify(93839)
			for idx, validTime := range validTimes {
				__antithesis_instrumentation__.Notify(93857)
				if idx != 0 {
					__antithesis_instrumentation__.Notify(93859)
					fmt.Fprintf(&buf, ",")
				} else {
					__antithesis_instrumentation__.Notify(93860)
				}
				__antithesis_instrumentation__.Notify(93858)
				fmt.Fprintf(&buf, "%s", validTime)
			}
			__antithesis_instrumentation__.Notify(93840)
			fmt.Fprintf(&buf, "->%s", mustGetStringValue(o.Value.RawBytes))
		case *observedScan:
			__antithesis_instrumentation__.Notify(93841)
			opCode := "s"
			if o.IsDeleteRange {
				__antithesis_instrumentation__.Notify(93861)
				opCode = "dr.s"
			} else {
				__antithesis_instrumentation__.Notify(93862)
			}
			__antithesis_instrumentation__.Notify(93842)
			if o.Reverse {
				__antithesis_instrumentation__.Notify(93863)
				opCode = "rs"
			} else {
				__antithesis_instrumentation__.Notify(93864)
			}
			__antithesis_instrumentation__.Notify(93843)
			var kvs strings.Builder
			for i, kv := range o.KVs {
				__antithesis_instrumentation__.Notify(93865)
				if i > 0 {
					__antithesis_instrumentation__.Notify(93867)
					kvs.WriteString(`, `)
				} else {
					__antithesis_instrumentation__.Notify(93868)
				}
				__antithesis_instrumentation__.Notify(93866)
				kvs.WriteString(kv.Key.String())
				if !o.IsDeleteRange {
					__antithesis_instrumentation__.Notify(93869)
					kvs.WriteByte(':')
					kvs.WriteString(mustGetStringValue(kv.Value.RawBytes))
				} else {
					__antithesis_instrumentation__.Notify(93870)
				}
			}
			__antithesis_instrumentation__.Notify(93844)
			fmt.Fprintf(&buf, "[%s]%s:%s->[%s]",
				opCode, o.Span, o.Valid, kvs.String())
		default:
			__antithesis_instrumentation__.Notify(93845)
			panic(errors.AssertionFailedf(`unknown observedOp: %T %s`, observed, observed))
		}
	}
	__antithesis_instrumentation__.Notify(93830)
	return buf.String()
}
