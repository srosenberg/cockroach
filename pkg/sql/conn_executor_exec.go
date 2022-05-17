package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/sql/delegate"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/execstats"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec/explain"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessionphase"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/fsm"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
	"go.opentelemetry.io/otel/attribute"
)

func (ex *connExecutor) execStmt(
	ctx context.Context,
	parserStmt parser.Statement,
	prepared *PreparedStatement,
	pinfo *tree.PlaceholderInfo,
	res RestrictedCommandResult,
	canAutoCommit bool,
) (fsm.Event, fsm.EventPayload, error) {
	__antithesis_instrumentation__.Notify(457836)
	ast := parserStmt.AST
	if log.V(2) || func() bool {
		__antithesis_instrumentation__.Notify(457842)
		return logStatementsExecuteEnabled.Get(&ex.server.cfg.Settings.SV) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(457843)
		return log.HasSpanOrEvent(ctx) == true
	}() == true {
		__antithesis_instrumentation__.Notify(457844)
		log.VEventf(ctx, 2, "executing: %s in state: %s", ast, ex.machine.CurState())
	} else {
		__antithesis_instrumentation__.Notify(457845)
	}
	__antithesis_instrumentation__.Notify(457837)

	ex.mu.IdleInSessionTimeout.Stop()
	ex.mu.IdleInTransactionSessionTimeout.Stop()

	if _, ok := ast.(tree.ObserverStatement); ok {
		__antithesis_instrumentation__.Notify(457846)
		ex.statsCollector.Reset(ex.applicationStats, ex.phaseTimes)
		err := ex.runObserverStatement(ctx, ast, res)

		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(457847)
	}
	__antithesis_instrumentation__.Notify(457838)

	var ev fsm.Event
	var payload fsm.EventPayload
	var err error

	switch ex.machine.CurState().(type) {
	case stateNoTxn:
		__antithesis_instrumentation__.Notify(457848)

		ev, payload = ex.execStmtInNoTxnState(ctx, ast)

	case stateOpen:
		__antithesis_instrumentation__.Notify(457849)
		if ex.server.cfg.Settings.CPUProfileType() == cluster.CPUProfileWithLabels {
			__antithesis_instrumentation__.Notify(457854)
			remoteAddr := "internal"
			if rAddr := ex.sessionData().RemoteAddr; rAddr != nil {
				__antithesis_instrumentation__.Notify(457857)
				remoteAddr = rAddr.String()
			} else {
				__antithesis_instrumentation__.Notify(457858)
			}
			__antithesis_instrumentation__.Notify(457855)
			var stmtNoConstants string
			if prepared != nil {
				__antithesis_instrumentation__.Notify(457859)
				stmtNoConstants = prepared.StatementNoConstants
			} else {
				__antithesis_instrumentation__.Notify(457860)
				stmtNoConstants = formatStatementHideConstants(ast)
			}
			__antithesis_instrumentation__.Notify(457856)
			labels := pprof.Labels(
				"appname", ex.sessionData().ApplicationName,
				"addr", remoteAddr,
				"stmt.tag", ast.StatementTag(),
				"stmt.no.constants", stmtNoConstants,
			)
			pprof.Do(ctx, labels, func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(457861)
				ev, payload, err = ex.execStmtInOpenState(ctx, parserStmt, prepared, pinfo, res, canAutoCommit)
			})
		} else {
			__antithesis_instrumentation__.Notify(457862)
			ev, payload, err = ex.execStmtInOpenState(ctx, parserStmt, prepared, pinfo, res, canAutoCommit)
		}
		__antithesis_instrumentation__.Notify(457850)
		switch ev.(type) {
		case eventNonRetriableErr:
			__antithesis_instrumentation__.Notify(457863)
			ex.recordFailure()
		}

	case stateAborted:
		__antithesis_instrumentation__.Notify(457851)
		ev, payload = ex.execStmtInAbortedState(ctx, ast, res)

	case stateCommitWait:
		__antithesis_instrumentation__.Notify(457852)
		ev, payload = ex.execStmtInCommitWaitState(ctx, ast, res)

	default:
		__antithesis_instrumentation__.Notify(457853)
		panic(errors.AssertionFailedf("unexpected txn state: %#v", ex.machine.CurState()))
	}
	__antithesis_instrumentation__.Notify(457839)

	if ex.sessionData().IdleInSessionTimeout > 0 {
		__antithesis_instrumentation__.Notify(457864)

		ex.mu.IdleInSessionTimeout = timeout{time.AfterFunc(
			ex.sessionData().IdleInSessionTimeout,
			ex.cancelSession,
		)}
	} else {
		__antithesis_instrumentation__.Notify(457865)
	}
	__antithesis_instrumentation__.Notify(457840)

	if ex.sessionData().IdleInTransactionSessionTimeout > 0 {
		__antithesis_instrumentation__.Notify(457866)
		startIdleInTransactionSessionTimeout := func() {
			__antithesis_instrumentation__.Notify(457868)
			switch ast.(type) {
			case *tree.CommitTransaction, *tree.RollbackTransaction:
				__antithesis_instrumentation__.Notify(457869)

			default:
				__antithesis_instrumentation__.Notify(457870)
				ex.mu.IdleInTransactionSessionTimeout = timeout{time.AfterFunc(
					ex.sessionData().IdleInTransactionSessionTimeout,
					ex.cancelSession,
				)}
			}
		}
		__antithesis_instrumentation__.Notify(457867)
		switch ex.machine.CurState().(type) {
		case stateAborted, stateCommitWait:
			__antithesis_instrumentation__.Notify(457871)
			startIdleInTransactionSessionTimeout()
		case stateOpen:
			__antithesis_instrumentation__.Notify(457872)

			if !ex.implicitTxn() {
				__antithesis_instrumentation__.Notify(457873)
				startIdleInTransactionSessionTimeout()
			} else {
				__antithesis_instrumentation__.Notify(457874)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(457875)
	}
	__antithesis_instrumentation__.Notify(457841)

	return ev, payload, err
}

func (ex *connExecutor) recordFailure() {
	__antithesis_instrumentation__.Notify(457876)
	ex.metrics.EngineMetrics.FailureCount.Inc(1)
}

func (ex *connExecutor) execPortal(
	ctx context.Context,
	portal PreparedPortal,
	portalName string,
	stmtRes CommandResult,
	pinfo *tree.PlaceholderInfo,
	canAutoCommit bool,
) (ev fsm.Event, payload fsm.EventPayload, err error) {
	__antithesis_instrumentation__.Notify(457877)
	switch ex.machine.CurState().(type) {
	case stateOpen:
		__antithesis_instrumentation__.Notify(457878)

		if portal.exhausted {
			__antithesis_instrumentation__.Notify(457882)
			return nil, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(457883)
		}
		__antithesis_instrumentation__.Notify(457879)
		ev, payload, err = ex.execStmt(ctx, portal.Stmt.Statement, portal.Stmt, pinfo, stmtRes, canAutoCommit)

		if _, ok := ex.extraTxnState.prepStmtsNamespace.portals[portalName]; ok {
			__antithesis_instrumentation__.Notify(457884)
			ex.exhaustPortal(portalName)
		} else {
			__antithesis_instrumentation__.Notify(457885)
		}
		__antithesis_instrumentation__.Notify(457880)
		return ev, payload, err

	default:
		__antithesis_instrumentation__.Notify(457881)
		return ex.execStmt(ctx, portal.Stmt.Statement, portal.Stmt, pinfo, stmtRes, canAutoCommit)
	}
}

func (ex *connExecutor) execStmtInOpenState(
	ctx context.Context,
	parserStmt parser.Statement,
	prepared *PreparedStatement,
	pinfo *tree.PlaceholderInfo,
	res RestrictedCommandResult,
	canAutoCommit bool,
) (retEv fsm.Event, retPayload fsm.EventPayload, retErr error) {
	__antithesis_instrumentation__.Notify(457886)
	ctx, sp := tracing.EnsureChildSpan(ctx, ex.server.cfg.AmbientCtx.Tracer, "sql query")

	sp.SetTag("statement", attribute.StringValue(parserStmt.SQL))
	defer sp.Finish()
	ast := parserStmt.AST
	ctx = withStatement(ctx, ast)

	makeErrEvent := func(err error) (fsm.Event, fsm.EventPayload, error) {
		__antithesis_instrumentation__.Notify(457909)
		ev, payload := ex.makeErrEvent(err, ast)
		return ev, payload, nil
	}
	__antithesis_instrumentation__.Notify(457887)

	var stmt Statement
	queryID := ex.generateID()

	err := ex.extraTxnState.descCollection.MaybeUpdateDeadline(ctx, ex.state.mu.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(457910)
		return makeErrEvent(err)
	} else {
		__antithesis_instrumentation__.Notify(457911)
	}
	__antithesis_instrumentation__.Notify(457888)
	os := ex.machine.CurState().(stateOpen)

	isExtendedProtocol := prepared != nil
	if isExtendedProtocol {
		__antithesis_instrumentation__.Notify(457912)
		stmt = makeStatementFromPrepared(prepared, queryID)
	} else {
		__antithesis_instrumentation__.Notify(457913)
		stmt = makeStatement(parserStmt, queryID)
	}
	__antithesis_instrumentation__.Notify(457889)

	ex.incrementStartedStmtCounter(ast)
	defer func() {
		__antithesis_instrumentation__.Notify(457914)
		if retErr == nil && func() bool {
			__antithesis_instrumentation__.Notify(457915)
			return !payloadHasError(retPayload) == true
		}() == true {
			__antithesis_instrumentation__.Notify(457916)
			ex.incrementExecutedStmtCounter(ast)
		} else {
			__antithesis_instrumentation__.Notify(457917)
		}
	}()
	__antithesis_instrumentation__.Notify(457890)

	ex.state.mu.Lock()
	ex.state.mu.stmtCount++
	ex.state.mu.Unlock()

	var timeoutTicker *time.Timer
	queryTimedOut := false

	var doneAfterFunc chan struct{}

	if !ex.planner.EvalContext().HasPlaceholders() {
		__antithesis_instrumentation__.Notify(457918)
		ex.planner.EvalContext().Placeholders = pinfo
	} else {
		__antithesis_instrumentation__.Notify(457919)
	}
	__antithesis_instrumentation__.Notify(457891)

	ex.addActiveQuery(ast, formatWithPlaceholders(ast, ex.planner.EvalContext()), queryID, ex.state.cancel)
	if ex.executorType != executorTypeInternal {
		__antithesis_instrumentation__.Notify(457920)
		ex.metrics.EngineMetrics.SQLActiveStatements.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(457921)
	}
	__antithesis_instrumentation__.Notify(457892)

	defer func(ctx context.Context, res RestrictedCommandResult) {
		__antithesis_instrumentation__.Notify(457922)
		if timeoutTicker != nil {
			__antithesis_instrumentation__.Notify(457926)
			if !timeoutTicker.Stop() {
				__antithesis_instrumentation__.Notify(457927)

				<-doneAfterFunc
			} else {
				__antithesis_instrumentation__.Notify(457928)
			}
		} else {
			__antithesis_instrumentation__.Notify(457929)
		}
		__antithesis_instrumentation__.Notify(457923)
		ex.removeActiveQuery(queryID, ast)
		if ex.executorType != executorTypeInternal {
			__antithesis_instrumentation__.Notify(457930)
			ex.metrics.EngineMetrics.SQLActiveStatements.Dec(1)
		} else {
			__antithesis_instrumentation__.Notify(457931)
		}
		__antithesis_instrumentation__.Notify(457924)

		if res != nil && func() bool {
			__antithesis_instrumentation__.Notify(457932)
			return ctx.Err() != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(457933)
			return res.Err() != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(457934)

			retEv = eventNonRetriableErr{
				IsCommit: fsm.FromBool(isCommit(ast)),
			}
			res.SetError(cancelchecker.QueryCanceledError)
			retPayload = eventNonRetriableErrPayload{err: cancelchecker.QueryCanceledError}
		} else {
			__antithesis_instrumentation__.Notify(457935)
		}
		__antithesis_instrumentation__.Notify(457925)

		if queryTimedOut {
			__antithesis_instrumentation__.Notify(457936)

			retEv = eventNonRetriableErr{
				IsCommit: fsm.FromBool(isCommit(ast)),
			}
			res.SetError(sqlerrors.QueryTimeoutError)
			retPayload = eventNonRetriableErrPayload{err: sqlerrors.QueryTimeoutError}
		} else {
			__antithesis_instrumentation__.Notify(457937)
		}
	}(ctx, res)
	__antithesis_instrumentation__.Notify(457893)

	p := &ex.planner
	stmtTS := ex.server.cfg.Clock.PhysicalTime()
	ex.statsCollector.Reset(ex.applicationStats, ex.phaseTimes)
	ex.resetPlanner(ctx, p, ex.state.mu.txn, stmtTS)
	p.sessionDataMutatorIterator.paramStatusUpdater = res
	p.noticeSender = res
	ih := &p.instrumentation

	if e, ok := ast.(*tree.ExplainAnalyze); ok {
		__antithesis_instrumentation__.Notify(457938)
		switch e.Mode {
		case tree.ExplainDebug:
			__antithesis_instrumentation__.Notify(457940)
			telemetry.Inc(sqltelemetry.ExplainAnalyzeDebugUseCounter)
			ih.SetOutputMode(explainAnalyzeDebugOutput, explain.Flags{})

		case tree.ExplainPlan:
			__antithesis_instrumentation__.Notify(457941)
			telemetry.Inc(sqltelemetry.ExplainAnalyzeUseCounter)
			flags := explain.MakeFlags(&e.ExplainOptions)
			if ex.server.cfg.TestingKnobs.DeterministicExplain {
				__antithesis_instrumentation__.Notify(457946)
				flags.Redact = explain.RedactAll
			} else {
				__antithesis_instrumentation__.Notify(457947)
			}
			__antithesis_instrumentation__.Notify(457942)
			ih.SetOutputMode(explainAnalyzePlanOutput, flags)

		case tree.ExplainDistSQL:
			__antithesis_instrumentation__.Notify(457943)
			telemetry.Inc(sqltelemetry.ExplainAnalyzeDistSQLUseCounter)
			flags := explain.MakeFlags(&e.ExplainOptions)
			if ex.server.cfg.TestingKnobs.DeterministicExplain {
				__antithesis_instrumentation__.Notify(457948)
				flags.Redact = explain.RedactAll
			} else {
				__antithesis_instrumentation__.Notify(457949)
			}
			__antithesis_instrumentation__.Notify(457944)
			ih.SetOutputMode(explainAnalyzeDistSQLOutput, flags)

		default:
			__antithesis_instrumentation__.Notify(457945)
			return makeErrEvent(errors.AssertionFailedf("unsupported EXPLAIN ANALYZE mode %s", e.Mode))
		}
		__antithesis_instrumentation__.Notify(457939)

		stmt.AST = e.Statement
		ast = e.Statement

		stmt.ExpectedTypes = nil
	} else {
		__antithesis_instrumentation__.Notify(457950)
	}
	__antithesis_instrumentation__.Notify(457894)

	if e, ok := ast.(*tree.Execute); ok {
		__antithesis_instrumentation__.Notify(457951)

		name := e.Name.String()
		ps, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[name]
		if !ok {
			__antithesis_instrumentation__.Notify(457955)
			err := pgerror.Newf(
				pgcode.InvalidSQLStatementName,
				"prepared statement %q does not exist", name,
			)
			return makeErrEvent(err)
		} else {
			__antithesis_instrumentation__.Notify(457956)
		}
		__antithesis_instrumentation__.Notify(457952)
		var err error
		pinfo, err = ex.planner.fillInPlaceholders(ctx, ps, name, e.Params)
		if err != nil {
			__antithesis_instrumentation__.Notify(457957)
			return makeErrEvent(err)
		} else {
			__antithesis_instrumentation__.Notify(457958)
		}
		__antithesis_instrumentation__.Notify(457953)

		stmt.Statement = ps.Statement
		stmt.Prepared = ps
		stmt.ExpectedTypes = ps.Columns
		stmt.StmtNoConstants = ps.StatementNoConstants
		res.ResetStmtType(ps.AST)

		if e.DiscardRows {
			__antithesis_instrumentation__.Notify(457959)
			ih.SetDiscardRows()
		} else {
			__antithesis_instrumentation__.Notify(457960)
		}
		__antithesis_instrumentation__.Notify(457954)
		ast = stmt.Statement.AST
	} else {
		__antithesis_instrumentation__.Notify(457961)
	}
	__antithesis_instrumentation__.Notify(457895)

	var needFinish bool
	ctx, needFinish = ih.Setup(
		ctx, ex.server.cfg, ex.statsCollector, p, ex.stmtDiagnosticsRecorder,
		stmt.StmtNoConstants, os.ImplicitTxn.Get(), ex.extraTxnState.shouldCollectTxnExecutionStats,
	)
	if needFinish {
		__antithesis_instrumentation__.Notify(457962)
		sql := stmt.SQL
		defer func() {
			__antithesis_instrumentation__.Notify(457964)
			retErr = ih.Finish(
				ex.server.cfg,
				ex.statsCollector,
				&ex.extraTxnState.accumulatedStats,
				ex.extraTxnState.shouldCollectTxnExecutionStats,
				p,
				ast,
				sql,
				res,
				retErr,
			)
		}()
		__antithesis_instrumentation__.Notify(457963)

		p.extendedEvalCtx.Context = ctx
	} else {
		__antithesis_instrumentation__.Notify(457965)
	}
	__antithesis_instrumentation__.Notify(457896)

	if ex.sessionData().StmtTimeout > 0 && func() bool {
		__antithesis_instrumentation__.Notify(457966)
		return ast.StatementTag() != "SET" == true
	}() == true {
		__antithesis_instrumentation__.Notify(457967)
		timerDuration :=
			ex.sessionData().StmtTimeout - timeutil.Since(ex.phaseTimes.GetSessionPhaseTime(sessionphase.SessionQueryReceived))

		if timerDuration < 0 {
			__antithesis_instrumentation__.Notify(457969)
			queryTimedOut = true
			return makeErrEvent(sqlerrors.QueryTimeoutError)
		} else {
			__antithesis_instrumentation__.Notify(457970)
		}
		__antithesis_instrumentation__.Notify(457968)
		doneAfterFunc = make(chan struct{}, 1)
		timeoutTicker = time.AfterFunc(
			timerDuration,
			func() {
				__antithesis_instrumentation__.Notify(457971)
				ex.cancelQuery(queryID)
				queryTimedOut = true
				doneAfterFunc <- struct{}{}
			})
	} else {
		__antithesis_instrumentation__.Notify(457972)
	}
	__antithesis_instrumentation__.Notify(457897)

	defer func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(457973)
		if filter := ex.server.cfg.TestingKnobs.StatementFilter; retErr == nil && func() bool {
			__antithesis_instrumentation__.Notify(457976)
			return filter != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(457977)
			var execErr error
			if perr, ok := retPayload.(payloadWithError); ok {
				__antithesis_instrumentation__.Notify(457979)
				execErr = perr.errorCause()
			} else {
				__antithesis_instrumentation__.Notify(457980)
			}
			__antithesis_instrumentation__.Notify(457978)
			filter(ctx, ex.sessionData(), ast.String(), execErr)
		} else {
			__antithesis_instrumentation__.Notify(457981)
		}
		__antithesis_instrumentation__.Notify(457974)

		if retEv != nil || func() bool {
			__antithesis_instrumentation__.Notify(457982)
			return retErr != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(457983)
			return
		} else {
			__antithesis_instrumentation__.Notify(457984)
		}
		__antithesis_instrumentation__.Notify(457975)
		if canAutoCommit && func() bool {
			__antithesis_instrumentation__.Notify(457985)
			return !isExtendedProtocol == true
		}() == true {
			__antithesis_instrumentation__.Notify(457986)
			retEv, retPayload = ex.handleAutoCommit(ctx, ast)
		} else {
			__antithesis_instrumentation__.Notify(457987)
		}
	}(ctx)
	__antithesis_instrumentation__.Notify(457898)

	switch s := ast.(type) {
	case *tree.BeginTransaction:
		__antithesis_instrumentation__.Notify(457988)

		if os.ImplicitTxn.Get() {
			__antithesis_instrumentation__.Notify(457999)
			ex.sessionDataStack.PushTopClone()
			return eventTxnUpgradeToExplicit{}, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(458000)
		}
		__antithesis_instrumentation__.Notify(457989)
		return makeErrEvent(errTransactionInProgress)

	case *tree.CommitTransaction:
		__antithesis_instrumentation__.Notify(457990)

		ev, payload := ex.commitSQLTransaction(ctx, ast, ex.commitSQLTransactionInternal)
		return ev, payload, nil

	case *tree.RollbackTransaction:
		__antithesis_instrumentation__.Notify(457991)

		ev, payload := ex.rollbackSQLTransaction(ctx, s)
		return ev, payload, nil

	case *tree.Savepoint:
		__antithesis_instrumentation__.Notify(457992)
		return ex.execSavepointInOpenState(ctx, s, res)

	case *tree.ReleaseSavepoint:
		__antithesis_instrumentation__.Notify(457993)
		ev, payload := ex.execRelease(ctx, s, res)
		return ev, payload, nil

	case *tree.RollbackToSavepoint:
		__antithesis_instrumentation__.Notify(457994)
		ev, payload := ex.execRollbackToSavepointInOpenState(ctx, s, res)
		return ev, payload, nil

	case *tree.Prepare:
		__antithesis_instrumentation__.Notify(457995)

		name := s.Name.String()
		if _, ok := ex.extraTxnState.prepStmtsNamespace.prepStmts[name]; ok {
			__antithesis_instrumentation__.Notify(458001)
			err := pgerror.Newf(
				pgcode.DuplicatePreparedStatement,
				"prepared statement %q already exists", name,
			)
			return makeErrEvent(err)
		} else {
			__antithesis_instrumentation__.Notify(458002)
		}
		__antithesis_instrumentation__.Notify(457996)
		var typeHints tree.PlaceholderTypes
		if len(s.Types) > 0 {
			__antithesis_instrumentation__.Notify(458003)
			if len(s.Types) > stmt.NumPlaceholders {
				__antithesis_instrumentation__.Notify(458005)
				err := pgerror.Newf(pgcode.Syntax, "too many types provided")
				return makeErrEvent(err)
			} else {
				__antithesis_instrumentation__.Notify(458006)
			}
			__antithesis_instrumentation__.Notify(458004)
			typeHints = make(tree.PlaceholderTypes, stmt.NumPlaceholders)
			for i, t := range s.Types {
				__antithesis_instrumentation__.Notify(458007)
				resolved, err := tree.ResolveType(ctx, t, ex.planner.semaCtx.GetTypeResolver())
				if err != nil {
					__antithesis_instrumentation__.Notify(458009)
					return makeErrEvent(err)
				} else {
					__antithesis_instrumentation__.Notify(458010)
				}
				__antithesis_instrumentation__.Notify(458008)
				typeHints[i] = resolved
			}
		} else {
			__antithesis_instrumentation__.Notify(458011)
		}
		__antithesis_instrumentation__.Notify(457997)
		prepStmt := makeStatement(
			parser.Statement{

				SQL:             tree.AsStringWithFlags(s.Statement, tree.FmtParsable),
				AST:             s.Statement,
				NumPlaceholders: stmt.NumPlaceholders,
				NumAnnotations:  stmt.NumAnnotations,
			},
			ex.generateID(),
		)
		var rawTypeHints []oid.Oid
		if _, err := ex.addPreparedStmt(
			ctx, name, prepStmt, typeHints, rawTypeHints, PreparedStatementOriginSQL,
		); err != nil {
			__antithesis_instrumentation__.Notify(458012)
			return makeErrEvent(err)
		} else {
			__antithesis_instrumentation__.Notify(458013)
		}
		__antithesis_instrumentation__.Notify(457998)
		return nil, nil, nil
	}
	__antithesis_instrumentation__.Notify(457899)

	p.semaCtx.Annotations = tree.MakeAnnotations(stmt.NumAnnotations)

	if err := ex.handleAOST(ctx, ast); err != nil {
		__antithesis_instrumentation__.Notify(458014)
		return makeErrEvent(err)
	} else {
		__antithesis_instrumentation__.Notify(458015)
	}
	__antithesis_instrumentation__.Notify(457900)

	prevSteppingMode := ex.state.mu.txn.ConfigureStepping(ctx, kv.SteppingEnabled)
	defer func() {
		__antithesis_instrumentation__.Notify(458016)
		_ = ex.state.mu.txn.ConfigureStepping(ctx, prevSteppingMode)
	}()
	__antithesis_instrumentation__.Notify(457901)

	if err := ex.state.mu.txn.Step(ctx); err != nil {
		__antithesis_instrumentation__.Notify(458017)
		return makeErrEvent(err)
	} else {
		__antithesis_instrumentation__.Notify(458018)
	}
	__antithesis_instrumentation__.Notify(457902)

	if err := p.semaCtx.Placeholders.Assign(pinfo, stmt.NumPlaceholders); err != nil {
		__antithesis_instrumentation__.Notify(458019)
		return makeErrEvent(err)
	} else {
		__antithesis_instrumentation__.Notify(458020)
	}
	__antithesis_instrumentation__.Notify(457903)
	p.extendedEvalCtx.Placeholders = &p.semaCtx.Placeholders
	p.extendedEvalCtx.Annotations = &p.semaCtx.Annotations
	p.stmt = stmt
	p.cancelChecker.Reset(ctx)

	p.autoCommit = canAutoCommit && func() bool {
		__antithesis_instrumentation__.Notify(458021)
		return !ex.server.cfg.TestingKnobs.DisableAutoCommitDuringExec == true
	}() == true
	p.extendedEvalCtx.TxnIsSingleStmt = canAutoCommit && func() bool {
		__antithesis_instrumentation__.Notify(458022)
		return !ex.extraTxnState.firstStmtExecuted == true
	}() == true
	ex.extraTxnState.firstStmtExecuted = true

	var stmtThresholdSpan *tracing.Span
	alreadyRecording := ex.transitionCtx.sessionTracing.Enabled()
	stmtTraceThreshold := TraceStmtThreshold.Get(&ex.planner.execCfg.Settings.SV)
	var stmtCtx context.Context

	if !alreadyRecording && func() bool {
		__antithesis_instrumentation__.Notify(458023)
		return stmtTraceThreshold > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(458024)
		stmtCtx, stmtThresholdSpan = tracing.EnsureChildSpan(ctx, ex.server.cfg.AmbientCtx.Tracer, "trace-stmt-threshold", tracing.WithRecording(tracing.RecordingVerbose))
	} else {
		__antithesis_instrumentation__.Notify(458025)
		stmtCtx = ctx
	}
	__antithesis_instrumentation__.Notify(457904)

	if err := ex.dispatchToExecutionEngine(stmtCtx, p, res); err != nil {
		__antithesis_instrumentation__.Notify(458026)
		stmtThresholdSpan.Finish()
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(458027)
	}
	__antithesis_instrumentation__.Notify(457905)

	if stmtThresholdSpan != nil {
		__antithesis_instrumentation__.Notify(458028)
		stmtDur := timeutil.Since(ex.phaseTimes.GetSessionPhaseTime(sessionphase.SessionQueryReceived))
		needRecording := stmtTraceThreshold < stmtDur
		if needRecording {
			__antithesis_instrumentation__.Notify(458029)
			rec := stmtThresholdSpan.FinishAndGetRecording(tracing.RecordingVerbose)

			logTraceAboveThreshold(
				ctx,
				rec,
				fmt.Sprintf("SQL stmt %s", stmt.AST.String()),
				stmtTraceThreshold,
				stmtDur,
			)
		} else {
			__antithesis_instrumentation__.Notify(458030)
			stmtThresholdSpan.Finish()
		}
	} else {
		__antithesis_instrumentation__.Notify(458031)
	}
	__antithesis_instrumentation__.Notify(457906)

	if err := res.Err(); err != nil {
		__antithesis_instrumentation__.Notify(458032)
		return makeErrEvent(err)
	} else {
		__antithesis_instrumentation__.Notify(458033)
	}
	__antithesis_instrumentation__.Notify(457907)

	txn := ex.state.mu.txn

	if !os.ImplicitTxn.Get() && func() bool {
		__antithesis_instrumentation__.Notify(458034)
		return txn.IsSerializablePushAndRefreshNotPossible() == true
	}() == true {
		__antithesis_instrumentation__.Notify(458035)
		rc, canAutoRetry := ex.getRewindTxnCapability()
		if canAutoRetry {
			__antithesis_instrumentation__.Notify(458037)
			ev := eventRetriableErr{
				IsCommit:     fsm.FromBool(isCommit(ast)),
				CanAutoRetry: fsm.FromBool(canAutoRetry),
			}
			txn.ManualRestart(ctx, ex.server.cfg.Clock.Now())
			payload := eventRetriableErrPayload{
				err:    txn.PrepareRetryableError(ctx, "serializable transaction timestamp pushed (detected by connExecutor)"),
				rewCap: rc,
			}
			return ev, payload, nil
		} else {
			__antithesis_instrumentation__.Notify(458038)
		}
		__antithesis_instrumentation__.Notify(458036)
		log.VEventf(ctx, 2, "push detected for non-refreshable txn but auto-retry not possible")
	} else {
		__antithesis_instrumentation__.Notify(458039)
	}
	__antithesis_instrumentation__.Notify(457908)

	return nil, nil, nil
}

func (ex *connExecutor) handleAOST(ctx context.Context, stmt tree.Statement) error {
	__antithesis_instrumentation__.Notify(458040)
	if _, isNoTxn := ex.machine.CurState().(stateNoTxn); isNoTxn {
		__antithesis_instrumentation__.Notify(458047)
		return errors.AssertionFailedf(
			"cannot handle AOST clause without a transaction",
		)
	} else {
		__antithesis_instrumentation__.Notify(458048)
	}
	__antithesis_instrumentation__.Notify(458041)
	p := &ex.planner
	asOf, err := p.isAsOf(ctx, stmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(458049)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458050)
	}
	__antithesis_instrumentation__.Notify(458042)
	if asOf == nil {
		__antithesis_instrumentation__.Notify(458051)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(458052)
	}
	__antithesis_instrumentation__.Notify(458043)
	if ex.implicitTxn() {
		__antithesis_instrumentation__.Notify(458053)
		if p.extendedEvalCtx.AsOfSystemTime == nil {
			__antithesis_instrumentation__.Notify(458057)
			p.extendedEvalCtx.AsOfSystemTime = asOf
			if !asOf.BoundedStaleness {
				__antithesis_instrumentation__.Notify(458059)
				p.extendedEvalCtx.SetTxnTimestamp(asOf.Timestamp.GoTime())
				if err := ex.state.setHistoricalTimestamp(ctx, asOf.Timestamp); err != nil {
					__antithesis_instrumentation__.Notify(458060)
					return err
				} else {
					__antithesis_instrumentation__.Notify(458061)
				}
			} else {
				__antithesis_instrumentation__.Notify(458062)
			}
			__antithesis_instrumentation__.Notify(458058)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(458063)
		}
		__antithesis_instrumentation__.Notify(458054)
		if *p.extendedEvalCtx.AsOfSystemTime == *asOf {
			__antithesis_instrumentation__.Notify(458064)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(458065)
		}
		__antithesis_instrumentation__.Notify(458055)
		if p.extendedEvalCtx.AsOfSystemTime.BoundedStaleness {
			__antithesis_instrumentation__.Notify(458066)
			if !p.extendedEvalCtx.AsOfSystemTime.MaxTimestampBound.IsEmpty() {
				__antithesis_instrumentation__.Notify(458068)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(458069)
			}
			__antithesis_instrumentation__.Notify(458067)
			return errors.AssertionFailedf("expected bounded_staleness set with a max_timestamp_bound")
		} else {
			__antithesis_instrumentation__.Notify(458070)
		}
		__antithesis_instrumentation__.Notify(458056)
		return errors.AssertionFailedf(
			"cannot specify AS OF SYSTEM TIME with different timestamps. expected: %s, got: %s",
			p.extendedEvalCtx.AsOfSystemTime.Timestamp, asOf.Timestamp,
		)
	} else {
		__antithesis_instrumentation__.Notify(458071)
	}
	__antithesis_instrumentation__.Notify(458044)

	if asOf.BoundedStaleness {
		__antithesis_instrumentation__.Notify(458072)
		return pgerror.Newf(
			pgcode.FeatureNotSupported,
			"cannot use a bounded staleness query in a transaction",
		)
	} else {
		__antithesis_instrumentation__.Notify(458073)
	}
	__antithesis_instrumentation__.Notify(458045)
	if readTs := ex.state.getReadTimestamp(); asOf.Timestamp != readTs {
		__antithesis_instrumentation__.Notify(458074)
		err = pgerror.Newf(pgcode.Syntax,
			"inconsistent AS OF SYSTEM TIME timestamp; expected: %s, got: %s", readTs, asOf.Timestamp)
		err = errors.WithHint(err, "try SET TRANSACTION AS OF SYSTEM TIME")
		return err
	} else {
		__antithesis_instrumentation__.Notify(458075)
	}
	__antithesis_instrumentation__.Notify(458046)
	p.extendedEvalCtx.AsOfSystemTime = asOf
	return nil
}

func formatWithPlaceholders(ast tree.Statement, evalCtx *tree.EvalContext) string {
	__antithesis_instrumentation__.Notify(458076)
	var fmtCtx *tree.FmtCtx
	fmtFlags := tree.FmtSimple

	if evalCtx.HasPlaceholders() {
		__antithesis_instrumentation__.Notify(458078)
		fmtCtx = evalCtx.FmtCtx(
			fmtFlags,
			tree.FmtPlaceholderFormat(func(ctx *tree.FmtCtx, placeholder *tree.Placeholder) {
				__antithesis_instrumentation__.Notify(458079)
				d, err := placeholder.Eval(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(458081)

					ctx.Printf("$%d", placeholder.Idx+1)
					return
				} else {
					__antithesis_instrumentation__.Notify(458082)
				}
				__antithesis_instrumentation__.Notify(458080)
				d.Format(ctx)
			}),
		)
	} else {
		__antithesis_instrumentation__.Notify(458083)
		fmtCtx = evalCtx.FmtCtx(fmtFlags)
	}
	__antithesis_instrumentation__.Notify(458077)

	fmtCtx.FormatNode(ast)

	return fmtCtx.CloseAndGetString()
}

func (ex *connExecutor) checkDescriptorTwoVersionInvariant(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(458084)
	var inRetryBackoff func()
	if knobs := ex.server.cfg.SchemaChangerTestingKnobs; knobs != nil {
		__antithesis_instrumentation__.Notify(458088)
		inRetryBackoff = knobs.TwoVersionLeaseViolation
	} else {
		__antithesis_instrumentation__.Notify(458089)
	}
	__antithesis_instrumentation__.Notify(458085)

	if err := descs.CheckSpanCountLimit(
		ctx,
		&ex.extraTxnState.descCollection,
		ex.server.cfg.SpanConfigSplitter,
		ex.server.cfg.SpanConfigLimiter,
		ex.state.mu.txn,
	); err != nil {
		__antithesis_instrumentation__.Notify(458090)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458091)
	}
	__antithesis_instrumentation__.Notify(458086)

	retryErr, err := descs.CheckTwoVersionInvariant(
		ctx,
		ex.server.cfg.Clock,
		ex.server.cfg.InternalExecutor,
		&ex.extraTxnState.descCollection,
		ex.state.mu.txn,
		inRetryBackoff,
	)
	if retryErr {
		__antithesis_instrumentation__.Notify(458092)
		if newTransactionErr := ex.resetTransactionOnSchemaChangeRetry(ctx); newTransactionErr != nil {
			__antithesis_instrumentation__.Notify(458093)
			return newTransactionErr
		} else {
			__antithesis_instrumentation__.Notify(458094)
		}
	} else {
		__antithesis_instrumentation__.Notify(458095)
	}
	__antithesis_instrumentation__.Notify(458087)
	return err
}

func (ex *connExecutor) resetTransactionOnSchemaChangeRetry(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(458096)
	ex.state.mu.Lock()
	defer ex.state.mu.Unlock()
	userPriority := ex.state.mu.txn.UserPriority()
	ex.state.mu.txn = kv.NewTxnWithSteppingEnabled(ctx, ex.transitionCtx.db,
		ex.transitionCtx.nodeIDOrZero, ex.QualityOfService())
	return ex.state.mu.txn.SetUserPriority(userPriority)
}

func (ex *connExecutor) commitSQLTransaction(
	ctx context.Context, ast tree.Statement, commitFn func(ctx context.Context) error,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458097)
	ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionStartTransactionCommit, timeutil.Now())
	if err := commitFn(ctx); err != nil {
		__antithesis_instrumentation__.Notify(458100)
		return ex.makeErrEvent(err, ast)
	} else {
		__antithesis_instrumentation__.Notify(458101)
	}
	__antithesis_instrumentation__.Notify(458098)
	ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionEndTransactionCommit, timeutil.Now())
	if err := ex.reportSessionDataChanges(func() error {
		__antithesis_instrumentation__.Notify(458102)
		ex.sessionDataStack.PopAll()
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(458103)
		return ex.makeErrEvent(err, ast)
	} else {
		__antithesis_instrumentation__.Notify(458104)
	}
	__antithesis_instrumentation__.Notify(458099)
	return eventTxnFinishCommitted{}, nil
}

func (ex *connExecutor) reportSessionDataChanges(fn func() error) error {
	__antithesis_instrumentation__.Notify(458105)
	before := ex.sessionDataStack.Top()
	if err := fn(); err != nil {
		__antithesis_instrumentation__.Notify(458110)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458111)
	}
	__antithesis_instrumentation__.Notify(458106)
	after := ex.sessionDataStack.Top()
	if ex.dataMutatorIterator.paramStatusUpdater != nil {
		__antithesis_instrumentation__.Notify(458112)
		for _, param := range bufferableParamStatusUpdates {
			__antithesis_instrumentation__.Notify(458113)
			_, v, err := getSessionVar(param.lowerName, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(458117)
				return err
			} else {
				__antithesis_instrumentation__.Notify(458118)
			}
			__antithesis_instrumentation__.Notify(458114)
			if v.Equal == nil {
				__antithesis_instrumentation__.Notify(458119)
				return errors.AssertionFailedf("Equal for %s must be set", param.name)
			} else {
				__antithesis_instrumentation__.Notify(458120)
			}
			__antithesis_instrumentation__.Notify(458115)
			if v.GetFromSessionData == nil {
				__antithesis_instrumentation__.Notify(458121)
				return errors.AssertionFailedf("GetFromSessionData for %s must be set", param.name)
			} else {
				__antithesis_instrumentation__.Notify(458122)
			}
			__antithesis_instrumentation__.Notify(458116)
			if !v.Equal(before, after) {
				__antithesis_instrumentation__.Notify(458123)
				ex.dataMutatorIterator.paramStatusUpdater.BufferParamStatusUpdate(
					param.name,
					v.GetFromSessionData(after),
				)
			} else {
				__antithesis_instrumentation__.Notify(458124)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(458125)
	}
	__antithesis_instrumentation__.Notify(458107)
	if before.DefaultIntSize != after.DefaultIntSize && func() bool {
		__antithesis_instrumentation__.Notify(458126)
		return ex.dataMutatorIterator.onDefaultIntSizeChange != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(458127)
		ex.dataMutatorIterator.onDefaultIntSizeChange(after.DefaultIntSize)
	} else {
		__antithesis_instrumentation__.Notify(458128)
	}
	__antithesis_instrumentation__.Notify(458108)
	if before.ApplicationName != after.ApplicationName && func() bool {
		__antithesis_instrumentation__.Notify(458129)
		return ex.dataMutatorIterator.onApplicationNameChange != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(458130)
		ex.dataMutatorIterator.onApplicationNameChange(after.ApplicationName)
	} else {
		__antithesis_instrumentation__.Notify(458131)
	}
	__antithesis_instrumentation__.Notify(458109)
	return nil
}

func (ex *connExecutor) commitSQLTransactionInternal(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(458132)
	ctx, sp := tracing.EnsureChildSpan(ctx, ex.server.cfg.AmbientCtx.Tracer, "commit sql txn")
	defer sp.Finish()

	if err := ex.createJobs(ctx); err != nil {
		__antithesis_instrumentation__.Notify(458139)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458140)
	}
	__antithesis_instrumentation__.Notify(458133)

	if ex.extraTxnState.schemaChangerState.mode != sessiondatapb.UseNewSchemaChangerOff {
		__antithesis_instrumentation__.Notify(458141)
		if err := ex.runPreCommitStages(ctx); err != nil {
			__antithesis_instrumentation__.Notify(458142)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458143)
		}
	} else {
		__antithesis_instrumentation__.Notify(458144)
	}
	__antithesis_instrumentation__.Notify(458134)

	if err := ex.extraTxnState.descCollection.ValidateUncommittedDescriptors(ctx, ex.state.mu.txn); err != nil {
		__antithesis_instrumentation__.Notify(458145)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458146)
	}
	__antithesis_instrumentation__.Notify(458135)

	if err := ex.checkDescriptorTwoVersionInvariant(ctx); err != nil {
		__antithesis_instrumentation__.Notify(458147)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458148)
	}
	__antithesis_instrumentation__.Notify(458136)

	if err := ex.state.mu.txn.Commit(ctx); err != nil {
		__antithesis_instrumentation__.Notify(458149)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458150)
	}
	__antithesis_instrumentation__.Notify(458137)

	if descs := ex.extraTxnState.descCollection.GetDescriptorsWithNewVersion(); descs != nil {
		__antithesis_instrumentation__.Notify(458151)
		ex.extraTxnState.descCollection.ReleaseLeases(ctx)
	} else {
		__antithesis_instrumentation__.Notify(458152)
	}
	__antithesis_instrumentation__.Notify(458138)
	return nil
}

func (ex *connExecutor) createJobs(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(458153)
	if len(ex.extraTxnState.schemaChangeJobRecords) == 0 {
		__antithesis_instrumentation__.Notify(458157)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(458158)
	}
	__antithesis_instrumentation__.Notify(458154)
	var records []*jobs.Record
	for _, record := range ex.extraTxnState.schemaChangeJobRecords {
		__antithesis_instrumentation__.Notify(458159)
		records = append(records, record)
	}
	__antithesis_instrumentation__.Notify(458155)
	jobIDs, err := ex.server.cfg.JobRegistry.CreateJobsWithTxn(ctx, ex.planner.extendedEvalCtx.Txn, records)
	if err != nil {
		__antithesis_instrumentation__.Notify(458160)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458161)
	}
	__antithesis_instrumentation__.Notify(458156)
	ex.planner.extendedEvalCtx.Jobs.add(jobIDs...)
	return nil
}

func (ex *connExecutor) rollbackSQLTransaction(
	ctx context.Context, stmt tree.Statement,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458162)
	if err := ex.state.mu.txn.Rollback(ctx); err != nil {
		__antithesis_instrumentation__.Notify(458165)
		log.Warningf(ctx, "txn rollback failed: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(458166)
	}
	__antithesis_instrumentation__.Notify(458163)
	if err := ex.reportSessionDataChanges(func() error {
		__antithesis_instrumentation__.Notify(458167)
		ex.sessionDataStack.PopAll()
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(458168)
		return ex.makeErrEvent(err, stmt)
	} else {
		__antithesis_instrumentation__.Notify(458169)
	}
	__antithesis_instrumentation__.Notify(458164)

	return eventTxnFinishAborted{}, nil
}

func (ex *connExecutor) dispatchToExecutionEngine(
	ctx context.Context, planner *planner, res RestrictedCommandResult,
) error {
	__antithesis_instrumentation__.Notify(458170)
	stmt := planner.stmt
	ex.sessionTracing.TracePlanStart(ctx, stmt.AST.StatementTag())
	ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.PlannerStartLogicalPlan, timeutil.Now())

	if adminAuditLog := adminAuditLogEnabled.Get(
		&ex.planner.execCfg.Settings.SV,
	); adminAuditLog {
		__antithesis_instrumentation__.Notify(458187)
		if !ex.extraTxnState.hasAdminRoleCache.IsSet {
			__antithesis_instrumentation__.Notify(458188)
			hasAdminRole, err := ex.planner.HasAdminRole(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(458190)
				return err
			} else {
				__antithesis_instrumentation__.Notify(458191)
			}
			__antithesis_instrumentation__.Notify(458189)
			ex.extraTxnState.hasAdminRoleCache.HasAdminRole = hasAdminRole
			ex.extraTxnState.hasAdminRoleCache.IsSet = true
		} else {
			__antithesis_instrumentation__.Notify(458192)
		}
	} else {
		__antithesis_instrumentation__.Notify(458193)
	}
	__antithesis_instrumentation__.Notify(458171)

	err := ex.makeExecPlan(ctx, planner)

	defer planner.curPlan.close(ctx)

	if planner.autoCommit {
		__antithesis_instrumentation__.Notify(458194)
		planner.curPlan.flags.Set(planFlagImplicitTxn)
	} else {
		__antithesis_instrumentation__.Notify(458195)
	}
	__antithesis_instrumentation__.Notify(458172)

	if planner.curPlan.avoidBuffering || func() bool {
		__antithesis_instrumentation__.Notify(458196)
		return ex.sessionData().AvoidBuffering == true
	}() == true {
		__antithesis_instrumentation__.Notify(458197)
		res.DisableBuffering()
	} else {
		__antithesis_instrumentation__.Notify(458198)
	}
	__antithesis_instrumentation__.Notify(458173)

	defer func() {
		__antithesis_instrumentation__.Notify(458199)
		planner.maybeLogStatement(
			ctx,
			ex.executorType,
			int(atomic.LoadInt32(ex.extraTxnState.atomicAutoRetryCounter)),
			ex.extraTxnState.txnCounter,
			res.RowsAffected(),
			res.Err(),
			ex.statsCollector.PhaseTimes().GetSessionPhaseTime(sessionphase.SessionQueryReceived),
			&ex.extraTxnState.hasAdminRoleCache,
			ex.server.TelemetryLoggingMetrics,
		)
	}()
	__antithesis_instrumentation__.Notify(458174)

	ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.PlannerEndLogicalPlan, timeutil.Now())
	ex.sessionTracing.TracePlanEnd(ctx, err)

	if err != nil {
		__antithesis_instrumentation__.Notify(458200)
		res.SetError(err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(458201)
	}
	__antithesis_instrumentation__.Notify(458175)

	var cols colinfo.ResultColumns
	if stmt.AST.StatementReturnType() == tree.Rows {
		__antithesis_instrumentation__.Notify(458202)
		cols = planner.curPlan.main.planColumns()
	} else {
		__antithesis_instrumentation__.Notify(458203)
	}
	__antithesis_instrumentation__.Notify(458176)
	if err := ex.initStatementResult(ctx, res, stmt.AST, cols); err != nil {
		__antithesis_instrumentation__.Notify(458204)
		res.SetError(err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(458205)
	}
	__antithesis_instrumentation__.Notify(458177)

	ex.sessionTracing.TracePlanCheckStart(ctx)
	distributePlan := getPlanDistribution(
		ctx, planner, planner.execCfg.NodeID, ex.sessionData().DistSQLMode, planner.curPlan.main,
	)
	ex.sessionTracing.TracePlanCheckEnd(ctx, nil, distributePlan.WillDistribute())

	if ex.server.cfg.TestingKnobs.BeforeExecute != nil {
		__antithesis_instrumentation__.Notify(458206)
		ex.server.cfg.TestingKnobs.BeforeExecute(ctx, stmt.String())
	} else {
		__antithesis_instrumentation__.Notify(458207)
	}
	__antithesis_instrumentation__.Notify(458178)

	ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.PlannerStartExecStmt, timeutil.Now())

	ex.mu.Lock()
	queryMeta, ok := ex.mu.ActiveQueries[stmt.QueryID]
	if !ok {
		__antithesis_instrumentation__.Notify(458208)
		ex.mu.Unlock()
		panic(errors.AssertionFailedf("query %d not in registry", stmt.QueryID))
	} else {
		__antithesis_instrumentation__.Notify(458209)
	}
	__antithesis_instrumentation__.Notify(458179)
	queryMeta.phase = executing

	queryMeta.isDistributed = distributePlan.WillDistribute()
	progAtomic := &queryMeta.progressAtomic
	ex.mu.Unlock()

	planner.curPlan.flags.Set(planFlagExecDone)
	if !planner.ExecCfg().Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(458210)
		planner.curPlan.flags.Set(planFlagTenant)
	} else {
		__antithesis_instrumentation__.Notify(458211)
	}
	__antithesis_instrumentation__.Notify(458180)

	switch distributePlan {
	case physicalplan.FullyDistributedPlan:
		__antithesis_instrumentation__.Notify(458212)
		planner.curPlan.flags.Set(planFlagFullyDistributed)
	case physicalplan.PartiallyDistributedPlan:
		__antithesis_instrumentation__.Notify(458213)
		planner.curPlan.flags.Set(planFlagPartiallyDistributed)
	default:
		__antithesis_instrumentation__.Notify(458214)
		planner.curPlan.flags.Set(planFlagNotDistributed)
	}
	__antithesis_instrumentation__.Notify(458181)

	ex.sessionTracing.TraceRetryInformation(ctx, int(atomic.LoadInt32(ex.extraTxnState.atomicAutoRetryCounter)), ex.extraTxnState.autoRetryReason)
	if ex.server.cfg.TestingKnobs.OnTxnRetry != nil && func() bool {
		__antithesis_instrumentation__.Notify(458215)
		return ex.extraTxnState.autoRetryReason != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(458216)
		ex.server.cfg.TestingKnobs.OnTxnRetry(ex.extraTxnState.autoRetryReason, planner.EvalContext())
	} else {
		__antithesis_instrumentation__.Notify(458217)
	}
	__antithesis_instrumentation__.Notify(458182)
	distribute := DistributionType(DistributionTypeNone)
	if distributePlan.WillDistribute() {
		__antithesis_instrumentation__.Notify(458218)
		distribute = DistributionTypeSystemTenantOnly
	} else {
		__antithesis_instrumentation__.Notify(458219)
	}
	__antithesis_instrumentation__.Notify(458183)
	ex.sessionTracing.TraceExecStart(ctx, "distributed")
	stats, err := ex.execWithDistSQLEngine(
		ctx, planner, stmt.AST.StatementReturnType(), res, distribute, progAtomic,
	)
	if res.Err() == nil {
		__antithesis_instrumentation__.Notify(458220)

		const numTxnRetryErrors = 3
		if ex.sessionData().InjectRetryErrorsEnabled && func() bool {
			__antithesis_instrumentation__.Notify(458221)
			return stmt.AST.StatementTag() != "SET" == true
		}() == true {
			__antithesis_instrumentation__.Notify(458222)
			if planner.Txn().Epoch() < ex.state.lastEpoch+numTxnRetryErrors {
				__antithesis_instrumentation__.Notify(458223)
				retryErr := planner.Txn().GenerateForcedRetryableError(
					ctx, "injected by `inject_retry_errors_enabled` session variable")
				res.SetError(retryErr)
			} else {
				__antithesis_instrumentation__.Notify(458224)
				ex.state.lastEpoch = planner.Txn().Epoch()
			}
		} else {
			__antithesis_instrumentation__.Notify(458225)
		}
	} else {
		__antithesis_instrumentation__.Notify(458226)
	}
	__antithesis_instrumentation__.Notify(458184)
	ex.sessionTracing.TraceExecEnd(ctx, res.Err(), res.RowsAffected())
	ex.statsCollector.PhaseTimes().SetSessionPhaseTime(sessionphase.PlannerEndExecStmt, timeutil.Now())

	ex.extraTxnState.rowsRead += stats.rowsRead
	ex.extraTxnState.bytesRead += stats.bytesRead
	ex.extraTxnState.rowsWritten += stats.rowsWritten

	ex.recordStatementSummary(
		ctx, planner,
		int(atomic.LoadInt32(ex.extraTxnState.atomicAutoRetryCounter)), res.RowsAffected(), res.Err(), stats,
	)
	if ex.server.cfg.TestingKnobs.AfterExecute != nil {
		__antithesis_instrumentation__.Notify(458227)
		ex.server.cfg.TestingKnobs.AfterExecute(ctx, stmt.String(), res.Err())
	} else {
		__antithesis_instrumentation__.Notify(458228)
	}
	__antithesis_instrumentation__.Notify(458185)

	if limitsErr := ex.handleTxnRowsWrittenReadLimits(ctx); limitsErr != nil && func() bool {
		__antithesis_instrumentation__.Notify(458229)
		return res.Err() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(458230)
		res.SetError(limitsErr)
	} else {
		__antithesis_instrumentation__.Notify(458231)
	}
	__antithesis_instrumentation__.Notify(458186)

	return err
}

type txnRowsWrittenLimitErr struct {
	eventpb.CommonTxnRowsLimitDetails
}

var _ error = &txnRowsWrittenLimitErr{}
var _ fmt.Formatter = &txnRowsWrittenLimitErr{}
var _ errors.SafeFormatter = &txnRowsWrittenLimitErr{}

func (e *txnRowsWrittenLimitErr) Error() string {
	__antithesis_instrumentation__.Notify(458232)
	return e.CommonTxnRowsLimitDetails.Error("written")
}

func (e *txnRowsWrittenLimitErr) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(458233)
	errors.FormatError(e, s, verb)
}

func (e *txnRowsWrittenLimitErr) SafeFormatError(p errors.Printer) (next error) {
	__antithesis_instrumentation__.Notify(458234)
	return e.CommonTxnRowsLimitDetails.SafeFormatError(p, "written")
}

type txnRowsReadLimitErr struct {
	eventpb.CommonTxnRowsLimitDetails
}

var _ error = &txnRowsReadLimitErr{}
var _ fmt.Formatter = &txnRowsReadLimitErr{}
var _ errors.SafeFormatter = &txnRowsReadLimitErr{}

func (e *txnRowsReadLimitErr) Error() string {
	__antithesis_instrumentation__.Notify(458235)
	return e.CommonTxnRowsLimitDetails.Error("read")
}

func (e *txnRowsReadLimitErr) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(458236)
	errors.FormatError(e, s, verb)
}

func (e *txnRowsReadLimitErr) SafeFormatError(p errors.Printer) (next error) {
	__antithesis_instrumentation__.Notify(458237)
	return e.CommonTxnRowsLimitDetails.SafeFormatError(p, "read")
}

func (ex *connExecutor) handleTxnRowsGuardrails(
	ctx context.Context,
	numRows, logLimit, errLimit int64,
	alreadyLogged *bool,
	isRead bool,
	logCounter, errCounter *metric.Counter,
) error {
	__antithesis_instrumentation__.Notify(458238)
	var err error
	shouldLog := logLimit != 0 && func() bool {
		__antithesis_instrumentation__.Notify(458244)
		return numRows > logLimit == true
	}() == true
	shouldErr := errLimit != 0 && func() bool {
		__antithesis_instrumentation__.Notify(458245)
		return numRows > errLimit == true
	}() == true
	if !shouldLog && func() bool {
		__antithesis_instrumentation__.Notify(458246)
		return !shouldErr == true
	}() == true {
		__antithesis_instrumentation__.Notify(458247)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(458248)
	}
	__antithesis_instrumentation__.Notify(458239)
	commonTxnRowsLimitDetails := eventpb.CommonTxnRowsLimitDetails{
		TxnID:     ex.state.mu.txn.ID().String(),
		SessionID: ex.sessionID.String(),
		NumRows:   numRows,
	}
	if shouldErr && func() bool {
		__antithesis_instrumentation__.Notify(458249)
		return ex.executorType == executorTypeInternal == true
	}() == true {
		__antithesis_instrumentation__.Notify(458250)

		shouldLog = true
		shouldErr = false
	} else {
		__antithesis_instrumentation__.Notify(458251)
	}
	__antithesis_instrumentation__.Notify(458240)
	if *alreadyLogged {
		__antithesis_instrumentation__.Notify(458252)

		shouldLog = false
	} else {
		__antithesis_instrumentation__.Notify(458253)
		*alreadyLogged = shouldLog
	}
	__antithesis_instrumentation__.Notify(458241)
	if shouldLog {
		__antithesis_instrumentation__.Notify(458254)
		commonSQLEventDetails := ex.planner.getCommonSQLEventDetails(defaultRedactionOptions)
		var event eventpb.EventPayload
		if ex.executorType == executorTypeInternal {
			__antithesis_instrumentation__.Notify(458255)
			if isRead {
				__antithesis_instrumentation__.Notify(458256)
				event = &eventpb.TxnRowsReadLimitInternal{
					CommonSQLEventDetails:     commonSQLEventDetails,
					CommonTxnRowsLimitDetails: commonTxnRowsLimitDetails,
				}
			} else {
				__antithesis_instrumentation__.Notify(458257)
				event = &eventpb.TxnRowsWrittenLimitInternal{
					CommonSQLEventDetails:     commonSQLEventDetails,
					CommonTxnRowsLimitDetails: commonTxnRowsLimitDetails,
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(458258)
			if isRead {
				__antithesis_instrumentation__.Notify(458260)
				event = &eventpb.TxnRowsReadLimit{
					CommonSQLEventDetails:     commonSQLEventDetails,
					CommonTxnRowsLimitDetails: commonTxnRowsLimitDetails,
				}
			} else {
				__antithesis_instrumentation__.Notify(458261)
				event = &eventpb.TxnRowsWrittenLimit{
					CommonSQLEventDetails:     commonSQLEventDetails,
					CommonTxnRowsLimitDetails: commonTxnRowsLimitDetails,
				}
			}
			__antithesis_instrumentation__.Notify(458259)
			log.StructuredEvent(ctx, event)
			logCounter.Inc(1)
		}
	} else {
		__antithesis_instrumentation__.Notify(458262)
	}
	__antithesis_instrumentation__.Notify(458242)
	if shouldErr {
		__antithesis_instrumentation__.Notify(458263)
		if isRead {
			__antithesis_instrumentation__.Notify(458265)
			err = &txnRowsReadLimitErr{CommonTxnRowsLimitDetails: commonTxnRowsLimitDetails}
		} else {
			__antithesis_instrumentation__.Notify(458266)
			err = &txnRowsWrittenLimitErr{CommonTxnRowsLimitDetails: commonTxnRowsLimitDetails}
		}
		__antithesis_instrumentation__.Notify(458264)
		err = pgerror.WithCandidateCode(err, pgcode.ProgramLimitExceeded)
		errCounter.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(458267)
	}
	__antithesis_instrumentation__.Notify(458243)
	return err
}

func (ex *connExecutor) handleTxnRowsWrittenReadLimits(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(458268)

	sd := ex.sessionData()
	writtenErr := ex.handleTxnRowsGuardrails(
		ctx,
		ex.extraTxnState.rowsWritten,
		sd.TxnRowsWrittenLog,
		sd.TxnRowsWrittenErr,
		&ex.extraTxnState.rowsWrittenLogged,
		false,
		ex.metrics.GuardrailMetrics.TxnRowsWrittenLogCount,
		ex.metrics.GuardrailMetrics.TxnRowsWrittenErrCount,
	)
	readErr := ex.handleTxnRowsGuardrails(
		ctx,
		ex.extraTxnState.rowsRead,
		sd.TxnRowsReadLog,
		sd.TxnRowsReadErr,
		&ex.extraTxnState.rowsReadLogged,
		true,
		ex.metrics.GuardrailMetrics.TxnRowsReadLogCount,
		ex.metrics.GuardrailMetrics.TxnRowsReadErrCount,
	)
	return errors.CombineErrors(writtenErr, readErr)
}

func (ex *connExecutor) makeExecPlan(ctx context.Context, planner *planner) error {
	__antithesis_instrumentation__.Notify(458269)
	if err := planner.makeOptimizerPlan(ctx); err != nil {
		__antithesis_instrumentation__.Notify(458273)
		log.VEventf(ctx, 1, "optimizer plan failed: %v", err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458274)
	}
	__antithesis_instrumentation__.Notify(458270)

	flags := planner.curPlan.flags

	if flags.IsSet(planFlagContainsFullIndexScan) || func() bool {
		__antithesis_instrumentation__.Notify(458275)
		return flags.IsSet(planFlagContainsFullTableScan) == true
	}() == true {
		__antithesis_instrumentation__.Notify(458276)
		if ex.executorType == executorTypeExec && func() bool {
			__antithesis_instrumentation__.Notify(458278)
			return planner.EvalContext().SessionData().DisallowFullTableScans == true
		}() == true {
			__antithesis_instrumentation__.Notify(458279)
			hasLargeScan := flags.IsSet(planFlagContainsLargeFullIndexScan) || func() bool {
				__antithesis_instrumentation__.Notify(458280)
				return flags.IsSet(planFlagContainsLargeFullTableScan) == true
			}() == true
			if hasLargeScan {
				__antithesis_instrumentation__.Notify(458281)

				ex.metrics.EngineMetrics.FullTableOrIndexScanRejectedCount.Inc(1)
				return errors.WithHint(
					pgerror.Newf(pgcode.TooManyRows,
						"query `%s` contains a full table/index scan which is explicitly disallowed",
						planner.stmt.SQL),
					"try overriding the `disallow_full_table_scans` or increasing the `large_full_scan_rows` cluster/session settings",
				)
			} else {
				__antithesis_instrumentation__.Notify(458282)
			}
		} else {
			__antithesis_instrumentation__.Notify(458283)
		}
		__antithesis_instrumentation__.Notify(458277)
		ex.metrics.EngineMetrics.FullTableOrIndexScanCount.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(458284)
	}
	__antithesis_instrumentation__.Notify(458271)

	if flags.IsSet(planFlagIsDDL) {
		__antithesis_instrumentation__.Notify(458285)
		ex.extraTxnState.numDDL++
	} else {
		__antithesis_instrumentation__.Notify(458286)
	}
	__antithesis_instrumentation__.Notify(458272)

	return nil
}

type topLevelQueryStats struct {
	bytesRead int64

	rowsRead int64

	rowsWritten int64
}

func (ex *connExecutor) execWithDistSQLEngine(
	ctx context.Context,
	planner *planner,
	stmtType tree.StatementReturnType,
	res RestrictedCommandResult,
	distribute DistributionType,
	progressAtomic *uint64,
) (topLevelQueryStats, error) {
	__antithesis_instrumentation__.Notify(458287)
	var testingPushCallback func(rowenc.EncDatumRow, *execinfrapb.ProducerMetadata)
	if ex.server.cfg.TestingKnobs.DistSQLReceiverPushCallbackFactory != nil {
		__antithesis_instrumentation__.Notify(458293)
		testingPushCallback = ex.server.cfg.TestingKnobs.DistSQLReceiverPushCallbackFactory(planner.stmt.SQL)
	} else {
		__antithesis_instrumentation__.Notify(458294)
	}
	__antithesis_instrumentation__.Notify(458288)
	recv := MakeDistSQLReceiver(
		ctx, res, stmtType,
		ex.server.cfg.RangeDescriptorCache,
		planner.txn,
		ex.server.cfg.Clock,
		&ex.sessionTracing,
		ex.server.cfg.ContentionRegistry,
		testingPushCallback,
	)
	recv.progressAtomic = progressAtomic
	defer recv.Release()

	evalCtx := planner.ExtendedEvalContext()
	planCtx := ex.server.cfg.DistSQLPlanner.NewPlanningCtx(ctx, evalCtx, planner,
		planner.txn, distribute)
	planCtx.stmtType = recv.stmtType
	if ex.server.cfg.TestingKnobs.TestingSaveFlows != nil {
		__antithesis_instrumentation__.Notify(458295)
		planCtx.saveFlows = ex.server.cfg.TestingKnobs.TestingSaveFlows(planner.stmt.SQL)
	} else {
		__antithesis_instrumentation__.Notify(458296)
		if planner.instrumentation.ShouldSaveFlows() {
			__antithesis_instrumentation__.Notify(458297)
			planCtx.saveFlows = planCtx.getDefaultSaveFlowsFunc(ctx, planner, planComponentTypeMainQuery)
		} else {
			__antithesis_instrumentation__.Notify(458298)
		}
	}
	__antithesis_instrumentation__.Notify(458289)
	planCtx.traceMetadata = planner.instrumentation.traceMetadata
	planCtx.collectExecStats = planner.instrumentation.ShouldCollectExecStats()

	var evalCtxFactory func() *extendedEvalContext
	if len(planner.curPlan.subqueryPlans) != 0 || func() bool {
		__antithesis_instrumentation__.Notify(458299)
		return len(planner.curPlan.cascades) != 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(458300)
		return len(planner.curPlan.checkPlans) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(458301)

		var factoryEvalCtx extendedEvalContext
		ex.initEvalCtx(ctx, &factoryEvalCtx, planner)
		evalCtxFactory = func() *extendedEvalContext {
			__antithesis_instrumentation__.Notify(458302)
			ex.resetEvalCtx(&factoryEvalCtx, planner.txn, planner.ExtendedEvalContext().StmtTimestamp)
			factoryEvalCtx.Placeholders = &planner.semaCtx.Placeholders
			factoryEvalCtx.Annotations = &planner.semaCtx.Annotations

			factoryEvalCtx.Context = evalCtx.Context
			return &factoryEvalCtx
		}
	} else {
		__antithesis_instrumentation__.Notify(458303)
	}
	__antithesis_instrumentation__.Notify(458290)

	if len(planner.curPlan.subqueryPlans) != 0 {
		__antithesis_instrumentation__.Notify(458304)

		subqueryResultMemAcc := planner.EvalContext().Mon.MakeBoundAccount()
		defer subqueryResultMemAcc.Close(ctx)
		if !ex.server.cfg.DistSQLPlanner.PlanAndRunSubqueries(
			ctx, planner, evalCtxFactory, planner.curPlan.subqueryPlans, recv, &subqueryResultMemAcc,
		) {
			__antithesis_instrumentation__.Notify(458305)
			return *recv.stats, recv.commErr
		} else {
			__antithesis_instrumentation__.Notify(458306)
		}
	} else {
		__antithesis_instrumentation__.Notify(458307)
	}
	__antithesis_instrumentation__.Notify(458291)
	recv.discardRows = planner.instrumentation.ShouldDiscardRows()

	cleanup := ex.server.cfg.DistSQLPlanner.PlanAndRun(
		ctx, evalCtx, planCtx, planner.txn, planner.curPlan.main, recv,
	)

	defer cleanup()
	if recv.commErr != nil || func() bool {
		__antithesis_instrumentation__.Notify(458308)
		return res.Err() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(458309)
		return *recv.stats, recv.commErr
	} else {
		__antithesis_instrumentation__.Notify(458310)
	}
	__antithesis_instrumentation__.Notify(458292)

	ex.server.cfg.DistSQLPlanner.PlanAndRunCascadesAndChecks(
		ctx, planner, evalCtxFactory, &planner.curPlan.planComponents, recv,
	)

	return *recv.stats, recv.commErr
}

func (ex *connExecutor) beginTransactionTimestampsAndReadMode(
	ctx context.Context, s *tree.BeginTransaction,
) (
	rwMode tree.ReadWriteMode,
	txnSQLTimestamp time.Time,
	historicalTimestamp *hlc.Timestamp,
	err error,
) {
	__antithesis_instrumentation__.Notify(458311)
	now := ex.server.cfg.Clock.PhysicalTime()
	var modes tree.TransactionModes
	if s != nil {
		__antithesis_instrumentation__.Notify(458316)
		modes = s.Modes
	} else {
		__antithesis_instrumentation__.Notify(458317)
	}
	__antithesis_instrumentation__.Notify(458312)
	asOfClause := ex.asOfClauseWithSessionDefault(modes.AsOf)
	if asOfClause.Expr == nil {
		__antithesis_instrumentation__.Notify(458318)
		rwMode = ex.readWriteModeWithSessionDefault(modes.ReadWriteMode)
		return rwMode, now, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(458319)
	}
	__antithesis_instrumentation__.Notify(458313)
	ex.statsCollector.Reset(ex.applicationStats, ex.phaseTimes)
	p := &ex.planner

	ex.resetPlanner(ctx, p, nil, now)
	asOf, err := p.EvalAsOfTimestamp(ctx, asOfClause)
	if err != nil {
		__antithesis_instrumentation__.Notify(458320)
		return 0, time.Time{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(458321)
	}
	__antithesis_instrumentation__.Notify(458314)

	if modes.ReadWriteMode == tree.ReadWrite {
		__antithesis_instrumentation__.Notify(458322)
		return 0, time.Time{}, nil, tree.ErrAsOfSpecifiedWithReadWrite
	} else {
		__antithesis_instrumentation__.Notify(458323)
	}
	__antithesis_instrumentation__.Notify(458315)
	return tree.ReadOnly, asOf.Timestamp.GoTime(), &asOf.Timestamp, nil
}

var eventStartImplicitTxn fsm.Event = eventTxnStart{ImplicitTxn: fsm.True}
var eventStartExplicitTxn fsm.Event = eventTxnStart{ImplicitTxn: fsm.False}

func (ex *connExecutor) execStmtInNoTxnState(
	ctx context.Context, ast tree.Statement,
) (_ fsm.Event, payload fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458324)
	switch s := ast.(type) {
	case *tree.BeginTransaction:
		__antithesis_instrumentation__.Notify(458325)
		ex.incrementStartedStmtCounter(ast)
		defer func() {
			__antithesis_instrumentation__.Notify(458331)
			if !payloadHasError(payload) {
				__antithesis_instrumentation__.Notify(458332)
				ex.incrementExecutedStmtCounter(ast)
			} else {
				__antithesis_instrumentation__.Notify(458333)
			}
		}()
		__antithesis_instrumentation__.Notify(458326)
		mode, sqlTs, historicalTs, err := ex.beginTransactionTimestampsAndReadMode(ctx, s)
		if err != nil {
			__antithesis_instrumentation__.Notify(458334)
			return ex.makeErrEvent(err, s)
		} else {
			__antithesis_instrumentation__.Notify(458335)
		}
		__antithesis_instrumentation__.Notify(458327)
		ex.sessionDataStack.PushTopClone()
		return eventStartExplicitTxn,
			makeEventTxnStartPayload(
				ex.txnPriorityWithSessionDefault(s.Modes.UserPriority),
				mode,
				sqlTs,
				historicalTs,
				ex.transitionCtx,
				ex.QualityOfService())
	case *tree.CommitTransaction, *tree.ReleaseSavepoint,
		*tree.RollbackTransaction, *tree.SetTransaction, *tree.Savepoint:
		__antithesis_instrumentation__.Notify(458328)
		return ex.makeErrEvent(errNoTransactionInProgress, ast)
	default:
		__antithesis_instrumentation__.Notify(458329)

		noBeginStmt := (*tree.BeginTransaction)(nil)
		mode, sqlTs, historicalTs, err := ex.beginTransactionTimestampsAndReadMode(ctx, noBeginStmt)
		if err != nil {
			__antithesis_instrumentation__.Notify(458336)
			return ex.makeErrEvent(err, s)
		} else {
			__antithesis_instrumentation__.Notify(458337)
		}
		__antithesis_instrumentation__.Notify(458330)
		return eventStartImplicitTxn,
			makeEventTxnStartPayload(
				ex.txnPriorityWithSessionDefault(tree.UnspecifiedUserPriority),
				mode,
				sqlTs,
				historicalTs,
				ex.transitionCtx,
				ex.QualityOfService())
	}
}

func (ex *connExecutor) beginImplicitTxn(
	ctx context.Context, ast tree.Statement,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458338)

	noBeginStmt := (*tree.BeginTransaction)(nil)
	mode, sqlTs, historicalTs, err := ex.beginTransactionTimestampsAndReadMode(ctx, noBeginStmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(458340)
		return ex.makeErrEvent(err, ast)
	} else {
		__antithesis_instrumentation__.Notify(458341)
	}
	__antithesis_instrumentation__.Notify(458339)
	return eventStartImplicitTxn,
		makeEventTxnStartPayload(
			ex.txnPriorityWithSessionDefault(tree.UnspecifiedUserPriority),
			mode,
			sqlTs,
			historicalTs,
			ex.transitionCtx,
			ex.QualityOfService(),
		)
}

func (ex *connExecutor) execStmtInAbortedState(
	ctx context.Context, ast tree.Statement, res RestrictedCommandResult,
) (_ fsm.Event, payload fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458342)
	ex.incrementStartedStmtCounter(ast)
	defer func() {
		__antithesis_instrumentation__.Notify(458345)
		if !payloadHasError(payload) {
			__antithesis_instrumentation__.Notify(458346)
			ex.incrementExecutedStmtCounter(ast)
		} else {
			__antithesis_instrumentation__.Notify(458347)
		}
	}()
	__antithesis_instrumentation__.Notify(458343)

	reject := func() (fsm.Event, fsm.EventPayload) {
		__antithesis_instrumentation__.Notify(458348)
		ev := eventNonRetriableErr{IsCommit: fsm.False}
		payload := eventNonRetriableErrPayload{
			err: sqlerrors.NewTransactionAbortedError(""),
		}
		return ev, payload
	}
	__antithesis_instrumentation__.Notify(458344)

	switch s := ast.(type) {
	case *tree.CommitTransaction, *tree.RollbackTransaction:
		__antithesis_instrumentation__.Notify(458349)
		if _, ok := s.(*tree.CommitTransaction); ok {
			__antithesis_instrumentation__.Notify(458355)

			res.ResetStmtType((*tree.RollbackTransaction)(nil))
		} else {
			__antithesis_instrumentation__.Notify(458356)
		}
		__antithesis_instrumentation__.Notify(458350)
		return ex.rollbackSQLTransaction(ctx, s)

	case *tree.RollbackToSavepoint:
		__antithesis_instrumentation__.Notify(458351)
		return ex.execRollbackToSavepointInAbortedState(ctx, s)

	case *tree.Savepoint:
		__antithesis_instrumentation__.Notify(458352)
		if ex.isCommitOnReleaseSavepoint(s.Name) {
			__antithesis_instrumentation__.Notify(458357)

			res.ResetStmtType((*tree.RollbackToSavepoint)(nil))
			return ex.execRollbackToSavepointInAbortedState(
				ctx, &tree.RollbackToSavepoint{Savepoint: s.Name})
		} else {
			__antithesis_instrumentation__.Notify(458358)
		}
		__antithesis_instrumentation__.Notify(458353)
		return reject()

	default:
		__antithesis_instrumentation__.Notify(458354)
		return reject()
	}
}

func (ex *connExecutor) execStmtInCommitWaitState(
	ctx context.Context, ast tree.Statement, res RestrictedCommandResult,
) (ev fsm.Event, payload fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458359)
	ex.incrementStartedStmtCounter(ast)
	defer func() {
		__antithesis_instrumentation__.Notify(458361)
		if !payloadHasError(payload) {
			__antithesis_instrumentation__.Notify(458362)
			ex.incrementExecutedStmtCounter(ast)
		} else {
			__antithesis_instrumentation__.Notify(458363)
		}
	}()
	__antithesis_instrumentation__.Notify(458360)
	switch ast.(type) {
	case *tree.CommitTransaction, *tree.RollbackTransaction:
		__antithesis_instrumentation__.Notify(458364)

		res.ResetStmtType((*tree.CommitTransaction)(nil))
		return ex.commitSQLTransaction(
			ctx,
			ast,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(458366)

				return nil
			},
		)
	default:
		__antithesis_instrumentation__.Notify(458365)
		ev = eventNonRetriableErr{IsCommit: fsm.False}
		payload = eventNonRetriableErrPayload{
			err: sqlerrors.NewTransactionCommittedError(),
		}
		return ev, payload
	}
}

func (ex *connExecutor) runObserverStatement(
	ctx context.Context, ast tree.Statement, res RestrictedCommandResult,
) error {
	__antithesis_instrumentation__.Notify(458367)
	switch sqlStmt := ast.(type) {
	case *tree.ShowTransactionStatus:
		__antithesis_instrumentation__.Notify(458368)
		return ex.runShowTransactionState(ctx, res)
	case *tree.ShowSavepointStatus:
		__antithesis_instrumentation__.Notify(458369)
		return ex.runShowSavepointState(ctx, res)
	case *tree.ShowSyntax:
		__antithesis_instrumentation__.Notify(458370)
		return ex.runShowSyntax(ctx, sqlStmt.Statement, res)
	case *tree.SetTracing:
		__antithesis_instrumentation__.Notify(458371)
		ex.runSetTracing(ctx, sqlStmt, res)
		return nil
	case *tree.ShowLastQueryStatistics:
		__antithesis_instrumentation__.Notify(458372)
		return ex.runShowLastQueryStatistics(ctx, res, sqlStmt)
	case *tree.ShowTransferState:
		__antithesis_instrumentation__.Notify(458373)
		return ex.runShowTransferState(ctx, res, sqlStmt)
	case *tree.ShowCompletions:
		__antithesis_instrumentation__.Notify(458374)
		return ex.runShowCompletions(ctx, sqlStmt, res)
	default:
		__antithesis_instrumentation__.Notify(458375)
		res.SetError(errors.AssertionFailedf("unrecognized observer statement type %T", ast))
		return nil
	}
}

func (ex *connExecutor) runShowSyntax(
	ctx context.Context, stmt string, res RestrictedCommandResult,
) error {
	__antithesis_instrumentation__.Notify(458376)
	res.SetColumns(ctx, colinfo.ShowSyntaxColumns)
	var commErr error
	parser.RunShowSyntax(ctx, stmt,
		func(ctx context.Context, field, msg string) {
			__antithesis_instrumentation__.Notify(458378)
			commErr = res.AddRow(ctx, tree.Datums{tree.NewDString(field), tree.NewDString(msg)})
		},
		func(ctx context.Context, err error) {
			__antithesis_instrumentation__.Notify(458379)
			sqltelemetry.RecordError(ctx, err, &ex.server.cfg.Settings.SV)
		},
	)
	__antithesis_instrumentation__.Notify(458377)
	return commErr
}

func (ex *connExecutor) runShowTransactionState(
	ctx context.Context, res RestrictedCommandResult,
) error {
	__antithesis_instrumentation__.Notify(458380)
	res.SetColumns(ctx, colinfo.ResultColumns{{Name: "TRANSACTION STATUS", Typ: types.String}})

	state := fmt.Sprintf("%s", ex.machine.CurState())
	return res.AddRow(ctx, tree.Datums{tree.NewDString(state)})
}

func (ex *connExecutor) sessionStateBase64() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(458381)

	_, isNoTxn := ex.machine.CurState().(stateNoTxn)
	state, err := serializeSessionState(
		!isNoTxn, ex.extraTxnState.prepStmtsNamespace, ex.sessionData(),
		ex.server.cfg,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(458383)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(458384)
	}
	__antithesis_instrumentation__.Notify(458382)
	return tree.NewDString(base64.StdEncoding.EncodeToString([]byte(*state))), nil
}

func (ex *connExecutor) sessionRevivalTokenBase64() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(458385)
	cm, err := ex.server.cfg.RPCContext.SecurityContext.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(458388)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(458389)
	}
	__antithesis_instrumentation__.Notify(458386)
	token, err := createSessionRevivalToken(
		AllowSessionRevival.Get(&ex.server.cfg.Settings.SV) && func() bool {
			__antithesis_instrumentation__.Notify(458390)
			return !ex.server.cfg.Codec.ForSystemTenant() == true
		}() == true,
		ex.sessionData(),
		cm,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(458391)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(458392)
	}
	__antithesis_instrumentation__.Notify(458387)
	return tree.NewDString(base64.StdEncoding.EncodeToString([]byte(*token))), nil
}

func (ex *connExecutor) runShowTransferState(
	ctx context.Context, res RestrictedCommandResult, stmt *tree.ShowTransferState,
) error {
	__antithesis_instrumentation__.Notify(458393)

	colNames := []string{
		"error", "session_state_base64", "session_revival_token_base64",
	}
	if stmt.TransferKey != nil {
		__antithesis_instrumentation__.Notify(458399)
		colNames = append(colNames, "transfer_key")
	} else {
		__antithesis_instrumentation__.Notify(458400)
	}
	__antithesis_instrumentation__.Notify(458394)
	cols := make(colinfo.ResultColumns, len(colNames))
	for i := 0; i < len(colNames); i++ {
		__antithesis_instrumentation__.Notify(458401)
		cols[i] = colinfo.ResultColumn{Name: colNames[i], Typ: types.String}
	}
	__antithesis_instrumentation__.Notify(458395)
	res.SetColumns(ctx, cols)

	var sessionState, sessionRevivalToken tree.Datum
	var row tree.Datums
	err := func() error {
		__antithesis_instrumentation__.Notify(458402)

		var err error
		if sessionState, err = ex.sessionStateBase64(); err != nil {
			__antithesis_instrumentation__.Notify(458405)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458406)
		}
		__antithesis_instrumentation__.Notify(458403)
		if sessionRevivalToken, err = ex.sessionRevivalTokenBase64(); err != nil {
			__antithesis_instrumentation__.Notify(458407)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458408)
		}
		__antithesis_instrumentation__.Notify(458404)
		return nil
	}()
	__antithesis_instrumentation__.Notify(458396)
	if err != nil {
		__antithesis_instrumentation__.Notify(458409)

		row = []tree.Datum{tree.NewDString(err.Error()), tree.DNull, tree.DNull}
	} else {
		__antithesis_instrumentation__.Notify(458410)
		row = []tree.Datum{tree.DNull, sessionState, sessionRevivalToken}
	}
	__antithesis_instrumentation__.Notify(458397)
	if stmt.TransferKey != nil {
		__antithesis_instrumentation__.Notify(458411)
		row = append(row, tree.NewDString(stmt.TransferKey.RawString()))
	} else {
		__antithesis_instrumentation__.Notify(458412)
	}
	__antithesis_instrumentation__.Notify(458398)
	return res.AddRow(ctx, row)
}

func (ex *connExecutor) runShowCompletions(
	ctx context.Context, n *tree.ShowCompletions, res RestrictedCommandResult,
) error {
	__antithesis_instrumentation__.Notify(458413)
	res.SetColumns(ctx, colinfo.ResultColumns{{Name: "COMPLETIONS", Typ: types.String}})
	offsetVal, ok := n.Offset.AsConstantInt()
	if !ok {
		__antithesis_instrumentation__.Notify(458418)
		return errors.Newf("invalid offset %v", n.Offset)
	} else {
		__antithesis_instrumentation__.Notify(458419)
	}
	__antithesis_instrumentation__.Notify(458414)
	offset, err := strconv.Atoi(offsetVal.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(458420)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458421)
	}
	__antithesis_instrumentation__.Notify(458415)
	completions, err := delegate.RunShowCompletions(n.Statement.RawString(), offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(458422)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458423)
	}
	__antithesis_instrumentation__.Notify(458416)

	for _, completion := range completions {
		__antithesis_instrumentation__.Notify(458424)
		err = res.AddRow(ctx, tree.Datums{tree.NewDString(completion)})
		if err != nil {
			__antithesis_instrumentation__.Notify(458425)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458426)
		}
	}
	__antithesis_instrumentation__.Notify(458417)
	return nil
}

var showQueryStatsFns = map[tree.Name]func(*sessionphase.Times) time.Duration{
	"parse_latency": func(phaseTimes *sessionphase.Times) time.Duration {
		__antithesis_instrumentation__.Notify(458427)
		return phaseTimes.GetParsingLatency()
	},
	"plan_latency": func(phaseTimes *sessionphase.Times) time.Duration {
		__antithesis_instrumentation__.Notify(458428)
		return phaseTimes.GetPlanningLatency()
	},
	"exec_latency": func(phaseTimes *sessionphase.Times) time.Duration {
		__antithesis_instrumentation__.Notify(458429)
		return phaseTimes.GetRunLatency()
	},

	"service_latency": func(phaseTimes *sessionphase.Times) time.Duration {
		__antithesis_instrumentation__.Notify(458430)
		return phaseTimes.GetServiceLatencyTotal()
	},
	"post_commit_jobs_latency": func(phaseTimes *sessionphase.Times) time.Duration {
		__antithesis_instrumentation__.Notify(458431)
		return phaseTimes.GetPostCommitJobsLatency()
	},
}

func (ex *connExecutor) runShowLastQueryStatistics(
	ctx context.Context, res RestrictedCommandResult, stmt *tree.ShowLastQueryStatistics,
) error {
	__antithesis_instrumentation__.Notify(458432)

	resColumns := make(colinfo.ResultColumns, len(stmt.Columns))
	for i, n := range stmt.Columns {
		__antithesis_instrumentation__.Notify(458435)
		resColumns[i] = colinfo.ResultColumn{Name: string(n), Typ: types.String}
	}
	__antithesis_instrumentation__.Notify(458433)
	res.SetColumns(ctx, resColumns)

	phaseTimes := ex.statsCollector.PreviousPhaseTimes()

	strs := make(tree.Datums, len(stmt.Columns))
	var buf bytes.Buffer
	for i, cname := range stmt.Columns {
		__antithesis_instrumentation__.Notify(458436)
		fn := showQueryStatsFns[cname]
		if fn == nil {
			__antithesis_instrumentation__.Notify(458437)

			strs[i] = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(458438)
			d := fn(phaseTimes)
			ival := tree.NewDInterval(duration.FromFloat64(d.Seconds()), types.DefaultIntervalTypeMetadata)
			buf.Reset()
			ival.Duration.FormatWithStyle(&buf, duration.IntervalStyle_POSTGRES)
			strs[i] = tree.NewDString(buf.String())
		}
	}
	__antithesis_instrumentation__.Notify(458434)

	return res.AddRow(ctx, strs)
}

func (ex *connExecutor) runSetTracing(
	ctx context.Context, n *tree.SetTracing, res RestrictedCommandResult,
) {
	__antithesis_instrumentation__.Notify(458439)
	if len(n.Values) == 0 {
		__antithesis_instrumentation__.Notify(458442)
		res.SetError(errors.AssertionFailedf("set tracing missing argument"))
		return
	} else {
		__antithesis_instrumentation__.Notify(458443)
	}
	__antithesis_instrumentation__.Notify(458440)

	modes := make([]string, len(n.Values))
	for i, v := range n.Values {
		__antithesis_instrumentation__.Notify(458444)
		v = paramparse.UnresolvedNameToStrVal(v)
		var strMode string
		switch val := v.(type) {
		case *tree.StrVal:
			__antithesis_instrumentation__.Notify(458446)
			strMode = val.RawString()
		case *tree.DBool:
			__antithesis_instrumentation__.Notify(458447)
			if *val {
				__antithesis_instrumentation__.Notify(458449)
				strMode = "on"
			} else {
				__antithesis_instrumentation__.Notify(458450)
				strMode = "off"
			}
		default:
			__antithesis_instrumentation__.Notify(458448)
			res.SetError(pgerror.New(pgcode.Syntax,
				"expected string or boolean for set tracing argument"))
			return
		}
		__antithesis_instrumentation__.Notify(458445)
		modes[i] = strMode
	}
	__antithesis_instrumentation__.Notify(458441)

	if err := ex.enableTracing(modes); err != nil {
		__antithesis_instrumentation__.Notify(458451)
		res.SetError(err)
	} else {
		__antithesis_instrumentation__.Notify(458452)
	}
}

func (ex *connExecutor) enableTracing(modes []string) error {
	__antithesis_instrumentation__.Notify(458453)
	traceKV := false
	recordingType := tracing.RecordingVerbose
	enableMode := true
	showResults := false

	for _, s := range modes {
		__antithesis_instrumentation__.Notify(458456)
		switch strings.ToLower(s) {
		case "results":
			__antithesis_instrumentation__.Notify(458457)
			showResults = true
		case "on":
			__antithesis_instrumentation__.Notify(458458)
			enableMode = true
		case "off":
			__antithesis_instrumentation__.Notify(458459)
			enableMode = false
		case "kv":
			__antithesis_instrumentation__.Notify(458460)
			traceKV = true
		case "cluster":
			__antithesis_instrumentation__.Notify(458461)
			recordingType = tracing.RecordingVerbose
		default:
			__antithesis_instrumentation__.Notify(458462)
			return pgerror.Newf(pgcode.Syntax,
				"set tracing: unknown mode %q", s)
		}
	}
	__antithesis_instrumentation__.Notify(458454)
	if !enableMode {
		__antithesis_instrumentation__.Notify(458463)
		return ex.sessionTracing.StopTracing()
	} else {
		__antithesis_instrumentation__.Notify(458464)
	}
	__antithesis_instrumentation__.Notify(458455)
	return ex.sessionTracing.StartTracing(recordingType, traceKV, showResults)
}

func (ex *connExecutor) addActiveQuery(
	ast tree.Statement, rawStmt string, queryID ClusterWideID, cancelFun context.CancelFunc,
) {
	__antithesis_instrumentation__.Notify(458465)
	_, hidden := ast.(tree.HiddenFromShowQueries)
	qm := &queryMeta{
		txnID:         ex.state.mu.txn.ID(),
		start:         ex.phaseTimes.GetSessionPhaseTime(sessionphase.SessionQueryReceived),
		rawStmt:       rawStmt,
		phase:         preparing,
		isDistributed: false,
		ctxCancel:     cancelFun,
		hidden:        hidden,
	}
	ex.mu.Lock()
	ex.mu.ActiveQueries[queryID] = qm
	ex.mu.Unlock()
}

func (ex *connExecutor) removeActiveQuery(queryID ClusterWideID, ast tree.Statement) {
	__antithesis_instrumentation__.Notify(458466)
	ex.mu.Lock()
	_, ok := ex.mu.ActiveQueries[queryID]
	if !ok {
		__antithesis_instrumentation__.Notify(458468)
		ex.mu.Unlock()
		panic(errors.AssertionFailedf("query %d missing from ActiveQueries", queryID))
	} else {
		__antithesis_instrumentation__.Notify(458469)
	}
	__antithesis_instrumentation__.Notify(458467)
	delete(ex.mu.ActiveQueries, queryID)
	ex.mu.LastActiveQuery = ast
	ex.mu.Unlock()
}

func (ex *connExecutor) handleAutoCommit(
	ctx context.Context, stmt tree.Statement,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458470)
	txn := ex.state.mu.txn
	if txn.IsCommitted() {
		__antithesis_instrumentation__.Notify(458475)
		log.Event(ctx, "statement execution committed the txn")
		return eventTxnFinishCommitted{}, nil
	} else {
		__antithesis_instrumentation__.Notify(458476)
	}
	__antithesis_instrumentation__.Notify(458471)

	if knob := ex.server.cfg.TestingKnobs.BeforeAutoCommit; knob != nil {
		__antithesis_instrumentation__.Notify(458477)
		if err := knob(ctx, stmt.String()); err != nil {
			__antithesis_instrumentation__.Notify(458478)
			return ex.makeErrEvent(err, stmt)
		} else {
			__antithesis_instrumentation__.Notify(458479)
		}
	} else {
		__antithesis_instrumentation__.Notify(458480)
	}
	__antithesis_instrumentation__.Notify(458472)

	err := ex.extraTxnState.descCollection.MaybeUpdateDeadline(ctx, ex.state.mu.txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(458481)
		return ex.makeErrEvent(err, stmt)
	} else {
		__antithesis_instrumentation__.Notify(458482)
	}
	__antithesis_instrumentation__.Notify(458473)
	ev, payload := ex.commitSQLTransaction(ctx, stmt, ex.commitSQLTransactionInternal)
	if perr, ok := payload.(payloadWithError); ok {
		__antithesis_instrumentation__.Notify(458483)
		err = perr.errorCause()
	} else {
		__antithesis_instrumentation__.Notify(458484)
	}
	__antithesis_instrumentation__.Notify(458474)
	log.VEventf(ctx, 2, "AutoCommit. err: %v", err)
	return ev, payload
}

func (ex *connExecutor) incrementStartedStmtCounter(ast tree.Statement) {
	__antithesis_instrumentation__.Notify(458485)
	ex.metrics.StartedStatementCounters.incrementCount(ex, ast)
}

func (ex *connExecutor) incrementExecutedStmtCounter(ast tree.Statement) {
	__antithesis_instrumentation__.Notify(458486)
	ex.metrics.ExecutedStatementCounters.incrementCount(ex, ast)
}

func payloadHasError(payload fsm.EventPayload) bool {
	__antithesis_instrumentation__.Notify(458487)
	_, hasErr := payload.(payloadWithError)
	return hasErr
}

func (ex *connExecutor) onTxnFinish(ctx context.Context, ev txnEvent) {
	__antithesis_instrumentation__.Notify(458488)
	if ex.extraTxnState.shouldExecuteOnTxnFinish {
		__antithesis_instrumentation__.Notify(458489)
		ex.extraTxnState.shouldExecuteOnTxnFinish = false
		txnStart := ex.extraTxnState.txnFinishClosure.txnStartTime
		implicit := ex.extraTxnState.txnFinishClosure.implicit
		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionEndExecTransaction, timeutil.Now())
		transactionFingerprintID :=
			roachpb.TransactionFingerprintID(ex.extraTxnState.transactionStatementsHash.Sum())
		if !implicit {
			__antithesis_instrumentation__.Notify(458492)
			ex.statsCollector.EndExplicitTransaction(
				ctx,
				transactionFingerprintID,
			)
		} else {
			__antithesis_instrumentation__.Notify(458493)
		}
		__antithesis_instrumentation__.Notify(458490)
		if ex.server.cfg.TestingKnobs.BeforeTxnStatsRecorded != nil {
			__antithesis_instrumentation__.Notify(458494)
			ex.server.cfg.TestingKnobs.BeforeTxnStatsRecorded(
				ex.sessionData(),
				ev.txnID,
				transactionFingerprintID,
			)
		} else {
			__antithesis_instrumentation__.Notify(458495)
		}
		__antithesis_instrumentation__.Notify(458491)
		err := ex.recordTransactionFinish(ctx, transactionFingerprintID, ev, implicit, txnStart)
		if err != nil {
			__antithesis_instrumentation__.Notify(458496)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(458498)
				log.Warningf(ctx, "failed to record transaction stats: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(458499)
			}
			__antithesis_instrumentation__.Notify(458497)
			ex.server.ServerMetrics.StatsMetrics.DiscardedStatsCount.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(458500)
		}
	} else {
		__antithesis_instrumentation__.Notify(458501)
	}
}

func (ex *connExecutor) onTxnRestart(ctx context.Context) {
	__antithesis_instrumentation__.Notify(458502)
	if ex.extraTxnState.shouldExecuteOnTxnRestart {
		__antithesis_instrumentation__.Notify(458503)
		ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionMostRecentStartExecTransaction, timeutil.Now())
		ex.extraTxnState.transactionStatementFingerprintIDs = nil
		ex.extraTxnState.transactionStatementsHash = util.MakeFNV64()
		ex.extraTxnState.numRows = 0

		ex.extraTxnState.accumulatedStats = execstats.QueryLevelStats{}
		ex.extraTxnState.rowsRead = 0
		ex.extraTxnState.bytesRead = 0
		ex.extraTxnState.rowsWritten = 0

		if ex.server.cfg.TestingKnobs.BeforeRestart != nil {
			__antithesis_instrumentation__.Notify(458504)
			ex.server.cfg.TestingKnobs.BeforeRestart(ctx, ex.extraTxnState.autoRetryReason)
		} else {
			__antithesis_instrumentation__.Notify(458505)
		}
	} else {
		__antithesis_instrumentation__.Notify(458506)
	}
}

func (ex *connExecutor) recordTransactionStart(txnID uuid.UUID) {
	__antithesis_instrumentation__.Notify(458507)

	ex.txnIDCacheWriter.Record(contentionpb.ResolvedTxnID{
		TxnID:            txnID,
		TxnFingerprintID: roachpb.InvalidTransactionFingerprintID,
	})

	ex.state.mu.RLock()
	txnStart := ex.state.mu.txnStart
	ex.state.mu.RUnlock()
	implicit := ex.implicitTxn()

	ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionTransactionReceived,
		ex.phaseTimes.GetSessionPhaseTime(sessionphase.SessionQueryReceived))
	ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionFirstStartExecTransaction, timeutil.Now())
	ex.phaseTimes.SetSessionPhaseTime(sessionphase.SessionMostRecentStartExecTransaction,
		ex.phaseTimes.GetSessionPhaseTime(sessionphase.SessionFirstStartExecTransaction))
	ex.extraTxnState.transactionStatementsHash = util.MakeFNV64()
	ex.extraTxnState.transactionStatementFingerprintIDs = nil
	ex.extraTxnState.numRows = 0
	ex.extraTxnState.shouldCollectTxnExecutionStats = false
	ex.extraTxnState.accumulatedStats = execstats.QueryLevelStats{}
	ex.extraTxnState.rowsRead = 0
	ex.extraTxnState.bytesRead = 0
	ex.extraTxnState.rowsWritten = 0
	ex.extraTxnState.rowsWrittenLogged = false
	ex.extraTxnState.rowsReadLogged = false
	if txnExecStatsSampleRate := collectTxnStatsSampleRate.Get(&ex.server.GetExecutorConfig().Settings.SV); txnExecStatsSampleRate > 0 {
		__antithesis_instrumentation__.Notify(458510)
		ex.extraTxnState.shouldCollectTxnExecutionStats = txnExecStatsSampleRate > ex.rng.Float64()
	} else {
		__antithesis_instrumentation__.Notify(458511)
	}
	__antithesis_instrumentation__.Notify(458508)

	if ex.executorType != executorTypeInternal {
		__antithesis_instrumentation__.Notify(458512)
		ex.metrics.EngineMetrics.SQLTxnsOpen.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(458513)
	}
	__antithesis_instrumentation__.Notify(458509)

	ex.extraTxnState.shouldExecuteOnTxnFinish = true
	ex.extraTxnState.txnFinishClosure.txnStartTime = txnStart
	ex.extraTxnState.txnFinishClosure.implicit = implicit
	ex.extraTxnState.shouldExecuteOnTxnRestart = true

	if !implicit {
		__antithesis_instrumentation__.Notify(458514)
		ex.statsCollector.StartExplicitTransaction()
	} else {
		__antithesis_instrumentation__.Notify(458515)
	}
}

func (ex *connExecutor) recordTransactionFinish(
	ctx context.Context,
	transactionFingerprintID roachpb.TransactionFingerprintID,
	ev txnEvent,
	implicit bool,
	txnStart time.Time,
) error {
	__antithesis_instrumentation__.Notify(458516)
	recordingStart := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(458520)
		recordingOverhead := timeutil.Since(recordingStart)
		ex.server.
			ServerMetrics.
			StatsMetrics.
			SQLTxnStatsCollectionOverhead.RecordValue(recordingOverhead.Nanoseconds())
	}()
	__antithesis_instrumentation__.Notify(458517)

	txnEnd := timeutil.Now()
	txnTime := txnEnd.Sub(txnStart)
	if ex.executorType != executorTypeInternal {
		__antithesis_instrumentation__.Notify(458521)
		ex.metrics.EngineMetrics.SQLTxnsOpen.Dec(1)
	} else {
		__antithesis_instrumentation__.Notify(458522)
	}
	__antithesis_instrumentation__.Notify(458518)
	ex.metrics.EngineMetrics.SQLTxnLatency.RecordValue(txnTime.Nanoseconds())

	ex.txnIDCacheWriter.Record(contentionpb.ResolvedTxnID{
		TxnID:            ev.txnID,
		TxnFingerprintID: transactionFingerprintID,
	})

	if len(ex.extraTxnState.transactionStatementFingerprintIDs) == 0 {
		__antithesis_instrumentation__.Notify(458523)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(458524)
	}
	__antithesis_instrumentation__.Notify(458519)

	txnServiceLat := ex.phaseTimes.GetTransactionServiceLatency()
	txnRetryLat := ex.phaseTimes.GetTransactionRetryLatency()
	commitLat := ex.phaseTimes.GetCommitLatency()

	recordedTxnStats := sqlstats.RecordedTxnStats{
		TransactionTimeSec:      txnTime.Seconds(),
		Committed:               ev.eventType == txnCommit,
		ImplicitTxn:             implicit,
		RetryCount:              int64(atomic.LoadInt32(ex.extraTxnState.atomicAutoRetryCounter)),
		StatementFingerprintIDs: ex.extraTxnState.transactionStatementFingerprintIDs,
		ServiceLatency:          txnServiceLat,
		RetryLatency:            txnRetryLat,
		CommitLatency:           commitLat,
		RowsAffected:            ex.extraTxnState.numRows,
		CollectedExecStats:      ex.extraTxnState.shouldCollectTxnExecutionStats,
		ExecStats:               ex.extraTxnState.accumulatedStats,
		RowsRead:                ex.extraTxnState.rowsRead,
		RowsWritten:             ex.extraTxnState.rowsWritten,
		BytesRead:               ex.extraTxnState.bytesRead,
	}

	return ex.statsCollector.RecordTransaction(
		ctx,
		transactionFingerprintID,
		recordedTxnStats,
	)
}

func logTraceAboveThreshold(
	ctx context.Context, r tracing.Recording, opName string, threshold, elapsed time.Duration,
) {
	__antithesis_instrumentation__.Notify(458525)
	if elapsed < threshold {
		__antithesis_instrumentation__.Notify(458529)
		return
	} else {
		__antithesis_instrumentation__.Notify(458530)
	}
	__antithesis_instrumentation__.Notify(458526)
	if r == nil {
		__antithesis_instrumentation__.Notify(458531)
		log.Warning(ctx, "missing trace when threshold tracing was enabled")
		return
	} else {
		__antithesis_instrumentation__.Notify(458532)
	}
	__antithesis_instrumentation__.Notify(458527)
	dump := r.String()
	if len(dump) == 0 {
		__antithesis_instrumentation__.Notify(458533)
		return
	} else {
		__antithesis_instrumentation__.Notify(458534)
	}
	__antithesis_instrumentation__.Notify(458528)

	log.Infof(ctx, "%s took %s, exceeding threshold of %s:\n%s", opName, elapsed, threshold, dump)
}
