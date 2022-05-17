package testcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"net"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/raft/v3"
)

type TestCluster struct {
	Servers []*server.TestServer
	Conns   []*gosql.DB
	stopper *stop.Stopper
	mu      struct {
		syncutil.Mutex
		serverStoppers []*stop.Stopper
	}
	serverArgs  []base.TestServerArgs
	clusterArgs base.TestClusterArgs

	t testing.TB
}

var _ serverutils.TestClusterInterface = &TestCluster{}

func (tc *TestCluster) NumServers() int {
	__antithesis_instrumentation__.Notify(646446)
	return len(tc.Servers)
}

func (tc *TestCluster) Server(idx int) serverutils.TestServerInterface {
	__antithesis_instrumentation__.Notify(646447)
	return tc.Servers[idx]
}

func (tc *TestCluster) ServerTyped(idx int) *server.TestServer {
	__antithesis_instrumentation__.Notify(646448)
	return tc.Servers[idx]
}

func (tc *TestCluster) ServerConn(idx int) *gosql.DB {
	__antithesis_instrumentation__.Notify(646449)
	return tc.Conns[idx]
}

func (tc *TestCluster) Stopper() *stop.Stopper {
	__antithesis_instrumentation__.Notify(646450)
	return tc.stopper
}

func (tc *TestCluster) stopServers(ctx context.Context) {
	__antithesis_instrumentation__.Notify(646451)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	log.Infof(ctx, "TestCluster quiescing nodes")
	var wg sync.WaitGroup
	wg.Add(len(tc.mu.serverStoppers))
	for i, s := range tc.mu.serverStoppers {
		__antithesis_instrumentation__.Notify(646455)
		go func(i int, s *stop.Stopper) {
			__antithesis_instrumentation__.Notify(646456)
			defer wg.Done()
			if s != nil {
				__antithesis_instrumentation__.Notify(646457)
				quiesceCtx := logtags.AddTag(ctx, "n", tc.Servers[i].NodeID())
				s.Quiesce(quiesceCtx)
			} else {
				__antithesis_instrumentation__.Notify(646458)
			}
		}(i, s)
	}
	__antithesis_instrumentation__.Notify(646452)
	wg.Wait()

	for i := 0; i < tc.NumServers(); i++ {
		__antithesis_instrumentation__.Notify(646459)
		tc.stopServerLocked(i)
	}
	__antithesis_instrumentation__.Notify(646453)

	for i := 0; i < tc.NumServers(); i++ {
		__antithesis_instrumentation__.Notify(646460)

		tracer := tc.Servers[i].Tracer()
		testutils.SucceedsSoon(tc.t, func() error {
			__antithesis_instrumentation__.Notify(646461)
			var sps []tracing.RegistrySpan
			_ = tracer.VisitSpans(func(span tracing.RegistrySpan) error {
				__antithesis_instrumentation__.Notify(646465)
				sps = append(sps, span)
				return nil
			})
			__antithesis_instrumentation__.Notify(646462)
			if len(sps) == 0 {
				__antithesis_instrumentation__.Notify(646466)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(646467)
			}
			__antithesis_instrumentation__.Notify(646463)
			var buf strings.Builder
			fmt.Fprintf(&buf, "unexpectedly found %d active spans:\n", len(sps))
			for _, sp := range sps {
				__antithesis_instrumentation__.Notify(646468)
				fmt.Fprintln(&buf, sp.GetFullRecording(tracing.RecordingVerbose))
				fmt.Fprintln(&buf)
			}
			__antithesis_instrumentation__.Notify(646464)
			return errors.Newf("%s", buf.String())
		})
	}
	__antithesis_instrumentation__.Notify(646454)

	runtime.GC()
}

func (tc *TestCluster) StopServer(idx int) {
	__antithesis_instrumentation__.Notify(646469)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.stopServerLocked(idx)
}

func (tc *TestCluster) stopServerLocked(idx int) {
	__antithesis_instrumentation__.Notify(646470)
	if tc.mu.serverStoppers[idx] != nil {
		__antithesis_instrumentation__.Notify(646471)
		tc.mu.serverStoppers[idx].Stop(context.TODO())
		tc.mu.serverStoppers[idx] = nil
	} else {
		__antithesis_instrumentation__.Notify(646472)
	}
}

func StartTestCluster(t testing.TB, nodes int, args base.TestClusterArgs) *TestCluster {
	__antithesis_instrumentation__.Notify(646473)
	cluster := NewTestCluster(t, nodes, args)
	cluster.Start(t)
	return cluster
}

func NewTestCluster(t testing.TB, nodes int, clusterArgs base.TestClusterArgs) *TestCluster {
	__antithesis_instrumentation__.Notify(646474)
	if nodes < 1 {
		__antithesis_instrumentation__.Notify(646481)
		t.Fatal("invalid cluster size: ", nodes)
	} else {
		__antithesis_instrumentation__.Notify(646482)
	}
	__antithesis_instrumentation__.Notify(646475)

	if err := checkServerArgsForCluster(
		clusterArgs.ServerArgs, clusterArgs.ReplicationMode, disallowJoinAddr,
	); err != nil {
		__antithesis_instrumentation__.Notify(646483)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646484)
	}
	__antithesis_instrumentation__.Notify(646476)
	for _, sargs := range clusterArgs.ServerArgsPerNode {
		__antithesis_instrumentation__.Notify(646485)
		if err := checkServerArgsForCluster(
			sargs, clusterArgs.ReplicationMode, allowJoinAddr,
		); err != nil {
			__antithesis_instrumentation__.Notify(646486)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646487)
		}
	}
	__antithesis_instrumentation__.Notify(646477)

	tc := &TestCluster{
		stopper:     stop.NewStopper(),
		clusterArgs: clusterArgs,
		t:           t,
	}

	noLocalities := true
	for _, arg := range tc.clusterArgs.ServerArgsPerNode {
		__antithesis_instrumentation__.Notify(646488)
		if len(arg.Locality.Tiers) > 0 {
			__antithesis_instrumentation__.Notify(646489)
			noLocalities = false
			break
		} else {
			__antithesis_instrumentation__.Notify(646490)
		}
	}
	__antithesis_instrumentation__.Notify(646478)
	if len(tc.clusterArgs.ServerArgs.Locality.Tiers) > 0 {
		__antithesis_instrumentation__.Notify(646491)
		noLocalities = false
	} else {
		__antithesis_instrumentation__.Notify(646492)
	}
	__antithesis_instrumentation__.Notify(646479)

	var firstListener net.Listener
	for i := 0; i < nodes; i++ {
		__antithesis_instrumentation__.Notify(646493)
		var serverArgs base.TestServerArgs
		if perNodeServerArgs, ok := tc.clusterArgs.ServerArgsPerNode[i]; ok {
			__antithesis_instrumentation__.Notify(646499)
			serverArgs = perNodeServerArgs
		} else {
			__antithesis_instrumentation__.Notify(646500)
			serverArgs = tc.clusterArgs.ServerArgs
		}
		__antithesis_instrumentation__.Notify(646494)

		if len(serverArgs.StoreSpecs) == 0 {
			__antithesis_instrumentation__.Notify(646501)
			serverArgs.StoreSpecs = []base.StoreSpec{base.DefaultTestStoreSpec}
		} else {
			__antithesis_instrumentation__.Notify(646502)
		}
		__antithesis_instrumentation__.Notify(646495)
		if knobs, ok := serverArgs.Knobs.Server.(*server.TestingKnobs); ok && func() bool {
			__antithesis_instrumentation__.Notify(646503)
			return knobs.StickyEngineRegistry != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(646504)
			for j := range serverArgs.StoreSpecs {
				__antithesis_instrumentation__.Notify(646505)
				if serverArgs.StoreSpecs[j].StickyInMemoryEngineID == "" {
					__antithesis_instrumentation__.Notify(646506)
					serverArgs.StoreSpecs[j].StickyInMemoryEngineID = fmt.Sprintf("auto-node%d-store%d", i+1, j+1)
				} else {
					__antithesis_instrumentation__.Notify(646507)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(646508)
		}
		__antithesis_instrumentation__.Notify(646496)

		if noLocalities {
			__antithesis_instrumentation__.Notify(646509)
			tiers := []roachpb.Tier{
				{Key: "region", Value: "test"},
				{Key: "dc", Value: fmt.Sprintf("dc%d", i+1)},
			}
			serverArgs.Locality = roachpb.Locality{Tiers: tiers}
		} else {
			__antithesis_instrumentation__.Notify(646510)
		}
		__antithesis_instrumentation__.Notify(646497)

		if i == 0 {
			__antithesis_instrumentation__.Notify(646511)
			if serverArgs.Listener != nil {
				__antithesis_instrumentation__.Notify(646512)

				firstListener = serverArgs.Listener
			} else {
				__antithesis_instrumentation__.Notify(646513)

				listener, err := net.Listen("tcp", "127.0.0.1:0")
				if err != nil {
					__antithesis_instrumentation__.Notify(646515)
					tc.Stopper().Stop(context.Background())
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(646516)
				}
				__antithesis_instrumentation__.Notify(646514)
				firstListener = listener
				serverArgs.Listener = listener
			}
		} else {
			__antithesis_instrumentation__.Notify(646517)
			if serverArgs.JoinAddr == "" {
				__antithesis_instrumentation__.Notify(646519)

				serverArgs.JoinAddr = firstListener.Addr().String()
			} else {
				__antithesis_instrumentation__.Notify(646520)
			}
			__antithesis_instrumentation__.Notify(646518)
			serverArgs.NoAutoInitializeCluster = true
		}
		__antithesis_instrumentation__.Notify(646498)

		if _, err := tc.AddServer(serverArgs); err != nil {
			__antithesis_instrumentation__.Notify(646521)
			tc.Stopper().Stop(context.Background())
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646522)
		}
	}
	__antithesis_instrumentation__.Notify(646480)

	return tc
}

func (tc *TestCluster) Start(t testing.TB) {
	__antithesis_instrumentation__.Notify(646523)
	nodes := len(tc.Servers)
	var errCh chan error
	if tc.clusterArgs.ParallelStart {
		__antithesis_instrumentation__.Notify(646531)
		errCh = make(chan error, nodes)
	} else {
		__antithesis_instrumentation__.Notify(646532)
	}
	__antithesis_instrumentation__.Notify(646524)

	disableLBS := false
	for i := 0; i < nodes; i++ {
		__antithesis_instrumentation__.Notify(646533)

		if tc.serverArgs[i].ScanInterval > 0 && func() bool {
			__antithesis_instrumentation__.Notify(646535)
			return tc.serverArgs[i].ScanInterval <= 100*time.Millisecond == true
		}() == true {
			__antithesis_instrumentation__.Notify(646536)
			disableLBS = true
		} else {
			__antithesis_instrumentation__.Notify(646537)
		}
		__antithesis_instrumentation__.Notify(646534)

		if tc.clusterArgs.ParallelStart {
			__antithesis_instrumentation__.Notify(646538)
			go func(i int) {
				__antithesis_instrumentation__.Notify(646539)
				errCh <- tc.startServer(i, tc.serverArgs[i])
			}(i)
		} else {
			__antithesis_instrumentation__.Notify(646540)
			if err := tc.startServer(i, tc.serverArgs[i]); err != nil {
				__antithesis_instrumentation__.Notify(646542)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(646543)
			}
			__antithesis_instrumentation__.Notify(646541)

			tc.WaitForNStores(t, i+1, tc.Servers[0].Gossip())
		}
	}
	__antithesis_instrumentation__.Notify(646525)

	if tc.clusterArgs.ParallelStart {
		__antithesis_instrumentation__.Notify(646544)
		for i := 0; i < nodes; i++ {
			__antithesis_instrumentation__.Notify(646546)
			if err := <-errCh; err != nil {
				__antithesis_instrumentation__.Notify(646547)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(646548)
			}
		}
		__antithesis_instrumentation__.Notify(646545)

		tc.WaitForNStores(t, tc.NumServers(), tc.Servers[0].Gossip())
	} else {
		__antithesis_instrumentation__.Notify(646549)
	}
	__antithesis_instrumentation__.Notify(646526)

	if tc.clusterArgs.ReplicationMode == base.ReplicationManual {
		__antithesis_instrumentation__.Notify(646550)

		if _, err := tc.Conns[0].Exec(`SET CLUSTER SETTING kv.range_merge.queue_enabled = false`); err != nil {
			__antithesis_instrumentation__.Notify(646551)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646552)
		}
	} else {
		__antithesis_instrumentation__.Notify(646553)
	}
	__antithesis_instrumentation__.Notify(646527)

	if disableLBS {
		__antithesis_instrumentation__.Notify(646554)
		if _, err := tc.Conns[0].Exec(`SET CLUSTER SETTING kv.range_split.by_load_enabled = false`); err != nil {
			__antithesis_instrumentation__.Notify(646555)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646556)
		}
	} else {
		__antithesis_instrumentation__.Notify(646557)
	}
	__antithesis_instrumentation__.Notify(646528)

	tc.stopper.AddCloser(stop.CloserFn(func() { __antithesis_instrumentation__.Notify(646558); tc.stopServers(context.TODO()) }))
	__antithesis_instrumentation__.Notify(646529)

	if tc.clusterArgs.ReplicationMode == base.ReplicationAuto {
		__antithesis_instrumentation__.Notify(646559)
		if err := tc.WaitForFullReplication(); err != nil {
			__antithesis_instrumentation__.Notify(646560)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646561)
		}
	} else {
		__antithesis_instrumentation__.Notify(646562)
	}
	__antithesis_instrumentation__.Notify(646530)

	tc.WaitForNodeStatuses(t)
}

type checkType bool

const (
	disallowJoinAddr checkType = false
	allowJoinAddr    checkType = true
)

func checkServerArgsForCluster(
	args base.TestServerArgs, replicationMode base.TestClusterReplicationMode, checkType checkType,
) error {
	__antithesis_instrumentation__.Notify(646563)
	if checkType == disallowJoinAddr && func() bool {
		__antithesis_instrumentation__.Notify(646568)
		return args.JoinAddr != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(646569)
		return errors.Errorf("can't specify a join addr when starting a cluster: %s",
			args.JoinAddr)
	} else {
		__antithesis_instrumentation__.Notify(646570)
	}
	__antithesis_instrumentation__.Notify(646564)
	if args.Stopper != nil {
		__antithesis_instrumentation__.Notify(646571)
		return errors.Errorf("can't set individual server stoppers when starting a cluster")
	} else {
		__antithesis_instrumentation__.Notify(646572)
	}
	__antithesis_instrumentation__.Notify(646565)
	if args.Knobs.Store != nil {
		__antithesis_instrumentation__.Notify(646573)
		storeKnobs := args.Knobs.Store.(*kvserver.StoreTestingKnobs)
		if storeKnobs.DisableSplitQueue || func() bool {
			__antithesis_instrumentation__.Notify(646574)
			return storeKnobs.DisableReplicateQueue == true
		}() == true {
			__antithesis_instrumentation__.Notify(646575)
			return errors.Errorf("can't disable an individual server's queues when starting a cluster; " +
				"the cluster controls replication")
		} else {
			__antithesis_instrumentation__.Notify(646576)
		}
	} else {
		__antithesis_instrumentation__.Notify(646577)
	}
	__antithesis_instrumentation__.Notify(646566)

	if replicationMode != base.ReplicationAuto && func() bool {
		__antithesis_instrumentation__.Notify(646578)
		return replicationMode != base.ReplicationManual == true
	}() == true {
		__antithesis_instrumentation__.Notify(646579)
		return errors.Errorf("unexpected replication mode: %s", replicationMode)
	} else {
		__antithesis_instrumentation__.Notify(646580)
	}
	__antithesis_instrumentation__.Notify(646567)

	return nil
}

func (tc *TestCluster) AddAndStartServer(t testing.TB, serverArgs base.TestServerArgs) {
	__antithesis_instrumentation__.Notify(646581)
	if serverArgs.JoinAddr == "" && func() bool {
		__antithesis_instrumentation__.Notify(646584)
		return len(tc.Servers) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(646585)
		serverArgs.JoinAddr = tc.Servers[0].ServingRPCAddr()
	} else {
		__antithesis_instrumentation__.Notify(646586)
	}
	__antithesis_instrumentation__.Notify(646582)
	_, err := tc.AddServer(serverArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(646587)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646588)
	}
	__antithesis_instrumentation__.Notify(646583)

	if err := tc.startServer(len(tc.Servers)-1, serverArgs); err != nil {
		__antithesis_instrumentation__.Notify(646589)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646590)
	}
}

func (tc *TestCluster) AddServer(serverArgs base.TestServerArgs) (*server.TestServer, error) {
	__antithesis_instrumentation__.Notify(646591)
	serverArgs.PartOfCluster = true
	if serverArgs.JoinAddr != "" {
		__antithesis_instrumentation__.Notify(646597)
		serverArgs.NoAutoInitializeCluster = true
	} else {
		__antithesis_instrumentation__.Notify(646598)
	}
	__antithesis_instrumentation__.Notify(646592)

	if err := checkServerArgsForCluster(
		serverArgs,
		tc.clusterArgs.ReplicationMode,

		allowJoinAddr,
	); err != nil {
		__antithesis_instrumentation__.Notify(646599)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646600)
	}
	__antithesis_instrumentation__.Notify(646593)
	if tc.clusterArgs.ReplicationMode == base.ReplicationManual {
		__antithesis_instrumentation__.Notify(646601)
		var stkCopy kvserver.StoreTestingKnobs
		if stk := serverArgs.Knobs.Store; stk != nil {
			__antithesis_instrumentation__.Notify(646603)
			stkCopy = *stk.(*kvserver.StoreTestingKnobs)
		} else {
			__antithesis_instrumentation__.Notify(646604)
		}
		__antithesis_instrumentation__.Notify(646602)
		stkCopy.DisableSplitQueue = true
		stkCopy.DisableMergeQueue = true
		stkCopy.DisableReplicateQueue = true
		serverArgs.Knobs.Store = &stkCopy
	} else {
		__antithesis_instrumentation__.Notify(646605)
	}
	__antithesis_instrumentation__.Notify(646594)

	if serverArgs.Listener != nil {
		__antithesis_instrumentation__.Notify(646606)

		if serverArgs.Knobs.Server == nil {
			__antithesis_instrumentation__.Notify(646608)
			serverArgs.Knobs.Server = &server.TestingKnobs{}
		} else {
			__antithesis_instrumentation__.Notify(646609)

			knobs := *serverArgs.Knobs.Server.(*server.TestingKnobs)
			serverArgs.Knobs.Server = &knobs
		}
		__antithesis_instrumentation__.Notify(646607)

		serverArgs.Knobs.Server.(*server.TestingKnobs).RPCListener = serverArgs.Listener
		serverArgs.Addr = serverArgs.Listener.Addr().String()
	} else {
		__antithesis_instrumentation__.Notify(646610)
	}
	__antithesis_instrumentation__.Notify(646595)

	srv, err := serverutils.NewServer(serverArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(646611)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646612)
	}
	__antithesis_instrumentation__.Notify(646596)
	s := srv.(*server.TestServer)

	tc.Servers = append(tc.Servers, s)
	tc.serverArgs = append(tc.serverArgs, serverArgs)

	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.mu.serverStoppers = append(tc.mu.serverStoppers, s.Stopper())
	return s, nil
}

func (tc *TestCluster) startServer(idx int, serverArgs base.TestServerArgs) error {
	__antithesis_instrumentation__.Notify(646613)
	server := tc.Servers[idx]
	if err := server.Start(context.Background()); err != nil {
		__antithesis_instrumentation__.Notify(646616)
		return err
	} else {
		__antithesis_instrumentation__.Notify(646617)
	}
	__antithesis_instrumentation__.Notify(646614)

	dbConn, err := serverutils.OpenDBConnE(
		server.ServingSQLAddr(), serverArgs.UseDatabase, serverArgs.Insecure, server.Stopper())
	if err != nil {
		__antithesis_instrumentation__.Notify(646618)
		return err
	} else {
		__antithesis_instrumentation__.Notify(646619)
	}
	__antithesis_instrumentation__.Notify(646615)

	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.Conns = append(tc.Conns, dbConn)
	return nil
}

func (tc *TestCluster) WaitForNStores(t testing.TB, n int, g *gossip.Gossip) {
	__antithesis_instrumentation__.Notify(646620)

	var storesMu syncutil.Mutex
	stores := map[roachpb.StoreID]struct{}{}
	storesDone := make(chan error)
	storesDoneOnce := storesDone
	unregister := g.RegisterCallback(gossip.MakePrefixPattern(gossip.KeyStorePrefix),
		func(_ string, content roachpb.Value) {
			__antithesis_instrumentation__.Notify(646622)
			storesMu.Lock()
			defer storesMu.Unlock()
			if storesDoneOnce == nil {
				__antithesis_instrumentation__.Notify(646625)
				return
			} else {
				__antithesis_instrumentation__.Notify(646626)
			}
			__antithesis_instrumentation__.Notify(646623)

			var desc roachpb.StoreDescriptor
			if err := content.GetProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(646627)
				storesDoneOnce <- err
				return
			} else {
				__antithesis_instrumentation__.Notify(646628)
			}
			__antithesis_instrumentation__.Notify(646624)

			stores[desc.StoreID] = struct{}{}
			if len(stores) == n {
				__antithesis_instrumentation__.Notify(646629)
				close(storesDoneOnce)
				storesDoneOnce = nil
			} else {
				__antithesis_instrumentation__.Notify(646630)
			}
		})
	__antithesis_instrumentation__.Notify(646621)
	defer unregister()

	for err := range storesDone {
		__antithesis_instrumentation__.Notify(646631)
		if err != nil {
			__antithesis_instrumentation__.Notify(646632)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646633)
		}
	}
}

func (tc *TestCluster) LookupRange(key roachpb.Key) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646634)
	return tc.Servers[0].LookupRange(key)
}

func (tc *TestCluster) LookupRangeOrFatal(t testing.TB, key roachpb.Key) roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(646635)
	t.Helper()
	desc, err := tc.LookupRange(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(646637)
		t.Fatalf(`looking up range for %s: %+v`, key, err)
	} else {
		__antithesis_instrumentation__.Notify(646638)
	}
	__antithesis_instrumentation__.Notify(646636)
	return desc
}

func (tc *TestCluster) SplitRangeWithExpiration(
	splitKey roachpb.Key, expirationTime hlc.Timestamp,
) (roachpb.RangeDescriptor, roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646639)
	return tc.Servers[0].SplitRangeWithExpiration(splitKey, expirationTime)
}

func (tc *TestCluster) SplitRange(
	splitKey roachpb.Key,
) (roachpb.RangeDescriptor, roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646640)
	return tc.Servers[0].SplitRange(splitKey)
}

func (tc *TestCluster) SplitRangeOrFatal(
	t testing.TB, splitKey roachpb.Key,
) (roachpb.RangeDescriptor, roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(646641)
	lhsDesc, rhsDesc, err := tc.Servers[0].SplitRange(splitKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(646643)
		t.Fatalf(`splitting at %s: %+v`, splitKey, err)
	} else {
		__antithesis_instrumentation__.Notify(646644)
	}
	__antithesis_instrumentation__.Notify(646642)
	return lhsDesc, rhsDesc
}

func (tc *TestCluster) Target(serverIdx int) roachpb.ReplicationTarget {
	__antithesis_instrumentation__.Notify(646645)
	s := tc.Servers[serverIdx]
	return roachpb.ReplicationTarget{
		NodeID:  s.GetNode().Descriptor.NodeID,
		StoreID: s.GetFirstStoreID(),
	}
}

func (tc *TestCluster) Targets(serverIdxs ...int) []roachpb.ReplicationTarget {
	__antithesis_instrumentation__.Notify(646646)
	ret := make([]roachpb.ReplicationTarget, 0, len(serverIdxs))
	for _, serverIdx := range serverIdxs {
		__antithesis_instrumentation__.Notify(646648)
		ret = append(ret, tc.Target(serverIdx))
	}
	__antithesis_instrumentation__.Notify(646647)
	return ret
}

func (tc *TestCluster) changeReplicas(
	changeType roachpb.ReplicaChangeType, startKey roachpb.RKey, targets ...roachpb.ReplicationTarget,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646649)
	tc.t.Helper()
	ctx := context.TODO()

	var returnErr error
	var desc *roachpb.RangeDescriptor
	if err := testutils.SucceedsSoonError(func() error {
		__antithesis_instrumentation__.Notify(646652)
		tc.t.Helper()
		var beforeDesc roachpb.RangeDescriptor
		if err := tc.Servers[0].DB().GetProto(
			ctx, keys.RangeDescriptorKey(startKey), &beforeDesc,
		); err != nil {
			__antithesis_instrumentation__.Notify(646655)
			return errors.Wrap(err, "range descriptor lookup error")
		} else {
			__antithesis_instrumentation__.Notify(646656)
		}
		__antithesis_instrumentation__.Notify(646653)
		var err error
		desc, err = tc.Servers[0].DB().AdminChangeReplicas(
			ctx, startKey.AsRawKey(), beforeDesc, roachpb.MakeReplicationChanges(changeType, targets...),
		)
		if kvserver.IsRetriableReplicationChangeError(err) {
			__antithesis_instrumentation__.Notify(646657)
			tc.t.Logf("encountered retriable replication change error: %v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646658)
		}
		__antithesis_instrumentation__.Notify(646654)

		returnErr = err
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(646659)
		returnErr = err
	} else {
		__antithesis_instrumentation__.Notify(646660)
	}
	__antithesis_instrumentation__.Notify(646650)

	if returnErr != nil {
		__antithesis_instrumentation__.Notify(646661)

		return roachpb.RangeDescriptor{}, errors.Handled(errors.Wrap(returnErr, "AdminChangeReplicas error"))
	} else {
		__antithesis_instrumentation__.Notify(646662)
	}
	__antithesis_instrumentation__.Notify(646651)
	return *desc, nil
}

func (tc *TestCluster) addReplica(
	startKey roachpb.Key, typ roachpb.ReplicaChangeType, targets ...roachpb.ReplicationTarget,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646663)
	rKey := keys.MustAddr(startKey)

	rangeDesc, err := tc.changeReplicas(
		typ, rKey, targets...,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(646666)
		return roachpb.RangeDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(646667)
	}
	__antithesis_instrumentation__.Notify(646664)

	if err := tc.waitForNewReplicas(startKey, false, targets...); err != nil {
		__antithesis_instrumentation__.Notify(646668)
		return roachpb.RangeDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(646669)
	}
	__antithesis_instrumentation__.Notify(646665)

	return rangeDesc, nil
}

func (tc *TestCluster) AddVoters(
	startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646670)
	return tc.addReplica(startKey, roachpb.ADD_VOTER, targets...)
}

func (tc *TestCluster) AddNonVoters(
	startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646671)
	return tc.addReplica(startKey, roachpb.ADD_NON_VOTER, targets...)
}

func (tc *TestCluster) AddNonVotersOrFatal(
	t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(646672)
	desc, err := tc.addReplica(startKey, roachpb.ADD_NON_VOTER, targets...)
	if err != nil {
		__antithesis_instrumentation__.Notify(646674)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646675)
	}
	__antithesis_instrumentation__.Notify(646673)
	return desc
}

func (tc *TestCluster) AddVotersMulti(
	kts ...serverutils.KeyAndTargets,
) ([]roachpb.RangeDescriptor, []error) {
	__antithesis_instrumentation__.Notify(646676)
	var descs []roachpb.RangeDescriptor
	var errs []error
	for _, kt := range kts {
		__antithesis_instrumentation__.Notify(646679)
		rKey := keys.MustAddr(kt.StartKey)

		rangeDesc, err := tc.changeReplicas(
			roachpb.ADD_VOTER, rKey, kt.Targets...,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(646681)
			errs = append(errs, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(646682)
		}
		__antithesis_instrumentation__.Notify(646680)

		descs = append(descs, rangeDesc)
	}
	__antithesis_instrumentation__.Notify(646677)

	for _, kt := range kts {
		__antithesis_instrumentation__.Notify(646683)
		if err := tc.waitForNewReplicas(kt.StartKey, false, kt.Targets...); err != nil {
			__antithesis_instrumentation__.Notify(646684)
			errs = append(errs, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(646685)
		}
	}
	__antithesis_instrumentation__.Notify(646678)

	return descs, errs
}

func (tc *TestCluster) WaitForVoters(
	startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) error {
	__antithesis_instrumentation__.Notify(646686)
	return tc.waitForNewReplicas(startKey, true, targets...)
}

func (tc *TestCluster) waitForNewReplicas(
	startKey roachpb.Key, waitForVoter bool, targets ...roachpb.ReplicationTarget,
) error {
	__antithesis_instrumentation__.Notify(646687)
	rKey := keys.MustAddr(startKey)
	errRetry := errors.Errorf("target not found")

	if err := retry.ForDuration(time.Second*25, func() error {
		__antithesis_instrumentation__.Notify(646689)
		for _, target := range targets {
			__antithesis_instrumentation__.Notify(646691)

			store, err := tc.findMemberStore(target.StoreID)
			if err != nil {
				__antithesis_instrumentation__.Notify(646694)
				log.Errorf(context.TODO(), "unexpected error: %s", err)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646695)
			}
			__antithesis_instrumentation__.Notify(646692)
			repl := store.LookupReplica(rKey)
			if repl == nil {
				__antithesis_instrumentation__.Notify(646696)
				return errors.Wrapf(errRetry, "for target %s", target)
			} else {
				__antithesis_instrumentation__.Notify(646697)
			}
			__antithesis_instrumentation__.Notify(646693)
			desc := repl.Desc()
			if replDesc, ok := desc.GetReplicaDescriptor(target.StoreID); !ok {
				__antithesis_instrumentation__.Notify(646698)
				return errors.Errorf("target store %d not yet in range descriptor %v", target.StoreID, desc)
			} else {
				__antithesis_instrumentation__.Notify(646699)
				if waitForVoter && func() bool {
					__antithesis_instrumentation__.Notify(646700)
					return replDesc.GetType() != roachpb.VOTER_FULL == true
				}() == true {
					__antithesis_instrumentation__.Notify(646701)
					return errors.Errorf("target store %d not yet voter in range descriptor %v", target.StoreID, desc)
				} else {
					__antithesis_instrumentation__.Notify(646702)
				}
			}
		}
		__antithesis_instrumentation__.Notify(646690)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(646703)
		return err
	} else {
		__antithesis_instrumentation__.Notify(646704)
	}
	__antithesis_instrumentation__.Notify(646688)

	return nil
}

func (tc *TestCluster) AddVotersOrFatal(
	t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(646705)
	t.Helper()
	desc, err := tc.AddVoters(startKey, targets...)
	if err != nil {
		__antithesis_instrumentation__.Notify(646707)
		t.Fatalf(`could not add %v replicas to range containing %s: %+v`,
			targets, startKey, err)
	} else {
		__antithesis_instrumentation__.Notify(646708)
	}
	__antithesis_instrumentation__.Notify(646706)
	return desc
}

func (tc *TestCluster) RemoveVoters(
	startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646709)
	return tc.changeReplicas(roachpb.REMOVE_VOTER, keys.MustAddr(startKey), targets...)
}

func (tc *TestCluster) RemoveVotersOrFatal(
	t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(646710)
	t.Helper()
	desc, err := tc.RemoveVoters(startKey, targets...)
	if err != nil {
		__antithesis_instrumentation__.Notify(646712)
		t.Fatalf(`could not remove %v replicas from range containing %s: %+v`,
			targets, startKey, err)
	} else {
		__antithesis_instrumentation__.Notify(646713)
	}
	__antithesis_instrumentation__.Notify(646711)
	return desc
}

func (tc *TestCluster) RemoveNonVoters(
	startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) (roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646714)
	return tc.changeReplicas(roachpb.REMOVE_NON_VOTER, keys.MustAddr(startKey), targets...)
}

func (tc *TestCluster) RemoveNonVotersOrFatal(
	t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
) roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(646715)
	desc, err := tc.RemoveNonVoters(startKey, targets...)
	if err != nil {
		__antithesis_instrumentation__.Notify(646717)
		t.Fatalf(`could not remove %v replicas from range containing %s: %+v`,
			targets, startKey, err)
	} else {
		__antithesis_instrumentation__.Notify(646718)
	}
	__antithesis_instrumentation__.Notify(646716)
	return desc
}

func (tc *TestCluster) SwapVoterWithNonVoter(
	startKey roachpb.Key, voterTarget, nonVoterTarget roachpb.ReplicationTarget,
) (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(646719)
	ctx := context.Background()
	key := keys.MustAddr(startKey)
	var beforeDesc roachpb.RangeDescriptor
	if err := tc.Servers[0].DB().GetProto(
		ctx, keys.RangeDescriptorKey(key), &beforeDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(646721)
		return nil, errors.Wrap(err, "range descriptor lookup error")
	} else {
		__antithesis_instrumentation__.Notify(646722)
	}
	__antithesis_instrumentation__.Notify(646720)
	changes := []roachpb.ReplicationChange{
		{ChangeType: roachpb.ADD_VOTER, Target: nonVoterTarget},
		{ChangeType: roachpb.REMOVE_NON_VOTER, Target: nonVoterTarget},
		{ChangeType: roachpb.ADD_NON_VOTER, Target: voterTarget},
		{ChangeType: roachpb.REMOVE_VOTER, Target: voterTarget},
	}

	return tc.Servers[0].DB().AdminChangeReplicas(ctx, key, beforeDesc, changes)
}

func (tc *TestCluster) SwapVoterWithNonVoterOrFatal(
	t *testing.T, startKey roachpb.Key, voterTarget, nonVoterTarget roachpb.ReplicationTarget,
) *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(646723)
	afterDesc, err := tc.SwapVoterWithNonVoter(startKey, voterTarget, nonVoterTarget)

	require.NoError(t, err)
	replDesc, ok := afterDesc.GetReplicaDescriptor(voterTarget.StoreID)
	require.True(t, ok)
	require.Equal(t, roachpb.NON_VOTER, replDesc.GetType())
	replDesc, ok = afterDesc.GetReplicaDescriptor(nonVoterTarget.StoreID)
	require.True(t, ok)
	require.Equal(t, roachpb.VOTER_FULL, replDesc.GetType())

	return afterDesc
}

func (tc *TestCluster) TransferRangeLease(
	rangeDesc roachpb.RangeDescriptor, dest roachpb.ReplicationTarget,
) error {
	__antithesis_instrumentation__.Notify(646724)
	err := tc.Servers[0].DB().AdminTransferLease(context.TODO(),
		rangeDesc.StartKey.AsRawKey(), dest.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(646726)
		return errors.Wrapf(err, "%q: transfer lease unexpected error", rangeDesc.StartKey)
	} else {
		__antithesis_instrumentation__.Notify(646727)
	}
	__antithesis_instrumentation__.Notify(646725)
	return nil
}

func (tc *TestCluster) TransferRangeLeaseOrFatal(
	t testing.TB, rangeDesc roachpb.RangeDescriptor, dest roachpb.ReplicationTarget,
) {
	__antithesis_instrumentation__.Notify(646728)
	if err := tc.TransferRangeLease(rangeDesc, dest); err != nil {
		__antithesis_instrumentation__.Notify(646729)
		t.Fatalf(`could not transfer lease for range %s error is %+v`, rangeDesc, err)
	} else {
		__antithesis_instrumentation__.Notify(646730)
	}
}

func (tc *TestCluster) RemoveLeaseHolderOrFatal(
	t testing.TB,
	rangeDesc roachpb.RangeDescriptor,
	src roachpb.ReplicationTarget,
	dest roachpb.ReplicationTarget,
) {
	__antithesis_instrumentation__.Notify(646731)
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(646732)
		if err := tc.TransferRangeLease(rangeDesc, dest); err != nil {
			__antithesis_instrumentation__.Notify(646735)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646736)
		}
		__antithesis_instrumentation__.Notify(646733)
		if _, err := tc.RemoveVoters(rangeDesc.StartKey.AsRawKey(), src); err != nil {
			__antithesis_instrumentation__.Notify(646737)
			if strings.Contains(err.Error(), "to remove self (leaseholder)") {
				__antithesis_instrumentation__.Notify(646739)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646740)
			}
			__antithesis_instrumentation__.Notify(646738)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(646741)
		}
		__antithesis_instrumentation__.Notify(646734)
		return nil
	})
}

func (tc *TestCluster) MoveRangeLeaseNonCooperatively(
	ctx context.Context,
	rangeDesc roachpb.RangeDescriptor,
	dest roachpb.ReplicationTarget,
	manual *hlc.HybridManualClock,
) (*roachpb.Lease, error) {
	__antithesis_instrumentation__.Notify(646742)
	knobs := tc.clusterArgs.ServerArgs.Knobs.Store.(*kvserver.StoreTestingKnobs)
	if !knobs.AllowLeaseRequestProposalsWhenNotLeader {
		__antithesis_instrumentation__.Notify(646747)

		return nil, errors.Errorf("must set StoreTestingKnobs.AllowLeaseRequestProposalsWhenNotLeader")
	} else {
		__antithesis_instrumentation__.Notify(646748)
	}
	__antithesis_instrumentation__.Notify(646743)

	destServer, err := tc.FindMemberServer(dest.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(646749)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646750)
	}
	__antithesis_instrumentation__.Notify(646744)
	destStore, err := destServer.Stores().GetStore(dest.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(646751)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646752)
	}
	__antithesis_instrumentation__.Notify(646745)

	const retryDur = testutils.DefaultSucceedsSoonDuration
	var newLease *roachpb.Lease
	err = retry.ForDuration(retryDur, func() error {
		__antithesis_instrumentation__.Notify(646753)

		prevLease, _, err := tc.FindRangeLease(rangeDesc, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(646761)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646762)
		}
		__antithesis_instrumentation__.Notify(646754)
		if prevLease.Replica.StoreID == dest.StoreID {
			__antithesis_instrumentation__.Notify(646763)
			newLease = &prevLease
			return nil
		} else {
			__antithesis_instrumentation__.Notify(646764)
		}
		__antithesis_instrumentation__.Notify(646755)

		lhStore, err := tc.findMemberStore(prevLease.Replica.StoreID)
		if err != nil {
			__antithesis_instrumentation__.Notify(646765)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646766)
		}
		__antithesis_instrumentation__.Notify(646756)
		log.Infof(ctx, "test: advancing clock to lease expiration")
		manual.Increment(lhStore.GetStoreConfig().LeaseExpiration())

		err = destServer.HeartbeatNodeLiveness()
		if err != nil {
			__antithesis_instrumentation__.Notify(646767)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646768)
		}
		__antithesis_instrumentation__.Notify(646757)

		r, err := destStore.GetReplica(rangeDesc.RangeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(646769)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646770)
		}
		__antithesis_instrumentation__.Notify(646758)
		ls, err := r.TestingAcquireLease(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(646771)
			log.Infof(ctx, "TestingAcquireLease failed: %s", err)
			if lErr := (*roachpb.NotLeaseHolderError)(nil); errors.As(err, &lErr) && func() bool {
				__antithesis_instrumentation__.Notify(646772)
				return lErr.Lease != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(646773)
				newLease = lErr.Lease
			} else {
				__antithesis_instrumentation__.Notify(646774)
				return err
			}
		} else {
			__antithesis_instrumentation__.Notify(646775)
			newLease = &ls.Lease
		}
		__antithesis_instrumentation__.Notify(646759)

		if newLease.Replica.StoreID != dest.StoreID {
			__antithesis_instrumentation__.Notify(646776)
			return errors.Errorf("LeaseInfoRequest succeeded, "+
				"but lease in wrong location, want %v, got %v", dest, newLease.Replica)
		} else {
			__antithesis_instrumentation__.Notify(646777)
		}
		__antithesis_instrumentation__.Notify(646760)
		return nil
	})
	__antithesis_instrumentation__.Notify(646746)
	log.Infof(ctx, "MoveRangeLeaseNonCooperatively: acquired lease: %s. err: %v", newLease, err)
	return newLease, err
}

func (tc *TestCluster) FindRangeLease(
	rangeDesc roachpb.RangeDescriptor, hint *roachpb.ReplicationTarget,
) (_ roachpb.Lease, now hlc.ClockTimestamp, _ error) {
	__antithesis_instrumentation__.Notify(646778)
	l, now, err := tc.FindRangeLeaseEx(context.TODO(), rangeDesc, hint)
	if err != nil {
		__antithesis_instrumentation__.Notify(646780)
		return roachpb.Lease{}, hlc.ClockTimestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(646781)
	}
	__antithesis_instrumentation__.Notify(646779)
	return l.CurrentOrProspective(), now, err
}

func (tc *TestCluster) FindRangeLeaseEx(
	ctx context.Context, rangeDesc roachpb.RangeDescriptor, hint *roachpb.ReplicationTarget,
) (_ server.LeaseInfo, now hlc.ClockTimestamp, _ error) {
	__antithesis_instrumentation__.Notify(646782)
	var queryPolicy server.LeaseInfoOpt
	if hint != nil {
		__antithesis_instrumentation__.Notify(646785)
		var ok bool
		if _, ok = rangeDesc.GetReplicaDescriptor(hint.StoreID); !ok {
			__antithesis_instrumentation__.Notify(646787)
			return server.LeaseInfo{}, hlc.ClockTimestamp{}, errors.Errorf(
				"bad hint: %+v; store doesn't have a replica of the range", hint)
		} else {
			__antithesis_instrumentation__.Notify(646788)
		}
		__antithesis_instrumentation__.Notify(646786)
		queryPolicy = server.QueryLocalNodeOnly
	} else {
		__antithesis_instrumentation__.Notify(646789)
		hint = &roachpb.ReplicationTarget{
			NodeID:  rangeDesc.Replicas().Descriptors()[0].NodeID,
			StoreID: rangeDesc.Replicas().Descriptors()[0].StoreID}
		queryPolicy = server.AllowQueryToBeForwardedToDifferentNode
	}
	__antithesis_instrumentation__.Notify(646783)

	hintServer, err := tc.FindMemberServer(hint.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(646790)
		return server.LeaseInfo{}, hlc.ClockTimestamp{}, errors.Wrapf(err, "bad hint: %+v; no such node", hint)
	} else {
		__antithesis_instrumentation__.Notify(646791)
	}
	__antithesis_instrumentation__.Notify(646784)

	return hintServer.GetRangeLease(ctx, rangeDesc.StartKey.AsRawKey(), queryPolicy)
}

func (tc *TestCluster) FindRangeLeaseHolder(
	rangeDesc roachpb.RangeDescriptor, hint *roachpb.ReplicationTarget,
) (roachpb.ReplicationTarget, error) {
	__antithesis_instrumentation__.Notify(646792)
	lease, now, err := tc.FindRangeLease(rangeDesc, hint)
	if err != nil {
		__antithesis_instrumentation__.Notify(646797)
		return roachpb.ReplicationTarget{}, err
	} else {
		__antithesis_instrumentation__.Notify(646798)
	}
	__antithesis_instrumentation__.Notify(646793)

	store, err := tc.findMemberStore(lease.Replica.StoreID)
	if err != nil {
		__antithesis_instrumentation__.Notify(646799)
		return roachpb.ReplicationTarget{}, err
	} else {
		__antithesis_instrumentation__.Notify(646800)
	}
	__antithesis_instrumentation__.Notify(646794)
	replica, err := store.GetReplica(rangeDesc.RangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(646801)
		return roachpb.ReplicationTarget{}, err
	} else {
		__antithesis_instrumentation__.Notify(646802)
	}
	__antithesis_instrumentation__.Notify(646795)
	if !replica.LeaseStatusAt(context.TODO(), now).IsValid() {
		__antithesis_instrumentation__.Notify(646803)
		return roachpb.ReplicationTarget{}, errors.New("no valid lease")
	} else {
		__antithesis_instrumentation__.Notify(646804)
	}
	__antithesis_instrumentation__.Notify(646796)
	replicaDesc := lease.Replica
	return roachpb.ReplicationTarget{NodeID: replicaDesc.NodeID, StoreID: replicaDesc.StoreID}, nil
}

func (tc *TestCluster) ScratchRange(t testing.TB) roachpb.Key {
	__antithesis_instrumentation__.Notify(646805)
	scratchKey, err := tc.Servers[0].ScratchRange()
	if err != nil {
		__antithesis_instrumentation__.Notify(646807)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646808)
	}
	__antithesis_instrumentation__.Notify(646806)
	return scratchKey
}

func (tc *TestCluster) ScratchRangeWithExpirationLease(t testing.TB) roachpb.Key {
	__antithesis_instrumentation__.Notify(646809)
	scratchKey, err := tc.Servers[0].ScratchRangeWithExpirationLease()
	if err != nil {
		__antithesis_instrumentation__.Notify(646811)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646812)
	}
	__antithesis_instrumentation__.Notify(646810)
	return scratchKey
}

func (tc *TestCluster) WaitForSplitAndInitialization(startKey roachpb.Key) error {
	__antithesis_instrumentation__.Notify(646813)
	return retry.ForDuration(testutils.DefaultSucceedsSoonDuration, func() error {
		__antithesis_instrumentation__.Notify(646814)
		desc, err := tc.LookupRange(startKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(646818)
			return errors.Wrapf(err, "unable to lookup range for %s", startKey)
		} else {
			__antithesis_instrumentation__.Notify(646819)
		}
		__antithesis_instrumentation__.Notify(646815)

		if !desc.StartKey.Equal(startKey) {
			__antithesis_instrumentation__.Notify(646820)
			return errors.Errorf("expected range start key %s; got %s",
				startKey, desc.StartKey)
		} else {
			__antithesis_instrumentation__.Notify(646821)
		}
		__antithesis_instrumentation__.Notify(646816)

		for _, rDesc := range desc.Replicas().Descriptors() {
			__antithesis_instrumentation__.Notify(646822)
			store, err := tc.findMemberStore(rDesc.StoreID)
			if err != nil {
				__antithesis_instrumentation__.Notify(646826)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646827)
			}
			__antithesis_instrumentation__.Notify(646823)
			repl, err := store.GetReplica(desc.RangeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(646828)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646829)
			}
			__antithesis_instrumentation__.Notify(646824)
			actualReplicaDesc, err := repl.GetReplicaDescriptor()
			if err != nil {
				__antithesis_instrumentation__.Notify(646830)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646831)
			}
			__antithesis_instrumentation__.Notify(646825)
			if !actualReplicaDesc.Equal(rDesc) {
				__antithesis_instrumentation__.Notify(646832)
				return errors.Errorf("expected replica %s; got %s", rDesc, actualReplicaDesc)
			} else {
				__antithesis_instrumentation__.Notify(646833)
			}
		}
		__antithesis_instrumentation__.Notify(646817)
		return nil
	})
}

func (tc *TestCluster) FindMemberServer(storeID roachpb.StoreID) (*server.TestServer, error) {
	__antithesis_instrumentation__.Notify(646834)
	for _, server := range tc.Servers {
		__antithesis_instrumentation__.Notify(646836)
		if server.Stores().HasStore(storeID) {
			__antithesis_instrumentation__.Notify(646837)
			return server, nil
		} else {
			__antithesis_instrumentation__.Notify(646838)
		}
	}
	__antithesis_instrumentation__.Notify(646835)
	return nil, errors.Errorf("store not found")
}

func (tc *TestCluster) findMemberStore(storeID roachpb.StoreID) (*kvserver.Store, error) {
	__antithesis_instrumentation__.Notify(646839)
	server, err := tc.FindMemberServer(storeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(646841)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(646842)
	}
	__antithesis_instrumentation__.Notify(646840)
	return server.Stores().GetStore(storeID)
}

func (tc *TestCluster) WaitForFullReplication() error {
	__antithesis_instrumentation__.Notify(646843)
	log.Infof(context.TODO(), "WaitForFullReplication")
	start := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(646847)
		end := timeutil.Now()
		log.Infof(context.TODO(), "WaitForFullReplication took: %s", end.Sub(start))
	}()
	__antithesis_instrumentation__.Notify(646844)

	if len(tc.Servers) < 3 {
		__antithesis_instrumentation__.Notify(646848)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(646849)
	}
	__antithesis_instrumentation__.Notify(646845)

	opts := retry.Options{
		InitialBackoff: time.Millisecond * 10,
		MaxBackoff:     time.Millisecond * 100,
		Multiplier:     2,
	}

	notReplicated := true
	for r := retry.Start(opts); r.Next() && func() bool {
		__antithesis_instrumentation__.Notify(646850)
		return notReplicated == true
	}() == true; {
		__antithesis_instrumentation__.Notify(646851)
		notReplicated = false
		for _, s := range tc.Servers {
			__antithesis_instrumentation__.Notify(646852)
			err := s.Stores().VisitStores(func(s *kvserver.Store) error {
				__antithesis_instrumentation__.Notify(646855)
				if n := s.ClusterNodeCount(); n != len(tc.Servers) {
					__antithesis_instrumentation__.Notify(646861)
					log.Infof(context.TODO(), "%s only sees %d/%d available nodes", s, n, len(tc.Servers))
					notReplicated = true
					return nil
				} else {
					__antithesis_instrumentation__.Notify(646862)
				}
				__antithesis_instrumentation__.Notify(646856)

				if err := s.ForceReplicationScanAndProcess(); err != nil {
					__antithesis_instrumentation__.Notify(646863)
					return err
				} else {
					__antithesis_instrumentation__.Notify(646864)
				}
				__antithesis_instrumentation__.Notify(646857)
				if err := s.ComputeMetrics(context.TODO(), 0); err != nil {
					__antithesis_instrumentation__.Notify(646865)

					log.Infof(context.TODO(), "%v", err)
					notReplicated = true
					return nil
				} else {
					__antithesis_instrumentation__.Notify(646866)
				}
				__antithesis_instrumentation__.Notify(646858)
				if n := s.Metrics().UnderReplicatedRangeCount.Value(); n > 0 {
					__antithesis_instrumentation__.Notify(646867)
					log.Infof(context.TODO(), "%s has %d underreplicated ranges", s, n)
					notReplicated = true
				} else {
					__antithesis_instrumentation__.Notify(646868)
				}
				__antithesis_instrumentation__.Notify(646859)
				if n := s.Metrics().OverReplicatedRangeCount.Value(); n > 0 {
					__antithesis_instrumentation__.Notify(646869)
					log.Infof(context.TODO(), "%s has %d overreplicated ranges", s, n)
					notReplicated = true
				} else {
					__antithesis_instrumentation__.Notify(646870)
				}
				__antithesis_instrumentation__.Notify(646860)
				return nil
			})
			__antithesis_instrumentation__.Notify(646853)
			if err != nil {
				__antithesis_instrumentation__.Notify(646871)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646872)
			}
			__antithesis_instrumentation__.Notify(646854)
			if notReplicated {
				__antithesis_instrumentation__.Notify(646873)
				break
			} else {
				__antithesis_instrumentation__.Notify(646874)
			}
		}
	}
	__antithesis_instrumentation__.Notify(646846)
	return nil
}

func (tc *TestCluster) WaitForNodeStatuses(t testing.TB) {
	__antithesis_instrumentation__.Notify(646875)
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(646876)
		client, err := tc.GetStatusClient(context.Background(), t, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(646882)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646883)
		}
		__antithesis_instrumentation__.Notify(646877)
		response, err := client.Nodes(context.Background(), &serverpb.NodesRequest{})
		if err != nil {
			__antithesis_instrumentation__.Notify(646884)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646885)
		}
		__antithesis_instrumentation__.Notify(646878)

		if len(response.Nodes) < tc.NumServers() {
			__antithesis_instrumentation__.Notify(646886)
			return fmt.Errorf("expected %d nodes registered, got %+v", tc.NumServers(), response)
		} else {
			__antithesis_instrumentation__.Notify(646887)
		}
		__antithesis_instrumentation__.Notify(646879)

		nodeIDs := make(map[roachpb.NodeID]bool)
		for _, node := range response.Nodes {
			__antithesis_instrumentation__.Notify(646888)
			if len(node.StoreStatuses) == 0 {
				__antithesis_instrumentation__.Notify(646890)
				return fmt.Errorf("missing StoreStatuses in NodeStatus: %+v", node)
			} else {
				__antithesis_instrumentation__.Notify(646891)
			}
			__antithesis_instrumentation__.Notify(646889)
			nodeIDs[node.Desc.NodeID] = true
		}
		__antithesis_instrumentation__.Notify(646880)
		for _, s := range tc.Servers {
			__antithesis_instrumentation__.Notify(646892)
			if id := s.GetNode().Descriptor.NodeID; !nodeIDs[id] {
				__antithesis_instrumentation__.Notify(646893)
				return fmt.Errorf("missing n%d in NodeStatus: %+v", id, response)
			} else {
				__antithesis_instrumentation__.Notify(646894)
			}
		}
		__antithesis_instrumentation__.Notify(646881)
		return nil
	})
}

func (tc *TestCluster) WaitForNodeLiveness(t testing.TB) {
	__antithesis_instrumentation__.Notify(646895)
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(646896)
		db := tc.Servers[0].DB()
		for _, s := range tc.Servers {
			__antithesis_instrumentation__.Notify(646898)
			key := keys.NodeLivenessKey(s.NodeID())
			var liveness livenesspb.Liveness
			if err := db.GetProto(context.Background(), key, &liveness); err != nil {
				__antithesis_instrumentation__.Notify(646901)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646902)
			}
			__antithesis_instrumentation__.Notify(646899)
			if (liveness == livenesspb.Liveness{}) {
				__antithesis_instrumentation__.Notify(646903)
				return fmt.Errorf("no liveness record")
			} else {
				__antithesis_instrumentation__.Notify(646904)
			}
			__antithesis_instrumentation__.Notify(646900)
			fmt.Printf("n%d: found liveness\n", s.NodeID())
		}
		__antithesis_instrumentation__.Notify(646897)
		return nil
	})
}

func (tc *TestCluster) ReplicationMode() base.TestClusterReplicationMode {
	__antithesis_instrumentation__.Notify(646905)
	return tc.clusterArgs.ReplicationMode
}

func (tc *TestCluster) ToggleReplicateQueues(active bool) {
	__antithesis_instrumentation__.Notify(646906)
	for _, s := range tc.Servers {
		__antithesis_instrumentation__.Notify(646907)
		_ = s.Stores().VisitStores(func(store *kvserver.Store) error {
			__antithesis_instrumentation__.Notify(646908)
			store.SetReplicateQueueActive(active)
			return nil
		})
	}
}

func (tc *TestCluster) ReadIntFromStores(key roachpb.Key) []int64 {
	__antithesis_instrumentation__.Notify(646909)
	results := make([]int64, len(tc.Servers))
	for i, server := range tc.Servers {
		__antithesis_instrumentation__.Notify(646911)
		err := server.Stores().VisitStores(func(s *kvserver.Store) error {
			__antithesis_instrumentation__.Notify(646913)
			val, _, err := storage.MVCCGet(context.Background(), s.Engine(), key,
				server.Clock().Now(), storage.MVCCGetOptions{})
			if err != nil {
				__antithesis_instrumentation__.Notify(646915)
				log.VEventf(context.Background(), 1, "store %d: error reading from key %s: %s", s.StoreID(), key, err)
			} else {
				__antithesis_instrumentation__.Notify(646916)
				if val == nil {
					__antithesis_instrumentation__.Notify(646917)
					log.VEventf(context.Background(), 1, "store %d: missing key %s", s.StoreID(), key)
				} else {
					__antithesis_instrumentation__.Notify(646918)
					results[i], err = val.GetInt()
					if err != nil {
						__antithesis_instrumentation__.Notify(646919)
						log.Errorf(context.Background(), "store %d: error decoding %s from key %s: %+v", s.StoreID(), val, key, err)
					} else {
						__antithesis_instrumentation__.Notify(646920)
					}
				}
			}
			__antithesis_instrumentation__.Notify(646914)
			return nil
		})
		__antithesis_instrumentation__.Notify(646912)
		if err != nil {
			__antithesis_instrumentation__.Notify(646921)
			log.VEventf(context.Background(), 1, "node %d: error reading from key %s: %s", server.NodeID(), key, err)
		} else {
			__antithesis_instrumentation__.Notify(646922)
		}
	}
	__antithesis_instrumentation__.Notify(646910)
	return results
}

func (tc *TestCluster) WaitForValues(t testing.TB, key roachpb.Key, expected []int64) {
	__antithesis_instrumentation__.Notify(646923)
	t.Helper()
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(646924)
		actual := tc.ReadIntFromStores(key)
		if !reflect.DeepEqual(expected, actual) {
			__antithesis_instrumentation__.Notify(646926)
			return errors.Errorf("expected %v, got %v", expected, actual)
		} else {
			__antithesis_instrumentation__.Notify(646927)
		}
		__antithesis_instrumentation__.Notify(646925)
		return nil
	})
}

func (tc *TestCluster) GetFirstStoreFromServer(t testing.TB, server int) *kvserver.Store {
	__antithesis_instrumentation__.Notify(646928)
	ts := tc.Servers[server]
	store, pErr := ts.Stores().GetStore(ts.GetFirstStoreID())
	if pErr != nil {
		__antithesis_instrumentation__.Notify(646930)
		t.Fatal(pErr)
	} else {
		__antithesis_instrumentation__.Notify(646931)
	}
	__antithesis_instrumentation__.Notify(646929)
	return store
}

func (tc *TestCluster) Restart() error {
	__antithesis_instrumentation__.Notify(646932)
	for i := range tc.Servers {
		__antithesis_instrumentation__.Notify(646934)
		tc.StopServer(i)
		if err := tc.RestartServer(i); err != nil {
			__antithesis_instrumentation__.Notify(646935)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646936)
		}
	}
	__antithesis_instrumentation__.Notify(646933)
	return nil
}

func (tc *TestCluster) RestartServer(idx int) error {
	__antithesis_instrumentation__.Notify(646937)
	return tc.RestartServerWithInspect(idx, nil)
}

func (tc *TestCluster) RestartServerWithInspect(idx int, inspect func(s *server.TestServer)) error {
	__antithesis_instrumentation__.Notify(646938)
	if !tc.ServerStopped(idx) {
		__antithesis_instrumentation__.Notify(646944)
		return errors.Errorf("server %d must be stopped before attempting to restart", idx)
	} else {
		__antithesis_instrumentation__.Notify(646945)
	}
	__antithesis_instrumentation__.Notify(646939)
	serverArgs := tc.serverArgs[idx]

	if idx == 0 {
		__antithesis_instrumentation__.Notify(646946)

		listener, err := net.Listen("tcp", serverArgs.Listener.Addr().String())
		if err != nil {
			__antithesis_instrumentation__.Notify(646948)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646949)
		}
		__antithesis_instrumentation__.Notify(646947)
		serverArgs.Listener = listener
		serverArgs.Knobs.Server.(*server.TestingKnobs).RPCListener = serverArgs.Listener
	} else {
		__antithesis_instrumentation__.Notify(646950)
		serverArgs.Addr = ""

		for i := range tc.Servers {
			__antithesis_instrumentation__.Notify(646951)
			if !tc.ServerStopped(i) {
				__antithesis_instrumentation__.Notify(646952)
				serverArgs.JoinAddr = tc.Servers[i].ServingRPCAddr()
			} else {
				__antithesis_instrumentation__.Notify(646953)
			}
		}
	}
	__antithesis_instrumentation__.Notify(646940)

	for i, specs := range serverArgs.StoreSpecs {
		__antithesis_instrumentation__.Notify(646954)
		if specs.InMemory && func() bool {
			__antithesis_instrumentation__.Notify(646955)
			return specs.StickyInMemoryEngineID == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(646956)
			return errors.Errorf("failed to restart Server %d, because a restart can only be used on a server with a sticky engine", i)
		} else {
			__antithesis_instrumentation__.Notify(646957)
		}
	}
	__antithesis_instrumentation__.Notify(646941)
	srv, err := serverutils.NewServer(serverArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(646958)
		return err
	} else {
		__antithesis_instrumentation__.Notify(646959)
	}
	__antithesis_instrumentation__.Notify(646942)
	s := srv.(*server.TestServer)

	ctx := context.Background()
	if err := func() error {
		__antithesis_instrumentation__.Notify(646960)
		func() {
			__antithesis_instrumentation__.Notify(646964)

			tc.mu.Lock()
			defer tc.mu.Unlock()
			tc.Servers[idx] = s
			tc.mu.serverStoppers[idx] = s.Stopper()

			if inspect != nil {
				__antithesis_instrumentation__.Notify(646965)
				inspect(s)
			} else {
				__antithesis_instrumentation__.Notify(646966)
			}
		}()
		__antithesis_instrumentation__.Notify(646961)

		if err := srv.Start(ctx); err != nil {
			__antithesis_instrumentation__.Notify(646967)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646968)
		}
		__antithesis_instrumentation__.Notify(646962)

		dbConn, err := serverutils.OpenDBConnE(srv.ServingSQLAddr(),
			serverArgs.UseDatabase, serverArgs.Insecure, srv.Stopper())
		if err != nil {
			__antithesis_instrumentation__.Notify(646969)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646970)
		}
		__antithesis_instrumentation__.Notify(646963)
		tc.Conns[idx] = dbConn
		return nil
	}(); err != nil {
		__antithesis_instrumentation__.Notify(646971)
		return err
	} else {
		__antithesis_instrumentation__.Notify(646972)
	}
	__antithesis_instrumentation__.Notify(646943)

	return contextutil.RunWithTimeout(
		ctx, "check-conn", 15*time.Second,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(646973)
			r := retry.StartWithCtx(ctx, retry.Options{
				InitialBackoff: 1 * time.Millisecond,
				MaxBackoff:     100 * time.Millisecond,
			})
			var err error
			for r.Next() {
				__antithesis_instrumentation__.Notify(646976)
				err = func() error {
					__antithesis_instrumentation__.Notify(646979)
					for idx, s := range tc.Servers {
						__antithesis_instrumentation__.Notify(646981)
						if tc.ServerStopped(idx) {
							__antithesis_instrumentation__.Notify(646983)
							continue
						} else {
							__antithesis_instrumentation__.Notify(646984)
						}
						__antithesis_instrumentation__.Notify(646982)
						for i := 0; i < rpc.NumConnectionClasses; i++ {
							__antithesis_instrumentation__.Notify(646985)
							class := rpc.ConnectionClass(i)
							if _, err := s.NodeDialer().(*nodedialer.Dialer).Dial(ctx, srv.NodeID(), class); err != nil {
								__antithesis_instrumentation__.Notify(646986)
								return errors.Wrapf(err, "connecting n%d->n%d (class %v)", s.NodeID(), srv.NodeID(), class)
							} else {
								__antithesis_instrumentation__.Notify(646987)
							}
						}
					}
					__antithesis_instrumentation__.Notify(646980)
					return nil
				}()
				__antithesis_instrumentation__.Notify(646977)
				if err != nil {
					__antithesis_instrumentation__.Notify(646988)
					continue
				} else {
					__antithesis_instrumentation__.Notify(646989)
				}
				__antithesis_instrumentation__.Notify(646978)
				return nil
			}
			__antithesis_instrumentation__.Notify(646974)
			if err != nil {
				__antithesis_instrumentation__.Notify(646990)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646991)
			}
			__antithesis_instrumentation__.Notify(646975)
			return ctx.Err()
		})
}

func (tc *TestCluster) ServerStopped(idx int) bool {
	__antithesis_instrumentation__.Notify(646992)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.serverStoppers[idx] == nil
}

func (tc *TestCluster) GetRaftLeader(t testing.TB, key roachpb.RKey) *kvserver.Replica {
	__antithesis_instrumentation__.Notify(646993)
	t.Helper()
	var raftLeaderRepl *kvserver.Replica
	testutils.SucceedsSoon(t, func() error {
		__antithesis_instrumentation__.Notify(646995)
		var latestTerm uint64
		for i := range tc.Servers {
			__antithesis_instrumentation__.Notify(646998)
			err := tc.Servers[i].Stores().VisitStores(func(store *kvserver.Store) error {
				__antithesis_instrumentation__.Notify(647000)
				repl := store.LookupReplica(key)
				if repl == nil {
					__antithesis_instrumentation__.Notify(647004)

					return nil
				} else {
					__antithesis_instrumentation__.Notify(647005)
				}
				__antithesis_instrumentation__.Notify(647001)
				raftStatus := repl.RaftStatus()
				if raftStatus == nil {
					__antithesis_instrumentation__.Notify(647006)
					return errors.Errorf("raft group is not initialized for replica with key %s", key)
				} else {
					__antithesis_instrumentation__.Notify(647007)
				}
				__antithesis_instrumentation__.Notify(647002)
				if raftStatus.Term > latestTerm || func() bool {
					__antithesis_instrumentation__.Notify(647008)
					return (raftLeaderRepl == nil && func() bool {
						__antithesis_instrumentation__.Notify(647009)
						return raftStatus.Term == latestTerm == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(647010)

					raftLeaderRepl = nil
					latestTerm = raftStatus.Term
					if raftStatus.RaftState == raft.StateLeader {
						__antithesis_instrumentation__.Notify(647011)
						raftLeaderRepl = repl
					} else {
						__antithesis_instrumentation__.Notify(647012)
					}
				} else {
					__antithesis_instrumentation__.Notify(647013)
				}
				__antithesis_instrumentation__.Notify(647003)
				return nil
			})
			__antithesis_instrumentation__.Notify(646999)
			if err != nil {
				__antithesis_instrumentation__.Notify(647014)
				return err
			} else {
				__antithesis_instrumentation__.Notify(647015)
			}
		}
		__antithesis_instrumentation__.Notify(646996)
		if latestTerm == 0 || func() bool {
			__antithesis_instrumentation__.Notify(647016)
			return raftLeaderRepl == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(647017)
			return errors.Errorf("could not find a raft leader for key %s", key)
		} else {
			__antithesis_instrumentation__.Notify(647018)
		}
		__antithesis_instrumentation__.Notify(646997)
		return nil
	})
	__antithesis_instrumentation__.Notify(646994)
	return raftLeaderRepl
}

func (tc *TestCluster) GetAdminClient(
	ctx context.Context, t testing.TB, serverIdx int,
) (serverpb.AdminClient, error) {
	__antithesis_instrumentation__.Notify(647019)
	srv := tc.Server(serverIdx)
	cc, err := srv.RPCContext().GRPCDialNode(srv.RPCAddr(), srv.NodeID(), rpc.DefaultClass).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(647021)
		return nil, errors.Wrap(err, "failed to create an admin client")
	} else {
		__antithesis_instrumentation__.Notify(647022)
	}
	__antithesis_instrumentation__.Notify(647020)
	return serverpb.NewAdminClient(cc), nil
}

func (tc *TestCluster) GetStatusClient(
	ctx context.Context, t testing.TB, serverIdx int,
) (serverpb.StatusClient, error) {
	__antithesis_instrumentation__.Notify(647023)
	srv := tc.Server(serverIdx)
	cc, err := srv.RPCContext().GRPCDialNode(srv.RPCAddr(), srv.NodeID(), rpc.DefaultClass).Connect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(647025)
		return nil, errors.Wrap(err, "failed to create a status client")
	} else {
		__antithesis_instrumentation__.Notify(647026)
	}
	__antithesis_instrumentation__.Notify(647024)
	return serverpb.NewStatusClient(cc), nil
}

type testClusterFactoryImpl struct{}

var TestClusterFactory serverutils.TestClusterFactory = testClusterFactoryImpl{}

func (testClusterFactoryImpl) NewTestCluster(
	t testing.TB, numNodes int, args base.TestClusterArgs,
) serverutils.TestClusterInterface {
	__antithesis_instrumentation__.Notify(647027)
	return NewTestCluster(t, numNodes, args)
}
