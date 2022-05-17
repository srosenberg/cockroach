package gossip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	circuit "github.com/cockroachdb/circuitbreaker"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
	"google.golang.org/grpc"
)

const (
	maxHops = 5

	minPeers = 3

	defaultStallInterval = 2 * time.Second

	defaultBootstrapInterval = 1 * time.Second

	defaultCullInterval = 60 * time.Second

	defaultClientsInterval = 2 * time.Second

	NodeDescriptorInterval = 1 * time.Hour

	NodeDescriptorTTL = 2 * NodeDescriptorInterval

	StoresInterval = 60 * time.Second

	StoreTTL = 2 * StoresInterval

	unknownNodeID roachpb.NodeID = 0
)

var (
	MetaConnectionsIncomingGauge = metric.Metadata{
		Name:        "gossip.connections.incoming",
		Help:        "Number of active incoming gossip connections",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	MetaConnectionsOutgoingGauge = metric.Metadata{
		Name:        "gossip.connections.outgoing",
		Help:        "Number of active outgoing gossip connections",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	MetaConnectionsRefused = metric.Metadata{
		Name:        "gossip.connections.refused",
		Help:        "Number of refused incoming gossip connections",
		Measurement: "Connections",
		Unit:        metric.Unit_COUNT,
	}
	MetaInfosSent = metric.Metadata{
		Name:        "gossip.infos.sent",
		Help:        "Number of sent gossip Info objects",
		Measurement: "Infos",
		Unit:        metric.Unit_COUNT,
	}
	MetaInfosReceived = metric.Metadata{
		Name:        "gossip.infos.received",
		Help:        "Number of received gossip Info objects",
		Measurement: "Infos",
		Unit:        metric.Unit_COUNT,
	}
	MetaBytesSent = metric.Metadata{
		Name:        "gossip.bytes.sent",
		Help:        "Number of sent gossip bytes",
		Measurement: "Gossip Bytes",
		Unit:        metric.Unit_BYTES,
	}
	MetaBytesReceived = metric.Metadata{
		Name:        "gossip.bytes.received",
		Help:        "Number of received gossip bytes",
		Measurement: "Gossip Bytes",
		Unit:        metric.Unit_BYTES,
	}
)

type KeyNotPresentError struct {
	key string
}

func (err KeyNotPresentError) Error() string {
	__antithesis_instrumentation__.Notify(65198)
	return fmt.Sprintf("KeyNotPresentError: gossip key %q does not exist or has expired", err.key)
}

func NewKeyNotPresentError(key string) error {
	__antithesis_instrumentation__.Notify(65199)
	return KeyNotPresentError{key: key}
}

func AddressResolver(gossip *Gossip) nodedialer.AddressResolver {
	__antithesis_instrumentation__.Notify(65200)
	return func(nodeID roachpb.NodeID) (net.Addr, error) {
		__antithesis_instrumentation__.Notify(65201)
		return gossip.GetNodeIDAddress(nodeID)
	}
}

type Storage interface {
	ReadBootstrapInfo(*BootstrapInfo) error

	WriteBootstrapInfo(*BootstrapInfo) error
}

type Gossip struct {
	started bool

	*server

	Connected     chan struct{}
	hasConnected  bool
	rpcContext    *rpc.Context
	outgoing      nodeSet
	storage       Storage
	bootstrapInfo BootstrapInfo
	bootstrapping map[string]struct{}
	hasCleanedBS  bool

	clientsMu struct {
		syncutil.Mutex
		clients []*client

		breakers map[string]*circuit.Breaker
	}

	disconnected chan *client
	stalled      bool
	stalledCh    chan struct{}

	stallInterval     time.Duration
	bootstrapInterval time.Duration
	cullInterval      time.Duration

	systemConfig         *config.SystemConfig
	systemConfigMu       syncutil.RWMutex
	systemConfigChannels []chan<- struct{}

	addressIdx     int
	addresses      []util.UnresolvedAddr
	addressesTried map[int]struct{}
	nodeDescs      map[roachpb.NodeID]*roachpb.NodeDescriptor

	storeMap map[roachpb.StoreID]roachpb.NodeID

	addressExists  map[util.UnresolvedAddr]bool
	bootstrapAddrs map[util.UnresolvedAddr]roachpb.NodeID

	locality roachpb.Locality

	lastConnectivity redact.RedactableString

	defaultZoneConfig *zonepb.ZoneConfig
}

func New(
	ambient log.AmbientContext,
	clusterID *base.ClusterIDContainer,
	nodeID *base.NodeIDContainer,
	rpcContext *rpc.Context,
	grpcServer *grpc.Server,
	stopper *stop.Stopper,
	registry *metric.Registry,
	locality roachpb.Locality,
	defaultZoneConfig *zonepb.ZoneConfig,
) *Gossip {
	__antithesis_instrumentation__.Notify(65202)
	ambient.SetEventLog("gossip", "gossip")
	g := &Gossip{
		server:            newServer(ambient, clusterID, nodeID, stopper, registry),
		Connected:         make(chan struct{}),
		rpcContext:        rpcContext,
		outgoing:          makeNodeSet(minPeers, metric.NewGauge(MetaConnectionsOutgoingGauge)),
		bootstrapping:     map[string]struct{}{},
		disconnected:      make(chan *client, 10),
		stalledCh:         make(chan struct{}, 1),
		stallInterval:     defaultStallInterval,
		bootstrapInterval: defaultBootstrapInterval,
		cullInterval:      defaultCullInterval,
		addressesTried:    map[int]struct{}{},
		nodeDescs:         map[roachpb.NodeID]*roachpb.NodeDescriptor{},
		storeMap:          make(map[roachpb.StoreID]roachpb.NodeID),
		addressExists:     map[util.UnresolvedAddr]bool{},
		bootstrapAddrs:    map[util.UnresolvedAddr]roachpb.NodeID{},
		locality:          locality,
		defaultZoneConfig: defaultZoneConfig,
	}

	stopper.AddCloser(stop.CloserFn(g.server.AmbientContext.FinishEventLog))

	registry.AddMetric(g.outgoing.gauge)
	g.clientsMu.breakers = map[string]*circuit.Breaker{}

	g.mu.Lock()

	g.mu.is.registerCallback(KeyDeprecatedSystemConfig, g.updateSystemConfig)

	g.mu.is.registerCallback(MakePrefixPattern(KeyNodeIDPrefix), g.updateNodeAddress)
	g.mu.is.registerCallback(MakePrefixPattern(KeyStorePrefix), g.updateStoreMap)

	g.mu.Unlock()

	if grpcServer != nil {
		__antithesis_instrumentation__.Notify(65204)
		RegisterGossipServer(grpcServer, g.server)
	} else {
		__antithesis_instrumentation__.Notify(65205)
	}
	__antithesis_instrumentation__.Notify(65203)
	return g
}

func NewTest(
	nodeID roachpb.NodeID,
	rpcContext *rpc.Context,
	grpcServer *grpc.Server,
	stopper *stop.Stopper,
	registry *metric.Registry,
	defaultZoneConfig *zonepb.ZoneConfig,
) *Gossip {
	__antithesis_instrumentation__.Notify(65206)
	return NewTestWithLocality(nodeID, rpcContext, grpcServer, stopper, registry, roachpb.Locality{}, defaultZoneConfig)
}

func NewTestWithLocality(
	nodeID roachpb.NodeID,
	rpcContext *rpc.Context,
	grpcServer *grpc.Server,
	stopper *stop.Stopper,
	registry *metric.Registry,
	locality roachpb.Locality,
	defaultZoneConfig *zonepb.ZoneConfig,
) *Gossip {
	__antithesis_instrumentation__.Notify(65207)
	c := &base.ClusterIDContainer{}
	n := &base.NodeIDContainer{}
	var ac log.AmbientContext
	ac.AddLogTag("n", n)
	gossip := New(ac, c, n, rpcContext, grpcServer, stopper, registry, locality, defaultZoneConfig)
	if nodeID != 0 {
		__antithesis_instrumentation__.Notify(65209)
		n.Set(context.TODO(), nodeID)
	} else {
		__antithesis_instrumentation__.Notify(65210)
	}
	__antithesis_instrumentation__.Notify(65208)
	return gossip
}

func (g *Gossip) AssertNotStarted(ctx context.Context) {
	__antithesis_instrumentation__.Notify(65211)
	if g.started {
		__antithesis_instrumentation__.Notify(65212)
		log.Fatalf(ctx, "gossip instance was already started")
	} else {
		__antithesis_instrumentation__.Notify(65213)
	}
}

func (g *Gossip) GetNodeMetrics() *Metrics {
	__antithesis_instrumentation__.Notify(65214)
	return g.server.GetNodeMetrics()
}

func (g *Gossip) SetNodeDescriptor(desc *roachpb.NodeDescriptor) error {
	__antithesis_instrumentation__.Notify(65215)
	ctx := g.AnnotateCtx(context.TODO())
	log.Infof(ctx, "NodeDescriptor set to %+v", desc)
	if desc.Address.IsEmpty() {
		__antithesis_instrumentation__.Notify(65218)
		log.Fatalf(ctx, "n%d address is empty", desc.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(65219)
	}
	__antithesis_instrumentation__.Notify(65216)
	if err := g.AddInfoProto(MakeNodeIDKey(desc.NodeID), desc, NodeDescriptorTTL); err != nil {
		__antithesis_instrumentation__.Notify(65220)
		return errors.Wrapf(err, "n%d: couldn't gossip descriptor", desc.NodeID)
	} else {
		__antithesis_instrumentation__.Notify(65221)
	}
	__antithesis_instrumentation__.Notify(65217)
	g.updateClients()
	return nil
}

func (g *Gossip) SetStallInterval(interval time.Duration) {
	__antithesis_instrumentation__.Notify(65222)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.stallInterval = interval
}

func (g *Gossip) SetBootstrapInterval(interval time.Duration) {
	__antithesis_instrumentation__.Notify(65223)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.bootstrapInterval = interval
}

func (g *Gossip) SetCullInterval(interval time.Duration) {
	__antithesis_instrumentation__.Notify(65224)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cullInterval = interval
}

func (g *Gossip) SetStorage(storage Storage) error {
	__antithesis_instrumentation__.Notify(65225)
	ctx := g.AnnotateCtx(context.TODO())

	var storedBI BootstrapInfo
	if err := storage.ReadBootstrapInfo(&storedBI); err != nil {
		__antithesis_instrumentation__.Notify(65233)
		log.Ops.Warningf(ctx, "failed to read gossip bootstrap info: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(65234)
	}
	__antithesis_instrumentation__.Notify(65226)

	g.mu.Lock()
	defer g.mu.Unlock()
	g.storage = storage

	existing := map[string]struct{}{}
	makeKey := func(a util.UnresolvedAddr) string {
		__antithesis_instrumentation__.Notify(65235)
		return fmt.Sprintf("%s,%s", a.Network(), a.String())
	}
	__antithesis_instrumentation__.Notify(65227)
	for _, addr := range g.bootstrapInfo.Addresses {
		__antithesis_instrumentation__.Notify(65236)
		existing[makeKey(addr)] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(65228)
	for _, addr := range storedBI.Addresses {
		__antithesis_instrumentation__.Notify(65237)

		if _, ok := existing[makeKey(addr)]; !ok && func() bool {
			__antithesis_instrumentation__.Notify(65238)
			return addr != g.mu.is.NodeAddr == true
		}() == true {
			__antithesis_instrumentation__.Notify(65239)
			g.maybeAddBootstrapAddressLocked(addr, unknownNodeID)
		} else {
			__antithesis_instrumentation__.Notify(65240)
		}
	}
	__antithesis_instrumentation__.Notify(65229)

	if numAddrs := len(g.bootstrapInfo.Addresses); numAddrs > len(storedBI.Addresses) {
		__antithesis_instrumentation__.Notify(65241)
		if err := g.storage.WriteBootstrapInfo(&g.bootstrapInfo); err != nil {
			__antithesis_instrumentation__.Notify(65242)
			log.Errorf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(65243)
		}
	} else {
		__antithesis_instrumentation__.Notify(65244)
	}
	__antithesis_instrumentation__.Notify(65230)

	newAddressFound := false
	for _, addr := range g.bootstrapInfo.Addresses {
		__antithesis_instrumentation__.Notify(65245)
		if !g.maybeAddAddressLocked(addr) {
			__antithesis_instrumentation__.Notify(65247)
			continue
		} else {
			__antithesis_instrumentation__.Notify(65248)
		}
		__antithesis_instrumentation__.Notify(65246)

		if !newAddressFound {
			__antithesis_instrumentation__.Notify(65249)
			newAddressFound = true
			g.addressIdx = len(g.addresses) - 1
		} else {
			__antithesis_instrumentation__.Notify(65250)
		}
	}
	__antithesis_instrumentation__.Notify(65231)

	if newAddressFound {
		__antithesis_instrumentation__.Notify(65251)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(65253)
			log.Ops.Infof(ctx, "found new addresses from storage; signaling bootstrap")
		} else {
			__antithesis_instrumentation__.Notify(65254)
		}
		__antithesis_instrumentation__.Notify(65252)
		g.signalStalledLocked()
	} else {
		__antithesis_instrumentation__.Notify(65255)
	}
	__antithesis_instrumentation__.Notify(65232)
	return nil
}

func (g *Gossip) setAddresses(addresses []util.UnresolvedAddr) {
	__antithesis_instrumentation__.Notify(65256)
	if addresses == nil {
		__antithesis_instrumentation__.Notify(65258)
		return
	} else {
		__antithesis_instrumentation__.Notify(65259)
	}
	__antithesis_instrumentation__.Notify(65257)

	g.mu.Lock()
	defer g.mu.Unlock()

	g.addressIdx = len(addresses) - 1
	g.addresses = addresses
	g.addressesTried = map[int]struct{}{}

	g.maybeSignalStatusChangeLocked()
}

func (g *Gossip) GetAddresses() []util.UnresolvedAddr {
	__antithesis_instrumentation__.Notify(65260)
	g.mu.RLock()
	defer g.mu.RUnlock()
	return append([]util.UnresolvedAddr(nil), g.addresses...)
}

func (g *Gossip) GetNodeIDAddress(nodeID roachpb.NodeID) (*util.UnresolvedAddr, error) {
	__antithesis_instrumentation__.Notify(65261)
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getNodeIDAddressLocked(nodeID)
}

func (g *Gossip) GetNodeIDSQLAddress(nodeID roachpb.NodeID) (*util.UnresolvedAddr, error) {
	__antithesis_instrumentation__.Notify(65262)
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getNodeIDSQLAddressLocked(nodeID)
}

func (g *Gossip) GetNodeIDHTTPAddress(nodeID roachpb.NodeID) (*util.UnresolvedAddr, error) {
	__antithesis_instrumentation__.Notify(65263)
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getNodeIDHTTPAddressLocked(nodeID)
}

func (g *Gossip) GetNodeDescriptor(nodeID roachpb.NodeID) (*roachpb.NodeDescriptor, error) {
	__antithesis_instrumentation__.Notify(65264)
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.getNodeDescriptorLocked(nodeID)
}

func (g *Gossip) LogStatus() {
	__antithesis_instrumentation__.Notify(65265)
	g.mu.RLock()
	n := len(g.nodeDescs)
	status := redact.SafeString("ok")
	if g.mu.is.getInfo(KeySentinel) == nil {
		__antithesis_instrumentation__.Notify(65268)
		status = redact.SafeString("stalled")
	} else {
		__antithesis_instrumentation__.Notify(65269)
	}
	__antithesis_instrumentation__.Notify(65266)
	g.mu.RUnlock()

	var connectivity redact.RedactableString
	if s := redact.Sprint(g.Connectivity()); s != g.lastConnectivity {
		__antithesis_instrumentation__.Notify(65270)
		g.lastConnectivity = s
		connectivity = s
	} else {
		__antithesis_instrumentation__.Notify(65271)
	}
	__antithesis_instrumentation__.Notify(65267)

	ctx := g.AnnotateCtx(context.TODO())
	log.Health.Infof(ctx, "gossip status (%s, %d node%s)\n%s%s%s",
		status, n, util.Pluralize(int64(n)),
		g.clientStatus(), g.server.status(),
		connectivity)
}

func (g *Gossip) clientStatus() ClientStatus {
	__antithesis_instrumentation__.Notify(65272)
	g.mu.RLock()
	defer g.mu.RUnlock()
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()

	var status ClientStatus

	status.MaxConns = int32(g.outgoing.maxSize)
	status.ConnStatus = make([]OutgoingConnStatus, 0, len(g.clientsMu.clients))
	for _, c := range g.clientsMu.clients {
		__antithesis_instrumentation__.Notify(65274)
		status.ConnStatus = append(status.ConnStatus, OutgoingConnStatus{
			ConnStatus: ConnStatus{
				NodeID:   c.peerID,
				Address:  c.addr.String(),
				AgeNanos: timeutil.Since(c.createdAt).Nanoseconds(),
			},
			MetricSnap: c.clientMetrics.Snapshot(),
		})
	}
	__antithesis_instrumentation__.Notify(65273)
	return status
}

func (g *Gossip) Connectivity() Connectivity {
	__antithesis_instrumentation__.Notify(65275)
	ctx := g.AnnotateCtx(context.TODO())
	var c Connectivity

	g.mu.RLock()

	if i := g.mu.is.getInfo(KeySentinel); i != nil {
		__antithesis_instrumentation__.Notify(65279)
		c.SentinelNodeID = i.NodeID
	} else {
		__antithesis_instrumentation__.Notify(65280)
	}
	__antithesis_instrumentation__.Notify(65276)

	for nodeID := range g.nodeDescs {
		__antithesis_instrumentation__.Notify(65281)
		i := g.mu.is.getInfo(MakeGossipClientsKey(nodeID))
		if i == nil {
			__antithesis_instrumentation__.Notify(65285)
			continue
		} else {
			__antithesis_instrumentation__.Notify(65286)
		}
		__antithesis_instrumentation__.Notify(65282)

		v, err := i.Value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(65287)
			log.Errorf(ctx, "unable to retrieve gossip value for %s: %v",
				MakeGossipClientsKey(nodeID), err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(65288)
		}
		__antithesis_instrumentation__.Notify(65283)
		if len(v) == 0 {
			__antithesis_instrumentation__.Notify(65289)
			continue
		} else {
			__antithesis_instrumentation__.Notify(65290)
		}
		__antithesis_instrumentation__.Notify(65284)

		for _, part := range strings.Split(string(v), ",") {
			__antithesis_instrumentation__.Notify(65291)
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(65293)
				log.Errorf(ctx, "unable to parse node ID: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(65294)
			}
			__antithesis_instrumentation__.Notify(65292)
			c.ClientConns = append(c.ClientConns, Connectivity_Conn{
				SourceID: nodeID,
				TargetID: roachpb.NodeID(id),
			})
		}
	}
	__antithesis_instrumentation__.Notify(65277)

	g.mu.RUnlock()

	sort.Slice(c.ClientConns, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(65295)
		a, b := &c.ClientConns[i], &c.ClientConns[j]
		if a.SourceID < b.SourceID {
			__antithesis_instrumentation__.Notify(65298)
			return true
		} else {
			__antithesis_instrumentation__.Notify(65299)
		}
		__antithesis_instrumentation__.Notify(65296)
		if a.SourceID > b.SourceID {
			__antithesis_instrumentation__.Notify(65300)
			return false
		} else {
			__antithesis_instrumentation__.Notify(65301)
		}
		__antithesis_instrumentation__.Notify(65297)
		return a.TargetID < b.TargetID
	})
	__antithesis_instrumentation__.Notify(65278)

	return c
}

func (g *Gossip) EnableSimulationCycler(enable bool) {
	__antithesis_instrumentation__.Notify(65302)
	g.mu.Lock()
	defer g.mu.Unlock()
	if enable {
		__antithesis_instrumentation__.Notify(65303)
		g.simulationCycler = sync.NewCond(&g.mu)
	} else {
		__antithesis_instrumentation__.Notify(65304)

		if g.simulationCycler != nil {
			__antithesis_instrumentation__.Notify(65305)
			g.simulationCycler.Broadcast()
			g.simulationCycler = nil
		} else {
			__antithesis_instrumentation__.Notify(65306)
		}
	}
}

func (g *Gossip) SimulationCycle() {
	__antithesis_instrumentation__.Notify(65307)
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.simulationCycler != nil {
		__antithesis_instrumentation__.Notify(65308)
		g.simulationCycler.Broadcast()
	} else {
		__antithesis_instrumentation__.Notify(65309)
	}
}

func (g *Gossip) maybeAddAddressLocked(addr util.UnresolvedAddr) bool {
	__antithesis_instrumentation__.Notify(65310)
	if g.addressExists[addr] {
		__antithesis_instrumentation__.Notify(65313)
		return false
	} else {
		__antithesis_instrumentation__.Notify(65314)
	}
	__antithesis_instrumentation__.Notify(65311)
	ctx := g.AnnotateCtx(context.TODO())
	if addr.Network() != "tcp" {
		__antithesis_instrumentation__.Notify(65315)
		log.Ops.Warningf(ctx, "unknown address network %q for %v", addr.Network(), addr)
		return false
	} else {
		__antithesis_instrumentation__.Notify(65316)
	}
	__antithesis_instrumentation__.Notify(65312)
	g.addresses = append(g.addresses, addr)
	g.addressExists[addr] = true
	log.Eventf(ctx, "add address %s", addr)
	return true
}

func (g *Gossip) maybeAddBootstrapAddressLocked(
	addr util.UnresolvedAddr, nodeID roachpb.NodeID,
) bool {
	__antithesis_instrumentation__.Notify(65317)
	if existingNodeID, ok := g.bootstrapAddrs[addr]; ok {
		__antithesis_instrumentation__.Notify(65319)
		if existingNodeID == unknownNodeID || func() bool {
			__antithesis_instrumentation__.Notify(65321)
			return existingNodeID != nodeID == true
		}() == true {
			__antithesis_instrumentation__.Notify(65322)
			g.bootstrapAddrs[addr] = nodeID
		} else {
			__antithesis_instrumentation__.Notify(65323)
		}
		__antithesis_instrumentation__.Notify(65320)
		return false
	} else {
		__antithesis_instrumentation__.Notify(65324)
	}
	__antithesis_instrumentation__.Notify(65318)
	g.bootstrapInfo.Addresses = append(g.bootstrapInfo.Addresses, addr)
	g.bootstrapAddrs[addr] = nodeID
	ctx := g.AnnotateCtx(context.TODO())
	log.Eventf(ctx, "add bootstrap %s", addr)
	return true
}

func (g *Gossip) maybeCleanupBootstrapAddressesLocked() {
	__antithesis_instrumentation__.Notify(65325)
	if g.storage == nil || func() bool {
		__antithesis_instrumentation__.Notify(65329)
		return g.hasCleanedBS == true
	}() == true {
		__antithesis_instrumentation__.Notify(65330)
		return
	} else {
		__antithesis_instrumentation__.Notify(65331)
	}
	__antithesis_instrumentation__.Notify(65326)
	defer func() { __antithesis_instrumentation__.Notify(65332); g.hasCleanedBS = true }()
	__antithesis_instrumentation__.Notify(65327)
	ctx := g.AnnotateCtx(context.TODO())
	log.Event(ctx, "cleaning up bootstrap addresses")

	g.addresses = g.addresses[:0]
	g.addressIdx = 0
	g.bootstrapInfo.Addresses = g.bootstrapInfo.Addresses[:0]
	g.bootstrapAddrs = map[util.UnresolvedAddr]roachpb.NodeID{}
	g.addressExists = map[util.UnresolvedAddr]bool{}
	g.addressesTried = map[int]struct{}{}

	var desc roachpb.NodeDescriptor
	if err := g.mu.is.visitInfos(func(key string, i *Info) error {
		__antithesis_instrumentation__.Notify(65333)
		if strings.HasPrefix(key, KeyNodeIDPrefix) {
			__antithesis_instrumentation__.Notify(65335)
			if err := i.Value.GetProto(&desc); err != nil {
				__antithesis_instrumentation__.Notify(65338)
				return err
			} else {
				__antithesis_instrumentation__.Notify(65339)
			}
			__antithesis_instrumentation__.Notify(65336)
			if desc.Address.IsEmpty() || func() bool {
				__antithesis_instrumentation__.Notify(65340)
				return desc.Address == g.mu.is.NodeAddr == true
			}() == true {
				__antithesis_instrumentation__.Notify(65341)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(65342)
			}
			__antithesis_instrumentation__.Notify(65337)
			g.maybeAddAddressLocked(desc.Address)
			g.maybeAddBootstrapAddressLocked(desc.Address, desc.NodeID)
		} else {
			__antithesis_instrumentation__.Notify(65343)
		}
		__antithesis_instrumentation__.Notify(65334)
		return nil
	}, true); err != nil {
		__antithesis_instrumentation__.Notify(65344)
		log.Errorf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(65345)
	}
	__antithesis_instrumentation__.Notify(65328)

	if err := g.storage.WriteBootstrapInfo(&g.bootstrapInfo); err != nil {
		__antithesis_instrumentation__.Notify(65346)
		log.Errorf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(65347)
	}
}

func maxPeers(nodeCount int) int {
	__antithesis_instrumentation__.Notify(65348)

	maxPeers := int(math.Ceil(math.Exp(math.Log(float64(nodeCount)) / float64(maxHops-2))))
	if maxPeers < minPeers {
		__antithesis_instrumentation__.Notify(65350)
		return minPeers
	} else {
		__antithesis_instrumentation__.Notify(65351)
	}
	__antithesis_instrumentation__.Notify(65349)
	return maxPeers
}

func (g *Gossip) updateNodeAddress(key string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(65352)
	ctx := g.AnnotateCtx(context.TODO())
	var desc roachpb.NodeDescriptor
	if err := content.GetProto(&desc); err != nil {
		__antithesis_instrumentation__.Notify(65359)
		log.Errorf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(65360)
	}
	__antithesis_instrumentation__.Notify(65353)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(65361)
		log.Infof(ctx, "updateNodeAddress called on %q with desc %+v", key, desc)
	} else {
		__antithesis_instrumentation__.Notify(65362)
	}
	__antithesis_instrumentation__.Notify(65354)

	g.mu.Lock()
	defer g.mu.Unlock()

	if desc.NodeID == 0 || func() bool {
		__antithesis_instrumentation__.Notify(65363)
		return desc.Address.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(65364)
		nodeID, err := NodeIDFromKey(key, KeyNodeIDPrefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(65366)
			log.Health.Errorf(ctx, "unable to update node address for removed node: %s", err)
			return
		} else {
			__antithesis_instrumentation__.Notify(65367)
		}
		__antithesis_instrumentation__.Notify(65365)
		log.Health.Infof(ctx, "removed n%d from gossip", nodeID)
		g.removeNodeDescriptorLocked(nodeID)
		return
	} else {
		__antithesis_instrumentation__.Notify(65368)
	}
	__antithesis_instrumentation__.Notify(65355)

	existingDesc, ok := g.nodeDescs[desc.NodeID]
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(65369)
		return !existingDesc.Equal(&desc) == true
	}() == true {
		__antithesis_instrumentation__.Notify(65370)
		g.nodeDescs[desc.NodeID] = &desc
	} else {
		__antithesis_instrumentation__.Notify(65371)
	}
	__antithesis_instrumentation__.Notify(65356)

	if ok && func() bool {
		__antithesis_instrumentation__.Notify(65372)
		return existingDesc.Address == desc.Address == true
	}() == true {
		__antithesis_instrumentation__.Notify(65373)
		return
	} else {
		__antithesis_instrumentation__.Notify(65374)
	}
	__antithesis_instrumentation__.Notify(65357)
	g.recomputeMaxPeersLocked()

	if desc.Address == g.mu.is.NodeAddr {
		__antithesis_instrumentation__.Notify(65375)
		return
	} else {
		__antithesis_instrumentation__.Notify(65376)
	}
	__antithesis_instrumentation__.Notify(65358)

	g.maybeAddAddressLocked(desc.Address)

	added := g.maybeAddBootstrapAddressLocked(desc.Address, desc.NodeID)
	if added && func() bool {
		__antithesis_instrumentation__.Notify(65377)
		return g.storage != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(65378)
		if err := g.storage.WriteBootstrapInfo(&g.bootstrapInfo); err != nil {
			__antithesis_instrumentation__.Notify(65379)
			log.Errorf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(65380)
		}
	} else {
		__antithesis_instrumentation__.Notify(65381)
	}
}

func (g *Gossip) removeNodeDescriptorLocked(nodeID roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(65382)
	delete(g.nodeDescs, nodeID)
	g.recomputeMaxPeersLocked()
}

func (g *Gossip) updateStoreMap(key string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(65383)
	ctx := g.AnnotateCtx(context.TODO())
	var desc roachpb.StoreDescriptor
	if err := content.GetProto(&desc); err != nil {
		__antithesis_instrumentation__.Notify(65386)
		log.Errorf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(65387)
	}
	__antithesis_instrumentation__.Notify(65384)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(65388)
		log.Infof(ctx, "updateStoreMap called on %q with desc %+v", key, desc)
	} else {
		__antithesis_instrumentation__.Notify(65389)
	}
	__antithesis_instrumentation__.Notify(65385)

	g.mu.Lock()
	defer g.mu.Unlock()
	g.storeMap[desc.StoreID] = desc.Node.NodeID
}

func (g *Gossip) updateClients() {
	__antithesis_instrumentation__.Notify(65390)
	nodeID := g.NodeID.Get()
	if nodeID == 0 {
		__antithesis_instrumentation__.Notify(65393)
		return
	} else {
		__antithesis_instrumentation__.Notify(65394)
	}
	__antithesis_instrumentation__.Notify(65391)

	var buf bytes.Buffer
	var sep string

	g.mu.RLock()
	g.clientsMu.Lock()
	for _, c := range g.clientsMu.clients {
		__antithesis_instrumentation__.Notify(65395)
		if c.peerID != 0 {
			__antithesis_instrumentation__.Notify(65396)
			fmt.Fprintf(&buf, "%s%d", sep, c.peerID)
			sep = ","
		} else {
			__antithesis_instrumentation__.Notify(65397)
		}
	}
	__antithesis_instrumentation__.Notify(65392)
	g.clientsMu.Unlock()
	g.mu.RUnlock()

	if err := g.AddInfo(MakeGossipClientsKey(nodeID), buf.Bytes(), 2*defaultClientsInterval); err != nil {
		__antithesis_instrumentation__.Notify(65398)
		log.Errorf(g.AnnotateCtx(context.Background()), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(65399)
	}
}

func (g *Gossip) recomputeMaxPeersLocked() {
	__antithesis_instrumentation__.Notify(65400)
	maxPeers := maxPeers(len(g.nodeDescs))
	g.mu.incoming.setMaxSize(maxPeers)
	g.outgoing.setMaxSize(maxPeers)
}

func (g *Gossip) getNodeDescriptorLocked(nodeID roachpb.NodeID) (*roachpb.NodeDescriptor, error) {
	__antithesis_instrumentation__.Notify(65401)
	if desc, ok := g.nodeDescs[nodeID]; ok {
		__antithesis_instrumentation__.Notify(65404)
		if desc.Address.IsEmpty() {
			__antithesis_instrumentation__.Notify(65406)
			log.Fatalf(g.AnnotateCtx(context.Background()), "n%d has an empty address", nodeID)
		} else {
			__antithesis_instrumentation__.Notify(65407)
		}
		__antithesis_instrumentation__.Notify(65405)
		return desc, nil
	} else {
		__antithesis_instrumentation__.Notify(65408)
	}
	__antithesis_instrumentation__.Notify(65402)

	nodeIDKey := MakeNodeIDKey(nodeID)

	if i := g.mu.is.getInfo(nodeIDKey); i != nil {
		__antithesis_instrumentation__.Notify(65409)
		if err := i.Value.Verify([]byte(nodeIDKey)); err != nil {
			__antithesis_instrumentation__.Notify(65413)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(65414)
		}
		__antithesis_instrumentation__.Notify(65410)
		nodeDescriptor := &roachpb.NodeDescriptor{}
		if err := i.Value.GetProto(nodeDescriptor); err != nil {
			__antithesis_instrumentation__.Notify(65415)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(65416)
		}
		__antithesis_instrumentation__.Notify(65411)

		if nodeDescriptor.NodeID == 0 || func() bool {
			__antithesis_instrumentation__.Notify(65417)
			return nodeDescriptor.Address.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(65418)
			return nil, errors.Errorf("n%d has been removed from the cluster", nodeID)
		} else {
			__antithesis_instrumentation__.Notify(65419)
		}
		__antithesis_instrumentation__.Notify(65412)

		return nodeDescriptor, nil
	} else {
		__antithesis_instrumentation__.Notify(65420)
	}
	__antithesis_instrumentation__.Notify(65403)

	return nil, errors.Errorf("unable to look up descriptor for n%d", nodeID)
}

func (g *Gossip) getNodeIDAddressLocked(nodeID roachpb.NodeID) (*util.UnresolvedAddr, error) {
	__antithesis_instrumentation__.Notify(65421)
	nd, err := g.getNodeDescriptorLocked(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(65423)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(65424)
	}
	__antithesis_instrumentation__.Notify(65422)
	return nd.AddressForLocality(g.locality), nil
}

func (g *Gossip) getNodeIDSQLAddressLocked(nodeID roachpb.NodeID) (*util.UnresolvedAddr, error) {
	__antithesis_instrumentation__.Notify(65425)
	nd, err := g.getNodeDescriptorLocked(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(65427)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(65428)
	}
	__antithesis_instrumentation__.Notify(65426)
	return &nd.SQLAddress, nil
}

func (g *Gossip) getNodeIDHTTPAddressLocked(nodeID roachpb.NodeID) (*util.UnresolvedAddr, error) {
	__antithesis_instrumentation__.Notify(65429)
	nd, err := g.getNodeDescriptorLocked(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(65431)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(65432)
	}
	__antithesis_instrumentation__.Notify(65430)
	return &nd.HTTPAddress, nil
}

func (g *Gossip) AddInfo(key string, val []byte, ttl time.Duration) error {
	__antithesis_instrumentation__.Notify(65433)
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.addInfoLocked(key, val, ttl)
}

func (g *Gossip) addInfoLocked(key string, val []byte, ttl time.Duration) error {
	__antithesis_instrumentation__.Notify(65434)
	err := g.mu.is.addInfo(key, g.mu.is.newInfo(val, ttl))
	if err == nil {
		__antithesis_instrumentation__.Notify(65436)
		g.signalConnectedLocked()
	} else {
		__antithesis_instrumentation__.Notify(65437)
	}
	__antithesis_instrumentation__.Notify(65435)
	return err
}

func (g *Gossip) AddInfoProto(key string, msg protoutil.Message, ttl time.Duration) error {
	__antithesis_instrumentation__.Notify(65438)
	bytes, err := protoutil.Marshal(msg)
	if err != nil {
		__antithesis_instrumentation__.Notify(65440)
		return err
	} else {
		__antithesis_instrumentation__.Notify(65441)
	}
	__antithesis_instrumentation__.Notify(65439)
	return g.AddInfo(key, bytes, ttl)
}

func (g *Gossip) AddClusterID(val uuid.UUID) error {
	__antithesis_instrumentation__.Notify(65442)
	return g.AddInfo(KeyClusterID, val.GetBytes(), 0)
}

func (g *Gossip) GetClusterID() (uuid.UUID, error) {
	__antithesis_instrumentation__.Notify(65443)
	uuidBytes, err := g.GetInfo(KeyClusterID)
	if err != nil {
		__antithesis_instrumentation__.Notify(65446)
		return uuid.Nil, errors.Wrap(err, "unable to ascertain cluster ID from gossip network")
	} else {
		__antithesis_instrumentation__.Notify(65447)
	}
	__antithesis_instrumentation__.Notify(65444)
	clusterID, err := uuid.FromBytes(uuidBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(65448)
		return uuid.Nil, errors.Wrap(err, "unable to parse cluster ID from gossip network")
	} else {
		__antithesis_instrumentation__.Notify(65449)
	}
	__antithesis_instrumentation__.Notify(65445)
	return clusterID, nil
}

func (g *Gossip) GetInfo(key string) ([]byte, error) {
	__antithesis_instrumentation__.Notify(65450)
	g.mu.RLock()
	i := g.mu.is.getInfo(key)
	g.mu.RUnlock()

	if i != nil {
		__antithesis_instrumentation__.Notify(65452)
		if err := i.Value.Verify([]byte(key)); err != nil {
			__antithesis_instrumentation__.Notify(65454)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(65455)
		}
		__antithesis_instrumentation__.Notify(65453)
		return i.Value.GetBytes()
	} else {
		__antithesis_instrumentation__.Notify(65456)
	}
	__antithesis_instrumentation__.Notify(65451)
	return nil, NewKeyNotPresentError(key)
}

func (g *Gossip) GetInfoProto(key string, msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(65457)
	bytes, err := g.GetInfo(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(65459)
		return err
	} else {
		__antithesis_instrumentation__.Notify(65460)
	}
	__antithesis_instrumentation__.Notify(65458)
	return protoutil.Unmarshal(bytes, msg)
}

func (g *Gossip) InfoOriginatedHere(key string) bool {
	__antithesis_instrumentation__.Notify(65461)
	g.mu.RLock()
	info := g.mu.is.getInfo(key)
	g.mu.RUnlock()
	return info != nil && func() bool {
		__antithesis_instrumentation__.Notify(65462)
		return info.NodeID == g.NodeID.Get() == true
	}() == true
}

func (g *Gossip) GetInfoStatus() InfoStatus {
	__antithesis_instrumentation__.Notify(65463)
	clientStatus := g.clientStatus()
	serverStatus := g.server.status()
	connectivity := g.Connectivity()

	g.mu.RLock()
	defer g.mu.RUnlock()
	is := InfoStatus{
		Infos:        make(map[string]Info),
		Client:       clientStatus,
		Server:       serverStatus,
		Connectivity: connectivity,
	}
	for k, v := range g.mu.is.Infos {
		__antithesis_instrumentation__.Notify(65465)
		is.Infos[k] = *protoutil.Clone(v).(*Info)
	}
	__antithesis_instrumentation__.Notify(65464)
	return is
}

func (g *Gossip) IterateInfos(prefix string, visit func(k string, info Info) error) error {
	__antithesis_instrumentation__.Notify(65466)
	g.mu.RLock()
	defer g.mu.RUnlock()
	for k, v := range g.mu.is.Infos {
		__antithesis_instrumentation__.Notify(65468)
		if strings.HasPrefix(k, prefix+separator) {
			__antithesis_instrumentation__.Notify(65469)
			if err := visit(k, *(protoutil.Clone(v).(*Info))); err != nil {
				__antithesis_instrumentation__.Notify(65470)
				return err
			} else {
				__antithesis_instrumentation__.Notify(65471)
			}
		} else {
			__antithesis_instrumentation__.Notify(65472)
		}
	}
	__antithesis_instrumentation__.Notify(65467)
	return nil
}

type Callback func(string, roachpb.Value)

type CallbackOption interface {
	apply(cb *callback)
}

type redundantCallbacks struct {
}

func (redundantCallbacks) apply(cb *callback) {
	__antithesis_instrumentation__.Notify(65473)
	cb.redundant = true
}

var Redundant redundantCallbacks

func (g *Gossip) RegisterCallback(pattern string, method Callback, opts ...CallbackOption) func() {
	__antithesis_instrumentation__.Notify(65474)
	g.mu.Lock()
	unregister := g.mu.is.registerCallback(pattern, method, opts...)
	g.mu.Unlock()
	return func() {
		__antithesis_instrumentation__.Notify(65475)
		g.mu.Lock()
		unregister()
		g.mu.Unlock()
	}
}

func (g *Gossip) DeprecatedGetSystemConfig() *config.SystemConfig {
	__antithesis_instrumentation__.Notify(65476)
	g.systemConfigMu.RLock()
	defer g.systemConfigMu.RUnlock()
	return g.systemConfig
}

func (g *Gossip) DeprecatedRegisterSystemConfigChannel() <-chan struct{} {
	__antithesis_instrumentation__.Notify(65477)

	c := make(chan struct{}, 1)

	g.systemConfigMu.Lock()
	defer g.systemConfigMu.Unlock()
	g.systemConfigChannels = append(g.systemConfigChannels, c)

	if g.systemConfig != nil {
		__antithesis_instrumentation__.Notify(65479)
		c <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(65480)
	}
	__antithesis_instrumentation__.Notify(65478)
	return c
}

func (g *Gossip) updateSystemConfig(key string, content roachpb.Value) {
	__antithesis_instrumentation__.Notify(65481)
	ctx := g.AnnotateCtx(context.TODO())
	if key != KeyDeprecatedSystemConfig {
		__antithesis_instrumentation__.Notify(65484)
		log.Fatalf(ctx, "wrong key received on SystemConfig callback: %s", key)
	} else {
		__antithesis_instrumentation__.Notify(65485)
	}
	__antithesis_instrumentation__.Notify(65482)
	cfg := config.NewSystemConfig(g.defaultZoneConfig)
	if err := content.GetProto(&cfg.SystemConfigEntries); err != nil {
		__antithesis_instrumentation__.Notify(65486)
		log.Errorf(ctx, "could not unmarshal system config on callback: %s", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(65487)
	}
	__antithesis_instrumentation__.Notify(65483)

	g.systemConfigMu.Lock()
	defer g.systemConfigMu.Unlock()
	g.systemConfig = cfg
	for _, c := range g.systemConfigChannels {
		__antithesis_instrumentation__.Notify(65488)
		select {
		case c <- struct{}{}:
			__antithesis_instrumentation__.Notify(65489)
		default:
			__antithesis_instrumentation__.Notify(65490)
		}
	}
}

func (g *Gossip) Incoming() []roachpb.NodeID {
	__antithesis_instrumentation__.Notify(65491)
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.mu.incoming.asSlice()
}

func (g *Gossip) Outgoing() []roachpb.NodeID {
	__antithesis_instrumentation__.Notify(65492)
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.outgoing.asSlice()
}

func (g *Gossip) MaxHops() uint32 {
	__antithesis_instrumentation__.Notify(65493)
	g.mu.Lock()
	defer g.mu.Unlock()
	_, maxHops := g.mu.is.mostDistant(func(_ roachpb.NodeID) bool { __antithesis_instrumentation__.Notify(65495); return false })
	__antithesis_instrumentation__.Notify(65494)
	return maxHops
}

func (g *Gossip) Start(advertAddr net.Addr, addresses []util.UnresolvedAddr) {
	__antithesis_instrumentation__.Notify(65496)
	g.AssertNotStarted(context.Background())
	g.started = true
	g.setAddresses(addresses)
	g.server.start(advertAddr)
	g.bootstrap()
	g.manage()
}

func (g *Gossip) hasIncomingLocked(nodeID roachpb.NodeID) bool {
	__antithesis_instrumentation__.Notify(65497)
	return g.mu.incoming.hasNode(nodeID)
}

func (g *Gossip) hasOutgoingLocked(nodeID roachpb.NodeID) bool {
	__antithesis_instrumentation__.Notify(65498)

	nodeAddr, err := g.getNodeIDAddressLocked(nodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(65501)

		ctx := g.AnnotateCtx(context.TODO())
		log.Errorf(ctx, "unable to get address for n%d: %s", nodeID, err)
		return g.outgoing.hasNode(nodeID)
	} else {
		__antithesis_instrumentation__.Notify(65502)
	}
	__antithesis_instrumentation__.Notify(65499)
	c := g.findClient(func(c *client) bool {
		__antithesis_instrumentation__.Notify(65503)
		return c.addr.String() == nodeAddr.String()
	})
	__antithesis_instrumentation__.Notify(65500)
	return c != nil
}

func (g *Gossip) getNextBootstrapAddressLocked() util.UnresolvedAddr {
	__antithesis_instrumentation__.Notify(65504)

	for range g.addresses {
		__antithesis_instrumentation__.Notify(65506)
		g.addressIdx++
		g.addressIdx %= len(g.addresses)
		defer func(idx int) { __antithesis_instrumentation__.Notify(65508); g.addressesTried[idx] = struct{}{} }(g.addressIdx)
		__antithesis_instrumentation__.Notify(65507)
		addr := g.addresses[g.addressIdx]
		addrStr := addr.String()
		if _, addrActive := g.bootstrapping[addrStr]; !addrActive {
			__antithesis_instrumentation__.Notify(65509)
			g.bootstrapping[addrStr] = struct{}{}
			return addr
		} else {
			__antithesis_instrumentation__.Notify(65510)
		}
	}
	__antithesis_instrumentation__.Notify(65505)
	return util.UnresolvedAddr{}
}

func (g *Gossip) bootstrap() {
	__antithesis_instrumentation__.Notify(65511)
	ctx := g.AnnotateCtx(context.Background())
	_ = g.server.stopper.RunAsyncTask(ctx, "gossip-bootstrap", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(65512)
		ctx = logtags.AddTag(ctx, "bootstrap", nil)
		var bootstrapTimer timeutil.Timer
		defer bootstrapTimer.Stop()
		for {
			__antithesis_instrumentation__.Notify(65513)
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(65516)
				g.mu.Lock()
				defer g.mu.Unlock()
				haveClients := g.outgoing.len() > 0
				haveSentinel := g.mu.is.getInfo(KeySentinel) != nil
				log.Eventf(ctx, "have clients: %t, have sentinel: %t", haveClients, haveSentinel)
				if !haveClients || func() bool {
					__antithesis_instrumentation__.Notify(65517)
					return !haveSentinel == true
				}() == true {
					__antithesis_instrumentation__.Notify(65518)

					if addr := g.getNextBootstrapAddressLocked(); !addr.IsEmpty() {
						__antithesis_instrumentation__.Notify(65519)
						g.startClientLocked(addr)
					} else {
						__antithesis_instrumentation__.Notify(65520)
						bootstrapAddrs := make([]string, 0, len(g.bootstrapping))
						for addr := range g.bootstrapping {
							__antithesis_instrumentation__.Notify(65522)
							bootstrapAddrs = append(bootstrapAddrs, addr)
						}
						__antithesis_instrumentation__.Notify(65521)
						log.Eventf(ctx, "no next bootstrap address; currently bootstrapping: %v", bootstrapAddrs)

						g.maybeSignalStatusChangeLocked()
					}
				} else {
					__antithesis_instrumentation__.Notify(65523)
				}
			}(ctx)
			__antithesis_instrumentation__.Notify(65514)

			bootstrapTimer.Reset(g.bootstrapInterval)
			log.Eventf(ctx, "sleeping %s until bootstrap", g.bootstrapInterval)
			select {
			case <-bootstrapTimer.C:
				__antithesis_instrumentation__.Notify(65524)
				bootstrapTimer.Read = true

			case <-g.server.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(65525)
				return
			}
			__antithesis_instrumentation__.Notify(65515)
			log.Eventf(ctx, "idling until bootstrap required")

			select {
			case <-g.stalledCh:
				__antithesis_instrumentation__.Notify(65526)
				log.Eventf(ctx, "detected stall; commencing bootstrap")

			case <-g.server.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(65527)
				return
			}
		}
	})
}

func (g *Gossip) manage() {
	__antithesis_instrumentation__.Notify(65528)
	ctx := g.AnnotateCtx(context.Background())
	_ = g.server.stopper.RunAsyncTask(ctx, "gossip-manage", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(65529)
		clientsTimer := timeutil.NewTimer()
		cullTimer := timeutil.NewTimer()
		stallTimer := timeutil.NewTimer()
		defer clientsTimer.Stop()
		defer cullTimer.Stop()
		defer stallTimer.Stop()

		clientsTimer.Reset(defaultClientsInterval)
		cullTimer.Reset(jitteredInterval(g.cullInterval))
		stallTimer.Reset(jitteredInterval(g.stallInterval))
		for {
			__antithesis_instrumentation__.Notify(65530)
			select {
			case <-g.server.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(65531)
				return
			case c := <-g.disconnected:
				__antithesis_instrumentation__.Notify(65532)
				g.doDisconnected(c)
			case <-g.tighten:
				__antithesis_instrumentation__.Notify(65533)
				g.tightenNetwork(ctx)
			case <-clientsTimer.C:
				__antithesis_instrumentation__.Notify(65534)
				clientsTimer.Read = true
				g.updateClients()
				clientsTimer.Reset(defaultClientsInterval)
			case <-cullTimer.C:
				__antithesis_instrumentation__.Notify(65535)
				cullTimer.Read = true
				cullTimer.Reset(jitteredInterval(g.cullInterval))
				func() {
					__antithesis_instrumentation__.Notify(65537)
					g.mu.Lock()
					if !g.outgoing.hasSpace() {
						__antithesis_instrumentation__.Notify(65539)
						leastUsefulID := g.mu.is.leastUseful(g.outgoing)

						if c := g.findClient(func(c *client) bool {
							__antithesis_instrumentation__.Notify(65540)
							return c.peerID == leastUsefulID
						}); c != nil {
							__antithesis_instrumentation__.Notify(65541)
							log.VEventf(ctx, 1, "closing least useful client %+v to tighten network graph", c)
							log.Health.Infof(ctx, "closing gossip client n%d %s", c.peerID, c.addr)
							c.close()

							defer func() {
								__antithesis_instrumentation__.Notify(65542)
								g.doDisconnected(<-g.disconnected)
							}()
						} else {
							__antithesis_instrumentation__.Notify(65543)
							if log.V(1) {
								__antithesis_instrumentation__.Notify(65544)
								g.clientsMu.Lock()
								log.Dev.Infof(ctx, "couldn't find least useful client among %+v", g.clientsMu.clients)
								g.clientsMu.Unlock()
							} else {
								__antithesis_instrumentation__.Notify(65545)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(65546)
					}
					__antithesis_instrumentation__.Notify(65538)
					g.mu.Unlock()
				}()
			case <-stallTimer.C:
				__antithesis_instrumentation__.Notify(65536)
				stallTimer.Read = true
				stallTimer.Reset(jitteredInterval(g.stallInterval))

				g.mu.Lock()
				g.maybeSignalStatusChangeLocked()
				g.mu.Unlock()
			}
		}
	})
}

func jitteredInterval(interval time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(65547)
	return time.Duration(float64(interval) * (0.75 + 0.5*rand.Float64()))
}

func (g *Gossip) tightenNetwork(ctx context.Context) {
	__antithesis_instrumentation__.Notify(65548)
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.outgoing.hasSpace() {
		__antithesis_instrumentation__.Notify(65549)
		distantNodeID, distantHops := g.mu.is.mostDistant(g.hasOutgoingLocked)
		log.VEventf(ctx, 2, "distantHops: %d from %d", distantHops, distantNodeID)
		if distantHops <= maxHops {
			__antithesis_instrumentation__.Notify(65551)
			return
		} else {
			__antithesis_instrumentation__.Notify(65552)
		}
		__antithesis_instrumentation__.Notify(65550)
		if nodeAddr, err := g.getNodeIDAddressLocked(distantNodeID); err != nil || func() bool {
			__antithesis_instrumentation__.Notify(65553)
			return nodeAddr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(65554)
			log.Health.Errorf(ctx, "unable to get address for n%d: %s", distantNodeID, err)
		} else {
			__antithesis_instrumentation__.Notify(65555)
			log.Health.Infof(ctx, "starting client to n%d (%d > %d) to tighten network graph",
				distantNodeID, distantHops, maxHops)
			log.Eventf(ctx, "tightening network with new client to %s", nodeAddr)
			g.startClientLocked(*nodeAddr)
		}
	} else {
		__antithesis_instrumentation__.Notify(65556)
	}
}

func (g *Gossip) doDisconnected(c *client) {
	__antithesis_instrumentation__.Notify(65557)
	defer g.updateClients()

	g.mu.Lock()
	defer g.mu.Unlock()
	g.removeClientLocked(c)

	if c.forwardAddr != nil {
		__antithesis_instrumentation__.Notify(65559)
		g.startClientLocked(*c.forwardAddr)
	} else {
		__antithesis_instrumentation__.Notify(65560)
	}
	__antithesis_instrumentation__.Notify(65558)
	g.maybeSignalStatusChangeLocked()
}

func (g *Gossip) maybeSignalStatusChangeLocked() {
	__antithesis_instrumentation__.Notify(65561)
	ctx := g.AnnotateCtx(context.TODO())
	orphaned := g.outgoing.len()+g.mu.incoming.len() == 0
	multiNode := len(g.bootstrapInfo.Addresses) > 0

	stalled := (orphaned && func() bool {
		__antithesis_instrumentation__.Notify(65563)
		return multiNode == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(65564)
		return g.mu.is.getInfo(KeySentinel) == nil == true
	}() == true
	if stalled {
		__antithesis_instrumentation__.Notify(65565)

		if !g.stalled {
			__antithesis_instrumentation__.Notify(65567)
			log.Eventf(ctx, "now stalled")
			if orphaned {
				__antithesis_instrumentation__.Notify(65568)
				if len(g.addresses) == 0 {
					__antithesis_instrumentation__.Notify(65569)
					log.Ops.Warningf(ctx, "no addresses found; use --join to specify a connected node")
				} else {
					__antithesis_instrumentation__.Notify(65570)
					log.Health.Warningf(ctx, "no incoming or outgoing connections")
				}
			} else {
				__antithesis_instrumentation__.Notify(65571)
				if len(g.addressesTried) == len(g.addresses) {
					__antithesis_instrumentation__.Notify(65572)
					log.Health.Warningf(ctx, "first range unavailable; addresses exhausted")
				} else {
					__antithesis_instrumentation__.Notify(65573)
					log.Health.Warningf(ctx, "first range unavailable; trying remaining addresses")
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(65574)
		}
		__antithesis_instrumentation__.Notify(65566)
		if len(g.addresses) > 0 {
			__antithesis_instrumentation__.Notify(65575)
			g.signalStalledLocked()
		} else {
			__antithesis_instrumentation__.Notify(65576)
		}
	} else {
		__antithesis_instrumentation__.Notify(65577)
		if g.stalled {
			__antithesis_instrumentation__.Notify(65579)
			log.Eventf(ctx, "connected")
			log.Ops.Infof(ctx, "node has connected to cluster via gossip")
			g.signalConnectedLocked()
		} else {
			__antithesis_instrumentation__.Notify(65580)
		}
		__antithesis_instrumentation__.Notify(65578)
		g.maybeCleanupBootstrapAddressesLocked()
	}
	__antithesis_instrumentation__.Notify(65562)
	g.stalled = stalled
}

func (g *Gossip) signalStalledLocked() {
	__antithesis_instrumentation__.Notify(65581)
	select {
	case g.stalledCh <- struct{}{}:
		__antithesis_instrumentation__.Notify(65582)
	default:
		__antithesis_instrumentation__.Notify(65583)
	}
}

func (g *Gossip) signalConnectedLocked() {
	__antithesis_instrumentation__.Notify(65584)

	if !g.hasConnected && func() bool {
		__antithesis_instrumentation__.Notify(65585)
		return g.mu.is.getInfo(KeyClusterID) != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(65586)
		g.hasConnected = true
		close(g.Connected)
	} else {
		__antithesis_instrumentation__.Notify(65587)
	}
}

func (g *Gossip) startClientLocked(addr util.UnresolvedAddr) {
	__antithesis_instrumentation__.Notify(65588)
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()
	breaker, ok := g.clientsMu.breakers[addr.String()]
	if !ok {
		__antithesis_instrumentation__.Notify(65590)
		name := fmt.Sprintf("gossip %v->%v", g.rpcContext.Config.Addr, addr)
		breaker = g.rpcContext.NewBreaker(name)
		g.clientsMu.breakers[addr.String()] = breaker
	} else {
		__antithesis_instrumentation__.Notify(65591)
	}
	__antithesis_instrumentation__.Notify(65589)
	ctx := g.AnnotateCtx(context.TODO())
	log.VEventf(ctx, 1, "starting new client to %s", addr)
	c := newClient(g.server.AmbientContext, &addr, g.serverMetrics)
	g.clientsMu.clients = append(g.clientsMu.clients, c)
	c.startLocked(g, g.disconnected, g.rpcContext, g.server.stopper, breaker)
}

func (g *Gossip) removeClientLocked(target *client) {
	__antithesis_instrumentation__.Notify(65592)
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()
	for i, candidate := range g.clientsMu.clients {
		__antithesis_instrumentation__.Notify(65593)
		if candidate == target {
			__antithesis_instrumentation__.Notify(65594)
			ctx := g.AnnotateCtx(context.TODO())
			log.VEventf(ctx, 1, "client %s disconnected", candidate.addr)
			g.clientsMu.clients = append(g.clientsMu.clients[:i], g.clientsMu.clients[i+1:]...)
			delete(g.bootstrapping, candidate.addr.String())
			g.outgoing.removeNode(candidate.peerID)
			break
		} else {
			__antithesis_instrumentation__.Notify(65595)
		}
	}
}

func (g *Gossip) findClient(match func(*client) bool) *client {
	__antithesis_instrumentation__.Notify(65596)
	g.clientsMu.Lock()
	defer g.clientsMu.Unlock()
	for _, c := range g.clientsMu.clients {
		__antithesis_instrumentation__.Notify(65598)
		if match(c) {
			__antithesis_instrumentation__.Notify(65599)
			return c
		} else {
			__antithesis_instrumentation__.Notify(65600)
		}
	}
	__antithesis_instrumentation__.Notify(65597)
	return nil
}

type firstRangeMissingError struct{}

func (f firstRangeMissingError) Error() string {
	__antithesis_instrumentation__.Notify(65601)
	return "the descriptor for the first range is not available via gossip"
}

func (g *Gossip) GetFirstRangeDescriptor() (*roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(65602)
	desc := &roachpb.RangeDescriptor{}
	if err := g.GetInfoProto(KeyFirstRangeDescriptor, desc); err != nil {
		__antithesis_instrumentation__.Notify(65604)
		return nil, firstRangeMissingError{}
	} else {
		__antithesis_instrumentation__.Notify(65605)
	}
	__antithesis_instrumentation__.Notify(65603)
	return desc, nil
}

func (g *Gossip) OnFirstRangeChanged(cb func(*roachpb.RangeDescriptor)) {
	__antithesis_instrumentation__.Notify(65606)
	g.RegisterCallback(KeyFirstRangeDescriptor, func(_ string, value roachpb.Value) {
		__antithesis_instrumentation__.Notify(65607)
		ctx := context.Background()
		desc := &roachpb.RangeDescriptor{}
		if err := value.GetProto(desc); err != nil {
			__antithesis_instrumentation__.Notify(65608)
			log.Errorf(ctx, "unable to parse gossiped first range descriptor: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(65609)
			cb(desc)
		}
	})
}

func MakeOptionalGossip(g *Gossip) OptionalGossip {
	__antithesis_instrumentation__.Notify(65610)
	return OptionalGossip{
		w: errorutil.MakeTenantSQLDeprecatedWrapper(g, g != nil),
	}
}

type OptionalGossip struct {
	w errorutil.TenantSQLDeprecatedWrapper
}

func (og OptionalGossip) OptionalErr(issue int) (*Gossip, error) {
	__antithesis_instrumentation__.Notify(65611)
	v, err := og.w.OptionalErr(issue)
	if err != nil {
		__antithesis_instrumentation__.Notify(65613)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(65614)
	}
	__antithesis_instrumentation__.Notify(65612)

	g, _ := v.(*Gossip)
	return g, nil
}

func (og OptionalGossip) Optional(issue int) (*Gossip, bool) {
	__antithesis_instrumentation__.Notify(65615)
	v, ok := og.w.Optional()
	if !ok {
		__antithesis_instrumentation__.Notify(65617)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(65618)
	}
	__antithesis_instrumentation__.Notify(65616)

	g, _ := v.(*Gossip)
	return g, true
}
