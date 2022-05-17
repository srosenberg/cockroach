package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func (p *planner) DeclareCursor(ctx context.Context, s *tree.DeclareCursor) (planNode, error) {
	__antithesis_instrumentation__.Notify(623747)
	if s.Hold {
		__antithesis_instrumentation__.Notify(623751)
		return nil, unimplemented.NewWithIssue(77101, "DECLARE CURSOR WITH HOLD")
	} else {
		__antithesis_instrumentation__.Notify(623752)
	}
	__antithesis_instrumentation__.Notify(623748)
	if s.Binary {
		__antithesis_instrumentation__.Notify(623753)
		return nil, unimplemented.NewWithIssue(77099, "DECLARE BINARY CURSOR")
	} else {
		__antithesis_instrumentation__.Notify(623754)
	}
	__antithesis_instrumentation__.Notify(623749)
	if s.Scroll == tree.Scroll {
		__antithesis_instrumentation__.Notify(623755)
		return nil, unimplemented.NewWithIssue(77102, "DECLARE SCROLL CURSOR")
	} else {
		__antithesis_instrumentation__.Notify(623756)
	}
	__antithesis_instrumentation__.Notify(623750)

	return &delayedNode{
		name: s.String(),
		constructor: func(ctx context.Context, p *planner) (_ planNode, _ error) {
			__antithesis_instrumentation__.Notify(623757)
			if p.extendedEvalCtx.TxnImplicit {
				__antithesis_instrumentation__.Notify(623766)
				return nil, pgerror.Newf(pgcode.NoActiveSQLTransaction, "DECLARE CURSOR can only be used in transaction blocks")
			} else {
				__antithesis_instrumentation__.Notify(623767)
			}
			__antithesis_instrumentation__.Notify(623758)

			ie := p.ExecCfg().InternalExecutorFactory(ctx, p.SessionData())
			cursorName := s.Name.String()
			if cursor, _ := p.sqlCursors.getCursor(cursorName); cursor != nil {
				__antithesis_instrumentation__.Notify(623768)
				return nil, pgerror.Newf(pgcode.DuplicateCursor, "cursor %q already exists", cursorName)
			} else {
				__antithesis_instrumentation__.Notify(623769)
			}
			__antithesis_instrumentation__.Notify(623759)

			if p.extendedEvalCtx.PreparedStatementState.HasPortal(cursorName) {
				__antithesis_instrumentation__.Notify(623770)
				return nil, pgerror.Newf(pgcode.DuplicateCursor, "cursor %q already exists as portal", cursorName)
			} else {
				__antithesis_instrumentation__.Notify(623771)
			}
			__antithesis_instrumentation__.Notify(623760)

			stmt := makeStatement(parser.Statement{AST: s.Select}, ClusterWideID{})
			pt := planTop{}
			pt.init(&stmt, &p.instrumentation)
			opc := &p.optPlanningCtx
			opc.p.stmt = stmt
			opc.reset()

			memo, err := opc.buildExecMemo(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(623772)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623773)
			}
			__antithesis_instrumentation__.Notify(623761)
			if err := opc.runExecBuilder(
				&pt,
				&stmt,
				newExecFactory(p),
				memo,
				p.EvalContext(),
				p.autoCommit,
			); err != nil {
				__antithesis_instrumentation__.Notify(623774)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623775)
			}
			__antithesis_instrumentation__.Notify(623762)
			if pt.flags.IsSet(planFlagContainsMutation) {
				__antithesis_instrumentation__.Notify(623776)

				return nil, pgerror.Newf(pgcode.FeatureNotSupported,
					"DECLARE CURSOR must not contain data-modifying statements in WITH")
			} else {
				__antithesis_instrumentation__.Notify(623777)
			}
			__antithesis_instrumentation__.Notify(623763)

			statement := s.Select.String()
			itCtx := context.Background()
			rows, err := ie.QueryIterator(itCtx, "sql-cursor", p.txn, statement)
			if err != nil {
				__antithesis_instrumentation__.Notify(623778)
				return nil, errors.Wrap(err, "failed to DECLARE CURSOR")
			} else {
				__antithesis_instrumentation__.Notify(623779)
			}
			__antithesis_instrumentation__.Notify(623764)
			inputState := p.txn.GetLeafTxnInputState(ctx)
			cursor := &sqlCursor{
				InternalRows: rows,
				readSeqNum:   inputState.ReadSeqNum,
				txn:          p.txn,
				statement:    statement,
				created:      timeutil.Now(),
			}
			if err := p.sqlCursors.addCursor(cursorName, cursor); err != nil {
				__antithesis_instrumentation__.Notify(623780)

				_ = cursor.Close()
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623781)
			}
			__antithesis_instrumentation__.Notify(623765)
			return newZeroNode(nil), nil
		},
	}, nil
}

var errBackwardScan = pgerror.Newf(pgcode.ObjectNotInPrerequisiteState, "cursor can only scan forward")

func (p *planner) FetchCursor(
	_ context.Context, s *tree.CursorStmt, isMove bool,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(623782)
	cursor, err := p.sqlCursors.getCursor(s.Name.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(623786)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623787)
	}
	__antithesis_instrumentation__.Notify(623783)
	if s.Count < 0 || func() bool {
		__antithesis_instrumentation__.Notify(623788)
		return s.FetchType == tree.FetchBackwardAll == true
	}() == true {
		__antithesis_instrumentation__.Notify(623789)
		return nil, errBackwardScan
	} else {
		__antithesis_instrumentation__.Notify(623790)
	}
	__antithesis_instrumentation__.Notify(623784)
	node := &fetchNode{
		n:         s.Count,
		fetchType: s.FetchType,
		cursor:    cursor,
		isMove:    isMove,
	}
	if s.FetchType != tree.FetchNormal {
		__antithesis_instrumentation__.Notify(623791)
		node.n = 0
		node.offset = s.Count
	} else {
		__antithesis_instrumentation__.Notify(623792)
	}
	__antithesis_instrumentation__.Notify(623785)
	return node, nil
}

type fetchNode struct {
	cursor *sqlCursor

	n int64

	offset    int64
	fetchType tree.FetchType

	isMove bool

	seeked bool

	origTxnSeqNum enginepb.TxnSeq
}

func (f *fetchNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(623793)
	state := f.cursor.txn.GetLeafTxnInputState(params.ctx)

	f.origTxnSeqNum = state.ReadSeqNum
	return f.cursor.txn.SetReadSeqNum(f.cursor.readSeqNum)
}

func (f *fetchNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623794)
	if f.fetchType == tree.FetchAll {
		__antithesis_instrumentation__.Notify(623798)
		return f.cursor.Next(params.ctx)
	} else {
		__antithesis_instrumentation__.Notify(623799)
	}
	__antithesis_instrumentation__.Notify(623795)

	if !f.seeked {
		__antithesis_instrumentation__.Notify(623800)

		f.seeked = true
		switch f.fetchType {
		case tree.FetchFirst:
			__antithesis_instrumentation__.Notify(623801)
			switch f.cursor.curRow {
			case 0:
				__antithesis_instrumentation__.Notify(623810)
				_, err := f.cursor.Next(params.ctx)
				return true, err
			case 1:
				__antithesis_instrumentation__.Notify(623811)
				return true, nil
			default:
				__antithesis_instrumentation__.Notify(623812)
			}
			__antithesis_instrumentation__.Notify(623802)
			return false, errBackwardScan
		case tree.FetchLast:
			__antithesis_instrumentation__.Notify(623803)
			return false, errBackwardScan
		case tree.FetchAbsolute:
			__antithesis_instrumentation__.Notify(623804)
			if f.cursor.curRow > f.offset {
				__antithesis_instrumentation__.Notify(623813)
				return false, errBackwardScan
			} else {
				__antithesis_instrumentation__.Notify(623814)
			}
			__antithesis_instrumentation__.Notify(623805)
			for f.cursor.curRow < f.offset {
				__antithesis_instrumentation__.Notify(623815)
				more, err := f.cursor.Next(params.ctx)
				if !more || func() bool {
					__antithesis_instrumentation__.Notify(623816)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(623817)
					return more, err
				} else {
					__antithesis_instrumentation__.Notify(623818)
				}
			}
			__antithesis_instrumentation__.Notify(623806)
			return true, nil
		case tree.FetchRelative:
			__antithesis_instrumentation__.Notify(623807)
			for i := int64(0); i < f.offset; i++ {
				__antithesis_instrumentation__.Notify(623819)
				more, err := f.cursor.Next(params.ctx)
				if !more || func() bool {
					__antithesis_instrumentation__.Notify(623820)
					return err != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(623821)
					return more, err
				} else {
					__antithesis_instrumentation__.Notify(623822)
				}
			}
			__antithesis_instrumentation__.Notify(623808)
			return true, nil
		default:
			__antithesis_instrumentation__.Notify(623809)
		}
	} else {
		__antithesis_instrumentation__.Notify(623823)
	}
	__antithesis_instrumentation__.Notify(623796)
	if f.n <= 0 {
		__antithesis_instrumentation__.Notify(623824)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(623825)
	}
	__antithesis_instrumentation__.Notify(623797)
	f.n--
	return f.cursor.Next(params.ctx)
}

func (f fetchNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623826)
	return f.cursor.Cur()
}

func (f fetchNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623827)

	if err := f.cursor.txn.SetReadSeqNum(f.origTxnSeqNum); err != nil {
		__antithesis_instrumentation__.Notify(623828)
		log.Warningf(ctx, "error resetting transaction read seq num after CURSOR operation: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(623829)
	}
}

func (p *planner) CloseCursor(ctx context.Context, n *tree.CloseCursor) (planNode, error) {
	__antithesis_instrumentation__.Notify(623830)
	return &delayedNode{
		name: n.String(),
		constructor: func(ctx context.Context, p *planner) (planNode, error) {
			__antithesis_instrumentation__.Notify(623831)
			return newZeroNode(nil), p.sqlCursors.closeCursor(n.Name.String())
		},
	}, nil
}

type sqlCursor struct {
	sqlutil.InternalRows

	txn *kv.Txn

	readSeqNum enginepb.TxnSeq
	statement  string
	created    time.Time
	curRow     int64
}

func (s *sqlCursor) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(623832)
	more, err := s.InternalRows.Next(ctx)
	if err == nil {
		__antithesis_instrumentation__.Notify(623834)
		s.curRow++
	} else {
		__antithesis_instrumentation__.Notify(623835)
	}
	__antithesis_instrumentation__.Notify(623833)
	return more, err
}

type sqlCursors interface {
	closeAll()

	closeCursor(string) error

	getCursor(string) (*sqlCursor, error)

	addCursor(string, *sqlCursor) error

	list() map[string]*sqlCursor
}

type cursorMap struct {
	cursors map[string]*sqlCursor
}

func (c *cursorMap) closeAll() {
	__antithesis_instrumentation__.Notify(623836)
	for _, c := range c.cursors {
		__antithesis_instrumentation__.Notify(623838)
		_ = c.Close()
	}
	__antithesis_instrumentation__.Notify(623837)
	c.cursors = nil
}

func (c *cursorMap) closeCursor(s string) error {
	__antithesis_instrumentation__.Notify(623839)
	cursor, ok := c.cursors[s]
	if !ok {
		__antithesis_instrumentation__.Notify(623841)
		return pgerror.Newf(pgcode.InvalidCursorName, "cursor %q does not exist", s)
	} else {
		__antithesis_instrumentation__.Notify(623842)
	}
	__antithesis_instrumentation__.Notify(623840)
	err := cursor.Close()
	delete(c.cursors, s)
	return err
}

func (c *cursorMap) getCursor(s string) (*sqlCursor, error) {
	__antithesis_instrumentation__.Notify(623843)
	cursor, ok := c.cursors[s]
	if !ok {
		__antithesis_instrumentation__.Notify(623845)
		return nil, pgerror.Newf(pgcode.InvalidCursorName, "cursor %q does not exist", s)
	} else {
		__antithesis_instrumentation__.Notify(623846)
	}
	__antithesis_instrumentation__.Notify(623844)
	return cursor, nil
}

func (c *cursorMap) addCursor(s string, cursor *sqlCursor) error {
	__antithesis_instrumentation__.Notify(623847)
	if c.cursors == nil {
		__antithesis_instrumentation__.Notify(623850)
		c.cursors = make(map[string]*sqlCursor)
	} else {
		__antithesis_instrumentation__.Notify(623851)
	}
	__antithesis_instrumentation__.Notify(623848)
	if _, ok := c.cursors[s]; ok {
		__antithesis_instrumentation__.Notify(623852)
		return pgerror.Newf(pgcode.DuplicateCursor, "cursor %q already exists", s)
	} else {
		__antithesis_instrumentation__.Notify(623853)
	}
	__antithesis_instrumentation__.Notify(623849)
	c.cursors[s] = cursor
	return nil
}

func (c *cursorMap) list() map[string]*sqlCursor {
	__antithesis_instrumentation__.Notify(623854)
	return c.cursors
}

type connExCursorAccessor struct {
	ex *connExecutor
}

func (c connExCursorAccessor) closeAll() {
	__antithesis_instrumentation__.Notify(623855)
	c.ex.extraTxnState.sqlCursors.closeAll()
}

func (c connExCursorAccessor) closeCursor(s string) error {
	__antithesis_instrumentation__.Notify(623856)
	return c.ex.extraTxnState.sqlCursors.closeCursor(s)
}

func (c connExCursorAccessor) getCursor(s string) (*sqlCursor, error) {
	__antithesis_instrumentation__.Notify(623857)
	return c.ex.extraTxnState.sqlCursors.getCursor(s)
}

func (c connExCursorAccessor) addCursor(s string, cursor *sqlCursor) error {
	__antithesis_instrumentation__.Notify(623858)
	return c.ex.extraTxnState.sqlCursors.addCursor(s, cursor)
}

func (c connExCursorAccessor) list() map[string]*sqlCursor {
	__antithesis_instrumentation__.Notify(623859)
	return c.ex.extraTxnState.sqlCursors.list()
}

func (p *planner) checkNoConflictingCursors(stmt tree.Statement) error {
	__antithesis_instrumentation__.Notify(623860)

	if len(p.sqlCursors.list()) > 0 {
		__antithesis_instrumentation__.Notify(623862)
		return unimplemented.NewWithIssue(74608, "cannot run schema change "+
			"in a transaction with open DECLARE cursors")
	} else {
		__antithesis_instrumentation__.Notify(623863)
	}
	__antithesis_instrumentation__.Notify(623861)
	return nil
}
