package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgproto3/v2"
)

const pgAcceptSSLRequest = 'S'

var pgSSLRequest = []int32{8, 80877103}

func updateMetricsAndSendErrToClient(err error, conn net.Conn, metrics *metrics) {
	__antithesis_instrumentation__.Notify(21789)
	metrics.updateForError(err)
	SendErrToClient(conn, err)
}

func toPgError(err error) *pgproto3.ErrorResponse {
	__antithesis_instrumentation__.Notify(21790)
	codeErr := (*codeError)(nil)
	if errors.As(err, &codeErr) {
		__antithesis_instrumentation__.Notify(21792)
		var msg string
		switch codeErr.code {

		case codeExpiredClientConnection,
			codeBackendDown,
			codeParamsRoutingFailed,
			codeClientDisconnected,
			codeBackendDisconnected,
			codeAuthFailed,
			codeProxyRefusedConnection,
			codeIdleDisconnect,
			codeUnavailable:
			__antithesis_instrumentation__.Notify(21795)
			msg = codeErr.Error()

		case codeUnexpectedInsecureStartupMessage:
			__antithesis_instrumentation__.Notify(21796)
			msg = "server requires encryption"
		default:
			__antithesis_instrumentation__.Notify(21797)
		}
		__antithesis_instrumentation__.Notify(21793)

		var pgCode string
		if codeErr.code == codeIdleDisconnect {
			__antithesis_instrumentation__.Notify(21798)
			pgCode = pgcode.AdminShutdown.String()
		} else {
			__antithesis_instrumentation__.Notify(21799)
			pgCode = pgcode.SQLserverRejectedEstablishmentOfSQLconnection.String()
		}
		__antithesis_instrumentation__.Notify(21794)

		return &pgproto3.ErrorResponse{
			Severity: "FATAL",
			Code:     pgCode,
			Message:  msg,
			Hint:     errors.FlattenHints(err),
		}
	} else {
		__antithesis_instrumentation__.Notify(21800)
	}
	__antithesis_instrumentation__.Notify(21791)

	return &pgproto3.ErrorResponse{
		Severity: "FATAL",
		Code:     pgcode.SQLserverRejectedEstablishmentOfSQLconnection.String(),
		Message:  "internal server error",
	}
}

var SendErrToClient = func(conn net.Conn, err error) {
	__antithesis_instrumentation__.Notify(21801)
	if err == nil || func() bool {
		__antithesis_instrumentation__.Notify(21803)
		return conn == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(21804)
		return
	} else {
		__antithesis_instrumentation__.Notify(21805)
	}
	__antithesis_instrumentation__.Notify(21802)
	_, _ = conn.Write(toPgError(err).Encode(nil))
}
