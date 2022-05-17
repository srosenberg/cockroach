package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/fsm"
)

const commitOnReleaseSavepointName = "cockroach_restart"

func (ex *connExecutor) execSavepointInOpenState(
	ctx context.Context, s *tree.Savepoint, res RestrictedCommandResult,
) (fsm.Event, fsm.EventPayload, error) {
	__antithesis_instrumentation__.Notify(458751)
	savepoints := &ex.extraTxnState.savepoints

	commitOnRelease := ex.isCommitOnReleaseSavepoint(s.Name)
	if commitOnRelease {
		__antithesis_instrumentation__.Notify(458754)

		active := ex.state.mu.txn.Active()
		l := len(*savepoints)

		if l == 1 && func() bool {
			__antithesis_instrumentation__.Notify(458757)
			return (*savepoints)[0].commitOnRelease == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(458758)
			return !active == true
		}() == true {
			__antithesis_instrumentation__.Notify(458759)
			return nil, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(458760)
		}
		__antithesis_instrumentation__.Notify(458755)

		err := func() error {
			__antithesis_instrumentation__.Notify(458761)
			if !savepoints.empty() {
				__antithesis_instrumentation__.Notify(458764)
				return pgerror.Newf(pgcode.Syntax,
					"SAVEPOINT \"%s\" cannot be nested",
					tree.ErrNameString(commitOnReleaseSavepointName))
			} else {
				__antithesis_instrumentation__.Notify(458765)
			}
			__antithesis_instrumentation__.Notify(458762)

			if ex.state.mu.txn.Active() {
				__antithesis_instrumentation__.Notify(458766)
				return pgerror.Newf(pgcode.Syntax,
					"SAVEPOINT \"%s\" needs to be the first statement in a transaction",
					tree.ErrNameString(commitOnReleaseSavepointName))
			} else {
				__antithesis_instrumentation__.Notify(458767)
			}
			__antithesis_instrumentation__.Notify(458763)
			return nil
		}()
		__antithesis_instrumentation__.Notify(458756)
		if err != nil {
			__antithesis_instrumentation__.Notify(458768)
			ev, payload := ex.makeErrEvent(err, s)
			return ev, payload, nil
		} else {
			__antithesis_instrumentation__.Notify(458769)
		}
	} else {
		__antithesis_instrumentation__.Notify(458770)
	}
	__antithesis_instrumentation__.Notify(458752)

	token, err := ex.state.mu.txn.CreateSavepoint(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(458771)
		ev, payload := ex.makeErrEvent(err, s)
		return ev, payload, nil
	} else {
		__antithesis_instrumentation__.Notify(458772)
	}
	__antithesis_instrumentation__.Notify(458753)

	sp := savepoint{
		name:            s.Name,
		commitOnRelease: commitOnRelease,
		kvToken:         token,
		numDDL:          ex.extraTxnState.numDDL,
	}
	savepoints.push(sp)
	ex.sessionDataStack.PushTopClone()

	return nil, nil, nil
}

func (ex *connExecutor) execRelease(
	ctx context.Context, s *tree.ReleaseSavepoint, res RestrictedCommandResult,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458773)
	env := &ex.extraTxnState.savepoints
	entry, idx := env.find(s.Savepoint)
	if entry == nil {
		__antithesis_instrumentation__.Notify(458778)
		ev, payload := ex.makeErrEvent(
			pgerror.Newf(pgcode.InvalidSavepointSpecification,
				"savepoint \"%s\" does not exist", &s.Savepoint), s)
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(458779)
	}
	__antithesis_instrumentation__.Notify(458774)

	currSessionData := ex.sessionDataStack.Top()

	env.popToIdx(idx - 1)

	numPoppedElems := (len(ex.extraTxnState.savepoints) - idx) + 1
	if err := ex.sessionDataStack.PopN(numPoppedElems); err != nil {
		__antithesis_instrumentation__.Notify(458780)
		return ex.makeErrEvent(err, s)
	} else {
		__antithesis_instrumentation__.Notify(458781)
	}
	__antithesis_instrumentation__.Notify(458775)
	ex.sessionDataStack.Push(currSessionData)

	if entry.commitOnRelease {
		__antithesis_instrumentation__.Notify(458782)
		res.ResetStmtType((*tree.CommitTransaction)(nil))
		err := ex.commitSQLTransactionInternal(ctx)
		if err == nil {
			__antithesis_instrumentation__.Notify(458785)
			return eventTxnReleased{}, nil
		} else {
			__antithesis_instrumentation__.Notify(458786)
		}
		__antithesis_instrumentation__.Notify(458783)

		if errIsRetriable(err) {
			__antithesis_instrumentation__.Notify(458787)

			env.push(*entry)
			ex.sessionDataStack.PushTopClone()

			rc, canAutoRetry := ex.getRewindTxnCapability()
			ev := eventRetriableErr{
				IsCommit:     fsm.FromBool(isCommit(s)),
				CanAutoRetry: fsm.FromBool(canAutoRetry),
			}
			payload := eventRetriableErrPayload{err: err, rewCap: rc}
			return ev, payload
		} else {
			__antithesis_instrumentation__.Notify(458788)
		}
		__antithesis_instrumentation__.Notify(458784)

		ex.rollbackSQLTransaction(ctx, s)
		ev := eventNonRetriableErr{IsCommit: fsm.FromBool(false)}
		payload := eventNonRetriableErrPayload{err: err}
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(458789)
	}
	__antithesis_instrumentation__.Notify(458776)

	if err := ex.state.mu.txn.ReleaseSavepoint(ctx, entry.kvToken); err != nil {
		__antithesis_instrumentation__.Notify(458790)
		ev, payload := ex.makeErrEvent(err, s)
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(458791)
	}
	__antithesis_instrumentation__.Notify(458777)

	return nil, nil
}

func (ex *connExecutor) execRollbackToSavepointInOpenState(
	ctx context.Context, s *tree.RollbackToSavepoint, res RestrictedCommandResult,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458792)
	entry, idx := ex.extraTxnState.savepoints.find(s.Savepoint)
	if entry == nil {
		__antithesis_instrumentation__.Notify(458798)
		ev, payload := ex.makeErrEvent(pgerror.Newf(pgcode.InvalidSavepointSpecification,
			"savepoint \"%s\" does not exist", &s.Savepoint), s)
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(458799)
	}
	__antithesis_instrumentation__.Notify(458793)

	if ev, payload, ok := ex.checkRollbackValidity(ctx, s, entry); !ok {
		__antithesis_instrumentation__.Notify(458800)
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(458801)
	}
	__antithesis_instrumentation__.Notify(458794)

	if err := ex.state.mu.txn.RollbackToSavepoint(ctx, entry.kvToken); err != nil {
		__antithesis_instrumentation__.Notify(458802)
		ev, payload := ex.makeErrEvent(err, s)
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(458803)
	}
	__antithesis_instrumentation__.Notify(458795)

	if err := ex.popSavepointsToIdx(s, idx); err != nil {
		__antithesis_instrumentation__.Notify(458804)
		return ex.makeErrEvent(err, s)
	} else {
		__antithesis_instrumentation__.Notify(458805)
	}
	__antithesis_instrumentation__.Notify(458796)

	if entry.kvToken.Initial() {
		__antithesis_instrumentation__.Notify(458806)
		return eventTxnRestart{}, nil
	} else {
		__antithesis_instrumentation__.Notify(458807)
	}
	__antithesis_instrumentation__.Notify(458797)

	return nil, nil
}

func (ex *connExecutor) checkRollbackValidity(
	ctx context.Context, s *tree.RollbackToSavepoint, entry *savepoint,
) (ev fsm.Event, payload fsm.EventPayload, ok bool) {
	__antithesis_instrumentation__.Notify(458808)
	if ex.extraTxnState.numDDL <= entry.numDDL {
		__antithesis_instrumentation__.Notify(458812)

		return ev, payload, true
	} else {
		__antithesis_instrumentation__.Notify(458813)
	}
	__antithesis_instrumentation__.Notify(458809)

	if !entry.kvToken.Initial() {
		__antithesis_instrumentation__.Notify(458814)

		ev, payload = ex.makeErrEvent(unimplemented.NewWithIssueDetail(10735, "rollback-after-ddl",
			"ROLLBACK TO SAVEPOINT not yet supported after DDL statements"), s)
		return ev, payload, false
	} else {
		__antithesis_instrumentation__.Notify(458815)
	}
	__antithesis_instrumentation__.Notify(458810)

	if ex.state.mu.txn.UserPriority() == roachpb.MaxUserPriority {
		__antithesis_instrumentation__.Notify(458816)

		ev, payload = ex.makeErrEvent(unimplemented.NewWithIssue(46414,
			"cannot use ROLLBACK TO SAVEPOINT in a HIGH PRIORITY transaction containing DDL"), s)
		return ev, payload, false
	} else {
		__antithesis_instrumentation__.Notify(458817)
	}
	__antithesis_instrumentation__.Notify(458811)

	return ev, payload, true
}

func (ex *connExecutor) execRollbackToSavepointInAbortedState(
	ctx context.Context, s *tree.RollbackToSavepoint,
) (fsm.Event, fsm.EventPayload) {
	__antithesis_instrumentation__.Notify(458818)
	makeErr := func(err error) (fsm.Event, fsm.EventPayload) {
		__antithesis_instrumentation__.Notify(458825)
		ev := eventNonRetriableErr{IsCommit: fsm.False}
		payload := eventNonRetriableErrPayload{
			err: err,
		}
		return ev, payload
	}
	__antithesis_instrumentation__.Notify(458819)

	entry, idx := ex.extraTxnState.savepoints.find(s.Savepoint)
	if entry == nil {
		__antithesis_instrumentation__.Notify(458826)
		return makeErr(pgerror.Newf(pgcode.InvalidSavepointSpecification,
			"savepoint \"%s\" does not exist", tree.ErrString(&s.Savepoint)))
	} else {
		__antithesis_instrumentation__.Notify(458827)
	}
	__antithesis_instrumentation__.Notify(458820)

	if ev, payload, ok := ex.checkRollbackValidity(ctx, s, entry); !ok {
		__antithesis_instrumentation__.Notify(458828)
		return ev, payload
	} else {
		__antithesis_instrumentation__.Notify(458829)
	}
	__antithesis_instrumentation__.Notify(458821)

	if err := ex.popSavepointsToIdx(s, idx); err != nil {
		__antithesis_instrumentation__.Notify(458830)
		return ex.makeErrEvent(err, s)
	} else {
		__antithesis_instrumentation__.Notify(458831)
	}
	__antithesis_instrumentation__.Notify(458822)

	if err := ex.state.mu.txn.RollbackToSavepoint(ctx, entry.kvToken); err != nil {
		__antithesis_instrumentation__.Notify(458832)
		return ex.makeErrEvent(err, s)
	} else {
		__antithesis_instrumentation__.Notify(458833)
	}
	__antithesis_instrumentation__.Notify(458823)

	if entry.kvToken.Initial() {
		__antithesis_instrumentation__.Notify(458834)
		return eventTxnRestart{}, nil
	} else {
		__antithesis_instrumentation__.Notify(458835)
	}
	__antithesis_instrumentation__.Notify(458824)
	return eventSavepointRollback{}, nil
}

func (ex *connExecutor) popSavepointsToIdx(stmt tree.Statement, idx int) error {
	__antithesis_instrumentation__.Notify(458836)
	if err := ex.reportSessionDataChanges(func() error {
		__antithesis_instrumentation__.Notify(458838)
		numPoppedElems := len(ex.extraTxnState.savepoints) - idx
		ex.extraTxnState.savepoints.popToIdx(idx)
		if err := ex.sessionDataStack.PopN(numPoppedElems); err != nil {
			__antithesis_instrumentation__.Notify(458840)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458841)
		}
		__antithesis_instrumentation__.Notify(458839)

		ex.sessionDataStack.PushTopClone()
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(458842)
		return err
	} else {
		__antithesis_instrumentation__.Notify(458843)
	}
	__antithesis_instrumentation__.Notify(458837)
	return nil
}

func (ex *connExecutor) isCommitOnReleaseSavepoint(savepoint tree.Name) bool {
	__antithesis_instrumentation__.Notify(458844)
	if ex.sessionData().ForceSavepointRestart {
		__antithesis_instrumentation__.Notify(458846)

		return true
	} else {
		__antithesis_instrumentation__.Notify(458847)
	}
	__antithesis_instrumentation__.Notify(458845)
	return strings.HasPrefix(string(savepoint), commitOnReleaseSavepointName)
}

type savepoint struct {
	name tree.Name

	commitOnRelease bool

	kvToken kv.SavepointToken

	numDDL int
}

type savepointStack []savepoint

func (stack savepointStack) empty() bool {
	__antithesis_instrumentation__.Notify(458848)
	return len(stack) == 0
}

func (stack *savepointStack) clear() {
	__antithesis_instrumentation__.Notify(458849)
	*stack = (*stack)[:0]
}

func (stack *savepointStack) push(s savepoint) {
	__antithesis_instrumentation__.Notify(458850)
	*stack = append(*stack, s)
}

func (stack savepointStack) find(sn tree.Name) (*savepoint, int) {
	__antithesis_instrumentation__.Notify(458851)
	for i := len(stack) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(458853)
		if stack[i].name == sn {
			__antithesis_instrumentation__.Notify(458854)
			return &stack[i], i
		} else {
			__antithesis_instrumentation__.Notify(458855)
		}
	}
	__antithesis_instrumentation__.Notify(458852)
	return nil, -1
}

func (stack *savepointStack) popToIdx(idx int) {
	__antithesis_instrumentation__.Notify(458856)
	*stack = (*stack)[:idx+1]
}

func (stack savepointStack) clone() savepointStack {
	__antithesis_instrumentation__.Notify(458857)
	if len(stack) == 0 {
		__antithesis_instrumentation__.Notify(458859)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(458860)
	}
	__antithesis_instrumentation__.Notify(458858)
	cpy := make(savepointStack, len(stack))
	copy(cpy, stack)
	return cpy
}

func (ex *connExecutor) runShowSavepointState(
	ctx context.Context, res RestrictedCommandResult,
) error {
	__antithesis_instrumentation__.Notify(458861)
	res.SetColumns(ctx, colinfo.ResultColumns{
		{Name: "savepoint_name", Typ: types.String},
		{Name: "is_initial_savepoint", Typ: types.Bool},
	})

	for _, entry := range ex.extraTxnState.savepoints {
		__antithesis_instrumentation__.Notify(458863)
		if err := res.AddRow(ctx, tree.Datums{
			tree.NewDString(string(entry.name)),
			tree.MakeDBool(tree.DBool(entry.kvToken.Initial())),
		}); err != nil {
			__antithesis_instrumentation__.Notify(458864)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458865)
		}
	}
	__antithesis_instrumentation__.Notify(458862)
	return nil
}
