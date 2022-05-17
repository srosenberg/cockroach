package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/cockroachdb/cmux"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

func startListenRPCAndSQL(
	ctx, workersCtx context.Context, cfg BaseConfig, stopper *stop.Stopper, grpc *grpcServer,
) (sqlListener net.Listener, startRPCServer func(ctx context.Context), err error) {
	__antithesis_instrumentation__.Notify(235261)
	rpcChanName := "rpc/sql"
	if cfg.SplitListenSQL {
		__antithesis_instrumentation__.Notify(235272)
		rpcChanName = "rpc"
	} else {
		__antithesis_instrumentation__.Notify(235273)
	}
	__antithesis_instrumentation__.Notify(235262)
	var ln net.Listener
	if k := cfg.TestingKnobs.Server; k != nil {
		__antithesis_instrumentation__.Notify(235274)
		knobs := k.(*TestingKnobs)
		ln = knobs.RPCListener
	} else {
		__antithesis_instrumentation__.Notify(235275)
	}
	__antithesis_instrumentation__.Notify(235263)
	if ln == nil {
		__antithesis_instrumentation__.Notify(235276)
		var err error
		ln, err = ListenAndUpdateAddrs(ctx, &cfg.Addr, &cfg.AdvertiseAddr, rpcChanName)
		if err != nil {
			__antithesis_instrumentation__.Notify(235278)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(235279)
		}
		__antithesis_instrumentation__.Notify(235277)
		log.Eventf(ctx, "listening on port %s", cfg.Addr)
	} else {
		__antithesis_instrumentation__.Notify(235280)
	}
	__antithesis_instrumentation__.Notify(235264)

	var pgL net.Listener
	if cfg.SplitListenSQL {
		__antithesis_instrumentation__.Notify(235281)
		pgL, err = ListenAndUpdateAddrs(ctx, &cfg.SQLAddr, &cfg.SQLAdvertiseAddr, "sql")
		if err != nil {
			__antithesis_instrumentation__.Notify(235285)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(235286)
		}
		__antithesis_instrumentation__.Notify(235282)

		waitQuiesce := func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(235287)
			<-stopper.ShouldQuiesce()

			if err := pgL.Close(); err != nil {
				__antithesis_instrumentation__.Notify(235288)
				log.Ops.Fatalf(ctx, "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(235289)
			}
		}
		__antithesis_instrumentation__.Notify(235283)
		if err := stopper.RunAsyncTask(workersCtx, "wait-quiesce", waitQuiesce); err != nil {
			__antithesis_instrumentation__.Notify(235290)
			waitQuiesce(workersCtx)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(235291)
		}
		__antithesis_instrumentation__.Notify(235284)
		log.Eventf(ctx, "listening on sql port %s", cfg.SQLAddr)
	} else {
		__antithesis_instrumentation__.Notify(235292)
	}
	__antithesis_instrumentation__.Notify(235265)

	var serveOnMux sync.Once

	m := cmux.New(ln)

	if !cfg.SplitListenSQL {
		__antithesis_instrumentation__.Notify(235293)

		pgL = m.Match(func(r io.Reader) bool {
			__antithesis_instrumentation__.Notify(235295)
			return pgwire.Match(r)
		})
		__antithesis_instrumentation__.Notify(235294)

		cfg.SQLAddr = cfg.Addr
		cfg.SQLAdvertiseAddr = cfg.AdvertiseAddr
	} else {
		__antithesis_instrumentation__.Notify(235296)
	}
	__antithesis_instrumentation__.Notify(235266)

	anyL := m.Match(cmux.Any())
	if serverTestKnobs, ok := cfg.TestingKnobs.Server.(*TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(235297)
		if serverTestKnobs.ContextTestingKnobs.ArtificialLatencyMap != nil {
			__antithesis_instrumentation__.Notify(235298)
			anyL = rpc.NewDelayingListener(anyL)
		} else {
			__antithesis_instrumentation__.Notify(235299)
		}
	} else {
		__antithesis_instrumentation__.Notify(235300)
	}
	__antithesis_instrumentation__.Notify(235267)

	waitForQuiesce := func(context.Context) {
		__antithesis_instrumentation__.Notify(235301)
		<-stopper.ShouldQuiesce()

		netutil.FatalIfUnexpected(anyL.Close())
	}
	__antithesis_instrumentation__.Notify(235268)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(235302)
		grpc.Stop()
		serveOnMux.Do(func() {
			__antithesis_instrumentation__.Notify(235303)

			netutil.FatalIfUnexpected(m.Serve())
		})
	}))
	__antithesis_instrumentation__.Notify(235269)
	if err := stopper.RunAsyncTask(
		workersCtx, "grpc-quiesce", waitForQuiesce,
	); err != nil {
		__antithesis_instrumentation__.Notify(235304)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(235305)
	}
	__antithesis_instrumentation__.Notify(235270)

	startRPCServer = func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(235306)

		_ = stopper.RunAsyncTask(workersCtx, "serve-grpc", func(context.Context) {
			__antithesis_instrumentation__.Notify(235308)
			netutil.FatalIfUnexpected(grpc.Serve(anyL))
		})
		__antithesis_instrumentation__.Notify(235307)

		_ = stopper.RunAsyncTask(ctx, "serve-mux", func(context.Context) {
			__antithesis_instrumentation__.Notify(235309)
			serveOnMux.Do(func() {
				__antithesis_instrumentation__.Notify(235310)
				netutil.FatalIfUnexpected(m.Serve())
			})
		})
	}
	__antithesis_instrumentation__.Notify(235271)

	return pgL, startRPCServer, nil
}
