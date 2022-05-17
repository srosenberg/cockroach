package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/tls"
	"net"

	"github.com/jackc/pgproto3/v2"
)

type FrontendAdmitInfo struct {
	conn          net.Conn
	msg           *pgproto3.StartupMessage
	err           error
	sniServerName string
}

var FrontendAdmit = func(
	conn net.Conn, incomingTLSConfig *tls.Config,
) *FrontendAdmitInfo {
	__antithesis_instrumentation__.Notify(21614)

	m, err := pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn).ReceiveStartupMessage()
	if err != nil {
		__antithesis_instrumentation__.Notify(21619)
		return &FrontendAdmitInfo{
			conn: conn, err: newErrorf(codeClientReadFailed, "while receiving startup message"),
		}
	} else {
		__antithesis_instrumentation__.Notify(21620)
	}
	__antithesis_instrumentation__.Notify(21615)

	if _, ok := m.(*pgproto3.CancelRequest); ok {
		__antithesis_instrumentation__.Notify(21621)
		return &FrontendAdmitInfo{conn: conn}
	} else {
		__antithesis_instrumentation__.Notify(21622)
	}
	__antithesis_instrumentation__.Notify(21616)

	var sniServerName string

	if incomingTLSConfig != nil {
		__antithesis_instrumentation__.Notify(21623)
		if _, ok := m.(*pgproto3.SSLRequest); !ok {
			__antithesis_instrumentation__.Notify(21627)
			code := codeUnexpectedInsecureStartupMessage
			return &FrontendAdmitInfo{conn: conn, err: newErrorf(code, "unsupported startup message: %T", m)}
		} else {
			__antithesis_instrumentation__.Notify(21628)
		}
		__antithesis_instrumentation__.Notify(21624)

		_, err = conn.Write([]byte{pgAcceptSSLRequest})
		if err != nil {
			__antithesis_instrumentation__.Notify(21629)
			return &FrontendAdmitInfo{conn: conn, err: newErrorf(codeClientWriteFailed, "acking SSLRequest: %v", err)}
		} else {
			__antithesis_instrumentation__.Notify(21630)
		}
		__antithesis_instrumentation__.Notify(21625)

		cfg := incomingTLSConfig.Clone()

		cfg.GetConfigForClient = func(h *tls.ClientHelloInfo) (*tls.Config, error) {
			__antithesis_instrumentation__.Notify(21631)
			sniServerName = h.ServerName
			return nil, nil
		}
		__antithesis_instrumentation__.Notify(21626)
		conn = tls.Server(conn, cfg)

		m, err = pgproto3.NewBackend(pgproto3.NewChunkReader(conn), conn).ReceiveStartupMessage()
		if err != nil {
			__antithesis_instrumentation__.Notify(21632)
			return &FrontendAdmitInfo{
				conn: conn,
				err:  newErrorf(codeClientReadFailed, "receiving post-TLS startup message: %v", err),
			}
		} else {
			__antithesis_instrumentation__.Notify(21633)
		}
	} else {
		__antithesis_instrumentation__.Notify(21634)
	}
	__antithesis_instrumentation__.Notify(21617)

	if startup, ok := m.(*pgproto3.StartupMessage); ok {
		__antithesis_instrumentation__.Notify(21635)

		startup.Parameters[remoteAddrStartupParam] = conn.RemoteAddr().String()

		if _, ok := startup.Parameters[sessionRevivalTokenStartupParam]; ok {
			__antithesis_instrumentation__.Notify(21637)
			return &FrontendAdmitInfo{
				conn: conn,
				err: newErrorf(
					codeUnexpectedStartupMessage,
					"parameter %s is not allowed",
					sessionRevivalTokenStartupParam,
				),
			}
		} else {
			__antithesis_instrumentation__.Notify(21638)
		}
		__antithesis_instrumentation__.Notify(21636)
		return &FrontendAdmitInfo{conn: conn, msg: startup, sniServerName: sniServerName}
	} else {
		__antithesis_instrumentation__.Notify(21639)
	}
	__antithesis_instrumentation__.Notify(21618)

	code := codeUnexpectedStartupMessage
	return &FrontendAdmitInfo{
		conn: conn,
		err:  newErrorf(code, "unsupported post-TLS startup message: %T", m),
	}
}
