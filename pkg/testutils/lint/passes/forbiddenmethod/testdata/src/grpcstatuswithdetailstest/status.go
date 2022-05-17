package grpcstatuswithdetailstest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "google.golang.org/grpc/status"

func F() {
	__antithesis_instrumentation__.Notify(644788)
	s := status.New(1, "message")
	_, _ = s.WithDetails()

	_, _ = s.WithDetails()
}
