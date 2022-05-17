package interceptor

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "net"

type PGConn struct {
	net.Conn
	*pgInterceptor
}

func NewPGConn(conn net.Conn) *PGConn {
	__antithesis_instrumentation__.Notify(21769)
	return &PGConn{
		Conn:          conn,
		pgInterceptor: newPgInterceptor(conn, defaultBufferSize),
	}
}

func (c *PGConn) ToFrontendConn() *FrontendConn {
	__antithesis_instrumentation__.Notify(21770)
	return &FrontendConn{Conn: c.Conn, interceptor: c.pgInterceptor}
}

func (c *PGConn) ToBackendConn() *BackendConn {
	__antithesis_instrumentation__.Notify(21771)
	return &BackendConn{Conn: c.Conn, interceptor: c.pgInterceptor}
}
