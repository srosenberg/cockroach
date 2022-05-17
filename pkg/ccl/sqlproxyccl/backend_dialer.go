package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"time"

	"github.com/jackc/pgproto3/v2"
)

var BackendDial = func(
	msg *pgproto3.StartupMessage, serverAddress string, tlsConfig *tls.Config,
) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(20945)

	conn, err := net.DialTimeout("tcp", serverAddress, time.Second*5)
	if err != nil {
		__antithesis_instrumentation__.Notify(20949)
		return nil, newErrorf(
			codeBackendDown, "unable to reach backend SQL server: %v", err,
		)
	} else {
		__antithesis_instrumentation__.Notify(20950)
	}
	__antithesis_instrumentation__.Notify(20946)
	conn, err = sslOverlay(conn, tlsConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(20951)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20952)
	}
	__antithesis_instrumentation__.Notify(20947)
	err = relayStartupMsg(conn, msg)
	if err != nil {
		__antithesis_instrumentation__.Notify(20953)
		return nil, newErrorf(
			codeBackendDown, "relaying StartupMessage to target server %v: %v",
			serverAddress, err)
	} else {
		__antithesis_instrumentation__.Notify(20954)
	}
	__antithesis_instrumentation__.Notify(20948)
	return conn, nil
}

func sslOverlay(conn net.Conn, tlsConfig *tls.Config) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(20955)
	if tlsConfig == nil {
		__antithesis_instrumentation__.Notify(20960)
		return conn, nil
	} else {
		__antithesis_instrumentation__.Notify(20961)
	}
	__antithesis_instrumentation__.Notify(20956)

	var err error

	if err := binary.Write(conn, binary.BigEndian, pgSSLRequest); err != nil {
		__antithesis_instrumentation__.Notify(20962)
		return nil, newErrorf(
			codeBackendDown, "sending SSLRequest to target server: %v", err,
		)
	} else {
		__antithesis_instrumentation__.Notify(20963)
	}
	__antithesis_instrumentation__.Notify(20957)

	response := make([]byte, 1)
	if _, err = io.ReadFull(conn, response); err != nil {
		__antithesis_instrumentation__.Notify(20964)
		return nil,
			newErrorf(codeBackendDown, "reading response to SSLRequest")
	} else {
		__antithesis_instrumentation__.Notify(20965)
	}
	__antithesis_instrumentation__.Notify(20958)

	if response[0] != pgAcceptSSLRequest {
		__antithesis_instrumentation__.Notify(20966)
		return nil, newErrorf(
			codeBackendRefusedTLS, "target server refused TLS connection",
		)
	} else {
		__antithesis_instrumentation__.Notify(20967)
	}
	__antithesis_instrumentation__.Notify(20959)

	outCfg := tlsConfig.Clone()
	return tls.Client(conn, outCfg), nil
}

func relayStartupMsg(conn net.Conn, msg *pgproto3.StartupMessage) (err error) {
	__antithesis_instrumentation__.Notify(20968)
	_, err = conn.Write(msg.Encode(nil))
	return
}
