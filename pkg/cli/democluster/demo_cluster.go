package democluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cli/clierror"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/distsql"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/cockroach/pkg/workload/workloadsql"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"golang.org/x/time/rate"
)

type transientCluster struct {
	demoCtx *Context

	connURL       string
	demoDir       string
	useSockets    bool
	stopper       *stop.Stopper
	firstServer   *server.TestServer
	servers       []*server.TestServer
	tenantServers []serverutils.TestTenantInterface
	defaultDB     string

	httpFirstPort int
	sqlFirstPort  int

	adminPassword string
	adminUser     security.SQLUsername

	stickyEngineRegistry server.StickyInMemEnginesRegistry

	getAdminClient   func(ctx context.Context, cfg server.Config) (serverpb.AdminClient, func(), error)
	drainAndShutdown func(ctx context.Context, adminClient serverpb.AdminClient) error

	infoLog  LoggerFn
	warnLog  LoggerFn
	shoutLog ShoutLoggerFn
}

const maxNodeInitTime = 30 * time.Second

const demoOrg = "Cockroach Demo"

type LoggerFn = func(context.Context, string, ...interface{})

type ShoutLoggerFn = func(context.Context, logpb.Severity, string, ...interface{})

func NewDemoCluster(
	ctx context.Context,
	demoCtx *Context,
	infoLog LoggerFn,
	warnLog LoggerFn,
	shoutLog ShoutLoggerFn,
	startStopper func(ctx context.Context) (*stop.Stopper, error),
	getAdminClient func(ctx context.Context, cfg server.Config) (serverpb.AdminClient, func(), error),
	drainAndShutdown func(ctx context.Context, s serverpb.AdminClient) error,
) (DemoCluster, error) {
	__antithesis_instrumentation__.Notify(31943)
	c := &transientCluster{
		demoCtx:          demoCtx,
		getAdminClient:   getAdminClient,
		drainAndShutdown: drainAndShutdown,
		infoLog:          infoLog,
		warnLog:          warnLog,
		shoutLog:         shoutLog,
	}

	c.useSockets = useUnixSocketsInDemo()

	if len(c.demoCtx.Localities) != 0 {
		__antithesis_instrumentation__.Notify(31949)

		if len(c.demoCtx.Localities) != c.demoCtx.NumNodes {
			__antithesis_instrumentation__.Notify(31950)
			return c, errors.Errorf("number of localities specified must equal number of nodes")
		} else {
			__antithesis_instrumentation__.Notify(31951)
		}
	} else {
		__antithesis_instrumentation__.Notify(31952)
		c.demoCtx.Localities = make([]roachpb.Locality, c.demoCtx.NumNodes)
		for i := 0; i < c.demoCtx.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(31953)
			c.demoCtx.Localities[i] = defaultLocalities[i%len(defaultLocalities)]
		}
	}
	__antithesis_instrumentation__.Notify(31944)

	var err error
	c.stopper, err = startStopper(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(31954)
		return c, err
	} else {
		__antithesis_instrumentation__.Notify(31955)
	}
	__antithesis_instrumentation__.Notify(31945)
	c.maybeWarnMemSize(ctx)

	if c.demoDir, err = ioutil.TempDir("", "demo"); err != nil {
		__antithesis_instrumentation__.Notify(31956)
		return c, err
	} else {
		__antithesis_instrumentation__.Notify(31957)
	}
	__antithesis_instrumentation__.Notify(31946)

	if !c.demoCtx.Insecure {
		__antithesis_instrumentation__.Notify(31958)
		if err := c.demoCtx.generateCerts(c.demoDir); err != nil {
			__antithesis_instrumentation__.Notify(31959)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(31960)
		}
	} else {
		__antithesis_instrumentation__.Notify(31961)
	}
	__antithesis_instrumentation__.Notify(31947)

	c.httpFirstPort = c.demoCtx.HTTPPort
	c.sqlFirstPort = c.demoCtx.SQLPort
	if c.demoCtx.Multitenant {
		__antithesis_instrumentation__.Notify(31962)

		c.httpFirstPort += c.demoCtx.NumNodes
		c.sqlFirstPort += c.demoCtx.NumNodes
	} else {
		__antithesis_instrumentation__.Notify(31963)
	}
	__antithesis_instrumentation__.Notify(31948)

	c.stickyEngineRegistry = server.NewStickyInMemEnginesRegistry()
	return c, nil
}

func (c *transientCluster) Start(
	ctx context.Context,
	runInitialSQL func(ctx context.Context, s *server.Server, startSingleNode bool, adminUser, adminPassword string) error,
) (err error) {
	__antithesis_instrumentation__.Notify(31964)
	ctx = logtags.AddTag(ctx, "start-demo-cluster", nil)

	c.defaultDB = catalogkeys.DefaultDatabaseName
	if c.demoCtx.WorkloadGenerator != nil {
		__antithesis_instrumentation__.Notify(31975)
		c.defaultDB = c.demoCtx.WorkloadGenerator.Meta().Name
	} else {
		__antithesis_instrumentation__.Notify(31976)
	}
	__antithesis_instrumentation__.Notify(31965)

	timeoutCh := time.After(maxNodeInitTime)

	errCh := make(chan error, c.demoCtx.NumNodes)

	rpcAddrReadyChs := make([]chan struct{}, c.demoCtx.NumNodes)

	latencyMapWaitChs := make([]chan struct{}, c.demoCtx.NumNodes)

	phaseCtx := logtags.AddTag(ctx, "phase", 1)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(31977)
		c.infoLog(ctx, "creating the first node")

		latencyMapWaitChs[0] = make(chan struct{})
		firstRPCAddrReadyCh, err := c.createAndAddNode(ctx, 0, latencyMapWaitChs[0], timeoutCh)
		if err != nil {
			__antithesis_instrumentation__.Notify(31979)
			return err
		} else {
			__antithesis_instrumentation__.Notify(31980)
		}
		__antithesis_instrumentation__.Notify(31978)
		rpcAddrReadyChs[0] = firstRPCAddrReadyCh
		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(31981)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31982)
	}
	__antithesis_instrumentation__.Notify(31966)

	phaseCtx = logtags.AddTag(ctx, "phase", 2)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(31983)
		c.infoLog(ctx, "starting first node")
		if err := c.startNodeAsync(ctx, 0, errCh, timeoutCh); err != nil {
			__antithesis_instrumentation__.Notify(31985)
			return err
		} else {
			__antithesis_instrumentation__.Notify(31986)
		}
		__antithesis_instrumentation__.Notify(31984)
		c.infoLog(ctx, "waiting for first node RPC address")
		return c.waitForRPCAddrReadinessOrError(ctx, 0, errCh, rpcAddrReadyChs, timeoutCh)
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(31987)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31988)
	}
	__antithesis_instrumentation__.Notify(31967)

	phaseCtx = logtags.AddTag(ctx, "phase", 3)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(31989)
		c.infoLog(ctx, "creating other nodes")

		for i := 1; i < c.demoCtx.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(31993)
			latencyMapWaitChs[i] = make(chan struct{})
			rpcAddrReady, err := c.createAndAddNode(ctx, i, latencyMapWaitChs[i], timeoutCh)
			if err != nil {
				__antithesis_instrumentation__.Notify(31995)
				return err
			} else {
				__antithesis_instrumentation__.Notify(31996)
			}
			__antithesis_instrumentation__.Notify(31994)
			rpcAddrReadyChs[i] = rpcAddrReady
		}
		__antithesis_instrumentation__.Notify(31990)

		c.stopper.AddCloser(stop.CloserFn(func() {
			__antithesis_instrumentation__.Notify(31997)
			c.stickyEngineRegistry.CloseAllStickyInMemEngines()
		}))
		__antithesis_instrumentation__.Notify(31991)

		for i := 1; i < c.demoCtx.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(31998)
			if err := c.startNodeAsync(ctx, i, errCh, timeoutCh); err != nil {
				__antithesis_instrumentation__.Notify(31999)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32000)
			}
		}
		__antithesis_instrumentation__.Notify(31992)
		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(32001)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32002)
	}
	__antithesis_instrumentation__.Notify(31968)

	phaseCtx = logtags.AddTag(ctx, "phase", 4)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(32003)
		c.infoLog(ctx, "waiting for remaining nodes to get their RPC address")

		for i := 0; i < c.demoCtx.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(32005)
			if err := c.waitForRPCAddrReadinessOrError(ctx, i, errCh, rpcAddrReadyChs, timeoutCh); err != nil {
				__antithesis_instrumentation__.Notify(32006)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32007)
			}
		}
		__antithesis_instrumentation__.Notify(32004)
		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(32008)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32009)
	}
	__antithesis_instrumentation__.Notify(31969)

	phaseCtx = logtags.AddTag(ctx, "phase", 5)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(32010)

		if c.demoCtx.SimulateLatency {
			__antithesis_instrumentation__.Notify(32012)

			c.infoLog(ctx, "initializing latency map")
			for i, serv := range c.servers {
				__antithesis_instrumentation__.Notify(32013)
				latencyMap := serv.Cfg.TestingKnobs.Server.(*server.TestingKnobs).ContextTestingKnobs.ArtificialLatencyMap
				srcLocality, ok := serv.Cfg.Locality.Find("region")
				if !ok {
					__antithesis_instrumentation__.Notify(32016)
					continue
				} else {
					__antithesis_instrumentation__.Notify(32017)
				}
				__antithesis_instrumentation__.Notify(32014)
				srcLocalityMap, ok := regionToRegionToLatency[srcLocality]
				if !ok {
					__antithesis_instrumentation__.Notify(32018)
					continue
				} else {
					__antithesis_instrumentation__.Notify(32019)
				}
				__antithesis_instrumentation__.Notify(32015)
				for j, dst := range c.servers {
					__antithesis_instrumentation__.Notify(32020)
					if i == j {
						__antithesis_instrumentation__.Notify(32023)
						continue
					} else {
						__antithesis_instrumentation__.Notify(32024)
					}
					__antithesis_instrumentation__.Notify(32021)
					dstLocality, ok := dst.Cfg.Locality.Find("region")
					if !ok {
						__antithesis_instrumentation__.Notify(32025)
						continue
					} else {
						__antithesis_instrumentation__.Notify(32026)
					}
					__antithesis_instrumentation__.Notify(32022)
					latency := srcLocalityMap[dstLocality]
					latencyMap[dst.ServingRPCAddr()] = latency
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(32027)
		}
		__antithesis_instrumentation__.Notify(32011)
		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(32028)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32029)
	}
	__antithesis_instrumentation__.Notify(31970)

	phaseCtx = logtags.AddTag(ctx, "phase", 6)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(32030)
		for i := 0; i < c.demoCtx.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(32032)
			c.infoLog(ctx, "letting server %d initialize", i)
			close(latencyMapWaitChs[i])
			if err := c.waitForNodeIDReadiness(ctx, i, errCh, timeoutCh); err != nil {
				__antithesis_instrumentation__.Notify(32034)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32035)
			}
			__antithesis_instrumentation__.Notify(32033)
			c.infoLog(ctx, "node n%d initialized", c.servers[i].NodeID())
		}
		__antithesis_instrumentation__.Notify(32031)
		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(32036)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32037)
	}
	__antithesis_instrumentation__.Notify(31971)

	phaseCtx = logtags.AddTag(ctx, "phase", 7)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(32038)
		for i := 0; i < c.demoCtx.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(32040)
			c.infoLog(ctx, "waiting for server %d SQL readiness", i)
			if err := c.waitForSQLReadiness(ctx, i, errCh, timeoutCh); err != nil {
				__antithesis_instrumentation__.Notify(32042)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32043)
			}
			__antithesis_instrumentation__.Notify(32041)
			c.infoLog(ctx, "node n%d ready", c.servers[i].NodeID())
		}
		__antithesis_instrumentation__.Notify(32039)
		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(32044)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32045)
	}
	__antithesis_instrumentation__.Notify(31972)

	const demoUsername = "demo"
	demoPassword := genDemoPassword(demoUsername)

	phaseCtx = logtags.AddTag(ctx, "phase", 8)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(32046)
		if c.demoCtx.Multitenant {
			__antithesis_instrumentation__.Notify(32048)
			c.infoLog(ctx, "starting tenant nodes")

			c.tenantServers = make([]serverutils.TestTenantInterface, c.demoCtx.NumNodes)
			for i := 0; i < c.demoCtx.NumNodes; i++ {
				__antithesis_instrumentation__.Notify(32049)
				latencyMap := c.servers[i].Cfg.TestingKnobs.Server.(*server.TestingKnobs).ContextTestingKnobs.ArtificialLatencyMap
				c.infoLog(ctx, "starting tenant node %d", i)
				tenantStopper := stop.NewStopper()
				ts, err := c.servers[i].StartTenant(ctx, base.TestTenantArgs{

					TenantID:         roachpb.MakeTenantID(uint64(i + 2)),
					Stopper:          tenantStopper,
					ForceInsecure:    c.demoCtx.Insecure,
					SSLCertsDir:      c.demoDir,
					StartingSQLPort:  c.demoCtx.SQLPort - 2,
					StartingHTTPPort: c.demoCtx.HTTPPort - 2,
					Locality:         c.demoCtx.Localities[i],
					TestingKnobs: base.TestingKnobs{
						Server: &server.TestingKnobs{
							ContextTestingKnobs: rpc.ContextTestingKnobs{
								ArtificialLatencyMap: latencyMap,
							},
						},
					},
				})
				c.stopper.AddCloser(stop.CloserFn(func() {
					__antithesis_instrumentation__.Notify(32052)
					stopCtx := context.Background()
					if ts != nil {
						__antithesis_instrumentation__.Notify(32054)
						stopCtx = ts.AnnotateCtx(stopCtx)
					} else {
						__antithesis_instrumentation__.Notify(32055)
					}
					__antithesis_instrumentation__.Notify(32053)
					tenantStopper.Stop(stopCtx)
				}))
				__antithesis_instrumentation__.Notify(32050)
				if err != nil {
					__antithesis_instrumentation__.Notify(32056)
					return err
				} else {
					__antithesis_instrumentation__.Notify(32057)
				}
				__antithesis_instrumentation__.Notify(32051)
				c.tenantServers[i] = ts
				c.infoLog(ctx, "started tenant %d: %s", i, ts.SQLAddr())

				ctx = ts.AnnotateCtx(ctx)

				if !c.demoCtx.Insecure {
					__antithesis_instrumentation__.Notify(32058)

					ie := ts.DistSQLServer().(*distsql.ServerImpl).ServerConfig.Executor
					_, err = ie.Exec(ctx, "tenant-password", nil,
						fmt.Sprintf("CREATE USER %s WITH PASSWORD %s", demoUsername, demoPassword))
					if err != nil {
						__antithesis_instrumentation__.Notify(32060)
						return err
					} else {
						__antithesis_instrumentation__.Notify(32061)
					}
					__antithesis_instrumentation__.Notify(32059)
					_, err = ie.Exec(ctx, "tenant-grant", nil, fmt.Sprintf("GRANT admin TO %s", demoUsername))
					if err != nil {
						__antithesis_instrumentation__.Notify(32062)
						return err
					} else {
						__antithesis_instrumentation__.Notify(32063)
					}
				} else {
					__antithesis_instrumentation__.Notify(32064)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(32065)
		}
		__antithesis_instrumentation__.Notify(32047)
		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(32066)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32067)
	}
	__antithesis_instrumentation__.Notify(31973)

	phaseCtx = logtags.AddTag(ctx, "phase", 9)
	if err := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(32068)

		c.infoLog(ctx, "running initial SQL for demo cluster")

		server := c.firstServer.Server
		ctx = server.AnnotateCtx(ctx)

		if err := runInitialSQL(ctx, server, c.demoCtx.NumNodes < 3, demoUsername, demoPassword); err != nil {
			__antithesis_instrumentation__.Notify(32074)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32075)
		}
		__antithesis_instrumentation__.Notify(32069)
		if c.demoCtx.Insecure {
			__antithesis_instrumentation__.Notify(32076)
			c.adminUser = security.RootUserName()
			c.adminPassword = "unused"
		} else {
			__antithesis_instrumentation__.Notify(32077)
			c.adminUser = security.MakeSQLUsernameFromPreNormalizedString(demoUsername)
			c.adminPassword = demoPassword
		}
		__antithesis_instrumentation__.Notify(32070)

		purl, err := c.getNetworkURLForServer(ctx, 0, true, c.demoCtx.Multitenant)
		if err != nil {
			__antithesis_instrumentation__.Notify(32078)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32079)
		}
		__antithesis_instrumentation__.Notify(32071)
		c.connURL = purl.ToPQ().String()

		if c.demoCtx.ListeningURLFile != "" {
			__antithesis_instrumentation__.Notify(32080)
			c.infoLog(ctx, "listening URL file: %s", c.demoCtx.ListeningURLFile)
			if err = ioutil.WriteFile(c.demoCtx.ListeningURLFile, []byte(fmt.Sprintf("%s\n", c.connURL)), 0644); err != nil {
				__antithesis_instrumentation__.Notify(32081)
				c.warnLog(ctx, "failed writing the URL: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(32082)
			}
		} else {
			__antithesis_instrumentation__.Notify(32083)
		}
		__antithesis_instrumentation__.Notify(32072)

		if !c.demoCtx.DisableTelemetry {
			__antithesis_instrumentation__.Notify(32084)
			c.infoLog(ctx, "starting telemetry")
			c.firstServer.StartDiagnostics(ctx)
		} else {
			__antithesis_instrumentation__.Notify(32085)
		}
		__antithesis_instrumentation__.Notify(32073)

		return nil
	}(phaseCtx); err != nil {
		__antithesis_instrumentation__.Notify(32086)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32087)
	}
	__antithesis_instrumentation__.Notify(31974)
	return nil
}

func (c *transientCluster) createAndAddNode(
	ctx context.Context, idx int, latencyMapWaitCh chan struct{}, timeoutCh <-chan time.Time,
) (rpcAddrReadyCh chan struct{}, err error) {
	__antithesis_instrumentation__.Notify(32088)
	var joinAddr string
	if idx > 0 {
		__antithesis_instrumentation__.Notify(32094)

		joinAddr = c.firstServer.ServingRPCAddr()
	} else {
		__antithesis_instrumentation__.Notify(32095)
	}
	__antithesis_instrumentation__.Notify(32089)
	nodeID := roachpb.NodeID(idx + 1)
	args := c.demoCtx.testServerArgsForTransientCluster(
		c.sockForServer(nodeID, ""), nodeID, joinAddr, c.demoDir,
		c.sqlFirstPort,
		c.httpFirstPort,
		c.stickyEngineRegistry,
	)
	if idx == 0 {
		__antithesis_instrumentation__.Notify(32096)

		args.NoAutoInitializeCluster = false
	} else {
		__antithesis_instrumentation__.Notify(32097)
	}
	__antithesis_instrumentation__.Notify(32090)

	serverKnobs := args.Knobs.Server.(*server.TestingKnobs)

	rpcAddrReadyCh = make(chan struct{})
	serverKnobs.SignalAfterGettingRPCAddress = rpcAddrReadyCh

	serverKnobs.PauseAfterGettingRPCAddress = latencyMapWaitCh

	if c.demoCtx.SimulateLatency {
		__antithesis_instrumentation__.Notify(32098)

		serverKnobs.ContextTestingKnobs = rpc.ContextTestingKnobs{
			ArtificialLatencyMap: make(map[string]int),
		}
	} else {
		__antithesis_instrumentation__.Notify(32099)
	}
	__antithesis_instrumentation__.Notify(32091)

	s, err := server.TestServerFactory.New(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(32100)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(32101)
	}
	__antithesis_instrumentation__.Notify(32092)
	serv := s.(*server.TestServer)

	c.stopper.AddCloser(stop.CloserFn(serv.Stop))

	if idx == 0 {
		__antithesis_instrumentation__.Notify(32102)

		c.firstServer = serv

		logcrash.SetGlobalSettings(&serv.ClusterSettings().SV)
	} else {
		__antithesis_instrumentation__.Notify(32103)
	}
	__antithesis_instrumentation__.Notify(32093)

	c.servers = append(c.servers, serv)

	return rpcAddrReadyCh, nil
}

func (c *transientCluster) startNodeAsync(
	ctx context.Context, idx int, errCh chan error, timeoutCh <-chan time.Time,
) error {
	__antithesis_instrumentation__.Notify(32104)
	if idx > len(c.servers) {
		__antithesis_instrumentation__.Notify(32106)
		return errors.AssertionFailedf("programming error: server %d not created yet", idx)
	} else {
		__antithesis_instrumentation__.Notify(32107)
	}
	__antithesis_instrumentation__.Notify(32105)

	serv := c.servers[idx]
	tag := fmt.Sprintf("start-n%d", idx+1)
	return c.stopper.RunAsyncTask(ctx, tag, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(32108)

		ctx = logtags.WithTags(context.Background(), logtags.FromContext(ctx))
		ctx = logtags.AddTag(ctx, tag, nil)

		err := serv.Start(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(32109)
			c.warnLog(ctx, "server %d failed to start: %v", idx, err)
			select {
			case errCh <- err:
				__antithesis_instrumentation__.Notify(32110)

			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(32111)
			case <-serv.Stopper().ShouldQuiesce():
				__antithesis_instrumentation__.Notify(32112)
			case <-c.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(32113)
			case <-timeoutCh:
				__antithesis_instrumentation__.Notify(32114)
			}
		} else {
			__antithesis_instrumentation__.Notify(32115)
		}
	})
}

func (c *transientCluster) waitForRPCAddrReadinessOrError(
	ctx context.Context,
	idx int,
	errCh chan error,
	rpcAddrReadyChs []chan struct{},
	timeoutCh <-chan time.Time,
) error {
	__antithesis_instrumentation__.Notify(32116)
	if idx >= len(rpcAddrReadyChs) || func() bool {
		__antithesis_instrumentation__.Notify(32118)
		return idx >= len(c.servers) == true
	}() == true {
		__antithesis_instrumentation__.Notify(32119)
		return errors.AssertionFailedf("programming error: server %d not created yet", idx)
	} else {
		__antithesis_instrumentation__.Notify(32120)
	}
	__antithesis_instrumentation__.Notify(32117)

	select {
	case <-rpcAddrReadyChs[idx]:
		__antithesis_instrumentation__.Notify(32121)

		return nil

	case err := <-errCh:
		__antithesis_instrumentation__.Notify(32122)
		return err
	case <-timeoutCh:
		__antithesis_instrumentation__.Notify(32123)
		return errors.Newf("demo startup timeout while waiting for server %d", idx)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(32124)
		return errors.CombineErrors(ctx.Err(), errors.Newf("server %d startup aborted due to context cancellation", idx))
	case <-c.servers[idx].Stopper().ShouldQuiesce():
		__antithesis_instrumentation__.Notify(32125)
		return errors.Newf("server %d stopped prematurely", idx)
	case <-c.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(32126)
		return errors.Newf("demo cluster stopped prematurely while starting server %d", idx)
	}
}

func (c *transientCluster) waitForNodeIDReadiness(
	ctx context.Context, idx int, errCh chan error, timeoutCh <-chan time.Time,
) error {
	__antithesis_instrumentation__.Notify(32127)
	retryOpts := retry.Options{InitialBackoff: 10 * time.Millisecond, MaxBackoff: time.Second}
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(32129)
		c.infoLog(ctx, "waiting for server %d to know its node ID", idx)
		select {

		case <-timeoutCh:
			__antithesis_instrumentation__.Notify(32131)
			return errors.Newf("initialization timeout while waiting for server %d node ID", idx)
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(32132)
			return errors.Wrapf(err, "server %d failed to start", idx)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(32133)
			return errors.CombineErrors(errors.Newf("context cancellation while waiting for server %d to have a node ID", idx), ctx.Err())
		case <-c.servers[idx].Stopper().ShouldQuiesce():
			__antithesis_instrumentation__.Notify(32134)
			return errors.Newf("server %s shut down prematurely", idx)
		case <-c.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(32135)
			return errors.Newf("demo cluster shut down prematurely while waiting for server %d to have a node ID", idx)

		default:
			__antithesis_instrumentation__.Notify(32136)
			if c.servers[idx].NodeID() == 0 {
				__antithesis_instrumentation__.Notify(32137)
				c.infoLog(ctx, "server %d does not know its node ID yet", idx)
				continue
			} else {
				__antithesis_instrumentation__.Notify(32138)
				c.infoLog(ctx, "server %d: n%d", idx, c.servers[idx].NodeID())
			}
		}
		__antithesis_instrumentation__.Notify(32130)
		break
	}
	__antithesis_instrumentation__.Notify(32128)
	return nil
}

func (c *transientCluster) waitForSQLReadiness(
	baseCtx context.Context, idx int, errCh chan error, timeoutCh <-chan time.Time,
) error {
	__antithesis_instrumentation__.Notify(32139)
	retryOpts := retry.Options{InitialBackoff: 10 * time.Millisecond, MaxBackoff: time.Second}
	for r := retry.StartWithCtx(baseCtx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(32141)
		ctx := logtags.AddTag(baseCtx, "n", c.servers[idx].NodeID())
		c.infoLog(ctx, "waiting for server %d to become ready", idx)
		select {

		case <-timeoutCh:
			__antithesis_instrumentation__.Notify(32143)
			return errors.Newf("initialization timeout while waiting for server %d readiness", idx)
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(32144)
			return errors.Wrapf(err, "server %d failed to start", idx)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(32145)
			return errors.CombineErrors(errors.Newf("context cancellation while waiting for server %d to become ready", idx), ctx.Err())
		case <-c.servers[idx].Stopper().ShouldQuiesce():
			__antithesis_instrumentation__.Notify(32146)
			return errors.Newf("server %s shut down prematurely", idx)
		case <-c.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(32147)
			return errors.Newf("demo cluster shut down prematurely while waiting for server %d to become ready", idx)
		default:
			__antithesis_instrumentation__.Notify(32148)
			if err := c.servers[idx].Readiness(ctx); err != nil {
				__antithesis_instrumentation__.Notify(32149)
				c.infoLog(ctx, "server %d not yet ready: %v", idx, err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(32150)
			}
		}
		__antithesis_instrumentation__.Notify(32142)
		break
	}
	__antithesis_instrumentation__.Notify(32140)
	return nil
}

func (demoCtx *Context) testServerArgsForTransientCluster(
	sock unixSocketDetails,
	nodeID roachpb.NodeID,
	joinAddr string,
	demoDir string,
	sqlBasePort, httpBasePort int,
	stickyEngineRegistry server.StickyInMemEnginesRegistry,
) base.TestServerArgs {
	__antithesis_instrumentation__.Notify(32151)

	storeSpec := base.DefaultTestStoreSpec
	storeSpec.StickyInMemoryEngineID = fmt.Sprintf("demo-node%d", nodeID)

	args := base.TestServerArgs{
		SocketFile:              sock.filename(),
		PartOfCluster:           true,
		Stopper:                 stop.NewStopper(),
		JoinAddr:                joinAddr,
		DisableTLSForHTTP:       true,
		StoreSpecs:              []base.StoreSpec{storeSpec},
		SQLMemoryPoolSize:       demoCtx.SQLPoolMemorySize,
		CacheSize:               demoCtx.CacheSize,
		NoAutoInitializeCluster: true,
		EnableDemoLoginEndpoint: true,

		TenantAddr: new(string),
		Knobs: base.TestingKnobs{
			Server: &server.TestingKnobs{
				StickyEngineRegistry: stickyEngineRegistry,
			},
			JobsTestingKnobs: &jobs.TestingKnobs{

				SchedulerDaemonInitialScanDelay: func() time.Duration {
					__antithesis_instrumentation__.Notify(32156)
					return time.Second * 15
				},
			},
		},
	}
	__antithesis_instrumentation__.Notify(32152)

	if !testingForceRandomizeDemoPorts {
		__antithesis_instrumentation__.Notify(32157)

		sqlPort := sqlBasePort + int(nodeID) - 1
		if sqlBasePort == 0 {
			__antithesis_instrumentation__.Notify(32160)
			sqlPort = 0
		} else {
			__antithesis_instrumentation__.Notify(32161)
		}
		__antithesis_instrumentation__.Notify(32158)
		httpPort := httpBasePort + int(nodeID) - 1
		if httpBasePort == 0 {
			__antithesis_instrumentation__.Notify(32162)
			httpPort = 0
		} else {
			__antithesis_instrumentation__.Notify(32163)
		}
		__antithesis_instrumentation__.Notify(32159)
		args.SQLAddr = fmt.Sprintf(":%d", sqlPort)
		args.HTTPAddr = fmt.Sprintf(":%d", httpPort)
	} else {
		__antithesis_instrumentation__.Notify(32164)
	}
	__antithesis_instrumentation__.Notify(32153)

	if demoCtx.Localities != nil {
		__antithesis_instrumentation__.Notify(32165)
		args.Locality = demoCtx.Localities[int(nodeID-1)]
	} else {
		__antithesis_instrumentation__.Notify(32166)
	}
	__antithesis_instrumentation__.Notify(32154)
	if demoCtx.Insecure {
		__antithesis_instrumentation__.Notify(32167)
		args.Insecure = true
	} else {
		__antithesis_instrumentation__.Notify(32168)
		args.Insecure = false
		args.SSLCertsDir = demoDir
	}
	__antithesis_instrumentation__.Notify(32155)

	return args
}

var testingForceRandomizeDemoPorts bool

func TestingForceRandomizeDemoPorts() func() {
	__antithesis_instrumentation__.Notify(32169)
	testingForceRandomizeDemoPorts = true
	return func() { __antithesis_instrumentation__.Notify(32170); testingForceRandomizeDemoPorts = false }
}

func (c *transientCluster) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(32171)
	if c.stopper != nil {
		__antithesis_instrumentation__.Notify(32173)
		c.stopper.Stop(ctx)
	} else {
		__antithesis_instrumentation__.Notify(32174)
	}
	__antithesis_instrumentation__.Notify(32172)
	if c.demoDir != "" {
		__antithesis_instrumentation__.Notify(32175)
		if err := clierror.CheckAndMaybeLog(os.RemoveAll(c.demoDir), c.shoutLog); err != nil {
			__antithesis_instrumentation__.Notify(32176)

			_ = err
		} else {
			__antithesis_instrumentation__.Notify(32177)
		}
	} else {
		__antithesis_instrumentation__.Notify(32178)
	}
}

func (c *transientCluster) DrainAndShutdown(ctx context.Context, nodeID int32) error {
	__antithesis_instrumentation__.Notify(32179)
	if c.demoCtx.SimulateLatency {
		__antithesis_instrumentation__.Notify(32186)
		return errors.Errorf("shutting down nodes is not supported in --%s configurations", cliflags.Global.Name)
	} else {
		__antithesis_instrumentation__.Notify(32187)
	}
	__antithesis_instrumentation__.Notify(32180)
	nodeIndex := int(nodeID - 1)

	if nodeIndex < 0 || func() bool {
		__antithesis_instrumentation__.Notify(32188)
		return nodeIndex >= len(c.servers) == true
	}() == true {
		__antithesis_instrumentation__.Notify(32189)
		return errors.Errorf("node %d does not exist", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(32190)
	}
	__antithesis_instrumentation__.Notify(32181)

	if nodeIndex == 0 {
		__antithesis_instrumentation__.Notify(32191)
		return errors.Errorf("cannot shutdown node %d", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(32192)
	}
	__antithesis_instrumentation__.Notify(32182)
	if c.servers[nodeIndex] == nil {
		__antithesis_instrumentation__.Notify(32193)
		return errors.Errorf("node %d is already shut down", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(32194)
	}
	__antithesis_instrumentation__.Notify(32183)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	adminClient, finish, err := c.getAdminClient(ctx, *(c.servers[nodeIndex].Cfg))
	if err != nil {
		__antithesis_instrumentation__.Notify(32195)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32196)
	}
	__antithesis_instrumentation__.Notify(32184)
	defer finish()

	if err := c.drainAndShutdown(ctx, adminClient); err != nil {
		__antithesis_instrumentation__.Notify(32197)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32198)
	}
	__antithesis_instrumentation__.Notify(32185)
	c.servers[nodeIndex] = nil
	return nil
}

func (c *transientCluster) Recommission(ctx context.Context, nodeID int32) error {
	__antithesis_instrumentation__.Notify(32199)
	nodeIndex := int(nodeID - 1)

	if nodeIndex < 0 || func() bool {
		__antithesis_instrumentation__.Notify(32203)
		return nodeIndex >= len(c.servers) == true
	}() == true {
		__antithesis_instrumentation__.Notify(32204)
		return errors.Errorf("node %d does not exist", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(32205)
	}
	__antithesis_instrumentation__.Notify(32200)

	req := &serverpb.DecommissionRequest{
		NodeIDs:          []roachpb.NodeID{roachpb.NodeID(nodeID)},
		TargetMembership: livenesspb.MembershipStatus_ACTIVE,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	adminClient, finish, err := c.getAdminClient(ctx, *(c.firstServer.Cfg))
	if err != nil {
		__antithesis_instrumentation__.Notify(32206)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32207)
	}
	__antithesis_instrumentation__.Notify(32201)

	defer finish()
	_, err = adminClient.Decommission(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(32208)
		return errors.Wrap(err, "while trying to mark as decommissioning")
	} else {
		__antithesis_instrumentation__.Notify(32209)
	}
	__antithesis_instrumentation__.Notify(32202)

	return nil
}

func (c *transientCluster) Decommission(ctx context.Context, nodeID int32) error {
	__antithesis_instrumentation__.Notify(32210)
	nodeIndex := int(nodeID - 1)

	if nodeIndex < 0 || func() bool {
		__antithesis_instrumentation__.Notify(32214)
		return nodeIndex >= len(c.servers) == true
	}() == true {
		__antithesis_instrumentation__.Notify(32215)
		return errors.Errorf("node %d does not exist", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(32216)
	}
	__antithesis_instrumentation__.Notify(32211)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	adminClient, finish, err := c.getAdminClient(ctx, *(c.firstServer.Cfg))
	if err != nil {
		__antithesis_instrumentation__.Notify(32217)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32218)
	}
	__antithesis_instrumentation__.Notify(32212)
	defer finish()

	{
		__antithesis_instrumentation__.Notify(32219)
		req := &serverpb.DecommissionRequest{
			NodeIDs:          []roachpb.NodeID{roachpb.NodeID(nodeID)},
			TargetMembership: livenesspb.MembershipStatus_DECOMMISSIONING,
		}
		_, err = adminClient.Decommission(ctx, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(32220)
			return errors.Wrap(err, "while trying to mark as decommissioning")
		} else {
			__antithesis_instrumentation__.Notify(32221)
		}
	}

	{
		__antithesis_instrumentation__.Notify(32222)
		req := &serverpb.DecommissionRequest{
			NodeIDs:          []roachpb.NodeID{roachpb.NodeID(nodeID)},
			TargetMembership: livenesspb.MembershipStatus_DECOMMISSIONED,
		}
		_, err = adminClient.Decommission(ctx, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(32223)
			return errors.Wrap(err, "while trying to mark as decommissioned")
		} else {
			__antithesis_instrumentation__.Notify(32224)
		}
	}
	__antithesis_instrumentation__.Notify(32213)

	return nil
}

func (c *transientCluster) RestartNode(ctx context.Context, nodeID int32) error {
	__antithesis_instrumentation__.Notify(32225)
	nodeIndex := int(nodeID - 1)

	if nodeIndex < 0 || func() bool {
		__antithesis_instrumentation__.Notify(32233)
		return nodeIndex >= len(c.servers) == true
	}() == true {
		__antithesis_instrumentation__.Notify(32234)
		return errors.Errorf("node %d does not exist", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(32235)
	}
	__antithesis_instrumentation__.Notify(32226)
	if c.servers[nodeIndex] != nil {
		__antithesis_instrumentation__.Notify(32236)
		return errors.Errorf("node %d is already running", nodeID)
	} else {
		__antithesis_instrumentation__.Notify(32237)
	}
	__antithesis_instrumentation__.Notify(32227)

	if c.demoCtx.SimulateLatency {
		__antithesis_instrumentation__.Notify(32238)
		return errors.Errorf("restarting nodes is not supported in --%s configurations", cliflags.Global.Name)
	} else {
		__antithesis_instrumentation__.Notify(32239)
	}
	__antithesis_instrumentation__.Notify(32228)
	args := c.demoCtx.testServerArgsForTransientCluster(
		c.sockForServer(roachpb.NodeID(nodeID), ""),
		roachpb.NodeID(nodeID),
		c.firstServer.ServingRPCAddr(), c.demoDir,
		c.sqlFirstPort, c.httpFirstPort, c.stickyEngineRegistry)
	s, err := server.TestServerFactory.New(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(32240)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32241)
	}
	__antithesis_instrumentation__.Notify(32229)
	serv := s.(*server.TestServer)

	readyCh := make(chan struct{})
	serv.Cfg.ReadyFn = func(bool) {
		__antithesis_instrumentation__.Notify(32242)
		close(readyCh)
	}
	__antithesis_instrumentation__.Notify(32230)

	if err := serv.Start(ctx); err != nil {
		__antithesis_instrumentation__.Notify(32243)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32244)
	}
	__antithesis_instrumentation__.Notify(32231)

	select {
	case <-readyCh:
		__antithesis_instrumentation__.Notify(32245)
	case <-time.After(maxNodeInitTime):
		__antithesis_instrumentation__.Notify(32246)
		return errors.Newf("could not initialize node %d in time", nodeID)
	}
	__antithesis_instrumentation__.Notify(32232)

	c.stopper.AddCloser(stop.CloserFn(serv.Stop))
	c.servers[nodeIndex] = serv
	return nil
}

func (c *transientCluster) AddNode(
	ctx context.Context, localityString string,
) (newNodeID int32, err error) {
	__antithesis_instrumentation__.Notify(32247)

	if c.demoCtx.SimulateLatency {
		__antithesis_instrumentation__.Notify(32252)
		return 0, errors.Errorf("adding nodes is not supported in --%s configurations", cliflags.Global.Name)
	} else {
		__antithesis_instrumentation__.Notify(32253)
	}
	__antithesis_instrumentation__.Notify(32248)

	re := regexp.MustCompile(`".+"|'.+'|[^'"]+`)
	if localityString != re.FindString(localityString) {
		__antithesis_instrumentation__.Notify(32254)
		return 0, errors.Errorf(`Invalid locality (missing " or '): %s`, localityString)
	} else {
		__antithesis_instrumentation__.Notify(32255)
	}
	__antithesis_instrumentation__.Notify(32249)
	trimmedString := strings.Trim(localityString, `"'`)

	var loc roachpb.Locality
	if err := loc.Set(trimmedString); err != nil {
		__antithesis_instrumentation__.Notify(32256)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(32257)
	}
	__antithesis_instrumentation__.Notify(32250)

	if len(c.demoCtx.Localities) != c.demoCtx.NumNodes || func() bool {
		__antithesis_instrumentation__.Notify(32258)
		return c.demoCtx.NumNodes != len(c.servers) == true
	}() == true {
		__antithesis_instrumentation__.Notify(32259)
		return 0, errors.Errorf("number of localities specified (%d) must equal number of "+
			"nodes (%d) and number of servers (%d)", len(c.demoCtx.Localities), c.demoCtx.NumNodes, len(c.servers))
	} else {
		__antithesis_instrumentation__.Notify(32260)
	}
	__antithesis_instrumentation__.Notify(32251)

	c.servers = append(c.servers, nil)
	c.demoCtx.Localities = append(c.demoCtx.Localities, loc)
	c.demoCtx.NumNodes++

	nodeID := int32(c.demoCtx.NumNodes)

	return nodeID, c.RestartNode(ctx, nodeID)
}

func (c *transientCluster) maybeWarnMemSize(ctx context.Context) {
	__antithesis_instrumentation__.Notify(32261)
	if maxMemory, err := status.GetTotalMemory(ctx); err == nil {
		__antithesis_instrumentation__.Notify(32262)
		requestedMem := (c.demoCtx.CacheSize + c.demoCtx.SQLPoolMemorySize) * int64(c.demoCtx.NumNodes)
		maxRecommendedMem := int64(.75 * float64(maxMemory))
		if requestedMem > maxRecommendedMem {
			__antithesis_instrumentation__.Notify(32263)
			c.shoutLog(
				ctx, severity.WARNING,
				`HIGH MEMORY USAGE
The sum of --max-sql-memory (%s) and --cache (%s) multiplied by the
number of nodes (%d) results in potentially high memory usage on your
device.
This server is running at increased risk of memory-related failures.`,
				humanizeutil.IBytes(c.demoCtx.SQLPoolMemorySize),
				humanizeutil.IBytes(c.demoCtx.CacheSize),
				c.demoCtx.NumNodes,
			)
		} else {
			__antithesis_instrumentation__.Notify(32264)
		}
	} else {
		__antithesis_instrumentation__.Notify(32265)
	}
}

func (demoCtx *Context) generateCerts(certsDir string) (err error) {
	__antithesis_instrumentation__.Notify(32266)
	caKeyPath := filepath.Join(certsDir, security.EmbeddedCAKey)

	if err := security.CreateCAPair(
		certsDir,
		caKeyPath,
		demoCtx.DefaultKeySize,
		demoCtx.DefaultCALifetime,
		false,
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(32271)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32272)
	}
	__antithesis_instrumentation__.Notify(32267)

	if err := security.CreateNodePair(
		certsDir,
		caKeyPath,
		demoCtx.DefaultKeySize,
		demoCtx.DefaultCertLifetime,
		false,
		[]string{"127.0.0.1"},
	); err != nil {
		__antithesis_instrumentation__.Notify(32273)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32274)
	}
	__antithesis_instrumentation__.Notify(32268)

	if err := security.CreateClientPair(
		certsDir,
		caKeyPath,
		demoCtx.DefaultKeySize,
		demoCtx.DefaultCertLifetime,
		false,
		security.RootUserName(),
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(32275)
		return err
	} else {
		__antithesis_instrumentation__.Notify(32276)
	}
	__antithesis_instrumentation__.Notify(32269)

	if demoCtx.Multitenant {
		__antithesis_instrumentation__.Notify(32277)
		tenantCAKeyPath := filepath.Join(certsDir, security.EmbeddedTenantCAKey)

		if err := security.CreateTenantCAPair(
			certsDir,
			tenantCAKeyPath,
			demoCtx.DefaultKeySize,

			demoCtx.DefaultCertLifetime+time.Hour,
			false,
			false,
		); err != nil {
			__antithesis_instrumentation__.Notify(32279)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32280)
		}
		__antithesis_instrumentation__.Notify(32278)

		for i := 0; i < demoCtx.NumNodes; i++ {
			__antithesis_instrumentation__.Notify(32281)

			hostAddrs := []string{
				"127.0.0.1",
				"::1",
				"localhost",
				"*.local",
			}
			pair, err := security.CreateTenantPair(
				certsDir,
				tenantCAKeyPath,
				demoCtx.DefaultKeySize,
				demoCtx.DefaultCertLifetime,
				uint64(i+2),
				hostAddrs,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(32284)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32285)
			}
			__antithesis_instrumentation__.Notify(32282)
			if err := security.WriteTenantPair(certsDir, pair, false); err != nil {
				__antithesis_instrumentation__.Notify(32286)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32287)
			}
			__antithesis_instrumentation__.Notify(32283)
			if err := security.CreateTenantSigningPair(
				certsDir, demoCtx.DefaultCertLifetime, false, uint64(i+2),
			); err != nil {
				__antithesis_instrumentation__.Notify(32288)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32289)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(32290)
	}
	__antithesis_instrumentation__.Notify(32270)
	return nil
}

func (c *transientCluster) getNetworkURLForServer(
	ctx context.Context, serverIdx int, includeAppName bool, isTenant bool,
) (*pgurl.URL, error) {
	__antithesis_instrumentation__.Notify(32291)
	u := pgurl.New()
	if includeAppName {
		__antithesis_instrumentation__.Notify(32296)
		if err := u.SetOption("application_name", catconstants.ReportableAppNamePrefix+"cockroach demo"); err != nil {
			__antithesis_instrumentation__.Notify(32297)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(32298)
		}
	} else {
		__antithesis_instrumentation__.Notify(32299)
	}
	__antithesis_instrumentation__.Notify(32292)
	sqlAddr := c.servers[serverIdx].ServingSQLAddr()
	database := c.defaultDB
	if isTenant {
		__antithesis_instrumentation__.Notify(32300)
		sqlAddr = c.tenantServers[serverIdx].SQLAddr()
	} else {
		__antithesis_instrumentation__.Notify(32301)
	}
	__antithesis_instrumentation__.Notify(32293)
	if !isTenant && func() bool {
		__antithesis_instrumentation__.Notify(32302)
		return c.demoCtx.Multitenant == true
	}() == true {
		__antithesis_instrumentation__.Notify(32303)
		database = catalogkeys.DefaultDatabaseName
	} else {
		__antithesis_instrumentation__.Notify(32304)
	}
	__antithesis_instrumentation__.Notify(32294)
	host, port, _ := addr.SplitHostPort(sqlAddr, "")
	u.
		WithNet(pgurl.NetTCP(host, port)).
		WithDatabase(database).
		WithUsername(c.adminUser.Normalized())

	if c.demoCtx.Insecure {
		__antithesis_instrumentation__.Notify(32305)
		u.WithInsecure()
	} else {
		__antithesis_instrumentation__.Notify(32306)
		u.
			WithAuthn(pgurl.AuthnPassword(true, c.adminPassword)).
			WithTransport(pgurl.TransportTLS(pgurl.TLSRequire, ""))
	}
	__antithesis_instrumentation__.Notify(32295)
	return u, nil
}

func (c *transientCluster) GetConnURL() string {
	__antithesis_instrumentation__.Notify(32307)
	return c.connURL
}

func (c *transientCluster) GetSQLCredentials() (
	adminUser security.SQLUsername,
	adminPassword, certsDir string,
) {
	__antithesis_instrumentation__.Notify(32308)
	return c.adminUser, c.adminPassword, c.demoDir
}

func (c *transientCluster) SetupWorkload(ctx context.Context, licenseDone <-chan error) error {
	__antithesis_instrumentation__.Notify(32309)
	gen := c.demoCtx.WorkloadGenerator

	if gen != nil {
		__antithesis_instrumentation__.Notify(32311)
		db, err := gosql.Open("postgres", c.connURL)
		if err != nil {
			__antithesis_instrumentation__.Notify(32317)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32318)
		}
		__antithesis_instrumentation__.Notify(32312)
		defer db.Close()

		if _, err := db.Exec(`CREATE DATABASE ` + gen.Meta().Name); err != nil {
			__antithesis_instrumentation__.Notify(32319)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32320)
		}
		__antithesis_instrumentation__.Notify(32313)

		ctx := context.TODO()
		var l workloadsql.InsertsDataLoader
		if c.demoCtx.IsInteractive() {
			__antithesis_instrumentation__.Notify(32321)
			fmt.Printf("#\n# Beginning initialization of the %s dataset, please wait...\n", gen.Meta().Name)
		} else {
			__antithesis_instrumentation__.Notify(32322)
		}
		__antithesis_instrumentation__.Notify(32314)
		if _, err := workloadsql.Setup(ctx, db, gen, l); err != nil {
			__antithesis_instrumentation__.Notify(32323)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32324)
		}
		__antithesis_instrumentation__.Notify(32315)

		if c.demoCtx.GeoPartitionedReplicas {
			__antithesis_instrumentation__.Notify(32325)

			if c.demoCtx.IsInteractive() {
				__antithesis_instrumentation__.Notify(32330)
				fmt.Println("#\n# Waiting for license acquisition to complete...")
			} else {
				__antithesis_instrumentation__.Notify(32331)
			}
			__antithesis_instrumentation__.Notify(32326)
			if err := <-licenseDone; err != nil {
				__antithesis_instrumentation__.Notify(32332)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32333)
			}
			__antithesis_instrumentation__.Notify(32327)
			if c.demoCtx.IsInteractive() {
				__antithesis_instrumentation__.Notify(32334)
				fmt.Println("#\n# Partitioning the demo database, please wait...")
			} else {
				__antithesis_instrumentation__.Notify(32335)
			}
			__antithesis_instrumentation__.Notify(32328)

			db, err := gosql.Open("postgres", c.connURL)
			if err != nil {
				__antithesis_instrumentation__.Notify(32336)
				return err
			} else {
				__antithesis_instrumentation__.Notify(32337)
			}
			__antithesis_instrumentation__.Notify(32329)
			defer db.Close()

			if err := gen.(workload.Hookser).Hooks().Partition(db); err != nil {
				__antithesis_instrumentation__.Notify(32338)
				return errors.Wrapf(err, "partitioning the demo database")
			} else {
				__antithesis_instrumentation__.Notify(32339)
			}
		} else {
			__antithesis_instrumentation__.Notify(32340)
		}
		__antithesis_instrumentation__.Notify(32316)

		if c.demoCtx.RunWorkload {
			__antithesis_instrumentation__.Notify(32341)
			var sqlURLs []string
			for i := range c.servers {
				__antithesis_instrumentation__.Notify(32343)
				sqlURL, err := c.getNetworkURLForServer(ctx, i, true, c.demoCtx.Multitenant)
				if err != nil {
					__antithesis_instrumentation__.Notify(32345)
					return err
				} else {
					__antithesis_instrumentation__.Notify(32346)
				}
				__antithesis_instrumentation__.Notify(32344)
				sqlURLs = append(sqlURLs, sqlURL.ToPQ().String())
			}
			__antithesis_instrumentation__.Notify(32342)
			if err := c.runWorkload(ctx, c.demoCtx.WorkloadGenerator, sqlURLs); err != nil {
				__antithesis_instrumentation__.Notify(32347)
				return errors.Wrapf(err, "starting background workload")
			} else {
				__antithesis_instrumentation__.Notify(32348)
			}
		} else {
			__antithesis_instrumentation__.Notify(32349)
		}
	} else {
		__antithesis_instrumentation__.Notify(32350)
	}
	__antithesis_instrumentation__.Notify(32310)

	return nil
}

func (c *transientCluster) runWorkload(
	ctx context.Context, gen workload.Generator, sqlUrls []string,
) error {
	__antithesis_instrumentation__.Notify(32351)
	opser, ok := gen.(workload.Opser)
	if !ok {
		__antithesis_instrumentation__.Notify(32355)
		return errors.Errorf("default dataset %s does not have a workload defined", gen.Meta().Name)
	} else {
		__antithesis_instrumentation__.Notify(32356)
	}
	__antithesis_instrumentation__.Notify(32352)

	reg := histogram.NewRegistry(
		time.Duration(100)*time.Millisecond,
		opser.Meta().Name,
	)
	ops, err := opser.Ops(ctx, sqlUrls, reg)
	if err != nil {
		__antithesis_instrumentation__.Notify(32357)
		return errors.Wrap(err, "unable to create workload")
	} else {
		__antithesis_instrumentation__.Notify(32358)
	}
	__antithesis_instrumentation__.Notify(32353)

	limiter := rate.NewLimiter(rate.Limit(c.demoCtx.WorkloadMaxQPS), 1)

	for _, workerFn := range ops.WorkerFns {
		__antithesis_instrumentation__.Notify(32359)
		workloadFun := func(f func(context.Context) error) func(context.Context) {
			__antithesis_instrumentation__.Notify(32361)
			return func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(32362)
				for {
					__antithesis_instrumentation__.Notify(32363)

					if err := limiter.Wait(ctx); err != nil {
						__antithesis_instrumentation__.Notify(32366)

						panic(err)
					} else {
						__antithesis_instrumentation__.Notify(32367)
					}
					__antithesis_instrumentation__.Notify(32364)
					if err := f(ctx); err != nil {
						__antithesis_instrumentation__.Notify(32368)

						c.warnLog(ctx, "error running workload query: %+v", err)
					} else {
						__antithesis_instrumentation__.Notify(32369)
					}
					__antithesis_instrumentation__.Notify(32365)
					select {
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(32370)
						c.warnLog(ctx, "workload terminating from context cancellation: %v", ctx.Err())
						return
					case <-c.stopper.ShouldQuiesce():
						__antithesis_instrumentation__.Notify(32371)
						c.warnLog(ctx, "demo cluster shutting down")
						return
					default:
						__antithesis_instrumentation__.Notify(32372)
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(32360)

		if err := c.stopper.RunAsyncTask(ctx, "workload", workloadFun(workerFn)); err != nil {
			__antithesis_instrumentation__.Notify(32373)
			return err
		} else {
			__antithesis_instrumentation__.Notify(32374)
		}
	}
	__antithesis_instrumentation__.Notify(32354)

	return nil
}

func (c *transientCluster) AcquireDemoLicense(ctx context.Context) (chan error, error) {
	__antithesis_instrumentation__.Notify(32375)

	licenseDone := make(chan error)
	if c.demoCtx.DisableLicenseAcquisition {
		__antithesis_instrumentation__.Notify(32377)

		close(licenseDone)
	} else {
		__antithesis_instrumentation__.Notify(32378)

		db, err := gosql.Open("postgres", c.connURL)
		if err != nil {
			__antithesis_instrumentation__.Notify(32380)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(32381)
		}
		__antithesis_instrumentation__.Notify(32379)
		go func() {
			__antithesis_instrumentation__.Notify(32382)
			defer db.Close()

			success, err := GetAndApplyLicense(db, c.firstServer.StorageClusterID(), demoOrg)
			if err != nil {
				__antithesis_instrumentation__.Notify(32385)
				select {
				case licenseDone <- err:
					__antithesis_instrumentation__.Notify(32387)

				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(32388)
				case <-c.firstServer.Stopper().ShouldQuiesce():
					__antithesis_instrumentation__.Notify(32389)
				case <-c.stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(32390)
				}
				__antithesis_instrumentation__.Notify(32386)
				return
			} else {
				__antithesis_instrumentation__.Notify(32391)
			}
			__antithesis_instrumentation__.Notify(32383)
			if !success {
				__antithesis_instrumentation__.Notify(32392)
				if c.demoCtx.GeoPartitionedReplicas {
					__antithesis_instrumentation__.Notify(32393)
					select {
					case licenseDone <- errors.WithDetailf(
						errors.New("unable to acquire a license for this demo"),
						"Enterprise features are needed for this demo (--%s).",
						cliflags.DemoGeoPartitionedReplicas.Name):
						__antithesis_instrumentation__.Notify(32395)

					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(32396)
					case <-c.firstServer.Stopper().ShouldQuiesce():
						__antithesis_instrumentation__.Notify(32397)
					case <-c.stopper.ShouldQuiesce():
						__antithesis_instrumentation__.Notify(32398)
					}
					__antithesis_instrumentation__.Notify(32394)
					return
				} else {
					__antithesis_instrumentation__.Notify(32399)
				}
			} else {
				__antithesis_instrumentation__.Notify(32400)
			}
			__antithesis_instrumentation__.Notify(32384)
			close(licenseDone)
		}()
	}
	__antithesis_instrumentation__.Notify(32376)

	return licenseDone, nil
}

func (c *transientCluster) sockForServer(
	nodeID roachpb.NodeID, databaseNameOverride string,
) unixSocketDetails {
	__antithesis_instrumentation__.Notify(32401)
	if !c.useSockets {
		__antithesis_instrumentation__.Notify(32404)
		return unixSocketDetails{}
	} else {
		__antithesis_instrumentation__.Notify(32405)
	}
	__antithesis_instrumentation__.Notify(32402)
	port := strconv.Itoa(c.sqlFirstPort + int(nodeID) - 1)
	databaseName := c.defaultDB
	if databaseNameOverride != "" {
		__antithesis_instrumentation__.Notify(32406)
		databaseName = databaseNameOverride
	} else {
		__antithesis_instrumentation__.Notify(32407)
	}
	__antithesis_instrumentation__.Notify(32403)
	return unixSocketDetails{
		socketDir: c.demoDir,
		port:      port,
		u: pgurl.New().
			WithNet(pgurl.NetUnix(c.demoDir, port)).
			WithUsername(c.adminUser.Normalized()).
			WithAuthn(pgurl.AuthnPassword(true, c.adminPassword)).
			WithDatabase(databaseName),
	}
}

type unixSocketDetails struct {
	socketDir string
	port      string
	u         *pgurl.URL
}

func (s unixSocketDetails) exists() bool {
	__antithesis_instrumentation__.Notify(32408)
	return s.socketDir != ""
}

func (s unixSocketDetails) filename() string {
	__antithesis_instrumentation__.Notify(32409)
	if !s.exists() {
		__antithesis_instrumentation__.Notify(32411)

		return ""
	} else {
		__antithesis_instrumentation__.Notify(32412)
	}
	__antithesis_instrumentation__.Notify(32410)
	return filepath.Join(s.socketDir, ".s.PGSQL."+s.port)
}

func (s unixSocketDetails) String() string {
	__antithesis_instrumentation__.Notify(32413)
	return s.u.ToPQ().String()
}

func (c *transientCluster) NumNodes() int {
	__antithesis_instrumentation__.Notify(32414)
	return len(c.servers)
}

func (c *transientCluster) GetLocality(nodeID int32) string {
	__antithesis_instrumentation__.Notify(32415)
	return c.demoCtx.Localities[nodeID-1].String()
}

func (c *transientCluster) ListDemoNodes(w, ew io.Writer, justOne bool) {
	__antithesis_instrumentation__.Notify(32416)
	numNodesLive := 0

	if c.demoCtx.Multitenant {
		__antithesis_instrumentation__.Notify(32421)
		fmt.Fprintln(w, "system tenant")
	} else {
		__antithesis_instrumentation__.Notify(32422)
	}
	__antithesis_instrumentation__.Notify(32417)
	for i, s := range c.servers {
		__antithesis_instrumentation__.Notify(32423)
		if s == nil {
			__antithesis_instrumentation__.Notify(32430)
			continue
		} else {
			__antithesis_instrumentation__.Notify(32431)
		}
		__antithesis_instrumentation__.Notify(32424)
		numNodesLive++
		if numNodesLive > 1 && func() bool {
			__antithesis_instrumentation__.Notify(32432)
			return justOne == true
		}() == true {
			__antithesis_instrumentation__.Notify(32433)

			continue
		} else {
			__antithesis_instrumentation__.Notify(32434)
		}
		__antithesis_instrumentation__.Notify(32425)

		nodeID := s.NodeID()
		if !justOne {
			__antithesis_instrumentation__.Notify(32435)

			fmt.Fprintf(w, "node %d:\n", nodeID)
		} else {
			__antithesis_instrumentation__.Notify(32436)
		}
		__antithesis_instrumentation__.Notify(32426)
		serverURL := s.Cfg.AdminURL()
		if !c.demoCtx.Insecure {
			__antithesis_instrumentation__.Notify(32437)

			pwauth := url.Values{
				"username": []string{c.adminUser.Normalized()},
				"password": []string{c.adminPassword},
			}
			serverURL.Path = server.DemoLoginPath
			serverURL.RawQuery = pwauth.Encode()
		} else {
			__antithesis_instrumentation__.Notify(32438)
		}
		__antithesis_instrumentation__.Notify(32427)
		fmt.Fprintln(w, "  (webui)   ", serverURL)

		netURL, err := c.getNetworkURLForServer(context.Background(), i,
			false, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(32439)
			fmt.Fprintln(ew, errors.Wrap(err, "retrieving network URL"))
		} else {
			__antithesis_instrumentation__.Notify(32440)
			fmt.Fprintln(w, "  (sql)     ", netURL.ToPQ())
			fmt.Fprintln(w, "  (sql/jdbc)", netURL.ToJDBC())
		}
		__antithesis_instrumentation__.Notify(32428)

		if c.useSockets {
			__antithesis_instrumentation__.Notify(32441)
			databaseNameOverride := ""
			if c.demoCtx.Multitenant {
				__antithesis_instrumentation__.Notify(32443)
				databaseNameOverride = catalogkeys.DefaultDatabaseName
			} else {
				__antithesis_instrumentation__.Notify(32444)
			}
			__antithesis_instrumentation__.Notify(32442)
			sock := c.sockForServer(nodeID, databaseNameOverride)
			fmt.Fprintln(w, "  (sql/unix)", sock)
		} else {
			__antithesis_instrumentation__.Notify(32445)
		}
		__antithesis_instrumentation__.Notify(32429)
		fmt.Fprintln(w)
	}
	__antithesis_instrumentation__.Notify(32418)

	if c.demoCtx.Multitenant {
		__antithesis_instrumentation__.Notify(32446)
		for i := range c.servers {
			__antithesis_instrumentation__.Notify(32447)
			fmt.Fprintf(w, "tenant %d:\n", i+1)
			tenantURL, err := c.getNetworkURLForServer(context.Background(), i,
				false, true)
			if err != nil {
				__antithesis_instrumentation__.Notify(32449)
				fmt.Fprintln(ew, errors.Wrap(err, "retrieving tenant network URL"))
			} else {
				__antithesis_instrumentation__.Notify(32450)
				fmt.Fprintln(w, "   (sql): ", tenantURL.ToPQ())
			}
			__antithesis_instrumentation__.Notify(32448)
			fmt.Fprintln(w)
		}
	} else {
		__antithesis_instrumentation__.Notify(32451)
	}
	__antithesis_instrumentation__.Notify(32419)

	if numNodesLive == 0 {
		__antithesis_instrumentation__.Notify(32452)
		fmt.Fprintln(w, "no demo nodes currently running")
	} else {
		__antithesis_instrumentation__.Notify(32453)
	}
	__antithesis_instrumentation__.Notify(32420)
	if justOne && func() bool {
		__antithesis_instrumentation__.Notify(32454)
		return numNodesLive > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(32455)
		fmt.Fprintln(w, `To display connection parameters for other nodes, use \demo ls.`)
	} else {
		__antithesis_instrumentation__.Notify(32456)
	}
}

func genDemoPassword(username string) string {
	__antithesis_instrumentation__.Notify(32457)
	mypid := os.Getpid()
	return fmt.Sprintf("%s%d", username, mypid)
}
