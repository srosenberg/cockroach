package interceptor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"net"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/jackc/pgproto3/v2"
)

type FrontendConn struct {
	net.Conn
	interceptor *pgInterceptor
}

func NewFrontendConn(conn net.Conn) *FrontendConn {
	__antithesis_instrumentation__.Notify(21762)
	return &FrontendConn{
		Conn:        conn,
		interceptor: newPgInterceptor(conn, defaultBufferSize),
	}
}

func (c *FrontendConn) PeekMsg() (typ pgwirebase.ServerMessageType, size int, err error) {
	__antithesis_instrumentation__.Notify(21763)
	byteType, size, err := c.interceptor.PeekMsg()
	return pgwirebase.ServerMessageType(byteType), size, err
}

func (c *FrontendConn) ReadMsg() (msg pgproto3.BackendMessage, err error) {
	__antithesis_instrumentation__.Notify(21764)
	msgBytes, err := c.interceptor.ReadMsg()
	if err != nil {
		__antithesis_instrumentation__.Notify(21766)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21767)
	}
	__antithesis_instrumentation__.Notify(21765)

	return pgproto3.NewFrontend(newChunkReader(msgBytes), &errWriter{}).Receive()
}

func (c *FrontendConn) ForwardMsg(dst io.Writer) (n int, err error) {
	__antithesis_instrumentation__.Notify(21768)
	return c.interceptor.ForwardMsg(dst)
}
