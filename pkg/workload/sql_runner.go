package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type SQLRunner struct {
	stmts []*stmt

	initialized bool
	method      method
	mcp         *MultiConnPool
}

type method int

const (
	prepare method = iota
	noprepare
	simple
)

var stringToMethod = map[string]method{
	"prepare":   prepare,
	"noprepare": noprepare,
	"simple":    simple,
}

func (sr *SQLRunner) Define(sql string) StmtHandle {
	__antithesis_instrumentation__.Notify(697346)
	if sr.initialized {
		__antithesis_instrumentation__.Notify(697348)
		panic("Define can't be called after Init")
	} else {
		__antithesis_instrumentation__.Notify(697349)
	}
	__antithesis_instrumentation__.Notify(697347)
	s := &stmt{sr: sr, sql: sql}
	sr.stmts = append(sr.stmts, s)
	return StmtHandle{s: s}
}

func (sr *SQLRunner) Init(
	ctx context.Context, name string, mcp *MultiConnPool, flags *ConnFlags,
) error {
	__antithesis_instrumentation__.Notify(697350)
	if sr.initialized {
		__antithesis_instrumentation__.Notify(697354)
		panic("already initialized")
	} else {
		__antithesis_instrumentation__.Notify(697355)
	}
	__antithesis_instrumentation__.Notify(697351)

	var ok bool
	sr.method, ok = stringToMethod[strings.ToLower(flags.Method)]
	if !ok {
		__antithesis_instrumentation__.Notify(697356)
		return errors.Errorf("unknown method %s", flags.Method)
	} else {
		__antithesis_instrumentation__.Notify(697357)
	}
	__antithesis_instrumentation__.Notify(697352)

	if sr.method == prepare {
		__antithesis_instrumentation__.Notify(697358)
		for i, s := range sr.stmts {
			__antithesis_instrumentation__.Notify(697359)
			stmtName := fmt.Sprintf("%s-%d", name, i+1)
			s.preparedName = stmtName
			mcp.AddPreparedStatement(stmtName, s.sql)
		}
	} else {
		__antithesis_instrumentation__.Notify(697360)
	}
	__antithesis_instrumentation__.Notify(697353)
	sr.mcp = mcp
	sr.initialized = true
	return nil
}

func (h StmtHandle) check() {
	__antithesis_instrumentation__.Notify(697361)
	if !h.s.sr.initialized {
		__antithesis_instrumentation__.Notify(697362)
		panic("SQLRunner.Init not called")
	} else {
		__antithesis_instrumentation__.Notify(697363)
	}
}

type stmt struct {
	sr  *SQLRunner
	sql string

	preparedName string
}

type StmtHandle struct {
	s *stmt
}

func (h StmtHandle) Exec(ctx context.Context, args ...interface{}) (pgconn.CommandTag, error) {
	__antithesis_instrumentation__.Notify(697364)
	h.check()
	p := h.s.sr.mcp.Get()
	switch h.s.sr.method {
	case prepare:
		__antithesis_instrumentation__.Notify(697365)
		return p.Exec(ctx, h.s.preparedName, args...)

	case noprepare:
		__antithesis_instrumentation__.Notify(697366)
		return p.Exec(ctx, h.s.sql, args...)

	case simple:
		__antithesis_instrumentation__.Notify(697367)
		return p.Exec(ctx, h.s.sql, prependQuerySimpleProtocol(args)...)

	default:
		__antithesis_instrumentation__.Notify(697368)
		panic("invalid method")
	}
}

func (h StmtHandle) ExecTx(
	ctx context.Context, tx pgx.Tx, args ...interface{},
) (pgconn.CommandTag, error) {
	__antithesis_instrumentation__.Notify(697369)
	h.check()
	switch h.s.sr.method {
	case prepare:
		__antithesis_instrumentation__.Notify(697370)
		return tx.Exec(ctx, h.s.preparedName, args...)

	case noprepare:
		__antithesis_instrumentation__.Notify(697371)
		return tx.Exec(ctx, h.s.sql, args...)

	case simple:
		__antithesis_instrumentation__.Notify(697372)
		return tx.Exec(ctx, h.s.sql, prependQuerySimpleProtocol(args)...)

	default:
		__antithesis_instrumentation__.Notify(697373)
		panic("invalid method")
	}
}

func (h StmtHandle) Query(ctx context.Context, args ...interface{}) (pgx.Rows, error) {
	__antithesis_instrumentation__.Notify(697374)
	h.check()
	p := h.s.sr.mcp.Get()
	switch h.s.sr.method {
	case prepare:
		__antithesis_instrumentation__.Notify(697375)
		return p.Query(ctx, h.s.preparedName, args...)

	case noprepare:
		__antithesis_instrumentation__.Notify(697376)
		return p.Query(ctx, h.s.sql, args...)

	case simple:
		__antithesis_instrumentation__.Notify(697377)
		return p.Query(ctx, h.s.sql, prependQuerySimpleProtocol(args)...)

	default:
		__antithesis_instrumentation__.Notify(697378)
		panic("invalid method")
	}
}

func (h StmtHandle) QueryTx(ctx context.Context, tx pgx.Tx, args ...interface{}) (pgx.Rows, error) {
	__antithesis_instrumentation__.Notify(697379)
	h.check()
	switch h.s.sr.method {
	case prepare:
		__antithesis_instrumentation__.Notify(697380)
		return tx.Query(ctx, h.s.preparedName, args...)

	case noprepare:
		__antithesis_instrumentation__.Notify(697381)
		return tx.Query(ctx, h.s.sql, args...)

	case simple:
		__antithesis_instrumentation__.Notify(697382)
		return tx.Query(ctx, h.s.sql, prependQuerySimpleProtocol(args)...)

	default:
		__antithesis_instrumentation__.Notify(697383)
		panic("invalid method")
	}
}

func (h StmtHandle) QueryRow(ctx context.Context, args ...interface{}) pgx.Row {
	__antithesis_instrumentation__.Notify(697384)
	h.check()
	p := h.s.sr.mcp.Get()
	switch h.s.sr.method {
	case prepare:
		__antithesis_instrumentation__.Notify(697385)
		return p.QueryRow(ctx, h.s.preparedName, args...)

	case noprepare:
		__antithesis_instrumentation__.Notify(697386)
		return p.QueryRow(ctx, h.s.sql, args...)

	case simple:
		__antithesis_instrumentation__.Notify(697387)
		return p.QueryRow(ctx, h.s.sql, prependQuerySimpleProtocol(args)...)

	default:
		__antithesis_instrumentation__.Notify(697388)
		panic("invalid method")
	}
}

func (h StmtHandle) QueryRowTx(ctx context.Context, tx pgx.Tx, args ...interface{}) pgx.Row {
	__antithesis_instrumentation__.Notify(697389)
	h.check()
	switch h.s.sr.method {
	case prepare:
		__antithesis_instrumentation__.Notify(697390)
		return tx.QueryRow(ctx, h.s.preparedName, args...)

	case noprepare:
		__antithesis_instrumentation__.Notify(697391)
		return tx.QueryRow(ctx, h.s.sql, args...)

	case simple:
		__antithesis_instrumentation__.Notify(697392)
		return tx.QueryRow(ctx, h.s.sql, prependQuerySimpleProtocol(args)...)

	default:
		__antithesis_instrumentation__.Notify(697393)
		panic("invalid method")
	}
}

func prependQuerySimpleProtocol(args []interface{}) []interface{} {
	__antithesis_instrumentation__.Notify(697394)
	args = append(args, pgx.QuerySimpleProtocol(true))
	copy(args[1:], args)
	args[0] = pgx.QuerySimpleProtocol(true)
	return args
}

var _ = StmtHandle.QueryRow
