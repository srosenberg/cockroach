package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"net"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/interceptor"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/throttler"
	pgproto3 "github.com/jackc/pgproto3/v2"
)

var authenticate = func(clientConn, crdbConn net.Conn, throttleHook func(throttler.AttemptStatus) error) error {
	__antithesis_instrumentation__.Notify(20882)
	fe := pgproto3.NewBackend(pgproto3.NewChunkReader(clientConn), clientConn)
	be := pgproto3.NewFrontend(pgproto3.NewChunkReader(crdbConn), crdbConn)

	feSend := func(msg pgproto3.BackendMessage) error {
		__antithesis_instrumentation__.Notify(20885)
		err := fe.Send(msg)
		if err != nil {
			__antithesis_instrumentation__.Notify(20887)
			return newErrorf(codeClientWriteFailed, "unable to send message %v to client: %v", msg, err)
		} else {
			__antithesis_instrumentation__.Notify(20888)
		}
		__antithesis_instrumentation__.Notify(20886)
		return nil
	}
	__antithesis_instrumentation__.Notify(20883)

	var i int
	for ; i < 20; i++ {
		__antithesis_instrumentation__.Notify(20889)

		backendMsg, err := be.Receive()
		if err != nil {
			__antithesis_instrumentation__.Notify(20891)
			return newErrorf(codeBackendReadFailed, "unable to receive message from backend: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(20892)
		}
		__antithesis_instrumentation__.Notify(20890)

		switch tp := backendMsg.(type) {

		case
			*pgproto3.AuthenticationCleartextPassword,
			*pgproto3.AuthenticationMD5Password,
			*pgproto3.AuthenticationSASLContinue,
			*pgproto3.AuthenticationSASLFinal,
			*pgproto3.AuthenticationSASL:
			__antithesis_instrumentation__.Notify(20893)
			if err = feSend(backendMsg); err != nil {
				__antithesis_instrumentation__.Notify(20906)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20907)
			}
			__antithesis_instrumentation__.Notify(20894)
			switch backendMsg.(type) {
			case *pgproto3.AuthenticationCleartextPassword:
				__antithesis_instrumentation__.Notify(20908)
				_ = fe.SetAuthType(pgproto3.AuthTypeCleartextPassword)
			case *pgproto3.AuthenticationMD5Password:
				__antithesis_instrumentation__.Notify(20909)
				_ = fe.SetAuthType(pgproto3.AuthTypeMD5Password)
			case *pgproto3.AuthenticationSASLContinue:
				__antithesis_instrumentation__.Notify(20910)
				_ = fe.SetAuthType(pgproto3.AuthTypeSASLContinue)
			case *pgproto3.AuthenticationSASL:
				__antithesis_instrumentation__.Notify(20911)
				_ = fe.SetAuthType(pgproto3.AuthTypeSASL)
			case *pgproto3.AuthenticationSASLFinal:
				__antithesis_instrumentation__.Notify(20912)

				continue
			}
			__antithesis_instrumentation__.Notify(20895)
			fntMsg, err := fe.Receive()
			if err != nil {
				__antithesis_instrumentation__.Notify(20913)
				return newErrorf(codeClientReadFailed, "unable to receive message from client: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(20914)
			}
			__antithesis_instrumentation__.Notify(20896)
			err = be.Send(fntMsg)
			if err != nil {
				__antithesis_instrumentation__.Notify(20915)
				return newErrorf(
					codeBackendWriteFailed, "unable to send message %v to backend: %v", fntMsg, err,
				)
			} else {
				__antithesis_instrumentation__.Notify(20916)
			}

		case *pgproto3.AuthenticationOk:
			__antithesis_instrumentation__.Notify(20897)
			throttleError := throttleHook(throttler.AttemptOK)
			if throttleError != nil {
				__antithesis_instrumentation__.Notify(20917)
				if err = feSend(toPgError(throttleError)); err != nil {
					__antithesis_instrumentation__.Notify(20919)
					return err
				} else {
					__antithesis_instrumentation__.Notify(20920)
				}
				__antithesis_instrumentation__.Notify(20918)
				return throttleError
			} else {
				__antithesis_instrumentation__.Notify(20921)
			}
			__antithesis_instrumentation__.Notify(20898)
			if err = feSend(backendMsg); err != nil {
				__antithesis_instrumentation__.Notify(20922)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20923)
			}

		case *pgproto3.ErrorResponse:
			__antithesis_instrumentation__.Notify(20899)
			throttleError := throttleHook(throttler.AttemptInvalidCredentials)
			if throttleError != nil {
				__antithesis_instrumentation__.Notify(20924)
				if err = feSend(toPgError(throttleError)); err != nil {
					__antithesis_instrumentation__.Notify(20926)
					return err
				} else {
					__antithesis_instrumentation__.Notify(20927)
				}
				__antithesis_instrumentation__.Notify(20925)
				return throttleError
			} else {
				__antithesis_instrumentation__.Notify(20928)
			}
			__antithesis_instrumentation__.Notify(20900)
			if err = feSend(backendMsg); err != nil {
				__antithesis_instrumentation__.Notify(20929)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20930)
			}
			__antithesis_instrumentation__.Notify(20901)
			return newErrorf(codeAuthFailed, "authentication failed: %s", tp.Message)

		case *pgproto3.ParameterStatus, *pgproto3.BackendKeyData:
			__antithesis_instrumentation__.Notify(20902)
			if err = feSend(backendMsg); err != nil {
				__antithesis_instrumentation__.Notify(20931)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20932)
			}

		case *pgproto3.ReadyForQuery:
			__antithesis_instrumentation__.Notify(20903)
			if err = feSend(backendMsg); err != nil {
				__antithesis_instrumentation__.Notify(20933)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20934)
			}
			__antithesis_instrumentation__.Notify(20904)
			return nil

		default:
			__antithesis_instrumentation__.Notify(20905)
			return newErrorf(codeBackendDisconnected, "received unexpected backend message type: %v", tp)
		}
	}
	__antithesis_instrumentation__.Notify(20884)
	return newErrorf(codeBackendDisconnected, "authentication took more than %d iterations", i)
}

var readTokenAuthResult = func(conn net.Conn) error {
	__antithesis_instrumentation__.Notify(20935)

	serverConn := interceptor.NewFrontendConn(conn)

	var i int
	for ; i < 20; i++ {
		__antithesis_instrumentation__.Notify(20937)
		backendMsg, err := serverConn.ReadMsg()
		if err != nil {
			__antithesis_instrumentation__.Notify(20939)
			return newErrorf(codeBackendReadFailed, "unable to receive message from backend: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(20940)
		}
		__antithesis_instrumentation__.Notify(20938)

		switch tp := backendMsg.(type) {
		case *pgproto3.AuthenticationOk, *pgproto3.ParameterStatus, *pgproto3.BackendKeyData:
			__antithesis_instrumentation__.Notify(20941)

		case *pgproto3.ErrorResponse:
			__antithesis_instrumentation__.Notify(20942)
			return newErrorf(codeAuthFailed, "authentication failed: %s", tp.Message)

		case *pgproto3.ReadyForQuery:
			__antithesis_instrumentation__.Notify(20943)
			return nil

		default:
			__antithesis_instrumentation__.Notify(20944)
			return newErrorf(codeBackendDisconnected, "received unexpected backend message type: %v", tp)
		}
	}
	__antithesis_instrumentation__.Notify(20936)

	return newErrorf(codeBackendDisconnected, "authentication took more than %d iterations", i)
}
