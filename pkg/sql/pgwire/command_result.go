package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type completionMsgType int

const (
	_ completionMsgType = iota
	commandComplete
	bindComplete
	closeComplete
	parseComplete
	emptyQueryResponse
	readyForQuery
	flush

	noCompletionMsg
)

type commandResult struct {
	conn *conn

	conv     sessiondatapb.DataConversionConfig
	location *time.Location

	pos sql.CmdPos

	buffer struct {
		notices            []pgnotice.Notice
		paramStatusUpdates []paramStatusUpdate
	}

	err error

	errExpected bool

	typ completionMsgType

	cmdCompleteTag string

	stmtType     tree.StatementReturnType
	descOpt      sql.RowDescOpt
	rowsAffected int

	formatCodes []pgwirebase.FormatCode

	types []*types.T

	bufferingDisabled bool

	released bool
}

type paramStatusUpdate struct {
	param string

	val string
}

var _ sql.CommandResult = &commandResult{}

func (r *commandResult) Close(ctx context.Context, t sql.TransactionStatusIndicator) {
	__antithesis_instrumentation__.Notify(559151)
	r.assertNotReleased()
	defer r.release()
	if r.errExpected && func() bool {
		__antithesis_instrumentation__.Notify(559156)
		return r.err == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(559157)
		panic("expected err to be set on result by Close, but wasn't")
	} else {
		__antithesis_instrumentation__.Notify(559158)
	}
	__antithesis_instrumentation__.Notify(559152)

	r.conn.writerState.fi.registerCmd(r.pos)
	if r.err != nil {
		__antithesis_instrumentation__.Notify(559159)
		r.conn.bufferErr(ctx, r.err)

		if r.typ != readyForQuery {
			__antithesis_instrumentation__.Notify(559160)
			return
		} else {
			__antithesis_instrumentation__.Notify(559161)
		}
	} else {
		__antithesis_instrumentation__.Notify(559162)
	}
	__antithesis_instrumentation__.Notify(559153)

	for _, notice := range r.buffer.notices {
		__antithesis_instrumentation__.Notify(559163)
		if err := r.conn.bufferNotice(ctx, notice); err != nil {
			__antithesis_instrumentation__.Notify(559164)
			panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err when sending notice"))
		} else {
			__antithesis_instrumentation__.Notify(559165)
		}
	}
	__antithesis_instrumentation__.Notify(559154)

	for _, paramStatusUpdate := range r.buffer.paramStatusUpdates {
		__antithesis_instrumentation__.Notify(559166)
		if err := r.conn.bufferParamStatus(
			paramStatusUpdate.param,
			paramStatusUpdate.val,
		); err != nil {
			__antithesis_instrumentation__.Notify(559167)
			panic(
				errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err when sending parameter status update"),
			)
		} else {
			__antithesis_instrumentation__.Notify(559168)
		}
	}
	__antithesis_instrumentation__.Notify(559155)

	switch r.typ {
	case commandComplete:
		__antithesis_instrumentation__.Notify(559169)
		tag := cookTag(
			r.cmdCompleteTag, r.conn.writerState.tagBuf[:0], r.stmtType, r.rowsAffected,
		)
		r.conn.bufferCommandComplete(tag)
	case parseComplete:
		__antithesis_instrumentation__.Notify(559170)
		r.conn.bufferParseComplete()
	case bindComplete:
		__antithesis_instrumentation__.Notify(559171)
		r.conn.bufferBindComplete()
	case closeComplete:
		__antithesis_instrumentation__.Notify(559172)
		r.conn.bufferCloseComplete()
	case readyForQuery:
		__antithesis_instrumentation__.Notify(559173)
		r.conn.bufferReadyForQuery(byte(t))

		_ = r.conn.Flush(r.pos)
	case emptyQueryResponse:
		__antithesis_instrumentation__.Notify(559174)
		r.conn.bufferEmptyQueryResponse()
	case flush:
		__antithesis_instrumentation__.Notify(559175)

		_ = r.conn.Flush(r.pos)
	case noCompletionMsg:
		__antithesis_instrumentation__.Notify(559176)

	default:
		__antithesis_instrumentation__.Notify(559177)
		panic(errors.AssertionFailedf("unknown type: %v", r.typ))
	}
}

func (r *commandResult) Discard() {
	__antithesis_instrumentation__.Notify(559178)
	r.assertNotReleased()
	defer r.release()
}

func (r *commandResult) Err() error {
	__antithesis_instrumentation__.Notify(559179)
	r.assertNotReleased()
	return r.err
}

func (r *commandResult) SetError(err error) {
	__antithesis_instrumentation__.Notify(559180)
	r.assertNotReleased()
	r.err = err
}

func (r *commandResult) addInternal(bufferData func()) error {
	__antithesis_instrumentation__.Notify(559181)
	r.assertNotReleased()
	if r.err != nil {
		__antithesis_instrumentation__.Notify(559186)
		panic(errors.NewAssertionErrorWithWrappedErrf(r.err, "can't call AddRow after having set error"))
	} else {
		__antithesis_instrumentation__.Notify(559187)
	}
	__antithesis_instrumentation__.Notify(559182)
	r.conn.writerState.fi.registerCmd(r.pos)
	if err := r.conn.GetErr(); err != nil {
		__antithesis_instrumentation__.Notify(559188)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559189)
	}
	__antithesis_instrumentation__.Notify(559183)
	if r.err != nil {
		__antithesis_instrumentation__.Notify(559190)
		panic("can't send row after error")
	} else {
		__antithesis_instrumentation__.Notify(559191)
	}
	__antithesis_instrumentation__.Notify(559184)

	bufferData()

	var err error
	if r.bufferingDisabled {
		__antithesis_instrumentation__.Notify(559192)
		err = r.conn.Flush(r.pos)
	} else {
		__antithesis_instrumentation__.Notify(559193)
		_, err = r.conn.maybeFlush(r.pos)
	}
	__antithesis_instrumentation__.Notify(559185)
	return err
}

func (r *commandResult) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(559194)
	return r.addInternal(func() {
		__antithesis_instrumentation__.Notify(559195)
		r.rowsAffected++
		r.conn.bufferRow(ctx, row, r.formatCodes, r.conv, r.location, r.types)
	})
}

func (r *commandResult) AddBatch(ctx context.Context, batch coldata.Batch) error {
	__antithesis_instrumentation__.Notify(559196)
	return r.addInternal(func() {
		__antithesis_instrumentation__.Notify(559197)
		r.rowsAffected += batch.Length()
		r.conn.bufferBatch(ctx, batch, r.formatCodes, r.conv, r.location)
	})
}

func (r *commandResult) SupportsAddBatch() bool {
	__antithesis_instrumentation__.Notify(559198)
	return true
}

func (r *commandResult) DisableBuffering() {
	__antithesis_instrumentation__.Notify(559199)
	r.assertNotReleased()
	r.bufferingDisabled = true
}

func (r *commandResult) BufferParamStatusUpdate(param string, val string) {
	__antithesis_instrumentation__.Notify(559200)
	r.buffer.paramStatusUpdates = append(
		r.buffer.paramStatusUpdates,
		paramStatusUpdate{param: param, val: val},
	)
}

func (r *commandResult) BufferNotice(notice pgnotice.Notice) {
	__antithesis_instrumentation__.Notify(559201)
	r.buffer.notices = append(r.buffer.notices, notice)
}

func (r *commandResult) SetColumns(ctx context.Context, cols colinfo.ResultColumns) {
	__antithesis_instrumentation__.Notify(559202)
	r.assertNotReleased()
	r.conn.writerState.fi.registerCmd(r.pos)
	if r.descOpt == sql.NeedRowDesc {
		__antithesis_instrumentation__.Notify(559204)
		_ = r.conn.writeRowDescription(ctx, cols, r.formatCodes, &r.conn.writerState.buf)
	} else {
		__antithesis_instrumentation__.Notify(559205)
	}
	__antithesis_instrumentation__.Notify(559203)
	r.types = make([]*types.T, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(559206)
		r.types[i] = col.Typ
	}
}

func (r *commandResult) SetInferredTypes(types []oid.Oid) {
	__antithesis_instrumentation__.Notify(559207)
	r.assertNotReleased()
	r.conn.writerState.fi.registerCmd(r.pos)
	r.conn.bufferParamDesc(types)
}

func (r *commandResult) SetNoDataRowDescription() {
	__antithesis_instrumentation__.Notify(559208)
	r.assertNotReleased()
	r.conn.writerState.fi.registerCmd(r.pos)
	r.conn.bufferNoDataMsg()
}

func (r *commandResult) SetPrepStmtOutput(ctx context.Context, cols colinfo.ResultColumns) {
	__antithesis_instrumentation__.Notify(559209)
	r.assertNotReleased()
	r.conn.writerState.fi.registerCmd(r.pos)
	_ = r.conn.writeRowDescription(ctx, cols, nil, &r.conn.writerState.buf)
}

func (r *commandResult) SetPortalOutput(
	ctx context.Context, cols colinfo.ResultColumns, formatCodes []pgwirebase.FormatCode,
) {
	__antithesis_instrumentation__.Notify(559210)
	r.assertNotReleased()
	r.conn.writerState.fi.registerCmd(r.pos)
	_ = r.conn.writeRowDescription(ctx, cols, formatCodes, &r.conn.writerState.buf)
}

func (r *commandResult) IncrementRowsAffected(ctx context.Context, n int) {
	__antithesis_instrumentation__.Notify(559211)
	r.assertNotReleased()
	r.rowsAffected += n
}

func (r *commandResult) RowsAffected() int {
	__antithesis_instrumentation__.Notify(559212)
	r.assertNotReleased()
	return r.rowsAffected
}

func (r *commandResult) ResetStmtType(stmt tree.Statement) {
	__antithesis_instrumentation__.Notify(559213)
	r.assertNotReleased()
	r.stmtType = stmt.StatementReturnType()
	r.cmdCompleteTag = stmt.StatementTag()
}

func (r *commandResult) release() {
	__antithesis_instrumentation__.Notify(559214)
	*r = commandResult{released: true}
}

func (r *commandResult) assertNotReleased() {
	__antithesis_instrumentation__.Notify(559215)
	if r.released {
		__antithesis_instrumentation__.Notify(559216)
		panic("commandResult used after being released")
	} else {
		__antithesis_instrumentation__.Notify(559217)
	}
}

func (c *conn) allocCommandResult() *commandResult {
	__antithesis_instrumentation__.Notify(559218)
	r := &c.res
	if r.released {
		__antithesis_instrumentation__.Notify(559220)
		r.released = false
	} else {
		__antithesis_instrumentation__.Notify(559221)

		r = new(commandResult)
	}
	__antithesis_instrumentation__.Notify(559219)
	return r
}

func (c *conn) newCommandResult(
	descOpt sql.RowDescOpt,
	pos sql.CmdPos,
	stmt tree.Statement,
	formatCodes []pgwirebase.FormatCode,
	conv sessiondatapb.DataConversionConfig,
	location *time.Location,
	limit int,
	portalName string,
	implicitTxn bool,
) sql.CommandResult {
	__antithesis_instrumentation__.Notify(559222)
	r := c.allocCommandResult()
	*r = commandResult{
		conn:           c,
		conv:           conv,
		location:       location,
		pos:            pos,
		typ:            commandComplete,
		cmdCompleteTag: stmt.StatementTag(),
		stmtType:       stmt.StatementReturnType(),
		descOpt:        descOpt,
		formatCodes:    formatCodes,
	}
	if limit == 0 {
		__antithesis_instrumentation__.Notify(559224)
		return r
	} else {
		__antithesis_instrumentation__.Notify(559225)
	}
	__antithesis_instrumentation__.Notify(559223)
	telemetry.Inc(sqltelemetry.PortalWithLimitRequestCounter)
	return &limitedCommandResult{
		limit:         limit,
		portalName:    portalName,
		implicitTxn:   implicitTxn,
		commandResult: r,
	}
}

func (c *conn) newMiscResult(pos sql.CmdPos, typ completionMsgType) *commandResult {
	__antithesis_instrumentation__.Notify(559226)
	r := c.allocCommandResult()
	*r = commandResult{
		conn: c,
		pos:  pos,
		typ:  typ,
	}
	return r
}

type limitedCommandResult struct {
	*commandResult
	portalName  string
	implicitTxn bool

	seenTuples int

	limit int
}

func (r *limitedCommandResult) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(559227)
	if err := r.commandResult.AddRow(ctx, row); err != nil {
		__antithesis_instrumentation__.Notify(559231)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559232)
	}
	__antithesis_instrumentation__.Notify(559228)
	r.seenTuples++

	if r.seenTuples == r.limit {
		__antithesis_instrumentation__.Notify(559233)

		r.conn.bufferPortalSuspended()
		if err := r.conn.Flush(r.pos); err != nil {
			__antithesis_instrumentation__.Notify(559235)
			return err
		} else {
			__antithesis_instrumentation__.Notify(559236)
		}
		__antithesis_instrumentation__.Notify(559234)
		r.seenTuples = 0

		return r.moreResultsNeeded(ctx)
	} else {
		__antithesis_instrumentation__.Notify(559237)
	}
	__antithesis_instrumentation__.Notify(559229)
	if _, err := r.conn.maybeFlush(r.pos); err != nil {
		__antithesis_instrumentation__.Notify(559238)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559239)
	}
	__antithesis_instrumentation__.Notify(559230)
	return nil
}

func (r *limitedCommandResult) SupportsAddBatch() bool {
	__antithesis_instrumentation__.Notify(559240)
	return false
}

func (r *limitedCommandResult) moreResultsNeeded(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(559241)

	prevPos := r.conn.stmtBuf.AdvanceOne()
	for {
		__antithesis_instrumentation__.Notify(559242)
		cmd, curPos, err := r.conn.stmtBuf.CurCmd()
		if err != nil {
			__antithesis_instrumentation__.Notify(559245)
			return err
		} else {
			__antithesis_instrumentation__.Notify(559246)
		}
		__antithesis_instrumentation__.Notify(559243)
		switch c := cmd.(type) {
		case sql.DeletePreparedStmt:
			__antithesis_instrumentation__.Notify(559247)

			if c.Type != pgwirebase.PreparePortal || func() bool {
				__antithesis_instrumentation__.Notify(559255)
				return c.Name != r.portalName == true
			}() == true {
				__antithesis_instrumentation__.Notify(559256)
				telemetry.Inc(sqltelemetry.InterleavedPortalRequestCounter)
				return errors.WithDetail(sql.ErrLimitedResultNotSupported,
					"cannot close a portal while a different one is open")
			} else {
				__antithesis_instrumentation__.Notify(559257)
			}
			__antithesis_instrumentation__.Notify(559248)
			return r.rewindAndClosePortal(ctx, prevPos)
		case sql.ExecPortal:
			__antithesis_instrumentation__.Notify(559249)

			if c.Name != r.portalName {
				__antithesis_instrumentation__.Notify(559258)
				telemetry.Inc(sqltelemetry.InterleavedPortalRequestCounter)
				return errors.WithDetail(sql.ErrLimitedResultNotSupported,
					"cannot execute a portal while a different one is open")
			} else {
				__antithesis_instrumentation__.Notify(559259)
			}
			__antithesis_instrumentation__.Notify(559250)
			r.limit = c.Limit

			r.rowsAffected = 0
			return nil
		case sql.Sync:
			__antithesis_instrumentation__.Notify(559251)
			if r.implicitTxn {
				__antithesis_instrumentation__.Notify(559260)

				return r.rewindAndClosePortal(ctx, prevPos)
			} else {
				__antithesis_instrumentation__.Notify(559261)
			}
			__antithesis_instrumentation__.Notify(559252)

			r.conn.stmtBuf.AdvanceOne()

			r.conn.stmtBuf.Ltrim(ctx, prevPos)

			r.conn.bufferReadyForQuery(byte(sql.InTxnBlock))
			if err := r.conn.Flush(r.pos); err != nil {
				__antithesis_instrumentation__.Notify(559262)
				return err
			} else {
				__antithesis_instrumentation__.Notify(559263)
			}
		default:
			__antithesis_instrumentation__.Notify(559253)

			if isCommit, err := r.isCommit(); err != nil {
				__antithesis_instrumentation__.Notify(559264)
				return err
			} else {
				__antithesis_instrumentation__.Notify(559265)
				if isCommit {
					__antithesis_instrumentation__.Notify(559266)
					return r.rewindAndClosePortal(ctx, prevPos)
				} else {
					__antithesis_instrumentation__.Notify(559267)
				}
			}
			__antithesis_instrumentation__.Notify(559254)

			telemetry.Inc(sqltelemetry.InterleavedPortalRequestCounter)
			return errors.WithDetail(sql.ErrLimitedResultNotSupported,
				fmt.Sprintf("cannot perform operation %T while a different portal is open", c))
		}
		__antithesis_instrumentation__.Notify(559244)
		prevPos = curPos
	}
}

func (r *limitedCommandResult) isCommit() (bool, error) {
	__antithesis_instrumentation__.Notify(559268)
	cmd, _, err := r.conn.stmtBuf.CurCmd()
	if err != nil {
		__antithesis_instrumentation__.Notify(559276)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(559277)
	}
	__antithesis_instrumentation__.Notify(559269)

	if execStmt, ok := cmd.(sql.ExecStmt); ok {
		__antithesis_instrumentation__.Notify(559278)
		if _, isCommit := execStmt.AST.(*tree.CommitTransaction); isCommit {
			__antithesis_instrumentation__.Notify(559279)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(559280)
		}
	} else {
		__antithesis_instrumentation__.Notify(559281)
	}
	__antithesis_instrumentation__.Notify(559270)

	commitStmtName := ""
	commitPortalName := ""

	if prepareStmt, ok := cmd.(sql.PrepareStmt); ok {
		__antithesis_instrumentation__.Notify(559282)
		if _, isCommit := prepareStmt.AST.(*tree.CommitTransaction); isCommit {
			__antithesis_instrumentation__.Notify(559283)
			commitStmtName = prepareStmt.Name
		} else {
			__antithesis_instrumentation__.Notify(559284)
			return false, nil
		}
	} else {
		__antithesis_instrumentation__.Notify(559285)
		return false, nil
	}
	__antithesis_instrumentation__.Notify(559271)

	r.conn.stmtBuf.AdvanceOne()
	cmd, _, err = r.conn.stmtBuf.CurCmd()
	if err != nil {
		__antithesis_instrumentation__.Notify(559286)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(559287)
	}
	__antithesis_instrumentation__.Notify(559272)

	if bindStmt, ok := cmd.(sql.BindStmt); ok {
		__antithesis_instrumentation__.Notify(559288)

		if bindStmt.PreparedStatementName == commitStmtName {
			__antithesis_instrumentation__.Notify(559289)
			commitPortalName = bindStmt.PortalName
		} else {
			__antithesis_instrumentation__.Notify(559290)
			return false, nil
		}
	} else {
		__antithesis_instrumentation__.Notify(559291)
		return false, nil
	}
	__antithesis_instrumentation__.Notify(559273)

	r.conn.stmtBuf.AdvanceOne()
	cmd, _, err = r.conn.stmtBuf.CurCmd()
	if err != nil {
		__antithesis_instrumentation__.Notify(559292)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(559293)
	}
	__antithesis_instrumentation__.Notify(559274)

	if execPortal, ok := cmd.(sql.ExecPortal); ok {
		__antithesis_instrumentation__.Notify(559294)

		if execPortal.Name == commitPortalName {
			__antithesis_instrumentation__.Notify(559295)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(559296)
		}
	} else {
		__antithesis_instrumentation__.Notify(559297)
	}
	__antithesis_instrumentation__.Notify(559275)
	return false, nil
}

func (r *limitedCommandResult) rewindAndClosePortal(
	ctx context.Context, rewindTo sql.CmdPos,
) error {
	__antithesis_instrumentation__.Notify(559298)

	r.typ = noCompletionMsg

	r.conn.stmtBuf.Rewind(ctx, rewindTo)
	return sql.ErrLimitedResultClosed
}
