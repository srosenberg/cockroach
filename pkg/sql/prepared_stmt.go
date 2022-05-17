package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/querycache"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

type PreparedStatementOrigin int

const (
	PreparedStatementOriginWire PreparedStatementOrigin = iota + 1

	PreparedStatementOriginSQL

	PreparedStatementOriginSessionMigration
)

type PreparedStatement struct {
	querycache.PrepareMetadata

	Memo *memo.Memo

	refCount int
	memAcc   mon.BoundAccount

	createdAt time.Time

	origin           PreparedStatementOrigin
	StatementSummary string
}

func (p *PreparedStatement) MemoryEstimate() int64 {
	__antithesis_instrumentation__.Notify(563231)

	size := p.PrepareMetadata.MemoryEstimate()
	if p.Memo != nil {
		__antithesis_instrumentation__.Notify(563233)
		size += p.Memo.MemoryEstimate()
	} else {
		__antithesis_instrumentation__.Notify(563234)
	}
	__antithesis_instrumentation__.Notify(563232)
	return size
}

func (p *PreparedStatement) decRef(ctx context.Context) {
	__antithesis_instrumentation__.Notify(563235)
	if p.refCount <= 0 {
		__antithesis_instrumentation__.Notify(563237)
		log.Fatal(ctx, "corrupt PreparedStatement refcount")
	} else {
		__antithesis_instrumentation__.Notify(563238)
	}
	__antithesis_instrumentation__.Notify(563236)
	p.refCount--
	if p.refCount == 0 {
		__antithesis_instrumentation__.Notify(563239)
		p.memAcc.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(563240)
	}
}

func (p *PreparedStatement) incRef(ctx context.Context) {
	__antithesis_instrumentation__.Notify(563241)
	if p.refCount <= 0 {
		__antithesis_instrumentation__.Notify(563243)
		log.Fatal(ctx, "corrupt PreparedStatement refcount")
	} else {
		__antithesis_instrumentation__.Notify(563244)
	}
	__antithesis_instrumentation__.Notify(563242)
	p.refCount++
}

type preparedStatementsAccessor interface {
	List() map[string]*PreparedStatement

	Get(name string) (*PreparedStatement, bool)

	Delete(ctx context.Context, name string) bool

	DeleteAll(ctx context.Context)
}

type PreparedPortal struct {
	Stmt  *PreparedStatement
	Qargs tree.QueryArguments

	OutFormats []pgwirebase.FormatCode

	exhausted bool
}

func (ex *connExecutor) makePreparedPortal(
	ctx context.Context,
	name string,
	stmt *PreparedStatement,
	qargs tree.QueryArguments,
	outFormats []pgwirebase.FormatCode,
) (PreparedPortal, error) {
	__antithesis_instrumentation__.Notify(563245)
	portal := PreparedPortal{
		Stmt:       stmt,
		Qargs:      qargs,
		OutFormats: outFormats,
	}
	return portal, portal.accountForCopy(ctx, &ex.extraTxnState.prepStmtsNamespaceMemAcc, name)
}

func (p PreparedPortal) accountForCopy(
	ctx context.Context, prepStmtsNamespaceMemAcc *mon.BoundAccount, portalName string,
) error {
	__antithesis_instrumentation__.Notify(563246)
	p.Stmt.incRef(ctx)
	return prepStmtsNamespaceMemAcc.Grow(ctx, p.size(portalName))
}

func (p PreparedPortal) close(
	ctx context.Context, prepStmtsNamespaceMemAcc *mon.BoundAccount, portalName string,
) {
	__antithesis_instrumentation__.Notify(563247)
	prepStmtsNamespaceMemAcc.Shrink(ctx, p.size(portalName))
	p.Stmt.decRef(ctx)
}

func (p PreparedPortal) size(portalName string) int64 {
	__antithesis_instrumentation__.Notify(563248)
	return int64(uintptr(len(portalName)) + unsafe.Sizeof(p))
}
