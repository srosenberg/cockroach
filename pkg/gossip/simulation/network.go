package simulation

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"google.golang.org/grpc"
)

type Node struct {
	Gossip    *gossip.Gossip
	Server    *grpc.Server
	Listener  net.Listener
	Registry  *metric.Registry
	Addresses []util.UnresolvedAddr
}

func (n *Node) Addr() net.Addr {
	__antithesis_instrumentation__.Notify(68115)
	return n.Listener.Addr()
}

type Network struct {
	Nodes           []*Node
	Stopper         *stop.Stopper
	RPCContext      *rpc.Context
	nodeIDAllocator roachpb.NodeID
	tlsConfig       *tls.Config
	started         bool
}

func NewNetwork(
	stopper *stop.Stopper, nodeCount int, createAddresses bool, defaultZoneConfig *zonepb.ZoneConfig,
) *Network {
	__antithesis_instrumentation__.Notify(68116)
	ctx := context.TODO()
	log.Infof(ctx, "simulating gossip network with %d nodes", nodeCount)

	n := &Network{
		Nodes:   []*Node{},
		Stopper: stopper,
	}
	n.RPCContext = rpc.NewContext(ctx,
		rpc.ContextOptions{
			TenantID: roachpb.SystemTenantID,
			Config:   &base.Config{Insecure: true},
			Clock:    hlc.NewClock(hlc.UnixNano, time.Nanosecond),
			Stopper:  n.Stopper,
			Settings: cluster.MakeTestingClusterSettings(),
		})
	var err error
	n.tlsConfig, err = n.RPCContext.GetServerTLSConfig()
	if err != nil {
		__antithesis_instrumentation__.Notify(68119)
		log.Fatalf(context.TODO(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(68120)
	}
	__antithesis_instrumentation__.Notify(68117)

	n.RPCContext.StorageClusterID.Set(context.TODO(), uuid.MakeV4())

	for i := 0; i < nodeCount; i++ {
		__antithesis_instrumentation__.Notify(68121)
		node, err := n.CreateNode(defaultZoneConfig)
		if err != nil {
			__antithesis_instrumentation__.Notify(68123)
			log.Fatalf(context.TODO(), "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(68124)
		}
		__antithesis_instrumentation__.Notify(68122)
		if createAddresses {
			__antithesis_instrumentation__.Notify(68125)
			node.Addresses = []util.UnresolvedAddr{
				util.MakeUnresolvedAddr("tcp", n.Nodes[0].Addr().String()),
			}
		} else {
			__antithesis_instrumentation__.Notify(68126)
		}
	}
	__antithesis_instrumentation__.Notify(68118)
	return n
}

func (n *Network) CreateNode(defaultZoneConfig *zonepb.ZoneConfig) (*Node, error) {
	__antithesis_instrumentation__.Notify(68127)
	server := rpc.NewServer(n.RPCContext)
	ln, err := net.Listen(util.IsolatedTestAddr.Network(), util.IsolatedTestAddr.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(68130)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(68131)
	}
	__antithesis_instrumentation__.Notify(68128)
	node := &Node{Server: server, Listener: ln, Registry: metric.NewRegistry()}
	node.Gossip = gossip.NewTest(0, n.RPCContext, server, n.Stopper, node.Registry, defaultZoneConfig)
	n.Stopper.AddCloser(stop.CloserFn(server.Stop))
	_ = n.Stopper.RunAsyncTask(context.TODO(), "node-wait-quiesce", func(context.Context) {
		__antithesis_instrumentation__.Notify(68132)
		<-n.Stopper.ShouldQuiesce()
		netutil.FatalIfUnexpected(ln.Close())
		node.Gossip.EnableSimulationCycler(false)
	})
	__antithesis_instrumentation__.Notify(68129)
	n.Nodes = append(n.Nodes, node)
	return node, nil
}

func (n *Network) StartNode(node *Node) error {
	__antithesis_instrumentation__.Notify(68133)
	node.Gossip.Start(node.Addr(), node.Addresses)
	node.Gossip.EnableSimulationCycler(true)
	n.nodeIDAllocator++
	node.Gossip.NodeID.Set(context.TODO(), n.nodeIDAllocator)
	if err := node.Gossip.SetNodeDescriptor(&roachpb.NodeDescriptor{
		NodeID:  node.Gossip.NodeID.Get(),
		Address: util.MakeUnresolvedAddr(node.Addr().Network(), node.Addr().String()),
	}); err != nil {
		__antithesis_instrumentation__.Notify(68136)
		return err
	} else {
		__antithesis_instrumentation__.Notify(68137)
	}
	__antithesis_instrumentation__.Notify(68134)
	if err := node.Gossip.AddInfo(node.Addr().String(),
		encoding.EncodeUint64Ascending(nil, 0), time.Hour); err != nil {
		__antithesis_instrumentation__.Notify(68138)
		return err
	} else {
		__antithesis_instrumentation__.Notify(68139)
	}
	__antithesis_instrumentation__.Notify(68135)
	bgCtx := context.TODO()
	return n.Stopper.RunAsyncTask(bgCtx, "start-node", func(context.Context) {
		__antithesis_instrumentation__.Notify(68140)
		netutil.FatalIfUnexpected(node.Server.Serve(node.Listener))
	})
}

func (n *Network) GetNodeFromID(nodeID roachpb.NodeID) (*Node, bool) {
	__antithesis_instrumentation__.Notify(68141)
	for _, node := range n.Nodes {
		__antithesis_instrumentation__.Notify(68143)
		if node.Gossip.NodeID.Get() == nodeID {
			__antithesis_instrumentation__.Notify(68144)
			return node, true
		} else {
			__antithesis_instrumentation__.Notify(68145)
		}
	}
	__antithesis_instrumentation__.Notify(68142)
	return nil, false
}

func (n *Network) SimulateNetwork(simCallback func(cycle int, network *Network) bool) {
	__antithesis_instrumentation__.Notify(68146)
	n.Start()
	nodes := n.Nodes
	for cycle := 1; ; cycle++ {
		__antithesis_instrumentation__.Notify(68148)

		if err := nodes[0].Gossip.AddInfo(
			gossip.KeySentinel,
			encoding.EncodeUint64Ascending(nil, uint64(cycle)),
			time.Hour,
		); err != nil {
			__antithesis_instrumentation__.Notify(68153)
			log.Fatalf(context.TODO(), "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(68154)
		}
		__antithesis_instrumentation__.Notify(68149)
		if err := nodes[0].Gossip.AddInfo(
			gossip.KeyClusterID,
			encoding.EncodeUint64Ascending(nil, uint64(cycle)),
			0*time.Second,
		); err != nil {
			__antithesis_instrumentation__.Notify(68155)
			log.Fatalf(context.TODO(), "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(68156)
		}
		__antithesis_instrumentation__.Notify(68150)

		for _, node := range nodes {
			__antithesis_instrumentation__.Notify(68157)
			if err := node.Gossip.AddInfo(
				node.Addr().String(),
				encoding.EncodeUint64Ascending(nil, uint64(cycle)),
				time.Hour,
			); err != nil {
				__antithesis_instrumentation__.Notify(68159)
				log.Fatalf(context.TODO(), "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(68160)
			}
			__antithesis_instrumentation__.Notify(68158)
			node.Gossip.SimulationCycle()
		}
		__antithesis_instrumentation__.Notify(68151)

		if !simCallback(cycle, n) {
			__antithesis_instrumentation__.Notify(68161)
			break
		} else {
			__antithesis_instrumentation__.Notify(68162)
		}
		__antithesis_instrumentation__.Notify(68152)
		time.Sleep(5 * time.Millisecond)
	}
	__antithesis_instrumentation__.Notify(68147)
	log.Infof(context.TODO(), "gossip network simulation: total infos sent=%d, received=%d", n.infosSent(), n.infosReceived())
}

func (n *Network) Start() {
	__antithesis_instrumentation__.Notify(68163)
	if n.started {
		__antithesis_instrumentation__.Notify(68165)
		return
	} else {
		__antithesis_instrumentation__.Notify(68166)
	}
	__antithesis_instrumentation__.Notify(68164)
	n.started = true
	for _, node := range n.Nodes {
		__antithesis_instrumentation__.Notify(68167)
		if err := n.StartNode(node); err != nil {
			__antithesis_instrumentation__.Notify(68168)
			log.Fatalf(context.TODO(), "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(68169)
		}
	}
}

func (n *Network) RunUntilFullyConnected() int {
	__antithesis_instrumentation__.Notify(68170)
	var connectedAtCycle int
	n.SimulateNetwork(func(cycle int, network *Network) bool {
		__antithesis_instrumentation__.Notify(68172)
		if network.IsNetworkConnected() {
			__antithesis_instrumentation__.Notify(68174)
			connectedAtCycle = cycle
			return false
		} else {
			__antithesis_instrumentation__.Notify(68175)
		}
		__antithesis_instrumentation__.Notify(68173)
		return true
	})
	__antithesis_instrumentation__.Notify(68171)
	return connectedAtCycle
}

func (n *Network) IsNetworkConnected() bool {
	__antithesis_instrumentation__.Notify(68176)
	for _, leftNode := range n.Nodes {
		__antithesis_instrumentation__.Notify(68178)
		for _, rightNode := range n.Nodes {
			__antithesis_instrumentation__.Notify(68179)
			if _, err := leftNode.Gossip.GetInfo(gossip.MakeNodeIDKey(rightNode.Gossip.NodeID.Get())); err != nil {
				__antithesis_instrumentation__.Notify(68180)
				return false
			} else {
				__antithesis_instrumentation__.Notify(68181)
			}
		}
	}
	__antithesis_instrumentation__.Notify(68177)
	return true
}

func (n *Network) infosSent() int {
	__antithesis_instrumentation__.Notify(68182)
	var count int64
	for _, node := range n.Nodes {
		__antithesis_instrumentation__.Notify(68184)
		count += node.Gossip.GetNodeMetrics().InfosSent.Counter.Count()
	}
	__antithesis_instrumentation__.Notify(68183)
	return int(count)
}

func (n *Network) infosReceived() int {
	__antithesis_instrumentation__.Notify(68185)
	var count int64
	for _, node := range n.Nodes {
		__antithesis_instrumentation__.Notify(68187)
		count += node.Gossip.GetNodeMetrics().InfosReceived.Counter.Count()
	}
	__antithesis_instrumentation__.Notify(68186)
	return int(count)
}
