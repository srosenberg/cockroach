package tenantdirsvr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func ListenAndServeInMemGRPC(
	ctx context.Context, stopper *stop.Stopper, server *grpc.Server,
) (*bufconn.Listener, error) {
	__antithesis_instrumentation__.Notify(23142)

	const defaultListenerBufSize = 1 << 20

	ln := bufconn.Listen(defaultListenerBufSize)

	stopper.AddCloser(stop.CloserFn(server.GracefulStop))
	waitQuiesce := func(context.Context) {
		__antithesis_instrumentation__.Notify(23146)
		<-stopper.ShouldQuiesce()
		fatalIfUnexpected(ln.Close())
	}
	__antithesis_instrumentation__.Notify(23143)
	if err := stopper.RunAsyncTask(ctx, "listen-quiesce", waitQuiesce); err != nil {
		__antithesis_instrumentation__.Notify(23147)
		waitQuiesce(ctx)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23148)
	}
	__antithesis_instrumentation__.Notify(23144)

	if err := stopper.RunAsyncTask(ctx, "serve", func(context.Context) {
		__antithesis_instrumentation__.Notify(23149)
		fatalIfUnexpected(server.Serve(ln))
	}); err != nil {
		__antithesis_instrumentation__.Notify(23150)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(23151)
	}
	__antithesis_instrumentation__.Notify(23145)
	return ln, nil
}

func fatalIfUnexpected(err error) {
	__antithesis_instrumentation__.Notify(23152)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(23153)
		return !isClosedConnection(err) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(23154)
		return !errors.Is(err, stop.ErrUnavailable) == true
	}() == true {
		__antithesis_instrumentation__.Notify(23155)
		log.Fatalf(context.TODO(), "%+v", err)
	} else {
		__antithesis_instrumentation__.Notify(23156)
	}
}

func isClosedConnection(err error) bool {
	__antithesis_instrumentation__.Notify(23157)
	return grpcutil.IsClosedConnection(err) || func() bool {
		__antithesis_instrumentation__.Notify(23158)
		return strings.Contains(err.Error(), "closed") == true
	}() == true
}
