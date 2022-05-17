package grpcconnclosetest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "google.golang.org/grpc"

func F() {
	__antithesis_instrumentation__.Notify(644787)
	cc := &grpc.ClientConn{}
	cc.Close()

	cc.Close()
}
