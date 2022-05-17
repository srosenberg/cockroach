package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/interceptor"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	pgproto3 "github.com/jackc/pgproto3/v2"
)

var defaultTransferTimeout = 15 * time.Second

var transferConnectionConnectorTestHook func(context.Context, string) (net.Conn, error) = nil

type transferContext struct {
	context.Context
	mu struct {
		syncutil.Mutex
		recoverableConn bool
	}
}

func newTransferContext(backgroundCtx context.Context) (*transferContext, context.CancelFunc) {
	__antithesis_instrumentation__.Notify(21120)
	transferCtx, cancel := context.WithTimeout(backgroundCtx, defaultTransferTimeout)
	ctx := &transferContext{
		Context: transferCtx,
	}
	ctx.mu.recoverableConn = true
	return ctx, cancel
}

func (t *transferContext) markRecoverable(r bool) {
	__antithesis_instrumentation__.Notify(21121)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mu.recoverableConn = r
}

func (t *transferContext) isRecoverable() bool {
	__antithesis_instrumentation__.Notify(21122)
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.mu.recoverableConn
}

func (f *forwarder) tryBeginTransfer() (started bool, cleanupFn func()) {
	__antithesis_instrumentation__.Notify(21123)
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.mu.isTransferring {
		__antithesis_instrumentation__.Notify(21126)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(21127)
	}
	__antithesis_instrumentation__.Notify(21124)

	request, response := f.mu.request, f.mu.response
	request.mu.Lock()
	response.mu.Lock()
	defer request.mu.Unlock()
	defer response.mu.Unlock()

	if !isSafeTransferPointLocked(request, response) {
		__antithesis_instrumentation__.Notify(21128)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(21129)
	}
	__antithesis_instrumentation__.Notify(21125)

	f.mu.isTransferring = true
	request.mu.suspendReq = true
	response.mu.suspendReq = true

	return true, func() {
		__antithesis_instrumentation__.Notify(21130)
		f.mu.Lock()
		defer f.mu.Unlock()
		f.mu.isTransferring = false
	}
}

var errTransferCannotStart = errors.New("transfer cannot be started")

func (f *forwarder) TransferConnection() (retErr error) {
	__antithesis_instrumentation__.Notify(21131)

	if f.ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(21139)
		return f.ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(21140)
	}
	__antithesis_instrumentation__.Notify(21132)

	started, cleanupFn := f.tryBeginTransfer()
	if !started {
		__antithesis_instrumentation__.Notify(21141)
		return errTransferCannotStart
	} else {
		__antithesis_instrumentation__.Notify(21142)
	}
	__antithesis_instrumentation__.Notify(21133)
	defer cleanupFn()

	f.metrics.ConnMigrationAttemptedCount.Inc(1)

	ctx, cancel := newTransferContext(f.ctx)
	defer cancel()

	go func() {
		__antithesis_instrumentation__.Notify(21143)
		<-ctx.Done()

		if !ctx.isRecoverable() {
			__antithesis_instrumentation__.Notify(21144)
			f.Close()
		} else {
			__antithesis_instrumentation__.Notify(21145)
		}
	}()
	__antithesis_instrumentation__.Notify(21134)

	tBegin := timeutil.Now()
	logCtx := logtags.WithTags(context.Background(), logtags.FromContext(f.ctx))
	defer func() {
		__antithesis_instrumentation__.Notify(21146)
		latencyDur := timeutil.Since(tBegin)
		f.metrics.ConnMigrationAttemptedLatency.RecordValue(latencyDur.Nanoseconds())

		if !ctx.isRecoverable() {
			__antithesis_instrumentation__.Notify(21147)
			log.Infof(logCtx, "transfer failed: connection closed, latency=%v, err=%v", latencyDur, retErr)
			f.metrics.ConnMigrationErrorFatalCount.Inc(1)
			f.Close()
		} else {
			__antithesis_instrumentation__.Notify(21148)

			if retErr == nil {
				__antithesis_instrumentation__.Notify(21150)
				log.Infof(logCtx, "transfer successful, latency=%v", latencyDur)
				f.metrics.ConnMigrationSuccessCount.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(21151)
				log.Infof(logCtx, "transfer failed: connection recovered, latency=%v, err=%v", latencyDur, retErr)
				f.metrics.ConnMigrationErrorRecoverableCount.Inc(1)
			}
			__antithesis_instrumentation__.Notify(21149)
			if err := f.resumeProcessors(); err != nil {
				__antithesis_instrumentation__.Notify(21152)
				log.Infof(logCtx, "unable to resume processors: %v", err)
				f.Close()
			} else {
				__antithesis_instrumentation__.Notify(21153)
			}
		}
	}()
	__antithesis_instrumentation__.Notify(21135)

	request, response := f.getProcessors()
	if err := request.suspend(ctx); err != nil {
		__antithesis_instrumentation__.Notify(21154)
		return errors.Wrap(err, "suspending request processor")
	} else {
		__antithesis_instrumentation__.Notify(21155)
	}
	__antithesis_instrumentation__.Notify(21136)
	if err := response.suspend(ctx); err != nil {
		__antithesis_instrumentation__.Notify(21156)
		return errors.Wrap(err, "suspending response processor")
	} else {
		__antithesis_instrumentation__.Notify(21157)
	}
	__antithesis_instrumentation__.Notify(21137)

	clientConn, serverConn := f.getConns()
	newServerConn, err := transferConnection(ctx, f.connector, f.metrics, clientConn, serverConn)
	if err != nil {
		__antithesis_instrumentation__.Notify(21158)
		return errors.Wrap(err, "transferring connection")
	} else {
		__antithesis_instrumentation__.Notify(21159)
	}
	__antithesis_instrumentation__.Notify(21138)

	f.replaceServerConn(newServerConn)
	return nil
}

func transferConnection(
	ctx *transferContext,
	connector *connector,
	metrics *metrics,
	clientConn, serverConn *interceptor.PGConn,
) (_ *interceptor.PGConn, retErr error) {
	__antithesis_instrumentation__.Notify(21160)
	ctx.markRecoverable(true)

	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(21169)
		return nil, ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(21170)
	}
	__antithesis_instrumentation__.Notify(21161)

	transferKey := uuid.MakeV4().String()

	ctx.markRecoverable(false)
	if err := runShowTransferState(serverConn, transferKey); err != nil {
		__antithesis_instrumentation__.Notify(21171)
		return nil, errors.Wrap(err, "sending transfer request")
	} else {
		__antithesis_instrumentation__.Notify(21172)
	}
	__antithesis_instrumentation__.Notify(21162)

	transferErr, state, revivalToken, err := waitForShowTransferState(
		ctx, serverConn.ToFrontendConn(), clientConn, transferKey, metrics)
	if err != nil {
		__antithesis_instrumentation__.Notify(21173)
		return nil, errors.Wrap(err, "waiting for transfer state")
	} else {
		__antithesis_instrumentation__.Notify(21174)
	}
	__antithesis_instrumentation__.Notify(21163)

	ctx.markRecoverable(true)

	if transferErr != "" {
		__antithesis_instrumentation__.Notify(21175)
		return nil, errors.Newf("%s", transferErr)
	} else {
		__antithesis_instrumentation__.Notify(21176)
	}
	__antithesis_instrumentation__.Notify(21164)

	connectFn := connector.OpenTenantConnWithToken
	if transferConnectionConnectorTestHook != nil {
		__antithesis_instrumentation__.Notify(21177)
		connectFn = transferConnectionConnectorTestHook
	} else {
		__antithesis_instrumentation__.Notify(21178)
	}
	__antithesis_instrumentation__.Notify(21165)
	netConn, err := connectFn(ctx, revivalToken)
	if err != nil {
		__antithesis_instrumentation__.Notify(21179)
		return nil, errors.Wrap(err, "opening connection")
	} else {
		__antithesis_instrumentation__.Notify(21180)
	}
	__antithesis_instrumentation__.Notify(21166)
	defer func() {
		__antithesis_instrumentation__.Notify(21181)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(21182)
			netConn.Close()
		} else {
			__antithesis_instrumentation__.Notify(21183)
		}
	}()
	__antithesis_instrumentation__.Notify(21167)
	newServerConn := interceptor.NewPGConn(netConn)

	if err := runAndWaitForDeserializeSession(
		ctx, newServerConn.ToFrontendConn(), state,
	); err != nil {
		__antithesis_instrumentation__.Notify(21184)
		return nil, errors.Wrap(err, "deserializing session")
	} else {
		__antithesis_instrumentation__.Notify(21185)
	}
	__antithesis_instrumentation__.Notify(21168)

	return newServerConn, nil
}

var isSafeTransferPointLocked = func(request *processor, response *processor) bool {
	__antithesis_instrumentation__.Notify(21186)

	if request.mu.lastMessageTransferredAt > response.mu.lastMessageTransferredAt {
		__antithesis_instrumentation__.Notify(21188)
		return false
	} else {
		__antithesis_instrumentation__.Notify(21189)
	}
	__antithesis_instrumentation__.Notify(21187)

	switch pgwirebase.ClientMessageType(request.mu.lastMessageType) {
	case pgwirebase.ClientMessageType(0),
		pgwirebase.ClientMsgSync,
		pgwirebase.ClientMsgSimpleQuery,
		pgwirebase.ClientMsgCopyDone,
		pgwirebase.ClientMsgCopyFail:
		__antithesis_instrumentation__.Notify(21190)

		serverMsg := pgwirebase.ServerMessageType(response.mu.lastMessageType)
		return serverMsg == pgwirebase.ServerMsgReady || func() bool {
			__antithesis_instrumentation__.Notify(21192)
			return serverMsg == pgwirebase.ServerMessageType(0) == true
		}() == true
	default:
		__antithesis_instrumentation__.Notify(21191)
		return false
	}
}

var runShowTransferState = func(w io.Writer, transferKey string) error {
	__antithesis_instrumentation__.Notify(21193)
	return writeQuery(w, "SHOW TRANSFER STATE WITH '%s'", transferKey)
}

var waitForShowTransferState = func(
	ctx context.Context,
	serverConn *interceptor.FrontendConn,
	clientConn io.Writer,
	transferKey string,
	metrics *metrics,
) (transferErr string, state string, revivalToken string, retErr error) {
	__antithesis_instrumentation__.Notify(21194)

	if err := waitForSmallRowDescription(
		ctx,
		serverConn,
		clientConn,
		func(msg *pgproto3.RowDescription) bool {
			__antithesis_instrumentation__.Notify(21199)

			if len(msg.Fields) != 4 {
				__antithesis_instrumentation__.Notify(21202)
				return false
			} else {
				__antithesis_instrumentation__.Notify(21203)
			}
			__antithesis_instrumentation__.Notify(21200)

			var transferStateCols = []string{
				"error",
				"session_state_base64",
				"session_revival_token_base64",
				"transfer_key",
			}
			for i, col := range transferStateCols {
				__antithesis_instrumentation__.Notify(21204)
				if string(msg.Fields[i].Name) != col {
					__antithesis_instrumentation__.Notify(21205)
					return false
				} else {
					__antithesis_instrumentation__.Notify(21206)
				}
			}
			__antithesis_instrumentation__.Notify(21201)
			return true
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(21207)
		return "", "", "", errors.Wrap(err, "waiting for RowDescription")
	} else {
		__antithesis_instrumentation__.Notify(21208)
	}
	__antithesis_instrumentation__.Notify(21195)

	if err := expectDataRow(ctx, serverConn, func(msg *pgproto3.DataRow, size int) bool {
		__antithesis_instrumentation__.Notify(21209)

		if len(msg.Values) != 4 {
			__antithesis_instrumentation__.Notify(21213)
			return false
		} else {
			__antithesis_instrumentation__.Notify(21214)
		}
		__antithesis_instrumentation__.Notify(21210)

		if string(msg.Values[3]) != transferKey {
			__antithesis_instrumentation__.Notify(21215)
			return false
		} else {
			__antithesis_instrumentation__.Notify(21216)
		}
		__antithesis_instrumentation__.Notify(21211)

		transferErr, state, revivalToken = string(msg.Values[0]), string(msg.Values[1]), string(msg.Values[2])

		if metrics != nil {
			__antithesis_instrumentation__.Notify(21217)
			metrics.ConnMigrationTransferResponseMessageSize.RecordValue(int64(size))
		} else {
			__antithesis_instrumentation__.Notify(21218)
		}
		__antithesis_instrumentation__.Notify(21212)
		return true
	}); err != nil {
		__antithesis_instrumentation__.Notify(21219)
		return "", "", "", errors.Wrap(err, "expecting DataRow")
	} else {
		__antithesis_instrumentation__.Notify(21220)
	}
	__antithesis_instrumentation__.Notify(21196)

	if err := expectCommandComplete(ctx, serverConn, "SHOW TRANSFER STATE 1"); err != nil {
		__antithesis_instrumentation__.Notify(21221)
		return "", "", "", errors.Wrap(err, "expecting CommandComplete")
	} else {
		__antithesis_instrumentation__.Notify(21222)
	}
	__antithesis_instrumentation__.Notify(21197)

	if err := expectReadyForQuery(ctx, serverConn); err != nil {
		__antithesis_instrumentation__.Notify(21223)
		return "", "", "", errors.Wrap(err, "expecting ReadyForQuery")
	} else {
		__antithesis_instrumentation__.Notify(21224)
	}
	__antithesis_instrumentation__.Notify(21198)

	return transferErr, state, revivalToken, nil
}

var runAndWaitForDeserializeSession = func(
	ctx context.Context, serverConn *interceptor.FrontendConn, state string,
) error {
	__antithesis_instrumentation__.Notify(21225)

	if err := writeQuery(serverConn,
		"SELECT crdb_internal.deserialize_session(decode('%s', 'base64'))", state); err != nil {
		__antithesis_instrumentation__.Notify(21231)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21232)
	}
	__antithesis_instrumentation__.Notify(21226)

	if err := waitForSmallRowDescription(
		ctx,
		serverConn,
		&errWriter{},
		func(msg *pgproto3.RowDescription) bool {
			__antithesis_instrumentation__.Notify(21233)
			return len(msg.Fields) == 1 && func() bool {
				__antithesis_instrumentation__.Notify(21234)
				return string(msg.Fields[0].Name) == "crdb_internal.deserialize_session" == true
			}() == true
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(21235)
		return errors.Wrap(err, "expecting RowDescription")
	} else {
		__antithesis_instrumentation__.Notify(21236)
	}
	__antithesis_instrumentation__.Notify(21227)

	if err := expectDataRow(ctx, serverConn, func(msg *pgproto3.DataRow, _ int) bool {
		__antithesis_instrumentation__.Notify(21237)
		return len(msg.Values) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(21238)
			return string(msg.Values[0]) == "t" == true
		}() == true
	}); err != nil {
		__antithesis_instrumentation__.Notify(21239)
		return errors.Wrap(err, "expecting DataRow")
	} else {
		__antithesis_instrumentation__.Notify(21240)
	}
	__antithesis_instrumentation__.Notify(21228)

	if err := expectCommandComplete(ctx, serverConn, "SELECT 1"); err != nil {
		__antithesis_instrumentation__.Notify(21241)
		return errors.Wrap(err, "expecting CommandComplete")
	} else {
		__antithesis_instrumentation__.Notify(21242)
	}
	__antithesis_instrumentation__.Notify(21229)

	if err := expectReadyForQuery(ctx, serverConn); err != nil {
		__antithesis_instrumentation__.Notify(21243)
		return errors.Wrap(err, "expecting ReadyForQuery")
	} else {
		__antithesis_instrumentation__.Notify(21244)
	}
	__antithesis_instrumentation__.Notify(21230)

	return nil
}

func writeQuery(w io.Writer, format string, a ...interface{}) error {
	__antithesis_instrumentation__.Notify(21245)
	query := &pgproto3.Query{String: fmt.Sprintf(format, a...)}
	_, err := w.Write(query.Encode(nil))
	return err
}

func waitForSmallRowDescription(
	ctx context.Context,
	serverConn *interceptor.FrontendConn,
	clientConn io.Writer,
	matchFn func(*pgproto3.RowDescription) bool,
) error {
	__antithesis_instrumentation__.Notify(21246)

	for {
		__antithesis_instrumentation__.Notify(21247)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(21255)
			return ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(21256)
		}
		__antithesis_instrumentation__.Notify(21248)

		typ, size, err := serverConn.PeekMsg()
		if err != nil {
			__antithesis_instrumentation__.Notify(21257)
			return errors.Wrap(err, "peeking message")
		} else {
			__antithesis_instrumentation__.Notify(21258)
		}
		__antithesis_instrumentation__.Notify(21249)

		if typ == pgwirebase.ServerMsgErrorResponse {
			__antithesis_instrumentation__.Notify(21259)

			msg, err := serverConn.ReadMsg()
			if err != nil {
				__antithesis_instrumentation__.Notify(21261)
				return errors.Wrap(err, "ambiguous ErrorResponse")
			} else {
				__antithesis_instrumentation__.Notify(21262)
			}
			__antithesis_instrumentation__.Notify(21260)
			return errors.Newf("ambiguous ErrorResponse: %v", jsonOrRaw(msg))
		} else {
			__antithesis_instrumentation__.Notify(21263)
		}
		__antithesis_instrumentation__.Notify(21250)

		const maxSmallMsgSize = 1 << 12
		if typ != pgwirebase.ServerMsgRowDescription || func() bool {
			__antithesis_instrumentation__.Notify(21264)
			return size > maxSmallMsgSize == true
		}() == true {
			__antithesis_instrumentation__.Notify(21265)
			if _, err := serverConn.ForwardMsg(clientConn); err != nil {
				__antithesis_instrumentation__.Notify(21267)
				return errors.Wrap(err, "forwarding message")
			} else {
				__antithesis_instrumentation__.Notify(21268)
			}
			__antithesis_instrumentation__.Notify(21266)
			continue
		} else {
			__antithesis_instrumentation__.Notify(21269)
		}
		__antithesis_instrumentation__.Notify(21251)

		msg, err := serverConn.ReadMsg()
		if err != nil {
			__antithesis_instrumentation__.Notify(21270)
			return errors.Wrap(err, "reading RowDescription")
		} else {
			__antithesis_instrumentation__.Notify(21271)
		}
		__antithesis_instrumentation__.Notify(21252)

		pgMsg, ok := msg.(*pgproto3.RowDescription)
		if !ok {
			__antithesis_instrumentation__.Notify(21272)

			return errors.Newf("unexpected message: %v", jsonOrRaw(msg))
		} else {
			__antithesis_instrumentation__.Notify(21273)
		}
		__antithesis_instrumentation__.Notify(21253)

		if matchFn(pgMsg) {
			__antithesis_instrumentation__.Notify(21274)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(21275)
		}
		__antithesis_instrumentation__.Notify(21254)

		if _, err := clientConn.Write(msg.Encode(nil)); err != nil {
			__antithesis_instrumentation__.Notify(21276)
			return errors.Wrap(err, "writing message")
		} else {
			__antithesis_instrumentation__.Notify(21277)
		}
	}
}

func expectDataRow(
	ctx context.Context,
	serverConn *interceptor.FrontendConn,
	validateFn func(*pgproto3.DataRow, int) bool,
) error {
	__antithesis_instrumentation__.Notify(21278)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(21284)
		return ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(21285)
	}
	__antithesis_instrumentation__.Notify(21279)
	_, size, err := serverConn.PeekMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21286)
		return errors.Wrap(err, "peeking message")
	} else {
		__antithesis_instrumentation__.Notify(21287)
	}
	__antithesis_instrumentation__.Notify(21280)
	msg, err := serverConn.ReadMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21288)
		return errors.Wrap(err, "reading message")
	} else {
		__antithesis_instrumentation__.Notify(21289)
	}
	__antithesis_instrumentation__.Notify(21281)
	pgMsg, ok := msg.(*pgproto3.DataRow)
	if !ok {
		__antithesis_instrumentation__.Notify(21290)
		return errors.Newf("unexpected message: %v", jsonOrRaw(msg))
	} else {
		__antithesis_instrumentation__.Notify(21291)
	}
	__antithesis_instrumentation__.Notify(21282)
	if !validateFn(pgMsg, size) {
		__antithesis_instrumentation__.Notify(21292)
		return errors.Newf("validation failed for message: %v", jsonOrRaw(msg))
	} else {
		__antithesis_instrumentation__.Notify(21293)
	}
	__antithesis_instrumentation__.Notify(21283)
	return nil
}

func expectCommandComplete(
	ctx context.Context, serverConn *interceptor.FrontendConn, tag string,
) error {
	__antithesis_instrumentation__.Notify(21294)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(21298)
		return ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(21299)
	}
	__antithesis_instrumentation__.Notify(21295)
	msg, err := serverConn.ReadMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21300)
		return errors.Wrap(err, "reading message")
	} else {
		__antithesis_instrumentation__.Notify(21301)
	}
	__antithesis_instrumentation__.Notify(21296)
	pgMsg, ok := msg.(*pgproto3.CommandComplete)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(21302)
		return string(pgMsg.CommandTag) != tag == true
	}() == true {
		__antithesis_instrumentation__.Notify(21303)
		return errors.Newf("unexpected message: %v", jsonOrRaw(msg))
	} else {
		__antithesis_instrumentation__.Notify(21304)
	}
	__antithesis_instrumentation__.Notify(21297)
	return nil
}

func expectReadyForQuery(ctx context.Context, serverConn *interceptor.FrontendConn) error {
	__antithesis_instrumentation__.Notify(21305)
	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(21309)
		return ctx.Err()
	} else {
		__antithesis_instrumentation__.Notify(21310)
	}
	__antithesis_instrumentation__.Notify(21306)
	msg, err := serverConn.ReadMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21311)
		return errors.Wrap(err, "reading message")
	} else {
		__antithesis_instrumentation__.Notify(21312)
	}
	__antithesis_instrumentation__.Notify(21307)
	_, ok := msg.(*pgproto3.ReadyForQuery)
	if !ok {
		__antithesis_instrumentation__.Notify(21313)
		return errors.Newf("unexpected message: %v", jsonOrRaw(msg))
	} else {
		__antithesis_instrumentation__.Notify(21314)
	}
	__antithesis_instrumentation__.Notify(21308)
	return nil
}

func jsonOrRaw(msg pgproto3.BackendMessage) string {
	__antithesis_instrumentation__.Notify(21315)
	m, err := json.Marshal(msg)
	if err != nil {
		__antithesis_instrumentation__.Notify(21317)
		return fmt.Sprintf("%v", msg)
	} else {
		__antithesis_instrumentation__.Notify(21318)
	}
	__antithesis_instrumentation__.Notify(21316)
	return string(m)
}

var _ io.Writer = &errWriter{}

type errWriter struct{}

func (w *errWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(21319)
	return 0, errors.AssertionFailedf("unexpected Write call")
}
