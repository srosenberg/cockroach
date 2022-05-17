package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirecancel"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/ring"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
	"github.com/lib/pq/oid"
	"go.opentelemetry.io/otel/attribute"
)

type conn struct {
	conn net.Conn

	sessionArgs sql.SessionArgs
	metrics     *ServerMetrics

	startTime time.Time

	rd bufio.Reader

	parser parser.Parser

	stmtBuf sql.StmtBuf

	res commandResult

	err atomic.Value

	writerState struct {
		fi flushInfo

		buf    bytes.Buffer
		tagBuf [64]byte
	}

	readBuf    pgwirebase.ReadBuffer
	msgBuilder writeBuffer

	vecsScratch coldata.TypedVecs

	sv *settings.Values

	alwaysLogAuthActivity bool

	afterReadMsgTestingKnob func(context.Context) error
}

func (s *Server) serveConn(
	ctx context.Context,
	netConn net.Conn,
	sArgs sql.SessionArgs,
	reserved mon.BoundAccount,
	connStart time.Time,
	authOpt authOptions,
) {
	__antithesis_instrumentation__.Notify(559299)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(559302)
		log.Infof(ctx, "new connection with options: %+v", sArgs)
	} else {
		__antithesis_instrumentation__.Notify(559303)
	}
	__antithesis_instrumentation__.Notify(559300)

	c := newConn(netConn, sArgs, &s.metrics, connStart, &s.execCfg.Settings.SV)
	c.alwaysLogAuthActivity = alwaysLogAuthActivity || func() bool {
		__antithesis_instrumentation__.Notify(559304)
		return atomic.LoadInt32(&s.testingAuthLogEnabled) > 0 == true
	}() == true
	if s.execCfg.PGWireTestingKnobs != nil {
		__antithesis_instrumentation__.Notify(559305)
		c.afterReadMsgTestingKnob = s.execCfg.PGWireTestingKnobs.AfterReadMsgTestingKnob
	} else {
		__antithesis_instrumentation__.Notify(559306)
	}
	__antithesis_instrumentation__.Notify(559301)

	c.serveImpl(ctx, s.IsDraining, s.SQLServer, reserved, authOpt)
}

var alwaysLogAuthActivity = envutil.EnvOrDefaultBool("COCKROACH_ALWAYS_LOG_AUTHN_EVENTS", false)

func newConn(
	netConn net.Conn,
	sArgs sql.SessionArgs,
	metrics *ServerMetrics,
	connStart time.Time,
	sv *settings.Values,
) *conn {
	__antithesis_instrumentation__.Notify(559307)
	c := &conn{
		conn:        netConn,
		sessionArgs: sArgs,
		metrics:     metrics,
		startTime:   connStart,
		rd:          *bufio.NewReader(netConn),
		sv:          sv,
		readBuf:     pgwirebase.MakeReadBuffer(pgwirebase.ReadBufferOptionWithClusterSettings(sv)),
	}
	c.stmtBuf.Init()
	c.res.released = true
	c.writerState.fi.buf = &c.writerState.buf
	c.writerState.fi.lastFlushed = -1
	c.msgBuilder.init(metrics.BytesOutCount)

	return c
}

func (c *conn) setErr(err error) {
	__antithesis_instrumentation__.Notify(559308)
	c.err.Store(err)
}

func (c *conn) GetErr() error {
	__antithesis_instrumentation__.Notify(559309)
	err := c.err.Load()
	if err != nil {
		__antithesis_instrumentation__.Notify(559311)
		return err.(error)
	} else {
		__antithesis_instrumentation__.Notify(559312)
	}
	__antithesis_instrumentation__.Notify(559310)
	return nil
}

func (c *conn) sendError(ctx context.Context, execCfg *sql.ExecutorConfig, err error) error {
	__antithesis_instrumentation__.Notify(559313)

	_ = writeErr(ctx, &execCfg.Settings.SV, err, &c.msgBuilder, c.conn)
	return err
}

func (c *conn) checkMaxConnections(ctx context.Context, sqlServer *sql.Server) error {
	__antithesis_instrumentation__.Notify(559314)
	if c.sessionArgs.IsSuperuser {
		__antithesis_instrumentation__.Notify(559318)

		sqlServer.IncrementConnectionCount()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(559319)
	}
	__antithesis_instrumentation__.Notify(559315)

	maxNumConnectionsValue := maxNumConnections.Get(&sqlServer.GetExecutorConfig().Settings.SV)
	if maxNumConnectionsValue < 0 {
		__antithesis_instrumentation__.Notify(559320)

		sqlServer.IncrementConnectionCount()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(559321)
	}
	__antithesis_instrumentation__.Notify(559316)
	if !sqlServer.IncrementConnectionCountIfLessThan(maxNumConnectionsValue) {
		__antithesis_instrumentation__.Notify(559322)
		return c.sendError(ctx, sqlServer.GetExecutorConfig(), errors.WithHintf(
			pgerror.New(pgcode.TooManyConnections, "sorry, too many clients already"),
			"the maximum number of allowed connections is %d and can be modified using the %s config key",
			maxNumConnectionsValue,
			maxNumConnections.Key(),
		))
	} else {
		__antithesis_instrumentation__.Notify(559323)
	}
	__antithesis_instrumentation__.Notify(559317)
	return nil
}

func (c *conn) authLogEnabled() bool {
	__antithesis_instrumentation__.Notify(559324)
	return c.alwaysLogAuthActivity || func() bool {
		__antithesis_instrumentation__.Notify(559325)
		return logSessionAuth.Get(c.sv) == true
	}() == true
}

const maxRepeatedErrorCount = 1 << 15

func (c *conn) serveImpl(
	ctx context.Context,
	draining func() bool,
	sqlServer *sql.Server,
	reserved mon.BoundAccount,
	authOpt authOptions,
) {
	__antithesis_instrumentation__.Notify(559326)
	defer func() { __antithesis_instrumentation__.Notify(559335); _ = c.conn.Close() }()
	__antithesis_instrumentation__.Notify(559327)

	if c.sessionArgs.User.IsRootUser() || func() bool {
		__antithesis_instrumentation__.Notify(559336)
		return c.sessionArgs.User.IsNodeUser() == true
	}() == true {
		__antithesis_instrumentation__.Notify(559337)
		ctx = logtags.AddTag(ctx, "user", redact.Safe(c.sessionArgs.User))
	} else {
		__antithesis_instrumentation__.Notify(559338)
		ctx = logtags.AddTag(ctx, "user", c.sessionArgs.User)
	}
	__antithesis_instrumentation__.Notify(559328)
	tracing.SpanFromContext(ctx).SetTag("user", attribute.StringValue(c.sessionArgs.User.Normalized()))

	inTestWithoutSQL := sqlServer == nil
	if !inTestWithoutSQL {
		__antithesis_instrumentation__.Notify(559339)
		sessionStart := timeutil.Now()
		defer func() {
			__antithesis_instrumentation__.Notify(559340)
			if c.authLogEnabled() {
				__antithesis_instrumentation__.Notify(559341)
				endTime := timeutil.Now()
				ev := &eventpb.ClientSessionEnd{
					CommonEventDetails:      eventpb.CommonEventDetails{Timestamp: endTime.UnixNano()},
					CommonConnectionDetails: authOpt.connDetails,
					Duration:                endTime.Sub(sessionStart).Nanoseconds(),
				}
				log.StructuredEvent(ctx, ev)
			} else {
				__antithesis_instrumentation__.Notify(559342)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(559343)
	}
	__antithesis_instrumentation__.Notify(559329)

	ctx, cancelConn := context.WithCancel(ctx)
	defer cancelConn()

	var sentDrainSignal bool

	c.conn = NewReadTimeoutConn(c.conn, func() error {
		__antithesis_instrumentation__.Notify(559344)

		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(559347)
			return ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(559348)
		}
		__antithesis_instrumentation__.Notify(559345)

		if !sentDrainSignal && func() bool {
			__antithesis_instrumentation__.Notify(559349)
			return draining() == true
		}() == true {
			__antithesis_instrumentation__.Notify(559350)
			_ = c.stmtBuf.Push(ctx, sql.DrainRequest{})
			sentDrainSignal = true
		} else {
			__antithesis_instrumentation__.Notify(559351)
		}
		__antithesis_instrumentation__.Notify(559346)
		return nil
	})
	__antithesis_instrumentation__.Notify(559330)
	c.rd = *bufio.NewReader(c.conn)

	logAuthn := !inTestWithoutSQL && func() bool {
		__antithesis_instrumentation__.Notify(559352)
		return c.authLogEnabled() == true
	}() == true

	authPipe := newAuthPipe(c, logAuthn, authOpt, c.sessionArgs.User)
	var authenticator authenticatorIO = authPipe

	var procCh <-chan error

	var atomicUnqualifiedIntSize = new(int32)
	onDefaultIntSizeChange := func(newSize int32) {
		__antithesis_instrumentation__.Notify(559353)
		atomic.StoreInt32(atomicUnqualifiedIntSize, newSize)
	}
	__antithesis_instrumentation__.Notify(559331)

	if sqlServer != nil {
		__antithesis_instrumentation__.Notify(559354)

		var ac AuthConn = authPipe
		procCh = c.processCommandsAsync(
			ctx,
			authOpt,
			ac,
			sqlServer,
			reserved,
			cancelConn,
			onDefaultIntSizeChange,
		)
	} else {
		__antithesis_instrumentation__.Notify(559355)

		var err error
		for param, value := range testingStatusReportParams {
			__antithesis_instrumentation__.Notify(559358)
			err = c.sendParamStatus(param, value)
			if err != nil {
				__antithesis_instrumentation__.Notify(559359)
				break
			} else {
				__antithesis_instrumentation__.Notify(559360)
			}
		}
		__antithesis_instrumentation__.Notify(559356)
		if err != nil {
			__antithesis_instrumentation__.Notify(559361)
			reserved.Close(ctx)
			return
		} else {
			__antithesis_instrumentation__.Notify(559362)
		}
		__antithesis_instrumentation__.Notify(559357)
		var ac AuthConn = authPipe

		ac.AuthOK(ctx)
		dummyCh := make(chan error)
		close(dummyCh)
		procCh = dummyCh

		if err := c.sendReadyForQuery(0); err != nil {
			__antithesis_instrumentation__.Notify(559363)
			reserved.Close(ctx)
			return
		} else {
			__antithesis_instrumentation__.Notify(559364)
		}
	}
	__antithesis_instrumentation__.Notify(559332)

	var terminateSeen bool
	var authDone, ignoreUntilSync bool
	var repeatedErrorCount int
	for {
		__antithesis_instrumentation__.Notify(559365)
		breakLoop, isSimpleQuery, err := func() (bool, bool, error) {
			__antithesis_instrumentation__.Notify(559368)
			typ, n, err := c.readBuf.ReadTypedMsg(&c.rd)
			c.metrics.BytesInCount.Inc(int64(n))
			if err == nil && func() bool {
				__antithesis_instrumentation__.Notify(559373)
				return c.afterReadMsgTestingKnob != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(559374)
				err = c.afterReadMsgTestingKnob(ctx)
			} else {
				__antithesis_instrumentation__.Notify(559375)
			}
			__antithesis_instrumentation__.Notify(559369)
			isSimpleQuery := typ == pgwirebase.ClientMsgSimpleQuery
			if err != nil {
				__antithesis_instrumentation__.Notify(559376)
				if pgwirebase.IsMessageTooBigError(err) {
					__antithesis_instrumentation__.Notify(559380)
					log.VInfof(ctx, 1, "pgwire: found big error message; attempting to slurp bytes and return error: %s", err)

					slurpN, slurpErr := c.readBuf.SlurpBytes(&c.rd, pgwirebase.GetMessageTooBigSize(err))
					c.metrics.BytesInCount.Inc(int64(slurpN))
					if slurpErr != nil {
						__antithesis_instrumentation__.Notify(559381)
						return false, isSimpleQuery, errors.Wrap(slurpErr, "pgwire: error slurping remaining bytes")
					} else {
						__antithesis_instrumentation__.Notify(559382)
					}
				} else {
					__antithesis_instrumentation__.Notify(559383)
				}
				__antithesis_instrumentation__.Notify(559377)

				if err := c.stmtBuf.Push(ctx, sql.SendError{Err: err}); err != nil {
					__antithesis_instrumentation__.Notify(559384)
					return false, isSimpleQuery, errors.New("pgwire: error writing too big error message to the client")
				} else {
					__antithesis_instrumentation__.Notify(559385)
				}
				__antithesis_instrumentation__.Notify(559378)

				if isSimpleQuery {
					__antithesis_instrumentation__.Notify(559386)
					if err := c.stmtBuf.Push(ctx, sql.Sync{}); err != nil {
						__antithesis_instrumentation__.Notify(559387)
						return false, isSimpleQuery, errors.New("pgwire: error writing sync to the client whilst message is too big")
					} else {
						__antithesis_instrumentation__.Notify(559388)
					}
				} else {
					__antithesis_instrumentation__.Notify(559389)
				}
				__antithesis_instrumentation__.Notify(559379)

				return false, isSimpleQuery, errors.Wrap(err, "pgwire: error reading input")
			} else {
				__antithesis_instrumentation__.Notify(559390)
			}
			__antithesis_instrumentation__.Notify(559370)
			timeReceived := timeutil.Now()
			log.VEventf(ctx, 2, "pgwire: processing %s", typ)

			if ignoreUntilSync {
				__antithesis_instrumentation__.Notify(559391)
				if typ != pgwirebase.ClientMsgSync {
					__antithesis_instrumentation__.Notify(559393)
					log.VInfof(ctx, 1, "pgwire: skipping non-sync message after encountering error")
					return false, isSimpleQuery, nil
				} else {
					__antithesis_instrumentation__.Notify(559394)
				}
				__antithesis_instrumentation__.Notify(559392)
				ignoreUntilSync = false
			} else {
				__antithesis_instrumentation__.Notify(559395)
			}
			__antithesis_instrumentation__.Notify(559371)

			if !authDone {
				__antithesis_instrumentation__.Notify(559396)
				if typ == pgwirebase.ClientMsgPassword {
					__antithesis_instrumentation__.Notify(559399)
					var pwd []byte
					if pwd, err = c.readBuf.GetBytes(n - 4); err != nil {
						__antithesis_instrumentation__.Notify(559402)
						return false, isSimpleQuery, err
					} else {
						__antithesis_instrumentation__.Notify(559403)
					}
					__antithesis_instrumentation__.Notify(559400)

					if err = authenticator.sendPwdData(pwd); err != nil {
						__antithesis_instrumentation__.Notify(559404)
						return false, isSimpleQuery, err
					} else {
						__antithesis_instrumentation__.Notify(559405)
					}
					__antithesis_instrumentation__.Notify(559401)
					return false, isSimpleQuery, nil
				} else {
					__antithesis_instrumentation__.Notify(559406)
				}
				__antithesis_instrumentation__.Notify(559397)

				if err = authenticator.authResult(); err != nil {
					__antithesis_instrumentation__.Notify(559407)

					return true, isSimpleQuery, nil
				} else {
					__antithesis_instrumentation__.Notify(559408)
				}
				__antithesis_instrumentation__.Notify(559398)
				authDone = true

				duration := timeutil.Since(c.startTime).Nanoseconds()
				c.metrics.ConnLatency.RecordValue(duration)
			} else {
				__antithesis_instrumentation__.Notify(559409)
			}
			__antithesis_instrumentation__.Notify(559372)

			switch typ {
			case pgwirebase.ClientMsgPassword:
				__antithesis_instrumentation__.Notify(559410)

				err = pgwirebase.NewProtocolViolationErrorf("unexpected authentication data")
				return true, isSimpleQuery, writeErr(
					ctx, &sqlServer.GetExecutorConfig().Settings.SV, err,
					&c.msgBuilder, &c.writerState.buf)
			case pgwirebase.ClientMsgSimpleQuery:
				__antithesis_instrumentation__.Notify(559411)
				if err = c.handleSimpleQuery(
					ctx, &c.readBuf, timeReceived, parser.NakedIntTypeFromDefaultIntSize(atomic.LoadInt32(atomicUnqualifiedIntSize)),
				); err != nil {
					__antithesis_instrumentation__.Notify(559424)
					return false, isSimpleQuery, err
				} else {
					__antithesis_instrumentation__.Notify(559425)
				}
				__antithesis_instrumentation__.Notify(559412)
				return false, isSimpleQuery, c.stmtBuf.Push(ctx, sql.Sync{})

			case pgwirebase.ClientMsgExecute:
				__antithesis_instrumentation__.Notify(559413)

				followedBySync := false
				if nextMsgType, err := c.rd.Peek(1); err == nil && func() bool {
					__antithesis_instrumentation__.Notify(559426)
					return pgwirebase.ClientMessageType(nextMsgType[0]) == pgwirebase.ClientMsgSync == true
				}() == true {
					__antithesis_instrumentation__.Notify(559427)
					followedBySync = true
				} else {
					__antithesis_instrumentation__.Notify(559428)
				}
				__antithesis_instrumentation__.Notify(559414)
				return false, isSimpleQuery, c.handleExecute(ctx, &c.readBuf, timeReceived, followedBySync)

			case pgwirebase.ClientMsgParse:
				__antithesis_instrumentation__.Notify(559415)
				return false, isSimpleQuery, c.handleParse(ctx, &c.readBuf, parser.NakedIntTypeFromDefaultIntSize(atomic.LoadInt32(atomicUnqualifiedIntSize)))

			case pgwirebase.ClientMsgDescribe:
				__antithesis_instrumentation__.Notify(559416)
				return false, isSimpleQuery, c.handleDescribe(ctx, &c.readBuf)

			case pgwirebase.ClientMsgBind:
				__antithesis_instrumentation__.Notify(559417)
				return false, isSimpleQuery, c.handleBind(ctx, &c.readBuf)

			case pgwirebase.ClientMsgClose:
				__antithesis_instrumentation__.Notify(559418)
				return false, isSimpleQuery, c.handleClose(ctx, &c.readBuf)

			case pgwirebase.ClientMsgTerminate:
				__antithesis_instrumentation__.Notify(559419)
				terminateSeen = true
				return true, isSimpleQuery, nil

			case pgwirebase.ClientMsgSync:
				__antithesis_instrumentation__.Notify(559420)

				return false, isSimpleQuery, c.stmtBuf.Push(ctx, sql.Sync{})

			case pgwirebase.ClientMsgFlush:
				__antithesis_instrumentation__.Notify(559421)
				return false, isSimpleQuery, c.handleFlush(ctx)

			case pgwirebase.ClientMsgCopyData, pgwirebase.ClientMsgCopyDone, pgwirebase.ClientMsgCopyFail:
				__antithesis_instrumentation__.Notify(559422)

				return false, isSimpleQuery, nil
			default:
				__antithesis_instrumentation__.Notify(559423)
				return false, isSimpleQuery, c.stmtBuf.Push(
					ctx,
					sql.SendError{Err: pgwirebase.NewUnrecognizedMsgTypeErr(typ)})
			}
		}()
		__antithesis_instrumentation__.Notify(559366)
		if err != nil {
			__antithesis_instrumentation__.Notify(559429)
			log.VEventf(ctx, 1, "pgwire: error processing message: %s", err)
			if !isSimpleQuery {
				__antithesis_instrumentation__.Notify(559431)

				ignoreUntilSync = true
			} else {
				__antithesis_instrumentation__.Notify(559432)
			}
			__antithesis_instrumentation__.Notify(559430)
			repeatedErrorCount++

			if netutil.IsClosedConnection(err) || func() bool {
				__antithesis_instrumentation__.Notify(559433)
				return errors.Is(err, context.Canceled) == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(559434)
				return repeatedErrorCount > maxRepeatedErrorCount == true
			}() == true {
				__antithesis_instrumentation__.Notify(559435)
				break
			} else {
				__antithesis_instrumentation__.Notify(559436)
			}
		} else {
			__antithesis_instrumentation__.Notify(559437)
			repeatedErrorCount = 0
		}
		__antithesis_instrumentation__.Notify(559367)
		if breakLoop {
			__antithesis_instrumentation__.Notify(559438)
			break
		} else {
			__antithesis_instrumentation__.Notify(559439)
		}
	}
	__antithesis_instrumentation__.Notify(559333)

	c.stmtBuf.Close()

	cancelConn()

	authenticator.noMorePwdData()

	<-procCh

	if terminateSeen {
		__antithesis_instrumentation__.Notify(559440)
		return
	} else {
		__antithesis_instrumentation__.Notify(559441)
	}
	__antithesis_instrumentation__.Notify(559334)

	if draining() {
		__antithesis_instrumentation__.Notify(559442)

		log.Ops.Info(ctx, "closing existing connection while server is draining")
		_ = writeErr(ctx, &sqlServer.GetExecutorConfig().Settings.SV,
			newAdminShutdownErr(ErrDrainingExistingConn), &c.msgBuilder, &c.writerState.buf)
		_, _ = c.writerState.buf.WriteTo(c.conn)
	} else {
		__antithesis_instrumentation__.Notify(559443)
	}
}

func (c *conn) processCommandsAsync(
	ctx context.Context,
	authOpt authOptions,
	ac AuthConn,
	sqlServer *sql.Server,
	reserved mon.BoundAccount,
	cancelConn func(),
	onDefaultIntSizeChange func(newSize int32),
) <-chan error {
	__antithesis_instrumentation__.Notify(559444)

	reservedOwned := true
	retCh := make(chan error, 1)
	go func() {
		__antithesis_instrumentation__.Notify(559446)
		var retErr error
		var connHandler sql.ConnectionHandler
		var authOK bool
		var connCloseAuthHandler func()
		defer func() {
			__antithesis_instrumentation__.Notify(559452)

			if reservedOwned {
				__antithesis_instrumentation__.Notify(559457)
				reserved.Close(ctx)
			} else {
				__antithesis_instrumentation__.Notify(559458)
			}
			__antithesis_instrumentation__.Notify(559453)

			cancelConn()

			pgwireKnobs := sqlServer.GetExecutorConfig().PGWireTestingKnobs
			if pgwireKnobs != nil && func() bool {
				__antithesis_instrumentation__.Notify(559459)
				return pgwireKnobs.CatchPanics == true
			}() == true {
				__antithesis_instrumentation__.Notify(559460)
				if r := recover(); r != nil {
					__antithesis_instrumentation__.Notify(559461)

					if err, ok := r.(error); ok {
						__antithesis_instrumentation__.Notify(559463)

						retErr = errors.Handled(err)
					} else {
						__antithesis_instrumentation__.Notify(559464)
						retErr = errors.Newf("%+v", r)
					}
					__antithesis_instrumentation__.Notify(559462)
					retErr = pgerror.WithCandidateCode(retErr, pgcode.CrashShutdown)

					retErr = errors.Wrap(retErr, "caught fatal error")
					_ = writeErr(
						ctx, &sqlServer.GetExecutorConfig().Settings.SV, retErr,
						&c.msgBuilder, &c.writerState.buf)
					_, _ = c.writerState.buf.WriteTo(c.conn)
					c.stmtBuf.Close()

					c.bufferReadyForQuery('I')
				} else {
					__antithesis_instrumentation__.Notify(559465)
				}
			} else {
				__antithesis_instrumentation__.Notify(559466)
			}
			__antithesis_instrumentation__.Notify(559454)
			if !authOK {
				__antithesis_instrumentation__.Notify(559467)
				ac.AuthFail(retErr)
			} else {
				__antithesis_instrumentation__.Notify(559468)
			}
			__antithesis_instrumentation__.Notify(559455)
			if connCloseAuthHandler != nil {
				__antithesis_instrumentation__.Notify(559469)
				connCloseAuthHandler()
			} else {
				__antithesis_instrumentation__.Notify(559470)
			}
			__antithesis_instrumentation__.Notify(559456)

			retCh <- retErr
		}()
		__antithesis_instrumentation__.Notify(559447)

		if connCloseAuthHandler, retErr = c.handleAuthentication(
			ctx, ac, authOpt, sqlServer.GetExecutorConfig(),
		); retErr != nil {
			__antithesis_instrumentation__.Notify(559471)

			return
		} else {
			__antithesis_instrumentation__.Notify(559472)
		}
		__antithesis_instrumentation__.Notify(559448)

		if retErr = c.checkMaxConnections(ctx, sqlServer); retErr != nil {
			__antithesis_instrumentation__.Notify(559473)
			return
		} else {
			__antithesis_instrumentation__.Notify(559474)
		}
		__antithesis_instrumentation__.Notify(559449)
		defer sqlServer.DecrementConnectionCount()

		if retErr = c.authOKMessage(); retErr != nil {
			__antithesis_instrumentation__.Notify(559475)
			return
		} else {
			__antithesis_instrumentation__.Notify(559476)
		}
		__antithesis_instrumentation__.Notify(559450)

		connHandler, retErr = c.sendInitialConnData(ctx, sqlServer, onDefaultIntSizeChange)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(559477)
			return
		} else {
			__antithesis_instrumentation__.Notify(559478)
		}
		__antithesis_instrumentation__.Notify(559451)

		ac.AuthOK(ctx)

		authOK = true

		reservedOwned = false
		retErr = sqlServer.ServeConn(ctx, connHandler, reserved, cancelConn)
	}()
	__antithesis_instrumentation__.Notify(559445)
	return retCh
}

func (c *conn) sendParamStatus(param, value string) error {
	__antithesis_instrumentation__.Notify(559479)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgParameterStatus)
	c.msgBuilder.writeTerminatedString(param)
	c.msgBuilder.writeTerminatedString(value)
	return c.msgBuilder.finishMsg(c.conn)
}

func (c *conn) bufferParamStatus(param, value string) error {
	__antithesis_instrumentation__.Notify(559480)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgParameterStatus)
	c.msgBuilder.writeTerminatedString(param)
	c.msgBuilder.writeTerminatedString(value)
	return c.msgBuilder.finishMsg(&c.writerState.buf)
}

func (c *conn) bufferNotice(ctx context.Context, noticeErr pgnotice.Notice) error {
	__antithesis_instrumentation__.Notify(559481)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgNoticeResponse)
	return writeErrFields(ctx, c.sv, noticeErr, &c.msgBuilder, &c.writerState.buf)
}

func (c *conn) sendInitialConnData(
	ctx context.Context, sqlServer *sql.Server, onDefaultIntSizeChange func(newSize int32),
) (sql.ConnectionHandler, error) {
	__antithesis_instrumentation__.Notify(559482)
	connHandler, err := sqlServer.SetupConn(
		ctx,
		c.sessionArgs,
		&c.stmtBuf,
		c,
		c.metrics.SQLMemMetrics,
		onDefaultIntSizeChange,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(559487)
		_ = writeErr(
			ctx, &sqlServer.GetExecutorConfig().Settings.SV, err, &c.msgBuilder, c.conn)
		return sql.ConnectionHandler{}, err
	} else {
		__antithesis_instrumentation__.Notify(559488)
	}
	__antithesis_instrumentation__.Notify(559483)

	for _, param := range statusReportParams {
		__antithesis_instrumentation__.Notify(559489)
		param := param
		value := connHandler.GetParamStatus(ctx, param)
		if err := c.sendParamStatus(param, value); err != nil {
			__antithesis_instrumentation__.Notify(559490)
			return sql.ConnectionHandler{}, err
		} else {
			__antithesis_instrumentation__.Notify(559491)
		}
	}
	__antithesis_instrumentation__.Notify(559484)

	if err := c.sendParamStatus("session_authorization", c.sessionArgs.User.Normalized()); err != nil {
		__antithesis_instrumentation__.Notify(559492)
		return sql.ConnectionHandler{}, err
	} else {
		__antithesis_instrumentation__.Notify(559493)
	}
	__antithesis_instrumentation__.Notify(559485)

	if err := c.sendReadyForQuery(connHandler.GetQueryCancelKey()); err != nil {
		__antithesis_instrumentation__.Notify(559494)
		return sql.ConnectionHandler{}, err
	} else {
		__antithesis_instrumentation__.Notify(559495)
	}
	__antithesis_instrumentation__.Notify(559486)
	return connHandler, nil
}

func (c *conn) sendReadyForQuery(queryCancelKey pgwirecancel.BackendKeyData) error {
	__antithesis_instrumentation__.Notify(559496)

	c.msgBuilder.initMsg(pgwirebase.ServerMsgBackendKeyData)
	c.msgBuilder.putInt64(int64(queryCancelKey))
	if err := c.msgBuilder.finishMsg(c.conn); err != nil {
		__antithesis_instrumentation__.Notify(559499)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559500)
	}
	__antithesis_instrumentation__.Notify(559497)

	c.msgBuilder.initMsg(pgwirebase.ServerMsgReady)
	c.msgBuilder.writeByte(byte(sql.IdleTxnBlock))
	if err := c.msgBuilder.finishMsg(c.conn); err != nil {
		__antithesis_instrumentation__.Notify(559501)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559502)
	}
	__antithesis_instrumentation__.Notify(559498)
	return nil
}

func (c *conn) handleSimpleQuery(
	ctx context.Context,
	buf *pgwirebase.ReadBuffer,
	timeReceived time.Time,
	unqualifiedIntSize *types.T,
) error {
	__antithesis_instrumentation__.Notify(559503)
	query, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559508)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559509)
	}
	__antithesis_instrumentation__.Notify(559504)

	startParse := timeutil.Now()
	stmts, err := c.parser.ParseWithInt(query, unqualifiedIntSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(559510)
		log.SqlExec.Errorf(ctx, "failed to parse simple query: %s", query)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559511)
	}
	__antithesis_instrumentation__.Notify(559505)
	endParse := timeutil.Now()

	if len(stmts) == 0 {
		__antithesis_instrumentation__.Notify(559512)
		return c.stmtBuf.Push(
			ctx, sql.ExecStmt{
				Statement:    parser.Statement{},
				TimeReceived: timeReceived,
				ParseStart:   startParse,
				ParseEnd:     endParse,
			})
	} else {
		__antithesis_instrumentation__.Notify(559513)
	}
	__antithesis_instrumentation__.Notify(559506)

	for i := range stmts {
		__antithesis_instrumentation__.Notify(559514)

		if cp, ok := stmts[i].AST.(*tree.CopyFrom); ok {
			__antithesis_instrumentation__.Notify(559516)
			if len(stmts) != 1 {
				__antithesis_instrumentation__.Notify(559519)

				return c.stmtBuf.Push(
					ctx,
					sql.SendError{
						Err: pgwirebase.NewProtocolViolationErrorf(
							"COPY together with other statements in a query string is not supported"),
					})
			} else {
				__antithesis_instrumentation__.Notify(559520)
			}
			__antithesis_instrumentation__.Notify(559517)
			copyDone := sync.WaitGroup{}
			copyDone.Add(1)
			if err := c.stmtBuf.Push(ctx, sql.CopyIn{Conn: c, Stmt: cp, CopyDone: &copyDone}); err != nil {
				__antithesis_instrumentation__.Notify(559521)
				return err
			} else {
				__antithesis_instrumentation__.Notify(559522)
			}
			__antithesis_instrumentation__.Notify(559518)
			copyDone.Wait()
			return nil
		} else {
			__antithesis_instrumentation__.Notify(559523)
		}
		__antithesis_instrumentation__.Notify(559515)

		if err := c.stmtBuf.Push(
			ctx,
			sql.ExecStmt{
				Statement:    stmts[i],
				TimeReceived: timeReceived,
				ParseStart:   startParse,
				ParseEnd:     endParse,
				LastInBatch:  i == len(stmts)-1,
			}); err != nil {
			__antithesis_instrumentation__.Notify(559524)
			return err
		} else {
			__antithesis_instrumentation__.Notify(559525)
		}
	}
	__antithesis_instrumentation__.Notify(559507)
	return nil
}

func (c *conn) handleParse(
	ctx context.Context, buf *pgwirebase.ReadBuffer, nakedIntSize *types.T,
) error {
	__antithesis_instrumentation__.Notify(559526)
	telemetry.Inc(sqltelemetry.ParseRequestCounter)
	name, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559537)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559538)
	}
	__antithesis_instrumentation__.Notify(559527)
	query, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559539)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559540)
	}
	__antithesis_instrumentation__.Notify(559528)

	numQArgTypes, err := buf.GetUint16()
	if err != nil {
		__antithesis_instrumentation__.Notify(559541)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559542)
	}
	__antithesis_instrumentation__.Notify(559529)
	inTypeHints := make([]oid.Oid, numQArgTypes)
	for i := range inTypeHints {
		__antithesis_instrumentation__.Notify(559543)
		typ, err := buf.GetUint32()
		if err != nil {
			__antithesis_instrumentation__.Notify(559545)
			return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
		} else {
			__antithesis_instrumentation__.Notify(559546)
		}
		__antithesis_instrumentation__.Notify(559544)
		inTypeHints[i] = oid.Oid(typ)
	}
	__antithesis_instrumentation__.Notify(559530)

	startParse := timeutil.Now()
	stmts, err := c.parser.ParseWithInt(query, nakedIntSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(559547)
		log.SqlExec.Errorf(ctx, "failed to parse: %s", query)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559548)
	}
	__antithesis_instrumentation__.Notify(559531)
	if len(stmts) > 1 {
		__antithesis_instrumentation__.Notify(559549)
		err := pgerror.WrongNumberOfPreparedStatements(len(stmts))
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559550)
	}
	__antithesis_instrumentation__.Notify(559532)
	var stmt parser.Statement
	if len(stmts) == 1 {
		__antithesis_instrumentation__.Notify(559551)
		stmt = stmts[0]
	} else {
		__antithesis_instrumentation__.Notify(559552)
	}
	__antithesis_instrumentation__.Notify(559533)

	if len(inTypeHints) > stmt.NumPlaceholders {
		__antithesis_instrumentation__.Notify(559553)
		err := pgwirebase.NewProtocolViolationErrorf(
			"received too many type hints: %d vs %d placeholders in query",
			len(inTypeHints), stmt.NumPlaceholders,
		)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559554)
	}
	__antithesis_instrumentation__.Notify(559534)

	var sqlTypeHints tree.PlaceholderTypes
	if len(inTypeHints) > 0 {
		__antithesis_instrumentation__.Notify(559555)

		sqlTypeHints = make(tree.PlaceholderTypes, stmt.NumPlaceholders)
		for i, t := range inTypeHints {
			__antithesis_instrumentation__.Notify(559556)
			if t == 0 {
				__antithesis_instrumentation__.Notify(559560)
				continue
			} else {
				__antithesis_instrumentation__.Notify(559561)
			}
			__antithesis_instrumentation__.Notify(559557)

			if t == oid.T_unknown || func() bool {
				__antithesis_instrumentation__.Notify(559562)
				return types.IsOIDUserDefinedType(t) == true
			}() == true {
				__antithesis_instrumentation__.Notify(559563)
				sqlTypeHints[i] = nil
				continue
			} else {
				__antithesis_instrumentation__.Notify(559564)
			}
			__antithesis_instrumentation__.Notify(559558)
			v, ok := types.OidToType[t]
			if !ok {
				__antithesis_instrumentation__.Notify(559565)
				err := pgwirebase.NewProtocolViolationErrorf("unknown oid type: %v", t)
				return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
			} else {
				__antithesis_instrumentation__.Notify(559566)
			}
			__antithesis_instrumentation__.Notify(559559)
			sqlTypeHints[i] = v
		}
	} else {
		__antithesis_instrumentation__.Notify(559567)
	}
	__antithesis_instrumentation__.Notify(559535)

	endParse := timeutil.Now()

	if _, ok := stmt.AST.(*tree.CopyFrom); ok {
		__antithesis_instrumentation__.Notify(559568)

		return c.stmtBuf.Push(ctx, sql.SendError{Err: fmt.Errorf("CopyFrom not supported in extended protocol mode")})
	} else {
		__antithesis_instrumentation__.Notify(559569)
	}
	__antithesis_instrumentation__.Notify(559536)

	return c.stmtBuf.Push(
		ctx,
		sql.PrepareStmt{
			Name:         name,
			Statement:    stmt,
			TypeHints:    sqlTypeHints,
			RawTypeHints: inTypeHints,
			ParseStart:   startParse,
			ParseEnd:     endParse,
		})
}

func (c *conn) handleDescribe(ctx context.Context, buf *pgwirebase.ReadBuffer) error {
	__antithesis_instrumentation__.Notify(559570)
	telemetry.Inc(sqltelemetry.DescribeRequestCounter)
	typ, err := buf.GetPrepareType()
	if err != nil {
		__antithesis_instrumentation__.Notify(559573)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559574)
	}
	__antithesis_instrumentation__.Notify(559571)
	name, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559575)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559576)
	}
	__antithesis_instrumentation__.Notify(559572)
	return c.stmtBuf.Push(
		ctx,
		sql.DescribeStmt{
			Name: name,
			Type: typ,
		})
}

func (c *conn) handleClose(ctx context.Context, buf *pgwirebase.ReadBuffer) error {
	__antithesis_instrumentation__.Notify(559577)
	telemetry.Inc(sqltelemetry.CloseRequestCounter)
	typ, err := buf.GetPrepareType()
	if err != nil {
		__antithesis_instrumentation__.Notify(559580)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559581)
	}
	__antithesis_instrumentation__.Notify(559578)
	name, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559582)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559583)
	}
	__antithesis_instrumentation__.Notify(559579)
	return c.stmtBuf.Push(
		ctx,
		sql.DeletePreparedStmt{
			Name: name,
			Type: typ,
		})
}

var formatCodesAllText = []pgwirebase.FormatCode{pgwirebase.FormatText}

func (c *conn) handleBind(ctx context.Context, buf *pgwirebase.ReadBuffer) error {
	__antithesis_instrumentation__.Notify(559584)
	telemetry.Inc(sqltelemetry.BindRequestCounter)
	portalName, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559593)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559594)
	}
	__antithesis_instrumentation__.Notify(559585)
	statementName, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559595)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559596)
	}
	__antithesis_instrumentation__.Notify(559586)

	numQArgFormatCodes, err := buf.GetUint16()
	if err != nil {
		__antithesis_instrumentation__.Notify(559597)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559598)
	}
	__antithesis_instrumentation__.Notify(559587)
	var qArgFormatCodes []pgwirebase.FormatCode
	switch numQArgFormatCodes {
	case 0:
		__antithesis_instrumentation__.Notify(559599)

		qArgFormatCodes = formatCodesAllText
	case 1:
		__antithesis_instrumentation__.Notify(559600)

		ch, err := buf.GetUint16()
		if err != nil {
			__antithesis_instrumentation__.Notify(559603)
			return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
		} else {
			__antithesis_instrumentation__.Notify(559604)
		}
		__antithesis_instrumentation__.Notify(559601)
		code := pgwirebase.FormatCode(ch)
		if code == pgwirebase.FormatText {
			__antithesis_instrumentation__.Notify(559605)
			qArgFormatCodes = formatCodesAllText
		} else {
			__antithesis_instrumentation__.Notify(559606)
			qArgFormatCodes = []pgwirebase.FormatCode{code}
		}
	default:
		__antithesis_instrumentation__.Notify(559602)
		qArgFormatCodes = make([]pgwirebase.FormatCode, numQArgFormatCodes)

		for i := range qArgFormatCodes {
			__antithesis_instrumentation__.Notify(559607)
			ch, err := buf.GetUint16()
			if err != nil {
				__antithesis_instrumentation__.Notify(559609)
				return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
			} else {
				__antithesis_instrumentation__.Notify(559610)
			}
			__antithesis_instrumentation__.Notify(559608)
			qArgFormatCodes[i] = pgwirebase.FormatCode(ch)
		}
	}
	__antithesis_instrumentation__.Notify(559588)

	numValues, err := buf.GetUint16()
	if err != nil {
		__antithesis_instrumentation__.Notify(559611)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559612)
	}
	__antithesis_instrumentation__.Notify(559589)
	qargs := make([][]byte, numValues)
	for i := 0; i < int(numValues); i++ {
		__antithesis_instrumentation__.Notify(559613)
		plen, err := buf.GetUint32()
		if err != nil {
			__antithesis_instrumentation__.Notify(559617)
			return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
		} else {
			__antithesis_instrumentation__.Notify(559618)
		}
		__antithesis_instrumentation__.Notify(559614)
		if int32(plen) == -1 {
			__antithesis_instrumentation__.Notify(559619)

			qargs[i] = nil
			continue
		} else {
			__antithesis_instrumentation__.Notify(559620)
		}
		__antithesis_instrumentation__.Notify(559615)
		b, err := buf.GetBytes(int(plen))
		if err != nil {
			__antithesis_instrumentation__.Notify(559621)
			return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
		} else {
			__antithesis_instrumentation__.Notify(559622)
		}
		__antithesis_instrumentation__.Notify(559616)
		qargs[i] = b
	}
	__antithesis_instrumentation__.Notify(559590)

	numColumnFormatCodes, err := buf.GetUint16()
	if err != nil {
		__antithesis_instrumentation__.Notify(559623)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559624)
	}
	__antithesis_instrumentation__.Notify(559591)
	var columnFormatCodes []pgwirebase.FormatCode
	switch numColumnFormatCodes {
	case 0:
		__antithesis_instrumentation__.Notify(559625)

		columnFormatCodes = formatCodesAllText
	case 1:
		__antithesis_instrumentation__.Notify(559626)

		ch, err := buf.GetUint16()
		if err != nil {
			__antithesis_instrumentation__.Notify(559629)
			return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
		} else {
			__antithesis_instrumentation__.Notify(559630)
		}
		__antithesis_instrumentation__.Notify(559627)
		code := pgwirebase.FormatCode(ch)
		if code == pgwirebase.FormatText {
			__antithesis_instrumentation__.Notify(559631)
			columnFormatCodes = formatCodesAllText
		} else {
			__antithesis_instrumentation__.Notify(559632)
			columnFormatCodes = []pgwirebase.FormatCode{code}
		}
	default:
		__antithesis_instrumentation__.Notify(559628)
		columnFormatCodes = make([]pgwirebase.FormatCode, numColumnFormatCodes)

		for i := range columnFormatCodes {
			__antithesis_instrumentation__.Notify(559633)
			ch, err := buf.GetUint16()
			if err != nil {
				__antithesis_instrumentation__.Notify(559635)
				return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
			} else {
				__antithesis_instrumentation__.Notify(559636)
			}
			__antithesis_instrumentation__.Notify(559634)
			columnFormatCodes[i] = pgwirebase.FormatCode(ch)
		}
	}
	__antithesis_instrumentation__.Notify(559592)
	return c.stmtBuf.Push(
		ctx,
		sql.BindStmt{
			PreparedStatementName: statementName,
			PortalName:            portalName,
			Args:                  qargs,
			ArgFormatCodes:        qArgFormatCodes,
			OutFormats:            columnFormatCodes,
		})
}

func (c *conn) handleExecute(
	ctx context.Context, buf *pgwirebase.ReadBuffer, timeReceived time.Time, followedBySync bool,
) error {
	__antithesis_instrumentation__.Notify(559637)
	telemetry.Inc(sqltelemetry.ExecuteRequestCounter)
	portalName, err := buf.GetString()
	if err != nil {
		__antithesis_instrumentation__.Notify(559640)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559641)
	}
	__antithesis_instrumentation__.Notify(559638)
	limit, err := buf.GetUint32()
	if err != nil {
		__antithesis_instrumentation__.Notify(559642)
		return c.stmtBuf.Push(ctx, sql.SendError{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(559643)
	}
	__antithesis_instrumentation__.Notify(559639)
	return c.stmtBuf.Push(ctx, sql.ExecPortal{
		Name:           portalName,
		TimeReceived:   timeReceived,
		Limit:          int(limit),
		FollowedBySync: followedBySync,
	})
}

func (c *conn) handleFlush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(559644)
	telemetry.Inc(sqltelemetry.FlushRequestCounter)
	return c.stmtBuf.Push(ctx, sql.Flush{})
}

func (c *conn) BeginCopyIn(
	ctx context.Context, columns []colinfo.ResultColumn, format pgwirebase.FormatCode,
) error {
	__antithesis_instrumentation__.Notify(559645)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgCopyInResponse)
	c.msgBuilder.writeByte(byte(format))
	c.msgBuilder.putInt16(int16(len(columns)))
	for range columns {
		__antithesis_instrumentation__.Notify(559647)
		c.msgBuilder.putInt16(int16(format))
	}
	__antithesis_instrumentation__.Notify(559646)
	return c.msgBuilder.finishMsg(c.conn)
}

func (c *conn) SendCommandComplete(tag []byte) error {
	__antithesis_instrumentation__.Notify(559648)
	c.bufferCommandComplete(tag)
	return nil
}

func (c *conn) Rd() pgwirebase.BufferedReader {
	__antithesis_instrumentation__.Notify(559649)
	return &pgwireReader{conn: c}
}

type flushInfo struct {
	buf *bytes.Buffer

	lastFlushed sql.CmdPos

	cmdStarts cmdIdxBuffer
}

type cmdIdx struct {
	pos sql.CmdPos
	idx int
}

var cmdIdxPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(559650)
		return &cmdIdx{}
	},
}

func (c *cmdIdx) release() {
	__antithesis_instrumentation__.Notify(559651)
	*c = cmdIdx{}
	cmdIdxPool.Put(c)
}

type cmdIdxBuffer struct {
	buffer ring.Buffer
}

func (b *cmdIdxBuffer) empty() bool {
	__antithesis_instrumentation__.Notify(559652)
	return b.buffer.Len() == 0
}

func (b *cmdIdxBuffer) addLast(pos sql.CmdPos, idx int) {
	__antithesis_instrumentation__.Notify(559653)
	cmdIdx := cmdIdxPool.Get().(*cmdIdx)
	cmdIdx.pos = pos
	cmdIdx.idx = idx
	b.buffer.AddLast(cmdIdx)
}

func (b *cmdIdxBuffer) removeLast() {
	__antithesis_instrumentation__.Notify(559654)
	b.getLast().release()
	b.buffer.RemoveLast()
}

func (b *cmdIdxBuffer) getLast() *cmdIdx {
	__antithesis_instrumentation__.Notify(559655)
	return b.buffer.GetLast().(*cmdIdx)
}

func (b *cmdIdxBuffer) clear() {
	__antithesis_instrumentation__.Notify(559656)
	for !b.empty() {
		__antithesis_instrumentation__.Notify(559657)
		b.removeLast()
	}
}

func (fi *flushInfo) registerCmd(pos sql.CmdPos) {
	__antithesis_instrumentation__.Notify(559658)
	if !fi.cmdStarts.empty() && func() bool {
		__antithesis_instrumentation__.Notify(559660)
		return fi.cmdStarts.getLast().pos >= pos == true
	}() == true {
		__antithesis_instrumentation__.Notify(559661)

		return
	} else {
		__antithesis_instrumentation__.Notify(559662)
	}
	__antithesis_instrumentation__.Notify(559659)
	fi.cmdStarts.addLast(pos, fi.buf.Len())
}

func cookTag(
	tagStr string, buf []byte, stmtType tree.StatementReturnType, rowsAffected int,
) []byte {
	__antithesis_instrumentation__.Notify(559663)
	if tagStr == "INSERT" {
		__antithesis_instrumentation__.Notify(559666)

		tagStr = "INSERT 0"
	} else {
		__antithesis_instrumentation__.Notify(559667)
	}
	__antithesis_instrumentation__.Notify(559664)
	tag := append(buf, tagStr...)

	switch stmtType {
	case tree.RowsAffected:
		__antithesis_instrumentation__.Notify(559668)
		tag = append(tag, ' ')
		tag = strconv.AppendInt(tag, int64(rowsAffected), 10)

	case tree.Rows:
		__antithesis_instrumentation__.Notify(559669)
		tag = append(tag, ' ')
		tag = strconv.AppendUint(tag, uint64(rowsAffected), 10)

	case tree.Ack, tree.DDL:
		__antithesis_instrumentation__.Notify(559670)
		if tagStr == "SELECT" {
			__antithesis_instrumentation__.Notify(559673)
			tag = append(tag, ' ')
			tag = strconv.AppendInt(tag, int64(rowsAffected), 10)
		} else {
			__antithesis_instrumentation__.Notify(559674)
		}

	case tree.CopyIn:
		__antithesis_instrumentation__.Notify(559671)

		panic(errors.AssertionFailedf("CopyIn statements should have been handled elsewhere " +
			"and not produce results"))
	default:
		__antithesis_instrumentation__.Notify(559672)
		panic(errors.AssertionFailedf("unexpected result type %v", stmtType))
	}
	__antithesis_instrumentation__.Notify(559665)

	return tag
}

func (c *conn) bufferRow(
	ctx context.Context,
	row tree.Datums,
	formatCodes []pgwirebase.FormatCode,
	conv sessiondatapb.DataConversionConfig,
	sessionLoc *time.Location,
	types []*types.T,
) {
	__antithesis_instrumentation__.Notify(559675)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgDataRow)
	c.msgBuilder.putInt16(int16(len(row)))
	for i, col := range row {
		__antithesis_instrumentation__.Notify(559677)
		fmtCode := pgwirebase.FormatText
		if formatCodes != nil {
			__antithesis_instrumentation__.Notify(559679)
			fmtCode = formatCodes[i]
		} else {
			__antithesis_instrumentation__.Notify(559680)
		}
		__antithesis_instrumentation__.Notify(559678)
		switch fmtCode {
		case pgwirebase.FormatText:
			__antithesis_instrumentation__.Notify(559681)
			c.msgBuilder.writeTextDatum(ctx, col, conv, sessionLoc, types[i])
		case pgwirebase.FormatBinary:
			__antithesis_instrumentation__.Notify(559682)
			c.msgBuilder.writeBinaryDatum(ctx, col, sessionLoc, types[i])
		default:
			__antithesis_instrumentation__.Notify(559683)
			c.msgBuilder.setError(errors.Errorf("unsupported format code %s", fmtCode))
		}
	}
	__antithesis_instrumentation__.Notify(559676)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559684)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559685)
	}
}

func (c *conn) bufferBatch(
	ctx context.Context,
	batch coldata.Batch,
	formatCodes []pgwirebase.FormatCode,
	conv sessiondatapb.DataConversionConfig,
	sessionLoc *time.Location,
) {
	__antithesis_instrumentation__.Notify(559686)
	sel := batch.Selection()
	n := batch.Length()
	if n > 0 {
		__antithesis_instrumentation__.Notify(559687)
		c.vecsScratch.SetBatch(batch)

		defer c.vecsScratch.Reset()
		width := int16(len(c.vecsScratch.Vecs))
		for i := 0; i < n; i++ {
			__antithesis_instrumentation__.Notify(559688)
			rowIdx := i
			if sel != nil {
				__antithesis_instrumentation__.Notify(559691)
				rowIdx = sel[rowIdx]
			} else {
				__antithesis_instrumentation__.Notify(559692)
			}
			__antithesis_instrumentation__.Notify(559689)
			c.msgBuilder.initMsg(pgwirebase.ServerMsgDataRow)
			c.msgBuilder.putInt16(width)
			for vecIdx := 0; vecIdx < len(c.vecsScratch.Vecs); vecIdx++ {
				__antithesis_instrumentation__.Notify(559693)
				fmtCode := pgwirebase.FormatText
				if formatCodes != nil {
					__antithesis_instrumentation__.Notify(559695)
					fmtCode = formatCodes[vecIdx]
				} else {
					__antithesis_instrumentation__.Notify(559696)
				}
				__antithesis_instrumentation__.Notify(559694)
				switch fmtCode {
				case pgwirebase.FormatText:
					__antithesis_instrumentation__.Notify(559697)
					c.msgBuilder.writeTextColumnarElement(ctx, &c.vecsScratch, vecIdx, rowIdx, conv, sessionLoc)
				case pgwirebase.FormatBinary:
					__antithesis_instrumentation__.Notify(559698)
					c.msgBuilder.writeBinaryColumnarElement(ctx, &c.vecsScratch, vecIdx, rowIdx, sessionLoc)
				default:
					__antithesis_instrumentation__.Notify(559699)
					c.msgBuilder.setError(errors.Errorf("unsupported format code %s", fmtCode))
				}
			}
			__antithesis_instrumentation__.Notify(559690)
			if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
				__antithesis_instrumentation__.Notify(559700)
				panic(fmt.Sprintf("unexpected err from buffer: %s", err))
			} else {
				__antithesis_instrumentation__.Notify(559701)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(559702)
	}
}

func (c *conn) bufferReadyForQuery(txnStatus byte) {
	__antithesis_instrumentation__.Notify(559703)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgReady)
	c.msgBuilder.writeByte(txnStatus)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559704)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559705)
	}
}

func (c *conn) bufferParseComplete() {
	__antithesis_instrumentation__.Notify(559706)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgParseComplete)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559707)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559708)
	}
}

func (c *conn) bufferBindComplete() {
	__antithesis_instrumentation__.Notify(559709)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgBindComplete)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559710)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559711)
	}
}

func (c *conn) bufferCloseComplete() {
	__antithesis_instrumentation__.Notify(559712)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgCloseComplete)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559713)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559714)
	}
}

func (c *conn) bufferCommandComplete(tag []byte) {
	__antithesis_instrumentation__.Notify(559715)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgCommandComplete)
	c.msgBuilder.write(tag)
	c.msgBuilder.nullTerminate()
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559716)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559717)
	}
}

func (c *conn) bufferPortalSuspended() {
	__antithesis_instrumentation__.Notify(559718)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgPortalSuspended)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559719)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559720)
	}
}

func (c *conn) bufferErr(ctx context.Context, err error) {
	__antithesis_instrumentation__.Notify(559721)
	if err := writeErr(ctx, c.sv,
		err, &c.msgBuilder, &c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559722)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559723)
	}
}

func (c *conn) bufferEmptyQueryResponse() {
	__antithesis_instrumentation__.Notify(559724)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgEmptyQuery)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559725)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559726)
	}
}

func writeErr(
	ctx context.Context, sv *settings.Values, err error, msgBuilder *writeBuffer, w io.Writer,
) error {
	__antithesis_instrumentation__.Notify(559727)

	sqltelemetry.RecordError(ctx, err, sv)
	msgBuilder.initMsg(pgwirebase.ServerMsgErrorResponse)
	return writeErrFields(ctx, sv, err, msgBuilder, w)
}

func writeErrFields(
	ctx context.Context, sv *settings.Values, err error, msgBuilder *writeBuffer, w io.Writer,
) error {
	__antithesis_instrumentation__.Notify(559728)
	pgErr := pgerror.Flatten(err)

	msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldSeverity)
	msgBuilder.writeTerminatedString(pgErr.Severity)

	msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldSQLState)
	msgBuilder.writeTerminatedString(pgErr.Code)

	if pgErr.Detail != "" {
		__antithesis_instrumentation__.Notify(559733)
		msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldDetail)
		msgBuilder.writeTerminatedString(pgErr.Detail)
	} else {
		__antithesis_instrumentation__.Notify(559734)
	}
	__antithesis_instrumentation__.Notify(559729)

	if pgErr.Hint != "" {
		__antithesis_instrumentation__.Notify(559735)
		msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldHint)
		msgBuilder.writeTerminatedString(pgErr.Hint)
	} else {
		__antithesis_instrumentation__.Notify(559736)
	}
	__antithesis_instrumentation__.Notify(559730)

	if pgErr.ConstraintName != "" {
		__antithesis_instrumentation__.Notify(559737)
		msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldConstraintName)
		msgBuilder.writeTerminatedString(pgErr.ConstraintName)
	} else {
		__antithesis_instrumentation__.Notify(559738)
	}
	__antithesis_instrumentation__.Notify(559731)

	if pgErr.Source != nil {
		__antithesis_instrumentation__.Notify(559739)
		errCtx := pgErr.Source
		if errCtx.File != "" {
			__antithesis_instrumentation__.Notify(559742)
			msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldSrcFile)
			msgBuilder.writeTerminatedString(errCtx.File)
		} else {
			__antithesis_instrumentation__.Notify(559743)
		}
		__antithesis_instrumentation__.Notify(559740)

		if errCtx.Line > 0 {
			__antithesis_instrumentation__.Notify(559744)
			msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldSrcLine)
			msgBuilder.writeTerminatedString(strconv.Itoa(int(errCtx.Line)))
		} else {
			__antithesis_instrumentation__.Notify(559745)
		}
		__antithesis_instrumentation__.Notify(559741)

		if errCtx.Function != "" {
			__antithesis_instrumentation__.Notify(559746)
			msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldSrcFunction)
			msgBuilder.writeTerminatedString(errCtx.Function)
		} else {
			__antithesis_instrumentation__.Notify(559747)
		}
	} else {
		__antithesis_instrumentation__.Notify(559748)
	}
	__antithesis_instrumentation__.Notify(559732)

	msgBuilder.putErrFieldMsg(pgwirebase.ServerErrFieldMsgPrimary)
	msgBuilder.writeTerminatedString(pgErr.Message)

	msgBuilder.nullTerminate()
	return msgBuilder.finishMsg(w)
}

func (c *conn) bufferParamDesc(types []oid.Oid) {
	__antithesis_instrumentation__.Notify(559749)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgParameterDescription)
	c.msgBuilder.putInt16(int16(len(types)))
	for _, t := range types {
		__antithesis_instrumentation__.Notify(559751)
		c.msgBuilder.putInt32(int32(t))
	}
	__antithesis_instrumentation__.Notify(559750)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559752)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559753)
	}
}

func (c *conn) bufferNoDataMsg() {
	__antithesis_instrumentation__.Notify(559754)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgNoData)
	if err := c.msgBuilder.finishMsg(&c.writerState.buf); err != nil {
		__antithesis_instrumentation__.Notify(559755)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "unexpected err from buffer"))
	} else {
		__antithesis_instrumentation__.Notify(559756)
	}
}

func (c *conn) writeRowDescription(
	ctx context.Context,
	columns []colinfo.ResultColumn,
	formatCodes []pgwirebase.FormatCode,
	w io.Writer,
) error {
	__antithesis_instrumentation__.Notify(559757)
	c.msgBuilder.initMsg(pgwirebase.ServerMsgRowDescription)
	c.msgBuilder.putInt16(int16(len(columns)))
	for i, column := range columns {
		__antithesis_instrumentation__.Notify(559760)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(559762)
			log.Infof(ctx, "pgwire: writing column %s of type: %s", column.Name, column.Typ)
		} else {
			__antithesis_instrumentation__.Notify(559763)
		}
		__antithesis_instrumentation__.Notify(559761)
		c.msgBuilder.writeTerminatedString(column.Name)
		typ := pgTypeForParserType(column.Typ)
		c.msgBuilder.putInt32(int32(column.TableID))
		c.msgBuilder.putInt16(int16(column.PGAttributeNum))
		c.msgBuilder.putInt32(int32(typ.oid))
		c.msgBuilder.putInt16(int16(typ.size))
		c.msgBuilder.putInt32(column.GetTypeModifier())
		if formatCodes == nil {
			__antithesis_instrumentation__.Notify(559764)
			c.msgBuilder.putInt16(int16(pgwirebase.FormatText))
		} else {
			__antithesis_instrumentation__.Notify(559765)
			c.msgBuilder.putInt16(int16(formatCodes[i]))
		}
	}
	__antithesis_instrumentation__.Notify(559758)
	if err := c.msgBuilder.finishMsg(w); err != nil {
		__antithesis_instrumentation__.Notify(559766)
		c.setErr(err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559767)
	}
	__antithesis_instrumentation__.Notify(559759)
	return nil
}

func (c *conn) Flush(pos sql.CmdPos) error {
	__antithesis_instrumentation__.Notify(559768)

	if err := c.GetErr(); err != nil {
		__antithesis_instrumentation__.Notify(559771)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559772)
	}
	__antithesis_instrumentation__.Notify(559769)

	c.writerState.fi.lastFlushed = pos

	c.writerState.fi.cmdStarts.clear()

	_, err := c.writerState.buf.WriteTo(c.conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(559773)
		c.setErr(err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(559774)
	}
	__antithesis_instrumentation__.Notify(559770)
	return nil
}

func (c *conn) maybeFlush(pos sql.CmdPos) (bool, error) {
	__antithesis_instrumentation__.Notify(559775)
	if int64(c.writerState.buf.Len()) <= c.sessionArgs.ConnResultsBufferSize {
		__antithesis_instrumentation__.Notify(559777)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(559778)
	}
	__antithesis_instrumentation__.Notify(559776)
	return true, c.Flush(pos)
}

func (c *conn) LockCommunication() sql.ClientLock {
	__antithesis_instrumentation__.Notify(559779)
	return (*clientConnLock)(&c.writerState.fi)
}

type clientConnLock flushInfo

var _ sql.ClientLock = &clientConnLock{}

func (cl *clientConnLock) Close() {
	__antithesis_instrumentation__.Notify(559780)

}

func (cl *clientConnLock) ClientPos() sql.CmdPos {
	__antithesis_instrumentation__.Notify(559781)
	return cl.lastFlushed
}

func (cl *clientConnLock) RTrim(ctx context.Context, pos sql.CmdPos) {
	__antithesis_instrumentation__.Notify(559782)
	if pos <= cl.lastFlushed {
		__antithesis_instrumentation__.Notify(559785)
		panic(errors.AssertionFailedf("asked to trim to pos: %d, below the last flush: %d", pos, cl.lastFlushed))
	} else {
		__antithesis_instrumentation__.Notify(559786)
	}
	__antithesis_instrumentation__.Notify(559783)

	truncateIdx := cl.buf.Len()

	for !cl.cmdStarts.empty() {
		__antithesis_instrumentation__.Notify(559787)
		cmdStart := cl.cmdStarts.getLast()
		if cmdStart.pos < pos {
			__antithesis_instrumentation__.Notify(559789)
			break
		} else {
			__antithesis_instrumentation__.Notify(559790)
		}
		__antithesis_instrumentation__.Notify(559788)
		truncateIdx = cmdStart.idx
		cl.cmdStarts.removeLast()
	}
	__antithesis_instrumentation__.Notify(559784)
	cl.buf.Truncate(truncateIdx)
}

func (c *conn) CreateStatementResult(
	stmt tree.Statement,
	descOpt sql.RowDescOpt,
	pos sql.CmdPos,
	formatCodes []pgwirebase.FormatCode,
	conv sessiondatapb.DataConversionConfig,
	location *time.Location,
	limit int,
	portalName string,
	implicitTxn bool,
) sql.CommandResult {
	__antithesis_instrumentation__.Notify(559791)
	return c.newCommandResult(descOpt, pos, stmt, formatCodes, conv, location, limit, portalName, implicitTxn)
}

func (c *conn) CreateSyncResult(pos sql.CmdPos) sql.SyncResult {
	__antithesis_instrumentation__.Notify(559792)
	return c.newMiscResult(pos, readyForQuery)
}

func (c *conn) CreateFlushResult(pos sql.CmdPos) sql.FlushResult {
	__antithesis_instrumentation__.Notify(559793)
	return c.newMiscResult(pos, flush)
}

func (c *conn) CreateDrainResult(pos sql.CmdPos) sql.DrainResult {
	__antithesis_instrumentation__.Notify(559794)
	return c.newMiscResult(pos, noCompletionMsg)
}

func (c *conn) CreateBindResult(pos sql.CmdPos) sql.BindResult {
	__antithesis_instrumentation__.Notify(559795)
	return c.newMiscResult(pos, bindComplete)
}

func (c *conn) CreatePrepareResult(pos sql.CmdPos) sql.ParseResult {
	__antithesis_instrumentation__.Notify(559796)
	return c.newMiscResult(pos, parseComplete)
}

func (c *conn) CreateDescribeResult(pos sql.CmdPos) sql.DescribeResult {
	__antithesis_instrumentation__.Notify(559797)
	return c.newMiscResult(pos, noCompletionMsg)
}

func (c *conn) CreateEmptyQueryResult(pos sql.CmdPos) sql.EmptyQueryResult {
	__antithesis_instrumentation__.Notify(559798)
	return c.newMiscResult(pos, emptyQueryResponse)
}

func (c *conn) CreateDeleteResult(pos sql.CmdPos) sql.DeleteResult {
	__antithesis_instrumentation__.Notify(559799)
	return c.newMiscResult(pos, closeComplete)
}

func (c *conn) CreateErrorResult(pos sql.CmdPos) sql.ErrorResult {
	__antithesis_instrumentation__.Notify(559800)
	res := c.newMiscResult(pos, noCompletionMsg)
	res.errExpected = true
	return res
}

func (c *conn) CreateCopyInResult(pos sql.CmdPos) sql.CopyInResult {
	__antithesis_instrumentation__.Notify(559801)
	return c.newMiscResult(pos, noCompletionMsg)
}

type pgwireReader struct {
	conn *conn
}

var _ pgwirebase.BufferedReader = &pgwireReader{}

func (r *pgwireReader) Read(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(559802)
	n, err := r.conn.rd.Read(p)
	r.conn.metrics.BytesInCount.Inc(int64(n))
	return n, err
}

func (r *pgwireReader) ReadString(delim byte) (string, error) {
	__antithesis_instrumentation__.Notify(559803)
	s, err := r.conn.rd.ReadString(delim)
	r.conn.metrics.BytesInCount.Inc(int64(len(s)))
	return s, err
}

func (r *pgwireReader) ReadByte() (byte, error) {
	__antithesis_instrumentation__.Notify(559804)
	b, err := r.conn.rd.ReadByte()
	if err == nil {
		__antithesis_instrumentation__.Notify(559806)
		r.conn.metrics.BytesInCount.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(559807)
	}
	__antithesis_instrumentation__.Notify(559805)
	return b, err
}

var statusReportParams = []string{
	"server_version",
	"server_encoding",
	"client_encoding",
	"application_name",

	"DateStyle",
	"IntervalStyle",
	"is_superuser",
	"TimeZone",
	"integer_datetimes",
	"standard_conforming_strings",
	"crdb_version",
}

var testingStatusReportParams = map[string]string{
	"client_encoding":             "UTF8",
	"standard_conforming_strings": "on",
}

type readTimeoutConn struct {
	net.Conn

	checkExitConds func() error
}

func NewReadTimeoutConn(c net.Conn, checkExitConds func() error) net.Conn {
	__antithesis_instrumentation__.Notify(559808)
	return &readTimeoutConn{
		Conn:           c,
		checkExitConds: checkExitConds,
	}
}

func (c *readTimeoutConn) Read(b []byte) (int, error) {
	__antithesis_instrumentation__.Notify(559809)

	const readTimeout = 1 * time.Second

	defer func() { __antithesis_instrumentation__.Notify(559811); _ = c.SetReadDeadline(time.Time{}) }()
	__antithesis_instrumentation__.Notify(559810)
	for {
		__antithesis_instrumentation__.Notify(559812)
		if err := c.checkExitConds(); err != nil {
			__antithesis_instrumentation__.Notify(559816)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(559817)
		}
		__antithesis_instrumentation__.Notify(559813)
		if err := c.SetReadDeadline(timeutil.Now().Add(readTimeout)); err != nil {
			__antithesis_instrumentation__.Notify(559818)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(559819)
		}
		__antithesis_instrumentation__.Notify(559814)
		n, err := c.Conn.Read(b)
		if err != nil {
			__antithesis_instrumentation__.Notify(559820)

			if ne := (net.Error)(nil); errors.As(err, &ne) && func() bool {
				__antithesis_instrumentation__.Notify(559821)
				return ne.Timeout() == true
			}() == true {
				__antithesis_instrumentation__.Notify(559822)
				continue
			} else {
				__antithesis_instrumentation__.Notify(559823)
			}
		} else {
			__antithesis_instrumentation__.Notify(559824)
		}
		__antithesis_instrumentation__.Notify(559815)
		return n, err
	}
}
