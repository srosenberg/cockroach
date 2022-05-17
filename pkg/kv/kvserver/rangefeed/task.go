package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type runnable interface {
	Run(context.Context)

	Cancel()
}

type initResolvedTSScan struct {
	p  *Processor
	is IntentScanner
}

func newInitResolvedTSScan(p *Processor, c IntentScanner) runnable {
	__antithesis_instrumentation__.Notify(114080)
	return &initResolvedTSScan{p: p, is: c}
}

func (s *initResolvedTSScan) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(114081)
	defer s.Cancel()
	if err := s.iterateAndConsume(ctx); err != nil {
		__antithesis_instrumentation__.Notify(114082)
		err = errors.Wrap(err, "initial resolved timestamp scan failed")
		log.Errorf(ctx, "%v", err)
		s.p.StopWithErr(roachpb.NewError(err))
	} else {
		__antithesis_instrumentation__.Notify(114083)

		s.p.setResolvedTSInitialized(ctx)
	}
}

func (s *initResolvedTSScan) iterateAndConsume(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(114084)
	startKey := s.p.Span.Key.AsRawKey()
	endKey := s.p.Span.EndKey.AsRawKey()
	return s.is.ConsumeIntents(ctx, startKey, endKey, func(op enginepb.MVCCWriteIntentOp) bool {
		__antithesis_instrumentation__.Notify(114085)
		var ops [1]enginepb.MVCCLogicalOp
		ops[0].SetValue(&op)
		return s.p.sendEvent(ctx, event{ops: ops[:]}, 0)
	})
}

func (s *initResolvedTSScan) Cancel() {
	__antithesis_instrumentation__.Notify(114086)
	s.is.Close()
}

type eventConsumer func(enginepb.MVCCWriteIntentOp) bool

type IntentScanner interface {
	ConsumeIntents(ctx context.Context, startKey roachpb.Key, endKey roachpb.Key, consumer eventConsumer) error

	Close()
}

type SeparatedIntentScanner struct {
	iter storage.EngineIterator
}

func NewSeparatedIntentScanner(iter storage.EngineIterator) IntentScanner {
	__antithesis_instrumentation__.Notify(114087)
	return &SeparatedIntentScanner{iter: iter}
}

func (s *SeparatedIntentScanner) ConsumeIntents(
	ctx context.Context, startKey roachpb.Key, _ roachpb.Key, consumer eventConsumer,
) error {
	__antithesis_instrumentation__.Notify(114088)
	ltStart, _ := keys.LockTableSingleKey(startKey, nil)
	var meta enginepb.MVCCMetadata
	for valid, err := s.iter.SeekEngineKeyGE(storage.EngineKey{Key: ltStart}); ; valid, err = s.iter.NextEngineKey() {
		__antithesis_instrumentation__.Notify(114090)
		if err != nil {
			__antithesis_instrumentation__.Notify(114096)
			return err
		} else {
			__antithesis_instrumentation__.Notify(114097)
			if !valid {
				__antithesis_instrumentation__.Notify(114098)

				break
			} else {
				__antithesis_instrumentation__.Notify(114099)
			}
		}
		__antithesis_instrumentation__.Notify(114091)

		engineKey, err := s.iter.EngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(114100)
			return err
		} else {
			__antithesis_instrumentation__.Notify(114101)
		}
		__antithesis_instrumentation__.Notify(114092)
		lockedKey, err := keys.DecodeLockTableSingleKey(engineKey.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(114102)
			return errors.Wrapf(err, "decoding LockTable key: %s", lockedKey)
		} else {
			__antithesis_instrumentation__.Notify(114103)
		}
		__antithesis_instrumentation__.Notify(114093)

		if err := protoutil.Unmarshal(s.iter.UnsafeValue(), &meta); err != nil {
			__antithesis_instrumentation__.Notify(114104)
			return errors.Wrapf(err, "unmarshaling mvcc meta for locked key %s", lockedKey)
		} else {
			__antithesis_instrumentation__.Notify(114105)
		}
		__antithesis_instrumentation__.Notify(114094)
		if meta.Txn == nil {
			__antithesis_instrumentation__.Notify(114106)
			return errors.Newf("expected transaction metadata but found none for %s", lockedKey)
		} else {
			__antithesis_instrumentation__.Notify(114107)
		}
		__antithesis_instrumentation__.Notify(114095)

		consumer(enginepb.MVCCWriteIntentOp{
			TxnID:           meta.Txn.ID,
			TxnKey:          meta.Txn.Key,
			TxnMinTimestamp: meta.Txn.MinTimestamp,
			Timestamp:       meta.Txn.WriteTimestamp,
		})
	}
	__antithesis_instrumentation__.Notify(114089)
	return nil
}

func (s *SeparatedIntentScanner) Close() {
	__antithesis_instrumentation__.Notify(114108)
	s.iter.Close()
}

type LegacyIntentScanner struct {
	iter storage.SimpleMVCCIterator
}

func NewLegacyIntentScanner(iter storage.SimpleMVCCIterator) IntentScanner {
	__antithesis_instrumentation__.Notify(114109)
	return &LegacyIntentScanner{iter: iter}
}

func (l *LegacyIntentScanner) ConsumeIntents(
	ctx context.Context, start roachpb.Key, end roachpb.Key, consumer eventConsumer,
) error {
	__antithesis_instrumentation__.Notify(114110)
	startKey := storage.MakeMVCCMetadataKey(start)
	endKey := storage.MakeMVCCMetadataKey(end)

	var meta enginepb.MVCCMetadata
	for l.iter.SeekGE(startKey); ; l.iter.NextKey() {
		__antithesis_instrumentation__.Notify(114112)
		if ok, err := l.iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(114116)
			return err
		} else {
			__antithesis_instrumentation__.Notify(114117)
			if !ok || func() bool {
				__antithesis_instrumentation__.Notify(114118)
				return !l.iter.UnsafeKey().Less(endKey) == true
			}() == true {
				__antithesis_instrumentation__.Notify(114119)
				break
			} else {
				__antithesis_instrumentation__.Notify(114120)
			}
		}
		__antithesis_instrumentation__.Notify(114113)

		unsafeKey := l.iter.UnsafeKey()
		if unsafeKey.IsValue() {
			__antithesis_instrumentation__.Notify(114121)
			continue
		} else {
			__antithesis_instrumentation__.Notify(114122)
		}
		__antithesis_instrumentation__.Notify(114114)

		if err := protoutil.Unmarshal(l.iter.UnsafeValue(), &meta); err != nil {
			__antithesis_instrumentation__.Notify(114123)
			return errors.Wrapf(err, "unmarshaling mvcc meta: %v", unsafeKey)
		} else {
			__antithesis_instrumentation__.Notify(114124)
		}
		__antithesis_instrumentation__.Notify(114115)

		if meta.Txn != nil {
			__antithesis_instrumentation__.Notify(114125)
			consumer(enginepb.MVCCWriteIntentOp{
				TxnID:           meta.Txn.ID,
				TxnKey:          meta.Txn.Key,
				TxnMinTimestamp: meta.Txn.MinTimestamp,
				Timestamp:       meta.Txn.WriteTimestamp,
			})
		} else {
			__antithesis_instrumentation__.Notify(114126)
		}
	}
	__antithesis_instrumentation__.Notify(114111)
	return nil
}

func (l *LegacyIntentScanner) Close() { __antithesis_instrumentation__.Notify(114127); l.iter.Close() }

type TxnPusher interface {
	PushTxns(context.Context, []enginepb.TxnMeta, hlc.Timestamp) ([]*roachpb.Transaction, error)

	ResolveIntents(ctx context.Context, intents []roachpb.LockUpdate) error
}

type txnPushAttempt struct {
	p     *Processor
	txns  []enginepb.TxnMeta
	ts    hlc.Timestamp
	doneC chan struct{}
}

func newTxnPushAttempt(
	p *Processor, txns []enginepb.TxnMeta, ts hlc.Timestamp, doneC chan struct{},
) runnable {
	__antithesis_instrumentation__.Notify(114128)
	return &txnPushAttempt{
		p:     p,
		txns:  txns,
		ts:    ts,
		doneC: doneC,
	}
}

func (a *txnPushAttempt) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(114129)
	defer a.Cancel()
	if err := a.pushOldTxns(ctx); err != nil {
		__antithesis_instrumentation__.Notify(114130)
		log.Errorf(ctx, "pushing old intents failed: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(114131)
	}
}

func (a *txnPushAttempt) pushOldTxns(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(114132)

	pushedTxns, err := a.p.TxnPusher.PushTxns(ctx, a.txns, a.ts)
	if err != nil {
		__antithesis_instrumentation__.Notify(114136)
		return err
	} else {
		__antithesis_instrumentation__.Notify(114137)
	}
	__antithesis_instrumentation__.Notify(114133)
	if len(pushedTxns) != len(a.txns) {
		__antithesis_instrumentation__.Notify(114138)

		return errors.AssertionFailedf("tried to push %d transactions, got response for %d",
			len(a.txns), len(pushedTxns))
	} else {
		__antithesis_instrumentation__.Notify(114139)
	}
	__antithesis_instrumentation__.Notify(114134)

	ops := make([]enginepb.MVCCLogicalOp, len(pushedTxns))
	var intentsToCleanup []roachpb.LockUpdate
	for i, txn := range pushedTxns {
		__antithesis_instrumentation__.Notify(114140)
		switch txn.Status {
		case roachpb.PENDING, roachpb.STAGING:
			__antithesis_instrumentation__.Notify(114141)

			ops[i].SetValue(&enginepb.MVCCUpdateIntentOp{
				TxnID:     txn.ID,
				Timestamp: txn.WriteTimestamp,
			})
		case roachpb.COMMITTED:
			__antithesis_instrumentation__.Notify(114142)

			ops[i].SetValue(&enginepb.MVCCUpdateIntentOp{
				TxnID:     txn.ID,
				Timestamp: txn.WriteTimestamp,
			})

			txnIntents := intentsInBound(txn, a.p.Span.AsRawSpanWithNoLocals())
			intentsToCleanup = append(intentsToCleanup, txnIntents...)
		case roachpb.ABORTED:
			__antithesis_instrumentation__.Notify(114143)

			ops[i].SetValue(&enginepb.MVCCAbortTxnOp{
				TxnID: txn.ID,
			})

			txnIntents := intentsInBound(txn, a.p.Span.AsRawSpanWithNoLocals())
			intentsToCleanup = append(intentsToCleanup, txnIntents...)
		default:
			__antithesis_instrumentation__.Notify(114144)
		}
	}
	__antithesis_instrumentation__.Notify(114135)

	a.p.sendEvent(ctx, event{ops: ops}, 0)

	return a.p.TxnPusher.ResolveIntents(ctx, intentsToCleanup)
}

func (a *txnPushAttempt) Cancel() {
	__antithesis_instrumentation__.Notify(114145)
	close(a.doneC)
}

func intentsInBound(txn *roachpb.Transaction, bound roachpb.Span) []roachpb.LockUpdate {
	__antithesis_instrumentation__.Notify(114146)
	var ret []roachpb.LockUpdate
	for _, sp := range txn.LockSpans {
		__antithesis_instrumentation__.Notify(114148)
		if in := sp.Intersect(bound); in.Valid() {
			__antithesis_instrumentation__.Notify(114149)
			ret = append(ret, roachpb.MakeLockUpdate(txn, in))
		} else {
			__antithesis_instrumentation__.Notify(114150)
		}
	}
	__antithesis_instrumentation__.Notify(114147)
	return ret
}
