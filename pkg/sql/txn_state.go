package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel/attribute"
)

type txnState struct {
	mu struct {
		syncutil.RWMutex

		txn *kv.Txn

		txnStart time.Time

		stmtCount int
	}

	connCtx context.Context

	Ctx context.Context

	recordingThreshold time.Duration
	recordingStart     time.Time

	cancel context.CancelFunc

	sqlTimestamp time.Time

	priority roachpb.UserPriority

	readOnly bool

	isHistorical bool

	lastEpoch enginepb.TxnEpoch

	mon *mon.BytesMonitor

	adv advanceInfo

	txnAbortCount *metric.Counter

	testingForceRealTracingSpans bool
}

type txnType int

const (
	implicitTxn txnType = iota

	explicitTxn
)

func (ts *txnState) resetForNewSQLTxn(
	connCtx context.Context,
	txnType txnType,
	sqlTimestamp time.Time,
	historicalTimestamp *hlc.Timestamp,
	priority roachpb.UserPriority,
	readOnly tree.ReadWriteMode,
	txn *kv.Txn,
	tranCtx transitionCtx,
	qualityOfService sessiondatapb.QoSLevel,
) (txnID uuid.UUID) {
	__antithesis_instrumentation__.Notify(628923)

	ts.sqlTimestamp = sqlTimestamp
	ts.isHistorical = false
	ts.lastEpoch = 0

	opName := sqlTxnName
	alreadyRecording := tranCtx.sessionTracing.Enabled()

	var txnCtx context.Context
	var sp *tracing.Span
	duration := traceTxnThreshold.Get(&tranCtx.settings.SV)
	if alreadyRecording || func() bool {
		__antithesis_instrumentation__.Notify(628930)
		return duration > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(628931)
		txnCtx, sp = tracing.EnsureChildSpan(connCtx, tranCtx.tracer, opName,
			tracing.WithRecording(tracing.RecordingVerbose))
	} else {
		__antithesis_instrumentation__.Notify(628932)
		if ts.testingForceRealTracingSpans {
			__antithesis_instrumentation__.Notify(628933)
			txnCtx, sp = tracing.EnsureChildSpan(connCtx, tranCtx.tracer, opName, tracing.WithForceRealSpan())
		} else {
			__antithesis_instrumentation__.Notify(628934)
			txnCtx, sp = tracing.EnsureChildSpan(connCtx, tranCtx.tracer, opName)
		}
	}
	__antithesis_instrumentation__.Notify(628924)
	if txnType == implicitTxn {
		__antithesis_instrumentation__.Notify(628935)
		sp.SetTag("implicit", attribute.StringValue("true"))
	} else {
		__antithesis_instrumentation__.Notify(628936)
	}
	__antithesis_instrumentation__.Notify(628925)

	if !alreadyRecording && func() bool {
		__antithesis_instrumentation__.Notify(628937)
		return (duration > 0) == true
	}() == true {
		__antithesis_instrumentation__.Notify(628938)
		ts.recordingThreshold = duration
		ts.recordingStart = timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(628939)
	}
	__antithesis_instrumentation__.Notify(628926)

	ts.Ctx, ts.cancel = contextutil.WithCancel(txnCtx)
	ts.mon.Start(ts.Ctx, tranCtx.connMon, mon.BoundAccount{})
	ts.mu.Lock()
	ts.mu.stmtCount = 0
	if txn == nil {
		__antithesis_instrumentation__.Notify(628940)
		ts.mu.txn = kv.NewTxnWithSteppingEnabled(ts.Ctx, tranCtx.db, tranCtx.nodeIDOrZero, qualityOfService)
		ts.mu.txn.SetDebugName(opName)
		if err := ts.setPriorityLocked(priority); err != nil {
			__antithesis_instrumentation__.Notify(628941)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(628942)
		}
	} else {
		__antithesis_instrumentation__.Notify(628943)
		if priority != roachpb.UnspecifiedUserPriority {
			__antithesis_instrumentation__.Notify(628945)
			panic(errors.AssertionFailedf("unexpected priority when using an existing txn: %s", priority))
		} else {
			__antithesis_instrumentation__.Notify(628946)
		}
		__antithesis_instrumentation__.Notify(628944)
		ts.mu.txn = txn
	}
	__antithesis_instrumentation__.Notify(628927)
	txnID = ts.mu.txn.ID()
	sp.SetTag("txn", attribute.StringValue(ts.mu.txn.ID().String()))
	ts.mu.txnStart = timeutil.Now()
	ts.mu.Unlock()
	if historicalTimestamp != nil {
		__antithesis_instrumentation__.Notify(628947)
		if err := ts.setHistoricalTimestamp(ts.Ctx, *historicalTimestamp); err != nil {
			__antithesis_instrumentation__.Notify(628948)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(628949)
		}
	} else {
		__antithesis_instrumentation__.Notify(628950)
	}
	__antithesis_instrumentation__.Notify(628928)
	if err := ts.setReadOnlyMode(readOnly); err != nil {
		__antithesis_instrumentation__.Notify(628951)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(628952)
	}
	__antithesis_instrumentation__.Notify(628929)

	return txnID
}

func (ts *txnState) finishSQLTxn() (txnID uuid.UUID) {
	__antithesis_instrumentation__.Notify(628953)
	ts.mon.Stop(ts.Ctx)
	if ts.cancel != nil {
		__antithesis_instrumentation__.Notify(628957)
		ts.cancel()
		ts.cancel = nil
	} else {
		__antithesis_instrumentation__.Notify(628958)
	}
	__antithesis_instrumentation__.Notify(628954)
	sp := tracing.SpanFromContext(ts.Ctx)
	if sp == nil {
		__antithesis_instrumentation__.Notify(628959)
		panic(errors.AssertionFailedf("No span in context? Was resetForNewSQLTxn() called previously?"))
	} else {
		__antithesis_instrumentation__.Notify(628960)
	}
	__antithesis_instrumentation__.Notify(628955)

	if ts.recordingThreshold > 0 {
		__antithesis_instrumentation__.Notify(628961)
		logTraceAboveThreshold(ts.Ctx, sp.GetRecording(sp.RecordingType()), "SQL txn", ts.recordingThreshold, timeutil.Since(ts.recordingStart))
	} else {
		__antithesis_instrumentation__.Notify(628962)
	}
	__antithesis_instrumentation__.Notify(628956)

	sp.Finish()
	ts.Ctx = nil
	ts.mu.Lock()
	txnID = ts.mu.txn.ID()
	ts.mu.txn = nil
	ts.mu.txnStart = time.Time{}
	ts.mu.Unlock()
	ts.recordingThreshold = 0
	return txnID
}

func (ts *txnState) finishExternalTxn() {
	__antithesis_instrumentation__.Notify(628963)
	if ts.Ctx == nil {
		__antithesis_instrumentation__.Notify(628967)
		ts.mon.Stop(ts.connCtx)
	} else {
		__antithesis_instrumentation__.Notify(628968)
		ts.mon.Stop(ts.Ctx)
	}
	__antithesis_instrumentation__.Notify(628964)
	if ts.cancel != nil {
		__antithesis_instrumentation__.Notify(628969)
		ts.cancel()
		ts.cancel = nil
	} else {
		__antithesis_instrumentation__.Notify(628970)
	}
	__antithesis_instrumentation__.Notify(628965)

	if ts.Ctx != nil {
		__antithesis_instrumentation__.Notify(628971)
		if sp := tracing.SpanFromContext(ts.Ctx); sp != nil {
			__antithesis_instrumentation__.Notify(628972)
			sp.Finish()
		} else {
			__antithesis_instrumentation__.Notify(628973)
		}
	} else {
		__antithesis_instrumentation__.Notify(628974)
	}
	__antithesis_instrumentation__.Notify(628966)
	ts.Ctx = nil
	ts.mu.Lock()
	ts.mu.txn = nil
	ts.mu.Unlock()
}

func (ts *txnState) setHistoricalTimestamp(
	ctx context.Context, historicalTimestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(628975)
	ts.mu.Lock()
	defer ts.mu.Unlock()
	if err := ts.mu.txn.SetFixedTimestamp(ctx, historicalTimestamp); err != nil {
		__antithesis_instrumentation__.Notify(628977)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628978)
	}
	__antithesis_instrumentation__.Notify(628976)
	ts.sqlTimestamp = historicalTimestamp.GoTime()
	ts.isHistorical = true
	return nil
}

func (ts *txnState) getReadTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(628979)
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	return ts.mu.txn.ReadTimestamp()
}

func (ts *txnState) setPriority(userPriority roachpb.UserPriority) error {
	__antithesis_instrumentation__.Notify(628980)
	ts.mu.Lock()
	defer ts.mu.Unlock()
	return ts.setPriorityLocked(userPriority)
}

func (ts *txnState) setPriorityLocked(userPriority roachpb.UserPriority) error {
	__antithesis_instrumentation__.Notify(628981)
	if err := ts.mu.txn.SetUserPriority(userPriority); err != nil {
		__antithesis_instrumentation__.Notify(628983)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628984)
	}
	__antithesis_instrumentation__.Notify(628982)
	ts.priority = userPriority
	return nil
}

func (ts *txnState) setReadOnlyMode(mode tree.ReadWriteMode) error {
	__antithesis_instrumentation__.Notify(628985)
	switch mode {
	case tree.UnspecifiedReadWriteMode:
		__antithesis_instrumentation__.Notify(628987)
		return nil
	case tree.ReadOnly:
		__antithesis_instrumentation__.Notify(628988)
		ts.readOnly = true
	case tree.ReadWrite:
		__antithesis_instrumentation__.Notify(628989)
		if ts.isHistorical {
			__antithesis_instrumentation__.Notify(628992)
			return tree.ErrAsOfSpecifiedWithReadWrite
		} else {
			__antithesis_instrumentation__.Notify(628993)
		}
		__antithesis_instrumentation__.Notify(628990)
		ts.readOnly = false
	default:
		__antithesis_instrumentation__.Notify(628991)
		return errors.AssertionFailedf("unknown read mode: %s", errors.Safe(mode))
	}
	__antithesis_instrumentation__.Notify(628986)
	return nil
}

type advanceCode int

const (
	advanceUnknown advanceCode = iota

	stayInPlace

	advanceOne

	skipBatch

	rewind
)

type txnEvent struct {
	eventType txnEventType

	txnID uuid.UUID
}

type txnEventType int

const (
	noEvent txnEventType = iota

	txnStart

	txnCommit

	txnRollback

	txnRestart
)

type advanceInfo struct {
	code advanceCode

	txnEvent txnEvent

	rewCap rewindCapability
}

type transitionCtx struct {
	db           *kv.DB
	nodeIDOrZero roachpb.NodeID
	clock        *hlc.Clock

	connMon *mon.BytesMonitor

	tracer *tracing.Tracer

	sessionTracing   *SessionTracing
	settings         *cluster.Settings
	execTestingKnobs ExecutorTestingKnobs
}

var noRewind = rewindCapability{}

func (ts *txnState) setAdvanceInfo(code advanceCode, rewCap rewindCapability, ev txnEvent) {
	__antithesis_instrumentation__.Notify(628994)
	if ts.adv.code != advanceUnknown {
		__antithesis_instrumentation__.Notify(628997)
		panic("previous advanceInfo has not been consume()d")
	} else {
		__antithesis_instrumentation__.Notify(628998)
	}
	__antithesis_instrumentation__.Notify(628995)
	if code != rewind && func() bool {
		__antithesis_instrumentation__.Notify(628999)
		return rewCap != noRewind == true
	}() == true {
		__antithesis_instrumentation__.Notify(629000)
		panic("if rewCap is specified, code needs to be rewind")
	} else {
		__antithesis_instrumentation__.Notify(629001)
	}
	__antithesis_instrumentation__.Notify(628996)
	ts.adv = advanceInfo{
		code:     code,
		rewCap:   rewCap,
		txnEvent: ev,
	}
}

func (ts *txnState) consumeAdvanceInfo() advanceInfo {
	__antithesis_instrumentation__.Notify(629002)
	adv := ts.adv
	ts.adv = advanceInfo{}
	return adv
}
