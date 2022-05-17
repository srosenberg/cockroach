package localcluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"go/build"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/gogo/protobuf/proto"

	_ "github.com/lib/pq"
)

func repoRoot() string {
	__antithesis_instrumentation__.Notify(826)
	root, err := build.Import("github.com/cockroachdb/cockroach", "", build.FindOnly)
	if err != nil {
		__antithesis_instrumentation__.Notify(828)
		panic(fmt.Sprintf("must run from within the cockroach repository: %s", err))
	} else {
		__antithesis_instrumentation__.Notify(829)
	}
	__antithesis_instrumentation__.Notify(827)
	return root.Dir
}

func SourceBinary() string {
	__antithesis_instrumentation__.Notify(830)
	return filepath.Join(repoRoot(), "cockroach")
}

const listeningURLFile = "cockroachdb-url"

func IsUnavailableError(err error) bool {
	__antithesis_instrumentation__.Notify(831)
	return strings.Contains(err.Error(), "grpc: the connection is unavailable")
}

type ClusterConfig struct {
	Ephemeral   bool
	Binary      string
	AllNodeArgs []string
	NumNodes    int
	DataDir     string
	LogDir      string
	PerNodeCfg  map[int]NodeConfig
	DB          string
	NumWorkers  int
	NoWait      bool
}

type NodeConfig struct {
	Binary            string
	DataDir           string
	LogDir            string
	Addr              string
	ExtraArgs         []string
	ExtraEnv          []string
	RPCPort, HTTPPort int
	DB                string
	NumWorkers        int
}

func MakePerNodeFixedPortsCfg(numNodes int) map[int]NodeConfig {
	__antithesis_instrumentation__.Notify(832)
	perNodeCfg := make(map[int]NodeConfig)

	for i := 0; i < numNodes; i++ {
		__antithesis_instrumentation__.Notify(834)
		perNodeCfg[i] = NodeConfig{
			RPCPort:  26257 + 2*i,
			HTTPPort: 26258 + 2*i,
		}
	}
	__antithesis_instrumentation__.Notify(833)

	return perNodeCfg
}

type Cluster struct {
	Cfg     ClusterConfig
	seq     *seqGen
	Nodes   []*Node
	stopper *stop.Stopper
	started time.Time
}

type seqGen int32

func (s *seqGen) Next() int32 {
	__antithesis_instrumentation__.Notify(835)
	return atomic.AddInt32((*int32)(s), 1)
}

func New(cfg ClusterConfig) *Cluster {
	__antithesis_instrumentation__.Notify(836)
	if cfg.Binary == "" {
		__antithesis_instrumentation__.Notify(838)
		cfg.Binary = SourceBinary()
	} else {
		__antithesis_instrumentation__.Notify(839)
	}
	__antithesis_instrumentation__.Notify(837)
	return &Cluster{
		Cfg:     cfg,
		seq:     new(seqGen),
		stopper: stop.NewStopper(),
	}
}

func (c *Cluster) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(840)
	c.started = timeutil.Now()

	chs := make([]<-chan error, c.Cfg.NumNodes)
	for i := 0; i < c.Cfg.NumNodes; i++ {
		__antithesis_instrumentation__.Notify(843)
		cfg := c.Cfg.PerNodeCfg[i]
		if cfg.Binary == "" {
			__antithesis_instrumentation__.Notify(850)
			cfg.Binary = c.Cfg.Binary
		} else {
			__antithesis_instrumentation__.Notify(851)
		}
		__antithesis_instrumentation__.Notify(844)
		if cfg.DataDir == "" {
			__antithesis_instrumentation__.Notify(852)
			cfg.DataDir = filepath.Join(c.Cfg.DataDir, fmt.Sprintf("%d", i+1))
		} else {
			__antithesis_instrumentation__.Notify(853)
		}
		__antithesis_instrumentation__.Notify(845)
		if cfg.LogDir == "" && func() bool {
			__antithesis_instrumentation__.Notify(854)
			return c.Cfg.LogDir != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(855)
			cfg.LogDir = filepath.Join(c.Cfg.LogDir, fmt.Sprintf("%d", i+1))
		} else {
			__antithesis_instrumentation__.Notify(856)
		}
		__antithesis_instrumentation__.Notify(846)
		if cfg.Addr == "" {
			__antithesis_instrumentation__.Notify(857)
			cfg.Addr = "127.0.0.1"
		} else {
			__antithesis_instrumentation__.Notify(858)
		}
		__antithesis_instrumentation__.Notify(847)
		if cfg.DB == "" {
			__antithesis_instrumentation__.Notify(859)
			cfg.DB = c.Cfg.DB
		} else {
			__antithesis_instrumentation__.Notify(860)
		}
		__antithesis_instrumentation__.Notify(848)
		if cfg.NumWorkers == 0 {
			__antithesis_instrumentation__.Notify(861)
			cfg.NumWorkers = c.Cfg.NumWorkers
		} else {
			__antithesis_instrumentation__.Notify(862)
		}
		__antithesis_instrumentation__.Notify(849)
		cfg.ExtraArgs = append(append([]string(nil), c.Cfg.AllNodeArgs...), cfg.ExtraArgs...)
		var node *Node
		node, chs[i] = c.makeNode(ctx, i, cfg)
		c.Nodes = append(c.Nodes, node)
		if i == 0 && func() bool {
			__antithesis_instrumentation__.Notify(863)
			return cfg.RPCPort == 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(864)
			return c.Cfg.NumNodes > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(865)

			if err := <-chs[0]; err != nil {
				__antithesis_instrumentation__.Notify(867)
				log.Fatalf(ctx, "while starting first node: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(868)
			}
			__antithesis_instrumentation__.Notify(866)
			ch := make(chan error)
			close(ch)
			chs[0] = ch
		} else {
			__antithesis_instrumentation__.Notify(869)
		}
	}
	__antithesis_instrumentation__.Notify(841)

	if !c.Cfg.NoWait {
		__antithesis_instrumentation__.Notify(870)
		for i := range chs {
			__antithesis_instrumentation__.Notify(871)
			if err := <-chs[i]; err != nil {
				__antithesis_instrumentation__.Notify(872)
				log.Fatalf(ctx, "node %d: %s", i+1, err)
			} else {
				__antithesis_instrumentation__.Notify(873)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(874)
	}
	__antithesis_instrumentation__.Notify(842)

	log.Infof(context.Background(), "started %.3fs", timeutil.Since(c.started).Seconds())

	if c.Cfg.NumNodes > 1 || func() bool {
		__antithesis_instrumentation__.Notify(875)
		return !c.Cfg.NoWait == true
	}() == true {
		__antithesis_instrumentation__.Notify(876)
		c.waitForFullReplication()
	} else {
		__antithesis_instrumentation__.Notify(877)

		log.Infof(ctx, "not waiting for initial replication")
	}
}

func (c *Cluster) Close() {
	__antithesis_instrumentation__.Notify(878)
	for _, n := range c.Nodes {
		__antithesis_instrumentation__.Notify(880)
		n.Kill()
	}
	__antithesis_instrumentation__.Notify(879)
	c.stopper.Stop(context.Background())
	if c.Cfg.Ephemeral {
		__antithesis_instrumentation__.Notify(881)
		_ = os.RemoveAll(c.Cfg.DataDir)
	} else {
		__antithesis_instrumentation__.Notify(882)
	}
}

func (c *Cluster) joins() []string {
	__antithesis_instrumentation__.Notify(883)
	type addrAndSeq struct {
		addr string
		seq  int32
	}

	var joins []addrAndSeq
	for _, node := range c.Nodes {
		__antithesis_instrumentation__.Notify(887)
		advertAddr := node.AdvertiseAddr()
		if advertAddr != "" {
			__antithesis_instrumentation__.Notify(888)
			joins = append(joins, addrAndSeq{
				addr: advertAddr,
				seq:  atomic.LoadInt32(&node.startSeq),
			})
		} else {
			__antithesis_instrumentation__.Notify(889)
		}
	}
	__antithesis_instrumentation__.Notify(884)
	sort.Slice(joins, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(890)
		return joins[i].seq < joins[j].seq
	})
	__antithesis_instrumentation__.Notify(885)

	if len(joins) == 0 {
		__antithesis_instrumentation__.Notify(891)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(892)
	}
	__antithesis_instrumentation__.Notify(886)

	return []string{joins[0].addr}
}

func (c *Cluster) IPAddr(nodeIdx int) string {
	__antithesis_instrumentation__.Notify(893)
	return c.Nodes[nodeIdx].IPAddr()
}

func (c *Cluster) RPCPort(nodeIdx int) string {
	__antithesis_instrumentation__.Notify(894)
	return c.Nodes[nodeIdx].RPCPort()
}

func (c *Cluster) makeNode(ctx context.Context, nodeIdx int, cfg NodeConfig) (*Node, <-chan error) {
	__antithesis_instrumentation__.Notify(895)
	baseCtx := &base.Config{
		User:     security.NodeUserName(),
		Insecure: true,
	}
	rpcCtx := rpc.NewContext(ctx, rpc.ContextOptions{
		TenantID: roachpb.SystemTenantID,
		Config:   baseCtx,
		Clock:    hlc.NewClock(hlc.UnixNano, 0),
		Stopper:  c.stopper,
		Settings: cluster.MakeTestingClusterSettings(),
	})

	n := &Node{
		Cfg:    cfg,
		rpcCtx: rpcCtx,
		seq:    c.seq,
	}

	args := []string{
		cfg.Binary,
		"start",
		"--insecure",

		fmt.Sprintf("--host=%s", n.IPAddr()),
		fmt.Sprintf("--port=%d", cfg.RPCPort),
		fmt.Sprintf("--http-port=%d", cfg.HTTPPort),
		fmt.Sprintf("--store=%s", cfg.DataDir),
		fmt.Sprintf("--listening-url-file=%s", n.listeningURLFile()),
		"--cache=256MiB",
	}

	if n.Cfg.LogDir != "" {
		__antithesis_instrumentation__.Notify(899)
		args = append(args, fmt.Sprintf("--log-dir=%s", n.Cfg.LogDir))
	} else {
		__antithesis_instrumentation__.Notify(900)
	}
	__antithesis_instrumentation__.Notify(896)

	n.Cfg.ExtraArgs = append(args, cfg.ExtraArgs...)

	if err := os.MkdirAll(n.logDir(), 0755); err != nil {
		__antithesis_instrumentation__.Notify(901)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(902)
	}
	__antithesis_instrumentation__.Notify(897)

	joins := c.joins()
	if nodeIdx > 0 && func() bool {
		__antithesis_instrumentation__.Notify(903)
		return len(joins) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(904)
		ch := make(chan error, 1)
		ch <- errors.Errorf("node %d started without join flags", nodeIdx+1)
		return nil, ch
	} else {
		__antithesis_instrumentation__.Notify(905)
	}
	__antithesis_instrumentation__.Notify(898)
	ch := n.StartAsync(ctx, joins...)
	return n, ch
}

func (c *Cluster) waitForFullReplication() {
	__antithesis_instrumentation__.Notify(906)
	for i := 1; true; i++ {
		__antithesis_instrumentation__.Notify(908)
		done, detail := c.isReplicated()
		if (done && func() bool {
			__antithesis_instrumentation__.Notify(911)
			return i >= 50 == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(912)
			return (i % 50) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(913)
			fmt.Print(detail)
			log.Infof(context.Background(), "waiting for replication")
		} else {
			__antithesis_instrumentation__.Notify(914)
		}
		__antithesis_instrumentation__.Notify(909)
		if done {
			__antithesis_instrumentation__.Notify(915)
			break
		} else {
			__antithesis_instrumentation__.Notify(916)
		}
		__antithesis_instrumentation__.Notify(910)
		time.Sleep(100 * time.Millisecond)
	}
	__antithesis_instrumentation__.Notify(907)

	log.Infof(context.Background(), "replicated %.3fs", timeutil.Since(c.started).Seconds())
}

func (c *Cluster) isReplicated() (bool, string) {
	__antithesis_instrumentation__.Notify(917)
	db := c.Nodes[0].DB()
	rows, err := db.Query(`SELECT range_id, start_key, end_key, array_length(replicas, 1) FROM crdb_internal.ranges`)
	if err != nil {
		__antithesis_instrumentation__.Notify(920)

		if testutils.IsError(err, "(table|relation) \"crdb_internal.ranges\" does not exist") {
			__antithesis_instrumentation__.Notify(922)
			return true, ""
		} else {
			__antithesis_instrumentation__.Notify(923)
		}
		__antithesis_instrumentation__.Notify(921)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(924)
	}
	__antithesis_instrumentation__.Notify(918)
	defer rows.Close()

	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
	done := true
	for rows.Next() {
		__antithesis_instrumentation__.Notify(925)
		var rangeID int64
		var startKey, endKey roachpb.Key
		var numReplicas int
		if err := rows.Scan(&rangeID, &startKey, &endKey, &numReplicas); err != nil {
			__antithesis_instrumentation__.Notify(927)
			log.Fatalf(context.Background(), "unable to scan range replicas: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(928)
		}
		__antithesis_instrumentation__.Notify(926)
		fmt.Fprintf(tw, "\t%s\t%s\t[%d]\t%d\n", startKey, endKey, rangeID, numReplicas)

		if numReplicas < 3 && func() bool {
			__antithesis_instrumentation__.Notify(929)
			return numReplicas != len(c.Nodes) == true
		}() == true {
			__antithesis_instrumentation__.Notify(930)
			done = false
		} else {
			__antithesis_instrumentation__.Notify(931)
		}
	}
	__antithesis_instrumentation__.Notify(919)
	_ = tw.Flush()
	return done, buf.String()
}

func (c *Cluster) UpdateZoneConfig(rangeMinBytes, rangeMaxBytes int64) {
	__antithesis_instrumentation__.Notify(932)
	zone := zonepb.DefaultZoneConfig()
	zone.RangeMinBytes = proto.Int64(rangeMinBytes)
	zone.RangeMaxBytes = proto.Int64(rangeMaxBytes)

	buf, err := protoutil.Marshal(&zone)
	if err != nil {
		__antithesis_instrumentation__.Notify(934)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(935)
	}
	__antithesis_instrumentation__.Notify(933)
	_, err = c.Nodes[0].DB().Exec(`UPSERT INTO system.zones (id, config) VALUES (0, $1)`, buf)
	if err != nil {
		__antithesis_instrumentation__.Notify(936)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(937)
	}
}

func (c *Cluster) Split(nodeIdx int, splitKey roachpb.Key) error {
	__antithesis_instrumentation__.Notify(938)
	return errors.Errorf("Split is unimplemented and should be re-implemented using SQL")
}

func (c *Cluster) TransferLease(nodeIdx int, r *rand.Rand, key roachpb.Key) (bool, error) {
	__antithesis_instrumentation__.Notify(939)
	return false, errors.Errorf("TransferLease is unimplemented and should be re-implemented using SQL")
}

func (c *Cluster) RandNode(f func(int) int) int {
	__antithesis_instrumentation__.Notify(940)
	for {
		__antithesis_instrumentation__.Notify(941)
		i := f(len(c.Nodes))
		if c.Nodes[i].Alive() {
			__antithesis_instrumentation__.Notify(942)
			return i
		} else {
			__antithesis_instrumentation__.Notify(943)
		}
	}
}

type Node struct {
	Cfg    NodeConfig
	rpcCtx *rpc.Context
	seq    *seqGen

	startSeq int32
	waitErr  atomic.Value

	syncutil.Mutex
	notRunning     chan struct{}
	cmd            *exec.Cmd
	rpcPort, pgURL string
	db             *gosql.DB
	statusClient   serverpb.StatusClient
}

func (n *Node) RPCPort() string {
	__antithesis_instrumentation__.Notify(944)
	if s := func() string {
		__antithesis_instrumentation__.Notify(947)

		n.Lock()
		defer n.Unlock()
		if n.rpcPort != "" && func() bool {
			__antithesis_instrumentation__.Notify(949)
			return n.rpcPort != "0" == true
		}() == true {
			__antithesis_instrumentation__.Notify(950)
			return n.rpcPort
		} else {
			__antithesis_instrumentation__.Notify(951)
		}
		__antithesis_instrumentation__.Notify(948)
		return ""
	}(); s != "" {
		__antithesis_instrumentation__.Notify(952)
		return s
	} else {
		__antithesis_instrumentation__.Notify(953)
	}
	__antithesis_instrumentation__.Notify(945)

	advAddr := readFileOrEmpty(n.advertiseAddrFile())
	if advAddr == "" {
		__antithesis_instrumentation__.Notify(954)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(955)
	}
	__antithesis_instrumentation__.Notify(946)
	_, p, _ := net.SplitHostPort(advAddr)
	return p
}

func (n *Node) RPCAddr() string {
	__antithesis_instrumentation__.Notify(956)
	port := n.RPCPort()
	if port == "" || func() bool {
		__antithesis_instrumentation__.Notify(958)
		return port == "0" == true
	}() == true {
		__antithesis_instrumentation__.Notify(959)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(960)
	}
	__antithesis_instrumentation__.Notify(957)
	return net.JoinHostPort(n.IPAddr(), port)
}

func (n *Node) HTTPAddr() string {
	__antithesis_instrumentation__.Notify(961)
	return readFileOrEmpty(n.httpAddrFile())
}

func (n *Node) PGUrl() string {
	__antithesis_instrumentation__.Notify(962)
	n.Lock()
	defer n.Unlock()
	return n.pgURL
}

func (n *Node) Alive() bool {
	__antithesis_instrumentation__.Notify(963)
	n.Lock()
	defer n.Unlock()
	return n.cmd != nil
}

func (n *Node) StatusClient() serverpb.StatusClient {
	__antithesis_instrumentation__.Notify(964)
	n.Lock()
	existingClient := n.statusClient
	n.Unlock()

	if existingClient != nil {
		__antithesis_instrumentation__.Notify(967)
		return existingClient
	} else {
		__antithesis_instrumentation__.Notify(968)
	}
	__antithesis_instrumentation__.Notify(965)

	conn, _, err := n.rpcCtx.GRPCDialRaw(n.RPCAddr())
	if err != nil {
		__antithesis_instrumentation__.Notify(969)
		log.Fatalf(context.Background(), "failed to initialize status client: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(970)
	}
	__antithesis_instrumentation__.Notify(966)
	return serverpb.NewStatusClient(conn)
}

func (n *Node) logDir() string {
	__antithesis_instrumentation__.Notify(971)
	if n.Cfg.LogDir == "" {
		__antithesis_instrumentation__.Notify(973)
		return filepath.Join(n.Cfg.DataDir, "logs")
	} else {
		__antithesis_instrumentation__.Notify(974)
	}
	__antithesis_instrumentation__.Notify(972)
	return n.Cfg.LogDir
}

func (n *Node) listeningURLFile() string {
	__antithesis_instrumentation__.Notify(975)
	return filepath.Join(n.Cfg.DataDir, listeningURLFile)
}

func (n *Node) Start(ctx context.Context, joins ...string) {
	__antithesis_instrumentation__.Notify(976)
	if err := <-n.StartAsync(ctx, joins...); err != nil {
		__antithesis_instrumentation__.Notify(977)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(978)
	}
}

func (n *Node) setNotRunningLocked(waitErr *exec.ExitError) {
	__antithesis_instrumentation__.Notify(979)
	_ = os.Remove(n.listeningURLFile())
	_ = os.Remove(n.advertiseAddrFile())
	_ = os.Remove(n.httpAddrFile())
	if n.notRunning != nil {
		__antithesis_instrumentation__.Notify(981)
		close(n.notRunning)
	} else {
		__antithesis_instrumentation__.Notify(982)
	}
	__antithesis_instrumentation__.Notify(980)
	n.notRunning = make(chan struct{})
	n.db = nil
	n.statusClient = nil
	n.cmd = nil
	n.rpcPort = ""
	n.waitErr.Store(waitErr)
	atomic.StoreInt32(&n.startSeq, 0)
}

func (n *Node) startAsyncInnerLocked(ctx context.Context, joins ...string) error {
	__antithesis_instrumentation__.Notify(983)
	n.setNotRunningLocked(nil)

	args := append([]string(nil), n.Cfg.ExtraArgs[1:]...)
	for _, join := range joins {
		__antithesis_instrumentation__.Notify(990)
		args = append(args, "--join", join)
	}
	__antithesis_instrumentation__.Notify(984)
	n.cmd = exec.Command(n.Cfg.ExtraArgs[0], args...)
	n.cmd.Env = os.Environ()
	n.cmd.Env = append(n.cmd.Env, "COCKROACH_SCAN_MAX_IDLE_TIME=5ms")
	n.cmd.Env = append(n.cmd.Env, n.Cfg.ExtraEnv...)

	atomic.StoreInt32(&n.startSeq, n.seq.Next())

	_ = os.MkdirAll(n.logDir(), 0755)

	stdoutPath := filepath.Join(n.logDir(), "stdout")
	stdout, err := os.OpenFile(stdoutPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		__antithesis_instrumentation__.Notify(991)
		return errors.Wrapf(err, "unable to open file %s", stdoutPath)
	} else {
		__antithesis_instrumentation__.Notify(992)
	}
	__antithesis_instrumentation__.Notify(985)

	n.cmd.Stdout = io.MultiWriter(stdout, os.Stdout)

	stderrPath := filepath.Join(n.logDir(), "stderr")
	stderr, err := os.OpenFile(stderrPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		__antithesis_instrumentation__.Notify(993)
		return errors.Wrapf(err, "unable to open file %s", stderrPath)
	} else {
		__antithesis_instrumentation__.Notify(994)
	}
	__antithesis_instrumentation__.Notify(986)
	n.cmd.Stderr = stderr

	if n.Cfg.RPCPort > 0 {
		__antithesis_instrumentation__.Notify(995)
		n.rpcPort = fmt.Sprintf("%d", n.Cfg.RPCPort)
	} else {
		__antithesis_instrumentation__.Notify(996)
	}
	__antithesis_instrumentation__.Notify(987)

	if err := n.cmd.Start(); err != nil {
		__antithesis_instrumentation__.Notify(997)
		if err := stdout.Close(); err != nil {
			__antithesis_instrumentation__.Notify(1000)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(1001)
		}
		__antithesis_instrumentation__.Notify(998)
		if err := stderr.Close(); err != nil {
			__antithesis_instrumentation__.Notify(1002)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(1003)
		}
		__antithesis_instrumentation__.Notify(999)
		return errors.Wrapf(err, "running %s %v", n.cmd.Path, n.cmd.Args)
	} else {
		__antithesis_instrumentation__.Notify(1004)
	}
	__antithesis_instrumentation__.Notify(988)

	log.Infof(ctx, "process %d starting: %s", n.cmd.Process.Pid, n.cmd.Args)

	go func(cmd *exec.Cmd) {
		__antithesis_instrumentation__.Notify(1005)
		waitErr := cmd.Wait()
		if waitErr != nil {
			__antithesis_instrumentation__.Notify(1009)
			log.Warningf(ctx, "%v", waitErr)
		} else {
			__antithesis_instrumentation__.Notify(1010)
		}
		__antithesis_instrumentation__.Notify(1006)
		if err := stdout.Close(); err != nil {
			__antithesis_instrumentation__.Notify(1011)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(1012)
		}
		__antithesis_instrumentation__.Notify(1007)
		if err := stderr.Close(); err != nil {
			__antithesis_instrumentation__.Notify(1013)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(1014)
		}
		__antithesis_instrumentation__.Notify(1008)

		log.Infof(ctx, "process %d: %s", cmd.Process.Pid, cmd.ProcessState)

		var execErr *exec.ExitError
		_ = errors.As(waitErr, &execErr)
		n.Lock()
		n.setNotRunningLocked(execErr)
		n.Unlock()
	}(n.cmd)
	__antithesis_instrumentation__.Notify(989)

	return nil
}

func (n *Node) StartAsync(ctx context.Context, joins ...string) <-chan error {
	__antithesis_instrumentation__.Notify(1015)
	ch := make(chan error, 1)

	if err := func() error {
		__antithesis_instrumentation__.Notify(1018)
		n.Lock()
		defer n.Unlock()
		if n.cmd != nil {
			__antithesis_instrumentation__.Notify(1020)
			return errors.New("server is already running")
		} else {
			__antithesis_instrumentation__.Notify(1021)
		}
		__antithesis_instrumentation__.Notify(1019)
		return n.startAsyncInnerLocked(ctx, joins...)
	}(); err != nil {
		__antithesis_instrumentation__.Notify(1022)
		ch <- err
		return ch
	} else {
		__antithesis_instrumentation__.Notify(1023)
	}
	__antithesis_instrumentation__.Notify(1016)

	go func() {
		__antithesis_instrumentation__.Notify(1024)

		ch <- n.waitUntilLive(time.Minute)
	}()
	__antithesis_instrumentation__.Notify(1017)

	return ch
}

func portFromURL(rawURL string) (string, *url.URL, error) {
	__antithesis_instrumentation__.Notify(1025)
	u, err := url.Parse(rawURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(1027)
		return "", nil, err
	} else {
		__antithesis_instrumentation__.Notify(1028)
	}
	__antithesis_instrumentation__.Notify(1026)

	_, port, err := net.SplitHostPort(u.Host)
	return port, u, err
}

func makeDB(url string, numWorkers int, dbName string) *gosql.DB {
	__antithesis_instrumentation__.Notify(1029)
	conn, err := gosql.Open("postgres", url)
	if err != nil {
		__antithesis_instrumentation__.Notify(1032)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(1033)
	}
	__antithesis_instrumentation__.Notify(1030)
	if numWorkers == 0 {
		__antithesis_instrumentation__.Notify(1034)
		numWorkers = 1
	} else {
		__antithesis_instrumentation__.Notify(1035)
	}
	__antithesis_instrumentation__.Notify(1031)
	conn.SetMaxOpenConns(numWorkers)
	conn.SetMaxIdleConns(numWorkers)
	return conn
}

func (n *Node) advertiseAddrFile() string {
	__antithesis_instrumentation__.Notify(1036)
	return filepath.Join(n.Cfg.DataDir, "cockroach.advertise-addr")
}

func (n *Node) httpAddrFile() string {
	__antithesis_instrumentation__.Notify(1037)
	return filepath.Join(n.Cfg.DataDir, "cockroach.http-addr")
}

func readFileOrEmpty(f string) string {
	__antithesis_instrumentation__.Notify(1038)
	c, err := ioutil.ReadFile(f)
	if err != nil {
		__antithesis_instrumentation__.Notify(1040)
		if !oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(1042)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(1043)
		}
		__antithesis_instrumentation__.Notify(1041)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(1044)
	}
	__antithesis_instrumentation__.Notify(1039)
	return string(c)
}

func (n *Node) AdvertiseAddr() (s string) {
	__antithesis_instrumentation__.Notify(1045)
	addr := readFileOrEmpty(n.advertiseAddrFile())
	if addr != "" {
		__antithesis_instrumentation__.Notify(1048)
		return addr
	} else {
		__antithesis_instrumentation__.Notify(1049)
	}
	__antithesis_instrumentation__.Notify(1046)

	if port := n.RPCPort(); port != "" {
		__antithesis_instrumentation__.Notify(1050)
		return net.JoinHostPort(n.IPAddr(), n.RPCPort())
	} else {
		__antithesis_instrumentation__.Notify(1051)
	}
	__antithesis_instrumentation__.Notify(1047)
	return addr
}

func (n *Node) waitUntilLive(dur time.Duration) error {
	__antithesis_instrumentation__.Notify(1052)
	ctx := context.Background()
	closer := make(chan struct{})
	defer time.AfterFunc(dur, func() { __antithesis_instrumentation__.Notify(1055); close(closer) }).Stop()
	__antithesis_instrumentation__.Notify(1053)
	opts := retry.Options{
		InitialBackoff: time.Millisecond,
		MaxBackoff:     500 * time.Millisecond,
		Multiplier:     2,
		Closer:         closer,
	}
	for r := retry.Start(opts); r.Next(); {
		__antithesis_instrumentation__.Notify(1056)
		var pid int
		n.Lock()
		if n.cmd != nil {
			__antithesis_instrumentation__.Notify(1064)
			pid = n.cmd.Process.Pid
		} else {
			__antithesis_instrumentation__.Notify(1065)
		}
		__antithesis_instrumentation__.Notify(1057)
		n.Unlock()
		if pid == 0 {
			__antithesis_instrumentation__.Notify(1066)
			log.Info(ctx, "process already quit")
			return nil
		} else {
			__antithesis_instrumentation__.Notify(1067)
		}
		__antithesis_instrumentation__.Notify(1058)

		urlBytes, err := ioutil.ReadFile(n.listeningURLFile())
		if err != nil {
			__antithesis_instrumentation__.Notify(1068)
			log.Infof(ctx, "%v", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(1069)
		}
		__antithesis_instrumentation__.Notify(1059)

		var pgURL *url.URL
		_, pgURL, err = portFromURL(string(urlBytes))
		if err != nil {
			__antithesis_instrumentation__.Notify(1070)
			log.Infof(ctx, "%v", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(1071)
		}
		__antithesis_instrumentation__.Notify(1060)

		if n.Cfg.RPCPort == 0 {
			__antithesis_instrumentation__.Notify(1072)
			n.Lock()
			n.rpcPort = pgURL.Port()
			n.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(1073)
		}
		__antithesis_instrumentation__.Notify(1061)

		pgURL.Path = n.Cfg.DB
		n.Lock()
		n.pgURL = pgURL.String()
		n.Unlock()

		var uiURL *url.URL

		defer func() {
			__antithesis_instrumentation__.Notify(1074)
			log.Infof(ctx, "process %d started (db: %s ui: %s)", pid, pgURL, uiURL)
		}()
		__antithesis_instrumentation__.Notify(1062)

		n.Lock()
		n.db = makeDB(n.pgURL, n.Cfg.NumWorkers, n.Cfg.DB)
		n.Unlock()

		{
			__antithesis_instrumentation__.Notify(1075)
			var uiStr string
			if err := n.db.QueryRow(
				`SELECT value FROM crdb_internal.node_runtime_info WHERE component='UI' AND field = 'URL'`,
			).Scan(&uiStr); err != nil {
				__antithesis_instrumentation__.Notify(1077)
				log.Infof(ctx, "%v", err)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(1078)
			}
			__antithesis_instrumentation__.Notify(1076)

			_, uiURL, err = portFromURL(uiStr)
			if err != nil {
				__antithesis_instrumentation__.Notify(1079)
				log.Infof(ctx, "%v", err)

			} else {
				__antithesis_instrumentation__.Notify(1080)
			}
		}
		__antithesis_instrumentation__.Notify(1063)
		return nil
	}
	__antithesis_instrumentation__.Notify(1054)
	return errors.Errorf("node %+v was unable to join cluster within %s", n.Cfg, dur)
}

func (n *Node) Kill() {
	__antithesis_instrumentation__.Notify(1081)
	n.Signal(os.Kill)

	for ok := false; !ok; {
		__antithesis_instrumentation__.Notify(1082)
		n.Lock()
		ok = n.cmd == nil
		n.Unlock()
	}
}

func (n *Node) IPAddr() string {
	__antithesis_instrumentation__.Notify(1083)
	return n.Cfg.Addr
}

func (n *Node) DB() *gosql.DB {
	__antithesis_instrumentation__.Notify(1084)
	n.Lock()
	defer n.Unlock()
	return n.db
}

func (n *Node) Signal(s os.Signal) {
	__antithesis_instrumentation__.Notify(1085)
	n.Lock()
	defer n.Unlock()
	if n.cmd == nil || func() bool {
		__antithesis_instrumentation__.Notify(1087)
		return n.cmd.Process == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(1088)
		return
	} else {
		__antithesis_instrumentation__.Notify(1089)
	}
	__antithesis_instrumentation__.Notify(1086)
	if err := n.cmd.Process.Signal(s); err != nil {
		__antithesis_instrumentation__.Notify(1090)
		log.Warningf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(1091)
	}
}

func (n *Node) Wait() *exec.ExitError {
	__antithesis_instrumentation__.Notify(1092)
	n.Lock()
	ch := n.notRunning
	n.Unlock()
	if ch == nil {
		__antithesis_instrumentation__.Notify(1094)
		log.Warning(context.Background(), "(*Node).Wait called when node was not running")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(1095)
	}
	__antithesis_instrumentation__.Notify(1093)
	<-ch
	ee, _ := n.waitErr.Load().(*exec.ExitError)
	return ee
}

var _ = (*Node)(nil).Wait
