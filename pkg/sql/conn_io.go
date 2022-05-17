package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/ring"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type CmdPos int64

type TransactionStatusIndicator byte

const (
	IdleTxnBlock TransactionStatusIndicator = 'I'

	InTxnBlock TransactionStatusIndicator = 'T'

	InFailedTxnBlock TransactionStatusIndicator = 'E'
)

type StmtBuf struct {
	mu struct {
		syncutil.Mutex

		closed bool

		cond *sync.Cond

		data ring.Buffer

		startPos CmdPos

		curPos CmdPos

		lastPos CmdPos
	}
}

type Command interface {
	fmt.Stringer

	command() string
}

type ExecStmt struct {
	parser.Statement

	TimeReceived time.Time

	ParseStart time.Time
	ParseEnd   time.Time

	LastInBatch bool
}

func (ExecStmt) command() string { __antithesis_instrumentation__.Notify(458911); return "exec stmt" }

func (e ExecStmt) String() string {
	__antithesis_instrumentation__.Notify(458912)

	s := "(empty)"

	if e.AST != nil {
		__antithesis_instrumentation__.Notify(458914)
		s = e.AST.String()
	} else {
		__antithesis_instrumentation__.Notify(458915)
	}
	__antithesis_instrumentation__.Notify(458913)
	return fmt.Sprintf("ExecStmt: %s", s)
}

var _ Command = ExecStmt{}

type ExecPortal struct {
	Name string

	Limit int

	TimeReceived time.Time

	FollowedBySync bool
}

func (ExecPortal) command() string {
	__antithesis_instrumentation__.Notify(458916)
	return "exec portal"
}

func (e ExecPortal) String() string {
	__antithesis_instrumentation__.Notify(458917)
	return fmt.Sprintf("ExecPortal name: %q", e.Name)
}

var _ Command = ExecPortal{}

type PrepareStmt struct {
	Name string

	parser.Statement

	TypeHints tree.PlaceholderTypes

	RawTypeHints []oid.Oid
	ParseStart   time.Time
	ParseEnd     time.Time
}

func (PrepareStmt) command() string {
	__antithesis_instrumentation__.Notify(458918)
	return "prepare stmt"
}

func (p PrepareStmt) String() string {
	__antithesis_instrumentation__.Notify(458919)

	s := "(empty)"

	if p.AST != nil {
		__antithesis_instrumentation__.Notify(458921)
		s = p.AST.String()
	} else {
		__antithesis_instrumentation__.Notify(458922)
	}
	__antithesis_instrumentation__.Notify(458920)
	return fmt.Sprintf("PrepareStmt: %s", s)
}

var _ Command = PrepareStmt{}

type DescribeStmt struct {
	Name string
	Type pgwirebase.PrepareType
}

func (DescribeStmt) command() string {
	__antithesis_instrumentation__.Notify(458923)
	return "describe stmt"
}

func (d DescribeStmt) String() string {
	__antithesis_instrumentation__.Notify(458924)
	return fmt.Sprintf("Describe: %q", d.Name)
}

var _ Command = DescribeStmt{}

type BindStmt struct {
	PreparedStatementName string
	PortalName            string

	OutFormats []pgwirebase.FormatCode

	Args [][]byte

	ArgFormatCodes []pgwirebase.FormatCode

	internalArgs []tree.Datum
}

func (BindStmt) command() string { __antithesis_instrumentation__.Notify(458925); return "bind stmt" }

func (b BindStmt) String() string {
	__antithesis_instrumentation__.Notify(458926)
	return fmt.Sprintf("BindStmt: %q->%q", b.PreparedStatementName, b.PortalName)
}

var _ Command = BindStmt{}

type DeletePreparedStmt struct {
	Name string
	Type pgwirebase.PrepareType
}

func (DeletePreparedStmt) command() string {
	__antithesis_instrumentation__.Notify(458927)
	return "delete stmt"
}

func (d DeletePreparedStmt) String() string {
	__antithesis_instrumentation__.Notify(458928)
	return fmt.Sprintf("DeletePreparedStmt: %q", d.Name)
}

var _ Command = DeletePreparedStmt{}

type Sync struct{}

func (Sync) command() string { __antithesis_instrumentation__.Notify(458929); return "sync" }

func (Sync) String() string {
	__antithesis_instrumentation__.Notify(458930)
	return "Sync"
}

var _ Command = Sync{}

type Flush struct{}

func (Flush) command() string { __antithesis_instrumentation__.Notify(458931); return "flush" }

func (Flush) String() string {
	__antithesis_instrumentation__.Notify(458932)
	return "Flush"
}

var _ Command = Flush{}

type CopyIn struct {
	Stmt *tree.CopyFrom

	Conn pgwirebase.Conn

	CopyDone *sync.WaitGroup
}

func (CopyIn) command() string { __antithesis_instrumentation__.Notify(458933); return "copy" }

func (c CopyIn) String() string {
	__antithesis_instrumentation__.Notify(458934)
	s := "(empty)"
	if c.Stmt != nil {
		__antithesis_instrumentation__.Notify(458936)
		s = c.Stmt.String()
	} else {
		__antithesis_instrumentation__.Notify(458937)
	}
	__antithesis_instrumentation__.Notify(458935)
	return fmt.Sprintf("CopyIn: %s", s)
}

var _ Command = CopyIn{}

type DrainRequest struct{}

func (DrainRequest) command() string { __antithesis_instrumentation__.Notify(458938); return "drain" }

func (DrainRequest) String() string {
	__antithesis_instrumentation__.Notify(458939)
	return "Drain"
}

var _ Command = DrainRequest{}

type SendError struct {
	Err error
}

func (SendError) command() string { __antithesis_instrumentation__.Notify(458940); return "send error" }

func (s SendError) String() string {
	__antithesis_instrumentation__.Notify(458941)
	return fmt.Sprintf("SendError: %s", s.Err)
}

var _ Command = SendError{}

func NewStmtBuf() *StmtBuf {
	__antithesis_instrumentation__.Notify(458942)
	var buf StmtBuf
	buf.Init()
	return &buf
}

func (buf *StmtBuf) Init() {
	__antithesis_instrumentation__.Notify(458943)
	buf.mu.lastPos = -1
	buf.mu.cond = sync.NewCond(&buf.mu.Mutex)
}

func (buf *StmtBuf) Close() {
	__antithesis_instrumentation__.Notify(458944)
	buf.mu.Lock()
	buf.mu.closed = true
	buf.mu.cond.Signal()
	buf.mu.Unlock()
}

func (buf *StmtBuf) Push(ctx context.Context, cmd Command) error {
	__antithesis_instrumentation__.Notify(458945)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	if buf.mu.closed {
		__antithesis_instrumentation__.Notify(458947)
		return errors.AssertionFailedf("buffer is closed")
	} else {
		__antithesis_instrumentation__.Notify(458948)
	}
	__antithesis_instrumentation__.Notify(458946)
	buf.mu.data.AddLast(cmd)
	buf.mu.lastPos++

	buf.mu.cond.Signal()
	return nil
}

func (buf *StmtBuf) CurCmd() (Command, CmdPos, error) {
	__antithesis_instrumentation__.Notify(458949)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	for {
		__antithesis_instrumentation__.Notify(458950)
		if buf.mu.closed {
			__antithesis_instrumentation__.Notify(458955)
			return nil, 0, io.EOF
		} else {
			__antithesis_instrumentation__.Notify(458956)
		}
		__antithesis_instrumentation__.Notify(458951)
		curPos := buf.mu.curPos
		cmdIdx, err := buf.translatePosLocked(curPos)
		if err != nil {
			__antithesis_instrumentation__.Notify(458957)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(458958)
		}
		__antithesis_instrumentation__.Notify(458952)
		len := buf.mu.data.Len()
		if cmdIdx < len {
			__antithesis_instrumentation__.Notify(458959)
			return buf.mu.data.Get(cmdIdx).(Command), curPos, nil
		} else {
			__antithesis_instrumentation__.Notify(458960)
		}
		__antithesis_instrumentation__.Notify(458953)
		if cmdIdx != len {
			__antithesis_instrumentation__.Notify(458961)
			return nil, 0, errors.AssertionFailedf(
				"can only wait for next command; corrupt cursor: %d", errors.Safe(curPos))
		} else {
			__antithesis_instrumentation__.Notify(458962)
		}
		__antithesis_instrumentation__.Notify(458954)

		buf.mu.cond.Wait()
	}
}

func (buf *StmtBuf) translatePosLocked(pos CmdPos) (int, error) {
	__antithesis_instrumentation__.Notify(458963)
	if pos < buf.mu.startPos {
		__antithesis_instrumentation__.Notify(458965)
		return 0, errors.AssertionFailedf(
			"position %d no longer in buffer (buffer starting at %d)",
			errors.Safe(pos), errors.Safe(buf.mu.startPos))
	} else {
		__antithesis_instrumentation__.Notify(458966)
	}
	__antithesis_instrumentation__.Notify(458964)
	return int(pos - buf.mu.startPos), nil
}

func (buf *StmtBuf) Ltrim(ctx context.Context, pos CmdPos) {
	__antithesis_instrumentation__.Notify(458967)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	if pos < buf.mu.startPos {
		__antithesis_instrumentation__.Notify(458970)
		log.Fatalf(ctx, "invalid ltrim position: %d. buf starting at: %d",
			pos, buf.mu.startPos)
	} else {
		__antithesis_instrumentation__.Notify(458971)
	}
	__antithesis_instrumentation__.Notify(458968)
	if buf.mu.curPos < pos {
		__antithesis_instrumentation__.Notify(458972)
		log.Fatalf(ctx, "invalid ltrim position: %d when cursor is: %d",
			pos, buf.mu.curPos)
	} else {
		__antithesis_instrumentation__.Notify(458973)
	}
	__antithesis_instrumentation__.Notify(458969)

	for {
		__antithesis_instrumentation__.Notify(458974)
		if buf.mu.startPos == pos {
			__antithesis_instrumentation__.Notify(458976)
			break
		} else {
			__antithesis_instrumentation__.Notify(458977)
		}
		__antithesis_instrumentation__.Notify(458975)
		buf.mu.data.RemoveFirst()
		buf.mu.startPos++
	}
}

func (buf *StmtBuf) AdvanceOne() CmdPos {
	__antithesis_instrumentation__.Notify(458978)
	buf.mu.Lock()
	prev := buf.mu.curPos
	buf.mu.curPos++
	buf.mu.Unlock()
	return prev
}

func (buf *StmtBuf) seekToNextBatch() error {
	__antithesis_instrumentation__.Notify(458979)
	buf.mu.Lock()
	curPos := buf.mu.curPos
	cmdIdx, err := buf.translatePosLocked(curPos)
	if err != nil {
		__antithesis_instrumentation__.Notify(458983)
		buf.mu.Unlock()
		return err
	} else {
		__antithesis_instrumentation__.Notify(458984)
	}
	__antithesis_instrumentation__.Notify(458980)
	if cmdIdx == buf.mu.data.Len() {
		__antithesis_instrumentation__.Notify(458985)
		buf.mu.Unlock()
		return errors.AssertionFailedf("invalid seek start point")
	} else {
		__antithesis_instrumentation__.Notify(458986)
	}
	__antithesis_instrumentation__.Notify(458981)
	buf.mu.Unlock()

	var foundSync bool
	for !foundSync {
		__antithesis_instrumentation__.Notify(458987)
		buf.AdvanceOne()
		_, pos, err := buf.CurCmd()
		if err != nil {
			__antithesis_instrumentation__.Notify(458991)
			return err
		} else {
			__antithesis_instrumentation__.Notify(458992)
		}
		__antithesis_instrumentation__.Notify(458988)
		buf.mu.Lock()
		cmdIdx, err := buf.translatePosLocked(pos)
		if err != nil {
			__antithesis_instrumentation__.Notify(458993)
			buf.mu.Unlock()
			return err
		} else {
			__antithesis_instrumentation__.Notify(458994)
		}
		__antithesis_instrumentation__.Notify(458989)

		if _, ok := buf.mu.data.Get(cmdIdx).(Sync); ok {
			__antithesis_instrumentation__.Notify(458995)
			foundSync = true
		} else {
			__antithesis_instrumentation__.Notify(458996)
		}
		__antithesis_instrumentation__.Notify(458990)

		buf.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(458982)
	return nil
}

func (buf *StmtBuf) Rewind(ctx context.Context, pos CmdPos) {
	__antithesis_instrumentation__.Notify(458997)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	if pos < buf.mu.startPos {
		__antithesis_instrumentation__.Notify(458999)
		log.Fatalf(ctx, "attempting to rewind below buffer start")
	} else {
		__antithesis_instrumentation__.Notify(459000)
	}
	__antithesis_instrumentation__.Notify(458998)
	buf.mu.curPos = pos
}

func (buf *StmtBuf) Len() int {
	__antithesis_instrumentation__.Notify(459001)
	buf.mu.Lock()
	defer buf.mu.Unlock()
	return buf.mu.data.Len()
}

type RowDescOpt bool

const (
	NeedRowDesc RowDescOpt = false

	DontNeedRowDesc RowDescOpt = true
)

type ClientComm interface {
	CreateStatementResult(
		stmt tree.Statement,
		descOpt RowDescOpt,
		pos CmdPos,
		formatCodes []pgwirebase.FormatCode,
		conv sessiondatapb.DataConversionConfig,
		location *time.Location,
		limit int,
		portalName string,
		implicitTxn bool,
	) CommandResult

	CreatePrepareResult(pos CmdPos) ParseResult

	CreateDescribeResult(pos CmdPos) DescribeResult

	CreateBindResult(pos CmdPos) BindResult

	CreateDeleteResult(pos CmdPos) DeleteResult

	CreateSyncResult(pos CmdPos) SyncResult

	CreateFlushResult(pos CmdPos) FlushResult

	CreateErrorResult(pos CmdPos) ErrorResult

	CreateEmptyQueryResult(pos CmdPos) EmptyQueryResult

	CreateCopyInResult(pos CmdPos) CopyInResult

	CreateDrainResult(pos CmdPos) DrainResult

	LockCommunication() ClientLock

	Flush(pos CmdPos) error
}

type CommandResult interface {
	RestrictedCommandResult
	CommandResultClose
}

type CommandResultErrBase interface {
	SetError(error)

	Err() error
}

type ResultBase interface {
	CommandResultErrBase
	CommandResultClose
}

type CommandResultClose interface {
	Close(context.Context, TransactionStatusIndicator)

	Discard()
}

type RestrictedCommandResult interface {
	CommandResultErrBase

	BufferParamStatusUpdate(string, string)

	BufferNotice(notice pgnotice.Notice)

	SetColumns(context.Context, colinfo.ResultColumns)

	ResetStmtType(stmt tree.Statement)

	AddRow(ctx context.Context, row tree.Datums) error

	AddBatch(ctx context.Context, batch coldata.Batch) error

	SupportsAddBatch() bool

	IncrementRowsAffected(ctx context.Context, n int)

	RowsAffected() int

	DisableBuffering()
}

type DescribeResult interface {
	ResultBase

	SetInferredTypes([]oid.Oid)

	SetNoDataRowDescription()

	SetPrepStmtOutput(context.Context, colinfo.ResultColumns)

	SetPortalOutput(context.Context, colinfo.ResultColumns, []pgwirebase.FormatCode)
}

type ParseResult interface {
	ResultBase
}

type BindResult interface {
	ResultBase
}

type ErrorResult interface {
	ResultBase
}

type DeleteResult interface {
	ResultBase
}

type SyncResult interface {
	ResultBase
}

type FlushResult interface {
	ResultBase
}

type DrainResult interface {
	ResultBase
}

type EmptyQueryResult interface {
	ResultBase
}

type CopyInResult interface {
	ResultBase
}

type ClientLock interface {
	Close()

	ClientPos() CmdPos

	RTrim(ctx context.Context, pos CmdPos)
}

type rewindCapability struct {
	cl  ClientLock
	buf *StmtBuf

	rewindPos CmdPos
}

func (rc *rewindCapability) rewindAndUnlock(ctx context.Context) {
	__antithesis_instrumentation__.Notify(459002)
	rc.cl.RTrim(ctx, rc.rewindPos)
	rc.buf.Rewind(ctx, rc.rewindPos)
	rc.cl.Close()
}

func (rc *rewindCapability) close() {
	__antithesis_instrumentation__.Notify(459003)
	rc.cl.Close()
}

type resCloseType bool

const closed resCloseType = true
const discarded resCloseType = false

type streamingCommandResult struct {
	w ieResultWriter

	err          error
	rowsAffected int

	closeCallback func(*streamingCommandResult, resCloseType)
}

var _ RestrictedCommandResult = &streamingCommandResult{}
var _ CommandResultClose = &streamingCommandResult{}

func (r *streamingCommandResult) SetColumns(ctx context.Context, cols colinfo.ResultColumns) {
	__antithesis_instrumentation__.Notify(459004)

	if cols == nil {
		__antithesis_instrumentation__.Notify(459006)
		cols = colinfo.ResultColumns{}
	} else {
		__antithesis_instrumentation__.Notify(459007)
	}
	__antithesis_instrumentation__.Notify(459005)
	_ = r.w.addResult(ctx, ieIteratorResult{cols: cols})
}

func (r *streamingCommandResult) BufferParamStatusUpdate(key string, val string) {
	__antithesis_instrumentation__.Notify(459008)
	panic("unimplemented")
}

func (r *streamingCommandResult) BufferNotice(notice pgnotice.Notice) {
	__antithesis_instrumentation__.Notify(459009)
	panic("unimplemented")
}

func (r *streamingCommandResult) ResetStmtType(stmt tree.Statement) {
	__antithesis_instrumentation__.Notify(459010)
	panic("unimplemented")
}

func (r *streamingCommandResult) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(459011)

	r.rowsAffected++
	rowCopy := make(tree.Datums, len(row))
	copy(rowCopy, row)
	return r.w.addResult(ctx, ieIteratorResult{row: rowCopy})
}

func (r *streamingCommandResult) AddBatch(context.Context, coldata.Batch) error {
	__antithesis_instrumentation__.Notify(459012)

	panic("unimplemented")
}

func (r *streamingCommandResult) SupportsAddBatch() bool {
	__antithesis_instrumentation__.Notify(459013)
	return false
}

func (r *streamingCommandResult) DisableBuffering() {
	__antithesis_instrumentation__.Notify(459014)
	panic("cannot disable buffering here")
}

func (r *streamingCommandResult) SetError(err error) {
	__antithesis_instrumentation__.Notify(459015)
	r.err = err

}

func (r *streamingCommandResult) Err() error {
	__antithesis_instrumentation__.Notify(459016)
	return r.err
}

func (r *streamingCommandResult) IncrementRowsAffected(ctx context.Context, n int) {
	__antithesis_instrumentation__.Notify(459017)
	r.rowsAffected += n

	if r.w != nil {
		__antithesis_instrumentation__.Notify(459018)
		_ = r.w.addResult(ctx, ieIteratorResult{rowsAffectedIncrement: &n})
	} else {
		__antithesis_instrumentation__.Notify(459019)
	}
}

func (r *streamingCommandResult) RowsAffected() int {
	__antithesis_instrumentation__.Notify(459020)
	return r.rowsAffected
}

func (r *streamingCommandResult) Close(context.Context, TransactionStatusIndicator) {
	__antithesis_instrumentation__.Notify(459021)
	if r.closeCallback != nil {
		__antithesis_instrumentation__.Notify(459022)
		r.closeCallback(r, closed)
	} else {
		__antithesis_instrumentation__.Notify(459023)
	}
}

func (r *streamingCommandResult) Discard() {
	__antithesis_instrumentation__.Notify(459024)
	if r.closeCallback != nil {
		__antithesis_instrumentation__.Notify(459025)
		r.closeCallback(r, discarded)
	} else {
		__antithesis_instrumentation__.Notify(459026)
	}
}

func (r *streamingCommandResult) SetInferredTypes([]oid.Oid) {
	__antithesis_instrumentation__.Notify(459027)
}

func (r *streamingCommandResult) SetNoDataRowDescription() {
	__antithesis_instrumentation__.Notify(459028)
}

func (r *streamingCommandResult) SetPrepStmtOutput(context.Context, colinfo.ResultColumns) {
	__antithesis_instrumentation__.Notify(459029)
}

func (r *streamingCommandResult) SetPortalOutput(
	context.Context, colinfo.ResultColumns, []pgwirebase.FormatCode,
) {
	__antithesis_instrumentation__.Notify(459030)
}
