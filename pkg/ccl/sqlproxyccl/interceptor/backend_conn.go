package interceptor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"net"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/jackc/pgproto3/v2"
)

type BackendConn struct {
	net.Conn
	interceptor *pgInterceptor
}

func NewBackendConn(conn net.Conn) *BackendConn {
	__antithesis_instrumentation__.Notify(21684)
	return &BackendConn{
		Conn:        conn,
		interceptor: newPgInterceptor(conn, defaultBufferSize),
	}
}

func (c *BackendConn) PeekMsg() (typ pgwirebase.ClientMessageType, size int, err error) {
	__antithesis_instrumentation__.Notify(21685)
	byteType, size, err := c.interceptor.PeekMsg()
	return pgwirebase.ClientMessageType(byteType), size, err
}

func (c *BackendConn) ReadMsg() (msg pgproto3.FrontendMessage, err error) {
	__antithesis_instrumentation__.Notify(21686)
	msgBytes, err := c.interceptor.ReadMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21688)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21689)
	}
	__antithesis_instrumentation__.Notify(21687)

	return pgproto3.NewBackend(newChunkReader(msgBytes), &errWriter{}).Receive()
}

func (c *BackendConn) ForwardMsg(dst io.Writer) (n int, err error) {
	__antithesis_instrumentation__.Notify(21690)
	return c.interceptor.ForwardMsg(dst)
}
