package grpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ClientConn struct{}

func (*ClientConn) Close() { __antithesis_instrumentation__.Notify(644784) }
