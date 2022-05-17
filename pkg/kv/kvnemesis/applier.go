package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"google.golang.org/protobuf/proto"
)

type Applier struct {
	env *Env
	dbs []*kv.DB
	mu  struct {
		dbIdx int
		syncutil.Mutex
		txns map[string]*kv.Txn
	}
}

func MakeApplier(env *Env, dbs ...*kv.DB) *Applier {
	__antithesis_instrumentation__.Notify(90070)
	a := &Applier{
		env: env,
		dbs: dbs,
	}
	a.mu.txns = make(map[string]*kv.Txn)
	return a
}

func (a *Applier) Apply(ctx context.Context, step *Step) (trace tracing.Recording, retErr error) {
	__antithesis_instrumentation__.Notify(90071)
	var db *kv.DB
	db, step.DBID = a.getNextDBRoundRobin()

	step.Before = db.Clock().Now()
	recCtx, collectAndFinish := tracing.ContextWithRecordingSpan(ctx, db.Tracer, "txn step")
	defer func() {
		__antithesis_instrumentation__.Notify(90073)
		step.After = db.Clock().Now()
		if p := recover(); p != nil {
			__antithesis_instrumentation__.Notify(90075)
			retErr = errors.Errorf(`panic applying step %s: %v`, step, p)
		} else {
			__antithesis_instrumentation__.Notify(90076)
		}
		__antithesis_instrumentation__.Notify(90074)
		trace = collectAndFinish()
	}()
	__antithesis_instrumentation__.Notify(90072)
	applyOp(recCtx, a.env, db, &step.Op)
	return collectAndFinish(), nil
}

func (a *Applier) getNextDBRoundRobin() (*kv.DB, int32) {
	__antithesis_instrumentation__.Notify(90077)
	a.mu.Lock()
	dbIdx := a.mu.dbIdx
	a.mu.dbIdx = (a.mu.dbIdx + 1) % len(a.dbs)
	a.mu.Unlock()
	return a.dbs[dbIdx], int32(dbIdx)
}

func applyOp(ctx context.Context, env *Env, db *kv.DB, op *Operation) {
	__antithesis_instrumentation__.Notify(90078)
	switch o := op.GetValue().(type) {
	case *GetOperation,
		*PutOperation,
		*ScanOperation,
		*BatchOperation,
		*DeleteOperation,
		*DeleteRangeOperation:
		__antithesis_instrumentation__.Notify(90079)
		applyClientOp(ctx, db, op, false)
	case *SplitOperation:
		__antithesis_instrumentation__.Notify(90080)
		err := db.AdminSplit(ctx, o.Key, hlc.MaxTimestamp)
		o.Result = resultError(ctx, err)
	case *MergeOperation:
		__antithesis_instrumentation__.Notify(90081)
		err := db.AdminMerge(ctx, o.Key)
		o.Result = resultError(ctx, err)
	case *ChangeReplicasOperation:
		__antithesis_instrumentation__.Notify(90082)
		desc := getRangeDesc(ctx, o.Key, db)
		_, err := db.AdminChangeReplicas(ctx, o.Key, desc, o.Changes)

		o.Result = resultError(ctx, err)
	case *TransferLeaseOperation:
		__antithesis_instrumentation__.Notify(90083)
		err := db.AdminTransferLease(ctx, o.Key, o.Target)
		o.Result = resultError(ctx, err)
	case *ChangeZoneOperation:
		__antithesis_instrumentation__.Notify(90084)
		err := updateZoneConfigInEnv(ctx, env, o.Type)
		o.Result = resultError(ctx, err)
	case *ClosureTxnOperation:
		__antithesis_instrumentation__.Notify(90085)

		retryOnAbort := retry.StartWithCtx(ctx, retry.Options{
			InitialBackoff: 1 * time.Millisecond,
			MaxBackoff:     250 * time.Millisecond,
		})
		var savedTxn *kv.Txn
		txnErr := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(90088)
			if savedTxn != nil && func() bool {
				__antithesis_instrumentation__.Notify(90092)
				return txn.TestingCloneTxn().Epoch == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(90093)

				retryOnAbort.Next()
			} else {
				__antithesis_instrumentation__.Notify(90094)
			}
			__antithesis_instrumentation__.Notify(90089)
			savedTxn = txn
			for i := range o.Ops {
				__antithesis_instrumentation__.Notify(90095)
				op := &o.Ops[i]
				applyClientOp(ctx, txn, op, true)

				if r := op.Result(); r.Type == ResultType_Error {
					__antithesis_instrumentation__.Notify(90096)
					return errors.DecodeError(ctx, *r.Err)
				} else {
					__antithesis_instrumentation__.Notify(90097)
				}
			}
			__antithesis_instrumentation__.Notify(90090)
			if o.CommitInBatch != nil {
				__antithesis_instrumentation__.Notify(90098)
				b := txn.NewBatch()
				applyBatchOp(ctx, b, txn.CommitInBatch, o.CommitInBatch, true)

				if r := o.CommitInBatch.Result; r.Type == ResultType_Error {
					__antithesis_instrumentation__.Notify(90099)
					return errors.DecodeError(ctx, *r.Err)
				} else {
					__antithesis_instrumentation__.Notify(90100)
				}
			} else {
				__antithesis_instrumentation__.Notify(90101)
			}
			__antithesis_instrumentation__.Notify(90091)
			switch o.Type {
			case ClosureTxnType_Commit:
				__antithesis_instrumentation__.Notify(90102)
				return nil
			case ClosureTxnType_Rollback:
				__antithesis_instrumentation__.Notify(90103)
				return errors.New("rollback")
			default:
				__antithesis_instrumentation__.Notify(90104)
				panic(errors.AssertionFailedf(`unknown closure txn type: %s`, o.Type))
			}
		})
		__antithesis_instrumentation__.Notify(90086)
		o.Result = resultError(ctx, txnErr)
		if txnErr == nil {
			__antithesis_instrumentation__.Notify(90105)
			o.Txn = savedTxn.TestingCloneTxn()
		} else {
			__antithesis_instrumentation__.Notify(90106)
		}
	default:
		__antithesis_instrumentation__.Notify(90087)
		panic(errors.AssertionFailedf(`unknown operation type: %T %v`, o, o))
	}
}

type clientI interface {
	Get(context.Context, interface{}) (kv.KeyValue, error)
	GetForUpdate(context.Context, interface{}) (kv.KeyValue, error)
	Put(context.Context, interface{}, interface{}) error
	Scan(context.Context, interface{}, interface{}, int64) ([]kv.KeyValue, error)
	ScanForUpdate(context.Context, interface{}, interface{}, int64) ([]kv.KeyValue, error)
	ReverseScan(context.Context, interface{}, interface{}, int64) ([]kv.KeyValue, error)
	ReverseScanForUpdate(context.Context, interface{}, interface{}, int64) ([]kv.KeyValue, error)
	Del(context.Context, ...interface{}) error
	DelRange(context.Context, interface{}, interface{}, bool) ([]roachpb.Key, error)
	Run(context.Context, *kv.Batch) error
}

func applyClientOp(ctx context.Context, db clientI, op *Operation, inTxn bool) {
	__antithesis_instrumentation__.Notify(90107)
	switch o := op.GetValue().(type) {
	case *GetOperation:
		__antithesis_instrumentation__.Notify(90108)
		fn := db.Get
		if o.ForUpdate {
			__antithesis_instrumentation__.Notify(90118)
			fn = db.GetForUpdate
		} else {
			__antithesis_instrumentation__.Notify(90119)
		}
		__antithesis_instrumentation__.Notify(90109)
		kv, err := fn(ctx, o.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(90120)
			o.Result = resultError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(90121)
			o.Result.Type = ResultType_Value
			if kv.Value != nil {
				__antithesis_instrumentation__.Notify(90122)
				o.Result.Value = kv.Value.RawBytes
			} else {
				__antithesis_instrumentation__.Notify(90123)
				o.Result.Value = nil
			}
		}
	case *PutOperation:
		__antithesis_instrumentation__.Notify(90110)
		err := db.Put(ctx, o.Key, o.Value)
		o.Result = resultError(ctx, err)
	case *ScanOperation:
		__antithesis_instrumentation__.Notify(90111)
		fn := db.Scan
		if o.Reverse && func() bool {
			__antithesis_instrumentation__.Notify(90124)
			return o.ForUpdate == true
		}() == true {
			__antithesis_instrumentation__.Notify(90125)
			fn = db.ReverseScanForUpdate
		} else {
			__antithesis_instrumentation__.Notify(90126)
			if o.Reverse {
				__antithesis_instrumentation__.Notify(90127)
				fn = db.ReverseScan
			} else {
				__antithesis_instrumentation__.Notify(90128)
				if o.ForUpdate {
					__antithesis_instrumentation__.Notify(90129)
					fn = db.ScanForUpdate
				} else {
					__antithesis_instrumentation__.Notify(90130)
				}
			}
		}
		__antithesis_instrumentation__.Notify(90112)
		kvs, err := fn(ctx, o.Key, o.EndKey, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(90131)
			o.Result = resultError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(90132)
			o.Result.Type = ResultType_Values
			o.Result.Values = make([]KeyValue, len(kvs))
			for i, kv := range kvs {
				__antithesis_instrumentation__.Notify(90133)
				o.Result.Values[i] = KeyValue{
					Key:   []byte(kv.Key),
					Value: kv.Value.RawBytes,
				}
			}
		}
	case *DeleteOperation:
		__antithesis_instrumentation__.Notify(90113)
		err := db.Del(ctx, o.Key)
		o.Result = resultError(ctx, err)
	case *DeleteRangeOperation:
		__antithesis_instrumentation__.Notify(90114)
		if !inTxn {
			__antithesis_instrumentation__.Notify(90134)
			panic(errors.AssertionFailedf(`non-transactional DelRange operations currently unsupported`))
		} else {
			__antithesis_instrumentation__.Notify(90135)
		}
		__antithesis_instrumentation__.Notify(90115)
		deletedKeys, err := db.DelRange(ctx, o.Key, o.EndKey, true)
		if err != nil {
			__antithesis_instrumentation__.Notify(90136)
			o.Result = resultError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(90137)
			o.Result.Type = ResultType_Keys
			o.Result.Keys = make([][]byte, len(deletedKeys))
			for i, deletedKey := range deletedKeys {
				__antithesis_instrumentation__.Notify(90138)
				o.Result.Keys[i] = deletedKey
			}
		}
	case *BatchOperation:
		__antithesis_instrumentation__.Notify(90116)
		b := &kv.Batch{}
		applyBatchOp(ctx, b, db.Run, o, inTxn)
	default:
		__antithesis_instrumentation__.Notify(90117)
		panic(errors.AssertionFailedf(`unknown batch operation type: %T %v`, o, o))
	}
}

func applyBatchOp(
	ctx context.Context,
	b *kv.Batch,
	runFn func(context.Context, *kv.Batch) error,
	o *BatchOperation,
	inTxn bool,
) {
	__antithesis_instrumentation__.Notify(90139)
	for i := range o.Ops {
		__antithesis_instrumentation__.Notify(90141)
		switch subO := o.Ops[i].GetValue().(type) {
		case *GetOperation:
			__antithesis_instrumentation__.Notify(90142)
			if subO.ForUpdate {
				__antithesis_instrumentation__.Notify(90149)
				b.GetForUpdate(subO.Key)
			} else {
				__antithesis_instrumentation__.Notify(90150)
				b.Get(subO.Key)
			}
		case *PutOperation:
			__antithesis_instrumentation__.Notify(90143)
			b.Put(subO.Key, subO.Value)
		case *ScanOperation:
			__antithesis_instrumentation__.Notify(90144)
			if subO.Reverse && func() bool {
				__antithesis_instrumentation__.Notify(90151)
				return subO.ForUpdate == true
			}() == true {
				__antithesis_instrumentation__.Notify(90152)
				b.ReverseScanForUpdate(subO.Key, subO.EndKey)
			} else {
				__antithesis_instrumentation__.Notify(90153)
				if subO.Reverse {
					__antithesis_instrumentation__.Notify(90154)
					b.ReverseScan(subO.Key, subO.EndKey)
				} else {
					__antithesis_instrumentation__.Notify(90155)
					if subO.ForUpdate {
						__antithesis_instrumentation__.Notify(90156)
						b.ScanForUpdate(subO.Key, subO.EndKey)
					} else {
						__antithesis_instrumentation__.Notify(90157)
						b.Scan(subO.Key, subO.EndKey)
					}
				}
			}
		case *DeleteOperation:
			__antithesis_instrumentation__.Notify(90145)
			b.Del(subO.Key)
		case *DeleteRangeOperation:
			__antithesis_instrumentation__.Notify(90146)
			if !inTxn {
				__antithesis_instrumentation__.Notify(90158)
				panic(errors.AssertionFailedf(`non-transactional batch DelRange operations currently unsupported`))
			} else {
				__antithesis_instrumentation__.Notify(90159)
			}
			__antithesis_instrumentation__.Notify(90147)
			b.DelRange(subO.Key, subO.EndKey, true)
		default:
			__antithesis_instrumentation__.Notify(90148)
			panic(errors.AssertionFailedf(`unknown batch operation type: %T %v`, subO, subO))
		}
	}
	__antithesis_instrumentation__.Notify(90140)
	runErr := runFn(ctx, b)
	o.Result = resultError(ctx, runErr)
	for i := range o.Ops {
		__antithesis_instrumentation__.Notify(90160)
		switch subO := o.Ops[i].GetValue().(type) {
		case *GetOperation:
			__antithesis_instrumentation__.Notify(90161)
			if b.Results[i].Err != nil {
				__antithesis_instrumentation__.Notify(90167)
				subO.Result = resultError(ctx, b.Results[i].Err)
			} else {
				__antithesis_instrumentation__.Notify(90168)
				subO.Result.Type = ResultType_Value
				result := b.Results[i].Rows[0]
				if result.Value != nil {
					__antithesis_instrumentation__.Notify(90169)
					subO.Result.Value = result.Value.RawBytes
				} else {
					__antithesis_instrumentation__.Notify(90170)
					subO.Result.Value = nil
				}
			}
		case *PutOperation:
			__antithesis_instrumentation__.Notify(90162)
			err := b.Results[i].Err
			subO.Result = resultError(ctx, err)
		case *ScanOperation:
			__antithesis_instrumentation__.Notify(90163)
			kvs, err := b.Results[i].Rows, b.Results[i].Err
			if err != nil {
				__antithesis_instrumentation__.Notify(90171)
				subO.Result = resultError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(90172)
				subO.Result.Type = ResultType_Values
				subO.Result.Values = make([]KeyValue, len(kvs))
				for j, kv := range kvs {
					__antithesis_instrumentation__.Notify(90173)
					subO.Result.Values[j] = KeyValue{
						Key:   []byte(kv.Key),
						Value: kv.Value.RawBytes,
					}
				}
			}
		case *DeleteOperation:
			__antithesis_instrumentation__.Notify(90164)
			err := b.Results[i].Err
			subO.Result = resultError(ctx, err)
		case *DeleteRangeOperation:
			__antithesis_instrumentation__.Notify(90165)
			keys, err := b.Results[i].Keys, b.Results[i].Err
			if err != nil {
				__antithesis_instrumentation__.Notify(90174)
				subO.Result = resultError(ctx, err)
			} else {
				__antithesis_instrumentation__.Notify(90175)
				subO.Result.Type = ResultType_Keys
				subO.Result.Keys = make([][]byte, len(keys))
				for j, key := range keys {
					__antithesis_instrumentation__.Notify(90176)
					subO.Result.Keys[j] = key
				}
			}
		default:
			__antithesis_instrumentation__.Notify(90166)
			panic(errors.AssertionFailedf(`unknown batch operation type: %T %v`, subO, subO))
		}
	}
}

func resultError(ctx context.Context, err error) Result {
	__antithesis_instrumentation__.Notify(90177)
	if err == nil {
		__antithesis_instrumentation__.Notify(90179)
		return Result{Type: ResultType_NoError}
	} else {
		__antithesis_instrumentation__.Notify(90180)
	}
	__antithesis_instrumentation__.Notify(90178)
	ee := errors.EncodeError(ctx, err)
	return Result{
		Type: ResultType_Error,
		Err:  &ee,
	}
}

func getRangeDesc(ctx context.Context, key roachpb.Key, dbs ...*kv.DB) roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(90181)
	var dbIdx int
	var opts = retry.Options{}
	for r := retry.StartWithCtx(ctx, opts); r.Next(); dbIdx = (dbIdx + 1) % len(dbs) {
		__antithesis_instrumentation__.Notify(90183)
		sender := dbs[dbIdx].NonTransactionalSender()
		descs, _, err := kv.RangeLookup(ctx, sender, key, roachpb.CONSISTENT, 0, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(90186)
			log.Infof(ctx, "looking up descriptor for %s: %+v", key, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(90187)
		}
		__antithesis_instrumentation__.Notify(90184)
		if len(descs) != 1 {
			__antithesis_instrumentation__.Notify(90188)
			log.Infof(ctx, "unexpected number of descriptors for %s: %d", key, len(descs))
			continue
		} else {
			__antithesis_instrumentation__.Notify(90189)
		}
		__antithesis_instrumentation__.Notify(90185)
		return descs[0]
	}
	__antithesis_instrumentation__.Notify(90182)
	panic(`unreachable`)
}

func newGetReplicasFn(dbs ...*kv.DB) GetReplicasFn {
	__antithesis_instrumentation__.Notify(90190)
	ctx := context.Background()
	return func(key roachpb.Key) []roachpb.ReplicationTarget {
		__antithesis_instrumentation__.Notify(90191)
		desc := getRangeDesc(ctx, key, dbs...)
		replicas := desc.Replicas().Descriptors()
		targets := make([]roachpb.ReplicationTarget, len(replicas))
		for i, replica := range replicas {
			__antithesis_instrumentation__.Notify(90193)
			targets[i] = roachpb.ReplicationTarget{
				NodeID:  replica.NodeID,
				StoreID: replica.StoreID,
			}
		}
		__antithesis_instrumentation__.Notify(90192)
		return targets
	}
}

func updateZoneConfig(zone *zonepb.ZoneConfig, change ChangeZoneType) {
	__antithesis_instrumentation__.Notify(90194)
	switch change {
	case ChangeZoneType_ToggleGlobalReads:
		__antithesis_instrumentation__.Notify(90195)
		cur := zone.GlobalReads != nil && func() bool {
			__antithesis_instrumentation__.Notify(90197)
			return *zone.GlobalReads == true
		}() == true
		zone.GlobalReads = proto.Bool(!cur)
	default:
		__antithesis_instrumentation__.Notify(90196)
		panic(errors.AssertionFailedf(`unknown ChangeZoneType: %v`, change))
	}
}

func updateZoneConfigInEnv(ctx context.Context, env *Env, change ChangeZoneType) error {
	__antithesis_instrumentation__.Notify(90198)
	return env.UpdateZoneConfig(ctx, int(GeneratorDataTableID), func(zone *zonepb.ZoneConfig) {
		__antithesis_instrumentation__.Notify(90199)
		updateZoneConfig(zone, change)
	})
}
