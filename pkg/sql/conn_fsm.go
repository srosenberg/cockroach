package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlfsm"
	"github.com/cockroachdb/cockroach/pkg/util/fsm"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

const (
	NoTxnStateStr         = sqlfsm.NoTxnStateStr
	OpenStateStr          = sqlfsm.OpenStateStr
	AbortedStateStr       = sqlfsm.AbortedStateStr
	CommitWaitStateStr    = sqlfsm.CommitWaitStateStr
	InternalErrorStateStr = sqlfsm.InternalErrorStateStr
)

type stateNoTxn struct{}

var _ fsm.State = &stateNoTxn{}

func (stateNoTxn) String() string {
	__antithesis_instrumentation__.Notify(458866)
	return NoTxnStateStr
}

type stateOpen struct {
	ImplicitTxn fsm.Bool
}

var _ fsm.State = &stateOpen{}

func (stateOpen) String() string {
	__antithesis_instrumentation__.Notify(458867)
	return OpenStateStr
}

type stateAborted struct{}

var _ fsm.State = &stateAborted{}

func (stateAborted) String() string {
	__antithesis_instrumentation__.Notify(458868)
	return AbortedStateStr
}

type stateCommitWait struct{}

var _ fsm.State = &stateCommitWait{}

func (stateCommitWait) String() string {
	__antithesis_instrumentation__.Notify(458869)
	return CommitWaitStateStr
}

type stateInternalError struct{}

var _ fsm.State = &stateInternalError{}

func (stateInternalError) String() string {
	__antithesis_instrumentation__.Notify(458870)
	return InternalErrorStateStr
}

func (stateNoTxn) State()         { __antithesis_instrumentation__.Notify(458871) }
func (stateOpen) State()          { __antithesis_instrumentation__.Notify(458872) }
func (stateAborted) State()       { __antithesis_instrumentation__.Notify(458873) }
func (stateCommitWait) State()    { __antithesis_instrumentation__.Notify(458874) }
func (stateInternalError) State() { __antithesis_instrumentation__.Notify(458875) }

type eventTxnStart struct {
	ImplicitTxn fsm.Bool
}
type eventTxnStartPayload struct {
	tranCtx transitionCtx

	pri roachpb.UserPriority

	txnSQLTimestamp     time.Time
	readOnly            tree.ReadWriteMode
	historicalTimestamp *hlc.Timestamp

	qualityOfService sessiondatapb.QoSLevel
}

func makeEventTxnStartPayload(
	pri roachpb.UserPriority,
	readOnly tree.ReadWriteMode,
	txnSQLTimestamp time.Time,
	historicalTimestamp *hlc.Timestamp,
	tranCtx transitionCtx,
	qualityOfService sessiondatapb.QoSLevel,
) eventTxnStartPayload {
	__antithesis_instrumentation__.Notify(458876)
	return eventTxnStartPayload{
		pri:                 pri,
		readOnly:            readOnly,
		txnSQLTimestamp:     txnSQLTimestamp,
		historicalTimestamp: historicalTimestamp,
		tranCtx:             tranCtx,
		qualityOfService:    qualityOfService,
	}
}

type eventTxnUpgradeToExplicit struct{}

type eventTxnFinishCommitted struct{}
type eventTxnFinishAborted struct{}

type eventSavepointRollback struct{}

type eventNonRetriableErr struct {
	IsCommit fsm.Bool
}

type eventNonRetriableErrPayload struct {
	err error
}

func (p eventNonRetriableErrPayload) errorCause() error {
	__antithesis_instrumentation__.Notify(458877)
	return p.err
}

var _ payloadWithError = eventNonRetriableErrPayload{}

type eventRetriableErr struct {
	CanAutoRetry fsm.Bool
	IsCommit     fsm.Bool
}

type eventRetriableErrPayload struct {
	err error

	rewCap rewindCapability
}

func (p eventRetriableErrPayload) errorCause() error {
	__antithesis_instrumentation__.Notify(458878)
	return p.err
}

var _ payloadWithError = eventRetriableErrPayload{}

type eventTxnRestart struct{}

type eventTxnReleased struct{}

type payloadWithError interface {
	errorCause() error
}

func (eventTxnStart) Event()             { __antithesis_instrumentation__.Notify(458879) }
func (eventTxnFinishCommitted) Event()   { __antithesis_instrumentation__.Notify(458880) }
func (eventTxnFinishAborted) Event()     { __antithesis_instrumentation__.Notify(458881) }
func (eventSavepointRollback) Event()    { __antithesis_instrumentation__.Notify(458882) }
func (eventNonRetriableErr) Event()      { __antithesis_instrumentation__.Notify(458883) }
func (eventRetriableErr) Event()         { __antithesis_instrumentation__.Notify(458884) }
func (eventTxnRestart) Event()           { __antithesis_instrumentation__.Notify(458885) }
func (eventTxnReleased) Event()          { __antithesis_instrumentation__.Notify(458886) }
func (eventTxnUpgradeToExplicit) Event() { __antithesis_instrumentation__.Notify(458887) }

var TxnStateTransitions = fsm.Compile(fsm.Pattern{

	stateNoTxn{}: {
		eventTxnStart{fsm.Var("implicitTxn")}: {
			Description: "BEGIN, or before a statement running as an implicit txn",
			Next:        stateOpen{ImplicitTxn: fsm.Var("implicitTxn")},
			Action:      noTxnToOpen,
		},
		eventNonRetriableErr{IsCommit: fsm.Any}: {

			Description: "anything but BEGIN or extended protocol command error",
			Next:        stateNoTxn{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458888)
				ts := args.Extended.(*txnState)
				ts.setAdvanceInfo(skipBatch, noRewind, txnEvent{eventType: noEvent})
				return nil
			},
		},
	},

	stateOpen{ImplicitTxn: fsm.Any}: {
		eventTxnFinishCommitted{}: {
			Description: "COMMIT, or after a statement running as an implicit txn",
			Next:        stateNoTxn{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458889)

				return args.Extended.(*txnState).finishTxn(txnCommit)
			},
		},
		eventTxnFinishAborted{}: {
			Description: "ROLLBACK, or after a statement running as an implicit txn fails",
			Next:        stateNoTxn{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458890)

				return args.Extended.(*txnState).finishTxn(txnRollback)
			},
		},

		eventRetriableErr{CanAutoRetry: fsm.False, IsCommit: fsm.True}: {
			Description: "Retriable err on COMMIT",
			Next:        stateNoTxn{},
			Action:      cleanupAndFinishOnError,
		},
		eventNonRetriableErr{IsCommit: fsm.True}: {
			Next:   stateNoTxn{},
			Action: cleanupAndFinishOnError,
		},
	},
	stateOpen{ImplicitTxn: fsm.Var("implicitTxn")}: {

		eventRetriableErr{CanAutoRetry: fsm.True, IsCommit: fsm.Any}: {

			Description: "Retriable err; will auto-retry",
			Next:        stateOpen{ImplicitTxn: fsm.Var("implicitTxn")},
			Action:      prepareTxnForRetryWithRewind,
		},
	},

	stateOpen{ImplicitTxn: fsm.True}: {
		eventRetriableErr{CanAutoRetry: fsm.False, IsCommit: fsm.False}: {
			Next:   stateNoTxn{},
			Action: cleanupAndFinishOnError,
		},
		eventNonRetriableErr{IsCommit: fsm.False}: {
			Next:   stateNoTxn{},
			Action: cleanupAndFinishOnError,
		},
		eventTxnUpgradeToExplicit{}: {
			Next: stateOpen{ImplicitTxn: fsm.False},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458891)
				args.Extended.(*txnState).setAdvanceInfo(
					advanceOne,
					noRewind,
					txnEvent{eventType: noEvent},
				)
				return nil
			},
		},
	},

	stateOpen{ImplicitTxn: fsm.False}: {
		eventNonRetriableErr{IsCommit: fsm.False}: {
			Next: stateAborted{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458892)
				ts := args.Extended.(*txnState)
				ts.setAdvanceInfo(skipBatch, noRewind, txnEvent{eventType: noEvent})
				return nil
			},
		},

		eventTxnRestart{}: {
			Description: "ROLLBACK TO SAVEPOINT cockroach_restart",
			Next:        stateOpen{ImplicitTxn: fsm.False},
			Action:      prepareTxnForRetry,
		},
		eventRetriableErr{CanAutoRetry: fsm.False, IsCommit: fsm.False}: {
			Next: stateAborted{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458893)
				args.Extended.(*txnState).setAdvanceInfo(
					skipBatch,
					noRewind,
					txnEvent{eventType: noEvent},
				)
				return nil
			},
		},
		eventTxnReleased{}: {
			Description: "RELEASE SAVEPOINT cockroach_restart",
			Next:        stateCommitWait{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458894)
				ts := args.Extended.(*txnState)
				ts.mu.Lock()
				txnID := ts.mu.txn.ID()
				ts.mu.Unlock()
				ts.setAdvanceInfo(
					advanceOne,
					noRewind,
					txnEvent{eventType: txnCommit, txnID: txnID},
				)
				return nil
			},
		},
	},

	stateAborted{}: {
		eventTxnFinishAborted{}: {
			Description: "ROLLBACK",
			Next:        stateNoTxn{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458895)
				ts := args.Extended.(*txnState)
				ts.txnAbortCount.Inc(1)

				return ts.finishTxn(txnRollback)
			},
		},

		eventNonRetriableErr{IsCommit: fsm.False}: {

			Description: "any other statement",
			Next:        stateAborted{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458896)
				args.Extended.(*txnState).setAdvanceInfo(
					skipBatch,
					noRewind,
					txnEvent{eventType: noEvent},
				)
				return nil
			},
		},

		eventNonRetriableErr{IsCommit: fsm.True}: {

			Description: "ConnExecutor closing",
			Next:        stateAborted{},
			Action:      cleanupAndFinishOnError,
		},

		eventSavepointRollback{}: {
			Description: "ROLLBACK TO SAVEPOINT (not cockroach_restart) success",
			Next:        stateOpen{ImplicitTxn: fsm.False},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458897)
				args.Extended.(*txnState).setAdvanceInfo(
					advanceOne,
					noRewind,
					txnEvent{eventType: noEvent},
				)
				return nil
			},
		},

		eventRetriableErr{CanAutoRetry: fsm.Any, IsCommit: fsm.Any}: {

			Description: "ROLLBACK TO SAVEPOINT (not cockroach_restart) failed because txn needs restart",
			Next:        stateAborted{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458898)
				args.Extended.(*txnState).setAdvanceInfo(
					skipBatch,
					noRewind,
					txnEvent{eventType: noEvent},
				)
				return nil
			},
		},

		eventTxnRestart{}: {
			Description: "ROLLBACK TO SAVEPOINT cockroach_restart",
			Next:        stateOpen{ImplicitTxn: fsm.False},
			Action:      prepareTxnForRetry,
		},
	},

	stateCommitWait{}: {
		eventTxnFinishCommitted{}: {
			Description: "COMMIT",
			Next:        stateNoTxn{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458899)

				return args.Extended.(*txnState).finishTxn(noEvent)
			},
		},
		eventNonRetriableErr{IsCommit: fsm.Any}: {

			Description: "any other statement",
			Next:        stateCommitWait{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458900)
				args.Extended.(*txnState).setAdvanceInfo(
					skipBatch,
					noRewind,
					txnEvent{eventType: noEvent},
				)
				return nil
			},
		},
	},
})

func noTxnToOpen(args fsm.Args) error {
	__antithesis_instrumentation__.Notify(458901)
	connCtx := args.Ctx
	ev := args.Event.(eventTxnStart)
	payload := args.Payload.(eventTxnStartPayload)
	ts := args.Extended.(*txnState)

	txnTyp := explicitTxn
	advCode := advanceOne
	if ev.ImplicitTxn.Get() {
		__antithesis_instrumentation__.Notify(458903)
		txnTyp = implicitTxn

		advCode = stayInPlace
	} else {
		__antithesis_instrumentation__.Notify(458904)
	}
	__antithesis_instrumentation__.Notify(458902)

	newTxnID := ts.resetForNewSQLTxn(
		connCtx,
		txnTyp,
		payload.txnSQLTimestamp,
		payload.historicalTimestamp,
		payload.pri,
		payload.readOnly,
		nil,
		payload.tranCtx,
		payload.qualityOfService,
	)
	ts.setAdvanceInfo(
		advCode,
		noRewind,
		txnEvent{eventType: txnStart, txnID: newTxnID},
	)
	return nil
}

func (ts *txnState) finishTxn(ev txnEventType) error {
	__antithesis_instrumentation__.Notify(458905)
	finishedTxnID := ts.finishSQLTxn()
	ts.setAdvanceInfo(advanceOne, noRewind, txnEvent{eventType: ev, txnID: finishedTxnID})
	return nil
}

func cleanupAndFinishOnError(args fsm.Args) error {
	__antithesis_instrumentation__.Notify(458906)
	ts := args.Extended.(*txnState)
	ts.mu.Lock()
	ts.mu.txn.CleanupOnError(ts.Ctx, args.Payload.(payloadWithError).errorCause())
	ts.mu.Unlock()
	finishedTxnID := ts.finishSQLTxn()
	ts.setAdvanceInfo(
		skipBatch,
		noRewind,
		txnEvent{eventType: txnRollback, txnID: finishedTxnID},
	)
	return nil
}

func prepareTxnForRetry(args fsm.Args) error {
	__antithesis_instrumentation__.Notify(458907)
	ts := args.Extended.(*txnState)
	ts.mu.Lock()
	ts.mu.txn.PrepareForRetry(ts.Ctx)
	ts.mu.Unlock()
	ts.setAdvanceInfo(
		advanceOne,
		noRewind,
		txnEvent{eventType: txnRestart},
	)
	return nil
}

func prepareTxnForRetryWithRewind(args fsm.Args) error {
	__antithesis_instrumentation__.Notify(458908)
	ts := args.Extended.(*txnState)
	ts.mu.Lock()
	ts.mu.txn.PrepareForRetry(ts.Ctx)
	ts.mu.Unlock()

	ts.setAdvanceInfo(
		rewind,
		args.Payload.(eventRetriableErrPayload).rewCap,
		txnEvent{eventType: txnRestart},
	)
	return nil
}

var BoundTxnStateTransitions = fsm.Compile(fsm.Pattern{
	stateOpen{ImplicitTxn: fsm.False}: {

		eventNonRetriableErr{IsCommit: fsm.Any}: {
			Next: stateInternalError{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458909)
				ts := args.Extended.(*txnState)
				finishedTxnID := ts.finishSQLTxn()
				ts.setAdvanceInfo(
					skipBatch,
					noRewind,
					txnEvent{eventType: txnRollback, txnID: finishedTxnID},
				)
				return nil
			},
		},
		eventRetriableErr{CanAutoRetry: fsm.Any, IsCommit: fsm.False}: {
			Next: stateInternalError{},
			Action: func(args fsm.Args) error {
				__antithesis_instrumentation__.Notify(458910)
				ts := args.Extended.(*txnState)
				finishedTxnID := ts.finishSQLTxn()
				ts.setAdvanceInfo(
					skipBatch,
					noRewind,
					txnEvent{eventType: txnRollback, txnID: finishedTxnID},
				)
				return nil
			},
		},
	},
})
