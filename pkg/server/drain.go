package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	queryWait = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.shutdown.query_wait",
		"the timeout for waiting for active queries to finish during a drain "+
			"(note that the --drain-wait parameter for cockroach node drain may need adjustment "+
			"after changing this setting)",
		10*time.Second,
		settings.NonNegativeDurationWithMaximum(10*time.Hour),
	).WithPublic()

	drainWait = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.shutdown.drain_wait",
		"the amount of time a server waits in an unready state before proceeding with a drain "+
			"(note that the --drain-wait parameter for cockroach node drain may need adjustment "+
			"after changing this setting. --drain-wait is to specify the duration of the "+
			"whole draining process, while server.shutdown.drain_wait is to set the "+
			"wait time for health probes to notice that the node is not ready.)",
		0*time.Second,
		settings.NonNegativeDurationWithMaximum(10*time.Hour),
	).WithPublic()

	connectionWait = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.shutdown.connection_wait",
		"the maximum amount of time a server waits for all SQL connections to "+
			"be closed before proceeding with a drain. "+
			"(note that the --drain-wait parameter for cockroach node drain may need adjustment "+
			"after changing this setting)",
		0*time.Second,
		settings.NonNegativeDurationWithMaximum(10*time.Hour),
	).WithPublic()
)

func (s *adminServer) Drain(req *serverpb.DrainRequest, stream serverpb.Admin_DrainServer) error {
	__antithesis_instrumentation__.Notify(193222)
	ctx := stream.Context()
	ctx = s.server.AnnotateCtx(ctx)

	nodeID, local, err := s.server.status.parseNodeID(req.NodeId)
	if err != nil {
		__antithesis_instrumentation__.Notify(193225)
		return status.Errorf(codes.InvalidArgument, err.Error())
	} else {
		__antithesis_instrumentation__.Notify(193226)
	}
	__antithesis_instrumentation__.Notify(193223)
	if !local {
		__antithesis_instrumentation__.Notify(193227)

		client, err := s.dialNode(ctx, nodeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(193229)
			return serverError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(193230)
		}
		__antithesis_instrumentation__.Notify(193228)
		return delegateDrain(ctx, req, client, stream)
	} else {
		__antithesis_instrumentation__.Notify(193231)
	}
	__antithesis_instrumentation__.Notify(193224)

	return s.server.drain.handleDrain(ctx, req, stream)
}

type drainServer struct {
	stopper      *stop.Stopper
	grpc         *grpcServer
	sqlServer    *SQLServer
	drainSleepFn func(time.Duration)

	kvServer struct {
		nodeLiveness *liveness.NodeLiveness
		node         *Node
	}
}

func newDrainServer(
	cfg BaseConfig, stopper *stop.Stopper, grpc *grpcServer, sqlServer *SQLServer,
) *drainServer {
	__antithesis_instrumentation__.Notify(193232)
	var drainSleepFn = time.Sleep
	if cfg.TestingKnobs.Server != nil {
		__antithesis_instrumentation__.Notify(193234)
		if cfg.TestingKnobs.Server.(*TestingKnobs).DrainSleepFn != nil {
			__antithesis_instrumentation__.Notify(193235)
			drainSleepFn = cfg.TestingKnobs.Server.(*TestingKnobs).DrainSleepFn
		} else {
			__antithesis_instrumentation__.Notify(193236)
		}
	} else {
		__antithesis_instrumentation__.Notify(193237)
	}
	__antithesis_instrumentation__.Notify(193233)
	return &drainServer{
		stopper:      stopper,
		grpc:         grpc,
		sqlServer:    sqlServer,
		drainSleepFn: drainSleepFn,
	}
}

func (s *drainServer) setNode(node *Node, nodeLiveness *liveness.NodeLiveness) {
	__antithesis_instrumentation__.Notify(193238)
	s.kvServer.node = node
	s.kvServer.nodeLiveness = nodeLiveness
}

func (s *drainServer) handleDrain(
	ctx context.Context, req *serverpb.DrainRequest, stream serverpb.Admin_DrainServer,
) error {
	__antithesis_instrumentation__.Notify(193239)
	log.Ops.Infof(ctx, "drain request received with doDrain = %v, shutdown = %v", req.DoDrain, req.Shutdown)

	res := serverpb.DrainResponse{}
	if req.DoDrain {
		__antithesis_instrumentation__.Notify(193243)
		remaining, info, err := s.runDrain(ctx, req.Verbose)
		if err != nil {
			__antithesis_instrumentation__.Notify(193245)
			log.Ops.Errorf(ctx, "drain failed: %v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(193246)
		}
		__antithesis_instrumentation__.Notify(193244)
		res.DrainRemainingIndicator = remaining
		res.DrainRemainingDescription = info.StripMarkers()
	} else {
		__antithesis_instrumentation__.Notify(193247)
	}
	__antithesis_instrumentation__.Notify(193240)
	if s.isDraining() {
		__antithesis_instrumentation__.Notify(193248)
		res.IsDraining = true
	} else {
		__antithesis_instrumentation__.Notify(193249)
	}
	__antithesis_instrumentation__.Notify(193241)

	if err := stream.Send(&res); err != nil {
		__antithesis_instrumentation__.Notify(193250)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193251)
	}
	__antithesis_instrumentation__.Notify(193242)

	return s.maybeShutdownAfterDrain(ctx, req)
}

func (s *drainServer) maybeShutdownAfterDrain(
	ctx context.Context, req *serverpb.DrainRequest,
) error {
	__antithesis_instrumentation__.Notify(193252)
	if !req.Shutdown {
		__antithesis_instrumentation__.Notify(193255)
		if req.DoDrain {
			__antithesis_instrumentation__.Notify(193257)

			log.Ops.Infof(ctx, "drain request completed without server shutdown")
		} else {
			__antithesis_instrumentation__.Notify(193258)
		}
		__antithesis_instrumentation__.Notify(193256)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(193259)
	}
	__antithesis_instrumentation__.Notify(193253)

	go func() {
		__antithesis_instrumentation__.Notify(193260)

		s.grpc.Stop()
		s.stopper.Stop(ctx)
	}()
	__antithesis_instrumentation__.Notify(193254)

	select {
	case <-s.stopper.IsStopped():
		__antithesis_instrumentation__.Notify(193261)
		return nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(193262)
		return ctx.Err()
	case <-time.After(10 * time.Second):
		__antithesis_instrumentation__.Notify(193263)

		log.Fatal(ctx, "timeout after drain")
		return errors.New("unreachable")
	}
}

func delegateDrain(
	ctx context.Context,
	req *serverpb.DrainRequest,
	client serverpb.AdminClient,
	stream serverpb.Admin_DrainServer,
) error {
	__antithesis_instrumentation__.Notify(193264)

	drainClient, err := client.Drain(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(193267)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193268)
	}
	__antithesis_instrumentation__.Notify(193265)

	for {
		__antithesis_instrumentation__.Notify(193269)

		resp, err := drainClient.Recv()
		if err != nil {
			__antithesis_instrumentation__.Notify(193271)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(193274)
				break
			} else {
				__antithesis_instrumentation__.Notify(193275)
			}
			__antithesis_instrumentation__.Notify(193272)
			if grpcutil.IsClosedConnection(err) {
				__antithesis_instrumentation__.Notify(193276)

				break
			} else {
				__antithesis_instrumentation__.Notify(193277)
			}
			__antithesis_instrumentation__.Notify(193273)

			return err
		} else {
			__antithesis_instrumentation__.Notify(193278)
		}
		__antithesis_instrumentation__.Notify(193270)

		if err := stream.Send(resp); err != nil {
			__antithesis_instrumentation__.Notify(193279)
			return err
		} else {
			__antithesis_instrumentation__.Notify(193280)
		}
	}
	__antithesis_instrumentation__.Notify(193266)
	return nil
}

func (s *drainServer) runDrain(
	ctx context.Context, verbose bool,
) (remaining uint64, info redact.RedactableString, err error) {
	__antithesis_instrumentation__.Notify(193281)
	reports := make(map[redact.SafeString]int)
	var mu syncutil.Mutex
	reporter := func(howMany int, what redact.SafeString) {
		__antithesis_instrumentation__.Notify(193285)
		if howMany > 0 {
			__antithesis_instrumentation__.Notify(193286)
			mu.Lock()
			reports[what] += howMany
			mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(193287)
		}
	}
	__antithesis_instrumentation__.Notify(193282)
	defer func() {
		__antithesis_instrumentation__.Notify(193288)

		var descBuf strings.Builder
		comma := redact.SafeString("")
		for what, howMany := range reports {
			__antithesis_instrumentation__.Notify(193290)
			remaining += uint64(howMany)
			redact.Fprintf(&descBuf, "%s%s: %d", comma, what, howMany)
			comma = ", "
		}
		__antithesis_instrumentation__.Notify(193289)
		info = redact.RedactableString(descBuf.String())
		log.Ops.Infof(ctx, "drain remaining: %d", remaining)
		if info != "" {
			__antithesis_instrumentation__.Notify(193291)
			log.Ops.Infof(ctx, "drain details: %s", info)
		} else {
			__antithesis_instrumentation__.Notify(193292)
		}
	}()
	__antithesis_instrumentation__.Notify(193283)

	if err = s.drainInner(ctx, reporter, verbose); err != nil {
		__antithesis_instrumentation__.Notify(193293)
		return 0, "", err
	} else {
		__antithesis_instrumentation__.Notify(193294)
	}
	__antithesis_instrumentation__.Notify(193284)

	return
}

func (s *drainServer) drainInner(
	ctx context.Context, reporter func(int, redact.SafeString), verbose bool,
) (err error) {
	__antithesis_instrumentation__.Notify(193295)

	if err = s.drainClients(ctx, reporter); err != nil {
		__antithesis_instrumentation__.Notify(193297)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193298)
	}
	__antithesis_instrumentation__.Notify(193296)

	return s.drainNode(ctx, reporter, verbose)
}

func (s *drainServer) isDraining() bool {
	__antithesis_instrumentation__.Notify(193299)
	return s.sqlServer.pgServer.IsDraining() || func() bool {
		__antithesis_instrumentation__.Notify(193300)
		return (s.kvServer.node != nil && func() bool {
			__antithesis_instrumentation__.Notify(193301)
			return s.kvServer.node.IsDraining() == true
		}() == true) == true
	}() == true
}

func (s *drainServer) drainClients(
	ctx context.Context, reporter func(int, redact.SafeString),
) error {
	__antithesis_instrumentation__.Notify(193302)
	shouldDelayDraining := !s.isDraining()

	s.grpc.setMode(modeDraining)
	s.sqlServer.isReady.Set(false)

	if err := s.logOpenConns(ctx); err != nil {
		__antithesis_instrumentation__.Notify(193307)
		log.Ops.Warningf(ctx, "error showing alive SQL connections: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(193308)
	}
	__antithesis_instrumentation__.Notify(193303)

	if shouldDelayDraining {
		__antithesis_instrumentation__.Notify(193309)
		log.Ops.Info(ctx, "waiting for health probes to notice that the node "+
			"is not ready for new sql connections")
		s.drainSleepFn(drainWait.Get(&s.sqlServer.execCfg.Settings.SV))
	} else {
		__antithesis_instrumentation__.Notify(193310)
	}
	__antithesis_instrumentation__.Notify(193304)

	if err := s.sqlServer.pgServer.WaitForSQLConnsToClose(ctx, connectionWait.Get(&s.sqlServer.execCfg.Settings.SV), s.stopper); err != nil {
		__antithesis_instrumentation__.Notify(193311)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193312)
	}
	__antithesis_instrumentation__.Notify(193305)

	queryMaxWait := queryWait.Get(&s.sqlServer.execCfg.Settings.SV)
	if err := s.sqlServer.pgServer.Drain(ctx, queryMaxWait, reporter, s.stopper); err != nil {
		__antithesis_instrumentation__.Notify(193313)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193314)
	}
	__antithesis_instrumentation__.Notify(193306)

	s.sqlServer.distSQLServer.Drain(ctx, queryMaxWait, reporter)

	s.sqlServer.pgServer.SQLServer.GetSQLStatsProvider().(*persistedsqlstats.PersistedSQLStats).Flush(ctx)

	s.sqlServer.leaseMgr.SetDraining(ctx, true, reporter)

	return nil
}

func (s *drainServer) drainNode(
	ctx context.Context, reporter func(int, redact.SafeString), verbose bool,
) (err error) {
	__antithesis_instrumentation__.Notify(193315)
	if s.kvServer.node == nil {
		__antithesis_instrumentation__.Notify(193318)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(193319)
	}
	__antithesis_instrumentation__.Notify(193316)

	if err = s.kvServer.nodeLiveness.SetDraining(ctx, true, reporter); err != nil {
		__antithesis_instrumentation__.Notify(193320)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193321)
	}
	__antithesis_instrumentation__.Notify(193317)

	return s.kvServer.node.SetDraining(true, reporter, verbose)
}

func (s *drainServer) logOpenConns(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(193322)
	return s.stopper.RunAsyncTask(ctx, "log-open-conns", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(193323)
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			__antithesis_instrumentation__.Notify(193324)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(193325)
				log.Ops.Infof(ctx, "number of open connections: %d\n", s.sqlServer.pgServer.GetConnCancelMapLen())
			case <-s.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(193326)
				return
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(193327)
				return
			}
		}
	})
}
