package cluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logflags"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/containerd/containerd/platforms"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	defaultImage  = "docker.io/library/ubuntu:focal-20210119"
	networkPrefix = "cockroachdb_acceptance"
)

const DefaultTCP nat.Port = base.DefaultPort + "/tcp"
const defaultHTTP nat.Port = base.DefaultHTTPPort + "/tcp"

const CockroachBinaryInContainer = "/cockroach/cockroach"

var cockroachImage = flag.String("i", defaultImage, "the docker image to run")
var cockroachEntry = flag.String("e", "", "the entry point for the image")
var waitOnStop = flag.Bool("w", false, "wait for the user to interrupt before tearing down the cluster")
var maxRangeBytes = *zonepb.DefaultZoneConfig().RangeMaxBytes

var CockroachBinary = flag.String("b", "", "the host-side binary to run")

func exists(path string) bool {
	__antithesis_instrumentation__.Notify(174)
	if _, err := os.Stat(path); oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(176)
		return false
	} else {
		__antithesis_instrumentation__.Notify(177)
	}
	__antithesis_instrumentation__.Notify(175)
	return true
}

func nodeStr(l *DockerCluster, i int) string {
	__antithesis_instrumentation__.Notify(178)
	return fmt.Sprintf("roach-%s-%d", l.clusterID, i)
}

func storeStr(node, store int) string {
	__antithesis_instrumentation__.Notify(179)
	return fmt.Sprintf("data%d.%d", node, store)
}

const (
	eventDie     = "die"
	eventRestart = "restart"
)

type Event struct {
	NodeIndex int
	Status    string
}

type testStore struct {
	index  int
	dir    string
	config StoreConfig
}

type testNode struct {
	*Container
	index   int
	nodeStr string
	config  NodeConfig
	stores  []testStore
}

type DockerCluster struct {
	client               client.APIClient
	mu                   syncutil.Mutex
	vols                 *Container
	config               TestConfig
	Nodes                []*testNode
	events               chan Event
	expectedEvents       chan Event
	oneshot              *Container
	stopper              *stop.Stopper
	monitorCtx           context.Context
	monitorCtxCancelFunc func()
	clusterID            string
	networkID            string
	networkName          string

	volumesDir string
}

func CreateDocker(
	ctx context.Context, cfg TestConfig, volumesDir string, stopper *stop.Stopper,
) *DockerCluster {
	__antithesis_instrumentation__.Notify(180)
	select {
	case <-stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(185)

		os.Exit(1)
	default:
		__antithesis_instrumentation__.Notify(186)
	}
	__antithesis_instrumentation__.Notify(181)

	if *cockroachImage == defaultImage && func() bool {
		__antithesis_instrumentation__.Notify(187)
		return !exists(*CockroachBinary) == true
	}() == true {
		__antithesis_instrumentation__.Notify(188)
		log.Fatalf(ctx, "\"%s\": does not exist", *CockroachBinary)
	} else {
		__antithesis_instrumentation__.Notify(189)
	}
	__antithesis_instrumentation__.Notify(182)

	cli, err := client.NewClientWithOpts(client.FromEnv)
	maybePanic(err)

	cli.NegotiateAPIVersion(ctx)

	clusterID := uuid.MakeV4()
	clusterIDS := clusterID.Short()

	if volumesDir == "" {
		__antithesis_instrumentation__.Notify(190)
		volumesDir, err = ioutil.TempDir("", fmt.Sprintf("cockroach-acceptance-%s", clusterIDS))
		maybePanic(err)
	} else {
		__antithesis_instrumentation__.Notify(191)
		volumesDir = filepath.Join(volumesDir, clusterIDS)
	}
	__antithesis_instrumentation__.Notify(183)
	if !filepath.IsAbs(volumesDir) {
		__antithesis_instrumentation__.Notify(192)
		pwd, err := os.Getwd()
		maybePanic(err)
		volumesDir = filepath.Join(pwd, volumesDir)
	} else {
		__antithesis_instrumentation__.Notify(193)
	}
	__antithesis_instrumentation__.Notify(184)
	maybePanic(os.MkdirAll(volumesDir, 0755))
	log.Infof(ctx, "cluster volume directory: %s", volumesDir)

	return &DockerCluster{
		clusterID: clusterIDS,
		client:    resilientDockerClient{APIClient: cli},
		config:    cfg,
		stopper:   stopper,

		events:         make(chan Event, 1000),
		expectedEvents: make(chan Event, 1000),
		volumesDir:     volumesDir,
	}
}

func (l *DockerCluster) expectEvent(c *Container, msgs ...string) {
	__antithesis_instrumentation__.Notify(194)
	for index, ctr := range l.Nodes {
		__antithesis_instrumentation__.Notify(195)
		if c.id != ctr.id {
			__antithesis_instrumentation__.Notify(198)
			continue
		} else {
			__antithesis_instrumentation__.Notify(199)
		}
		__antithesis_instrumentation__.Notify(196)
		for _, status := range msgs {
			__antithesis_instrumentation__.Notify(200)
			select {
			case l.expectedEvents <- Event{NodeIndex: index, Status: status}:
				__antithesis_instrumentation__.Notify(201)
			default:
				__antithesis_instrumentation__.Notify(202)
				panic("expectedEvents channel filled up")
			}
		}
		__antithesis_instrumentation__.Notify(197)
		break
	}
}

func (l *DockerCluster) OneShot(
	ctx context.Context,
	ref string,
	ipo types.ImagePullOptions,
	containerConfig container.Config,
	hostConfig container.HostConfig,
	platformSpec specs.Platform,
	name string,
) error {
	__antithesis_instrumentation__.Notify(203)
	if err := pullImage(ctx, l, ref, ipo); err != nil {
		__antithesis_instrumentation__.Notify(208)
		return err
	} else {
		__antithesis_instrumentation__.Notify(209)
	}
	__antithesis_instrumentation__.Notify(204)
	hostConfig.VolumesFrom = []string{l.vols.id}
	c, err := createContainer(ctx, l, containerConfig, hostConfig, platformSpec, name)
	if err != nil {
		__antithesis_instrumentation__.Notify(210)
		return err
	} else {
		__antithesis_instrumentation__.Notify(211)
	}
	__antithesis_instrumentation__.Notify(205)
	l.oneshot = c
	defer func() {
		__antithesis_instrumentation__.Notify(212)
		if err := l.oneshot.Remove(ctx); err != nil {
			__antithesis_instrumentation__.Notify(214)
			log.Errorf(ctx, "ContainerRemove: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(215)
		}
		__antithesis_instrumentation__.Notify(213)
		l.oneshot = nil
	}()
	__antithesis_instrumentation__.Notify(206)

	if err := l.oneshot.Start(ctx); err != nil {
		__antithesis_instrumentation__.Notify(216)
		return err
	} else {
		__antithesis_instrumentation__.Notify(217)
	}
	__antithesis_instrumentation__.Notify(207)
	return l.oneshot.Wait(ctx, container.WaitConditionNotRunning)
}

func (l *DockerCluster) stopOnPanic(ctx context.Context) {
	__antithesis_instrumentation__.Notify(218)
	if r := recover(); r != nil {
		__antithesis_instrumentation__.Notify(219)
		l.stop(ctx)
		if r != l {
			__antithesis_instrumentation__.Notify(221)
			panic(r)
		} else {
			__antithesis_instrumentation__.Notify(222)
		}
		__antithesis_instrumentation__.Notify(220)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(223)
	}
}

func (l *DockerCluster) panicOnStop() {
	__antithesis_instrumentation__.Notify(224)
	if l.stopper == nil {
		__antithesis_instrumentation__.Notify(226)
		panic(l)
	} else {
		__antithesis_instrumentation__.Notify(227)
	}
	__antithesis_instrumentation__.Notify(225)

	select {
	case <-l.stopper.IsStopped():
		__antithesis_instrumentation__.Notify(228)
		l.stopper = nil
		panic(l)
	default:
		__antithesis_instrumentation__.Notify(229)
	}
}

func (l *DockerCluster) createNetwork(ctx context.Context) {
	__antithesis_instrumentation__.Notify(230)
	l.panicOnStop()

	l.networkName = fmt.Sprintf("%s-%s", networkPrefix, l.clusterID)
	log.Infof(ctx, "creating docker network with name: %s", l.networkName)
	net, err := l.client.NetworkInspect(ctx, l.networkName, types.NetworkInspectOptions{})
	if err == nil {
		__antithesis_instrumentation__.Notify(233)

		for containerID := range net.Containers {
			__antithesis_instrumentation__.Notify(235)

			maybePanic(l.client.ContainerKill(ctx, containerID, "9"))
		}
		__antithesis_instrumentation__.Notify(234)
		maybePanic(l.client.NetworkRemove(ctx, l.networkName))
	} else {
		__antithesis_instrumentation__.Notify(236)
		if !client.IsErrNotFound(err) {
			__antithesis_instrumentation__.Notify(237)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(238)
		}
	}
	__antithesis_instrumentation__.Notify(231)

	resp, err := l.client.NetworkCreate(ctx, l.networkName, types.NetworkCreate{
		Driver: "bridge",

		CheckDuplicate: true,
	})
	maybePanic(err)
	if resp.Warning != "" {
		__antithesis_instrumentation__.Notify(239)
		log.Warningf(ctx, "creating network: %s", resp.Warning)
	} else {
		__antithesis_instrumentation__.Notify(240)
	}
	__antithesis_instrumentation__.Notify(232)
	l.networkID = resp.ID
}

func (l *DockerCluster) initCluster(ctx context.Context) {
	__antithesis_instrumentation__.Notify(241)
	configJSON, err := json.Marshal(l.config)
	maybePanic(err)
	log.Infof(ctx, "Initializing Cluster %s:\n%s", l.config.Name, configJSON)
	l.panicOnStop()

	pwd, err := os.Getwd()
	maybePanic(err)

	binds := []string{
		filepath.Join(pwd, certsDir) + ":/certs",
		filepath.Join(pwd, "..") + ":/go/src/github.com/cockroachdb/cockroach",
		filepath.Join(l.volumesDir, "logs") + ":/logs",
	}

	if *cockroachImage == defaultImage {
		__antithesis_instrumentation__.Notify(245)
		path, err := filepath.Abs(*CockroachBinary)
		maybePanic(err)
		binds = append(binds, path+":"+CockroachBinaryInContainer)
	} else {
		__antithesis_instrumentation__.Notify(246)
	}
	__antithesis_instrumentation__.Notify(242)

	l.Nodes = []*testNode{}

	for i, nc := range l.config.Nodes {
		__antithesis_instrumentation__.Notify(247)
		newTestNode := &testNode{
			config:  nc,
			index:   i,
			nodeStr: nodeStr(l, i),
		}
		for j, sc := range nc.Stores {
			__antithesis_instrumentation__.Notify(249)
			hostDir := filepath.Join(l.volumesDir, storeStr(i, j))
			containerDir := "/" + storeStr(i, j)
			binds = append(binds, hostDir+":"+containerDir)
			newTestNode.stores = append(newTestNode.stores,
				testStore{
					config: sc,
					index:  j,
					dir:    containerDir,
				})
		}
		__antithesis_instrumentation__.Notify(248)
		l.Nodes = append(l.Nodes, newTestNode)
	}
	__antithesis_instrumentation__.Notify(243)

	if *cockroachImage == defaultImage {
		__antithesis_instrumentation__.Notify(250)
		maybePanic(pullImage(ctx, l, defaultImage, types.ImagePullOptions{}))
	} else {
		__antithesis_instrumentation__.Notify(251)
	}
	__antithesis_instrumentation__.Notify(244)
	c, err := createContainer(
		ctx,
		l,
		container.Config{
			Image:      *cockroachImage,
			Entrypoint: []string{"/bin/true"},
		}, container.HostConfig{
			Binds:           binds,
			PublishAllPorts: true,
		},
		platforms.DefaultSpec(),
		fmt.Sprintf("volumes-%s", l.clusterID),
	)
	maybePanic(err)

	l.vols = c
	maybePanic(c.Start(ctx))
	maybePanic(c.Wait(ctx, container.WaitConditionNotRunning))
}

func cockroachEntrypoint() []string {
	__antithesis_instrumentation__.Notify(252)
	var entrypoint []string
	if *cockroachImage == defaultImage {
		__antithesis_instrumentation__.Notify(254)
		entrypoint = append(entrypoint, CockroachBinaryInContainer)
	} else {
		__antithesis_instrumentation__.Notify(255)
		if *cockroachEntry != "" {
			__antithesis_instrumentation__.Notify(256)
			entrypoint = append(entrypoint, *cockroachEntry)
		} else {
			__antithesis_instrumentation__.Notify(257)
		}
	}
	__antithesis_instrumentation__.Notify(253)
	return entrypoint
}

func (l *DockerCluster) createRoach(
	ctx context.Context, node *testNode, vols *Container, env []string, cmd ...string,
) {
	__antithesis_instrumentation__.Notify(258)
	l.panicOnStop()

	hostConfig := container.HostConfig{
		PublishAllPorts: true,
		NetworkMode:     container.NetworkMode(l.networkID),
	}

	if vols != nil {
		__antithesis_instrumentation__.Notify(261)
		hostConfig.VolumesFrom = append(hostConfig.VolumesFrom, vols.id)
	} else {
		__antithesis_instrumentation__.Notify(262)
	}
	__antithesis_instrumentation__.Notify(259)

	var hostname string
	if node.index >= 0 {
		__antithesis_instrumentation__.Notify(263)
		hostname = fmt.Sprintf("roach-%s-%d", l.clusterID, node.index)
	} else {
		__antithesis_instrumentation__.Notify(264)
	}
	__antithesis_instrumentation__.Notify(260)
	log.Infof(ctx, "creating docker container with name: %s", hostname)
	var err error
	node.Container, err = createContainer(
		ctx,
		l,
		container.Config{
			Hostname: hostname,
			Image:    *cockroachImage,
			ExposedPorts: map[nat.Port]struct{}{
				DefaultTCP:  {},
				defaultHTTP: {},
			},
			Entrypoint: cockroachEntrypoint(),
			Env:        env,
			Cmd:        cmd,
			Labels: map[string]string{

				"Hostname":              hostname,
				"Roach":                 "",
				"Acceptance-cluster-id": l.clusterID,
			},
		},
		hostConfig,
		platforms.DefaultSpec(),
		node.nodeStr,
	)
	maybePanic(err)
}

func (l *DockerCluster) createNodeCerts() {
	__antithesis_instrumentation__.Notify(265)

	nodes := []string{"localhost", "cockroach", dockerIP().String()}
	for _, node := range l.Nodes {
		__antithesis_instrumentation__.Notify(267)
		nodes = append(nodes, node.nodeStr)
	}
	__antithesis_instrumentation__.Notify(266)
	maybePanic(security.CreateNodePair(
		certsDir,
		filepath.Join(certsDir, security.EmbeddedCAKey),
		keyLen, 48*time.Hour, true, nodes))
}

func (l *DockerCluster) startNode(ctx context.Context, node *testNode) {
	__antithesis_instrumentation__.Notify(268)
	cmd := []string{
		"start",
		"--certs-dir=/certs/",
		"--listen-addr=" + node.nodeStr,
		"--vmodule=*=1",
	}

	vmoduleFlag := flag.Lookup(logflags.VModuleName)
	if vmoduleFlag.Value.String() != "" {
		__antithesis_instrumentation__.Notify(271)
		cmd = append(cmd, fmt.Sprintf("--%s=%s", vmoduleFlag.Name, vmoduleFlag.Value.String()))
	} else {
		__antithesis_instrumentation__.Notify(272)
	}
	__antithesis_instrumentation__.Notify(269)

	for _, store := range node.stores {
		__antithesis_instrumentation__.Notify(273)
		storeSpec := base.StoreSpec{
			Path: store.dir,
			Size: base.SizeSpec{InBytes: int64(store.config.MaxRanges) * maxRangeBytes},
		}
		cmd = append(cmd, fmt.Sprintf("--store=%s", storeSpec))
	}
	__antithesis_instrumentation__.Notify(270)

	firstNodeAddr := l.Nodes[0].nodeStr
	cmd = append(cmd, "--join="+net.JoinHostPort(firstNodeAddr, base.DefaultPort))

	dockerLogDir := "/logs/" + node.nodeStr
	localLogDir := filepath.Join(l.volumesDir, "logs", node.nodeStr)
	cmd = append(
		cmd,
		"--logtostderr=ERROR",
		"--log-dir="+dockerLogDir)
	env := []string{
		"COCKROACH_SCAN_MAX_IDLE_TIME=200ms",
		"COCKROACH_SKIP_UPDATE_CHECK=1",
		"COCKROACH_CRASH_REPORTS=",
	}
	l.createRoach(ctx, node, l.vols, env, cmd...)
	maybePanic(node.Start(ctx))
	httpAddr := node.Addr(ctx, defaultHTTP)

	log.Infof(ctx, `*** started %[1]s ***
  ui:        %[2]s
  trace:     %[2]s/debug/requests
  logs:      %[3]s/cockroach.INFO
  pprof:     docker exec -it %[4]s pprof https+insecure://$(hostname):%[5]s/debug/pprof/heap
  cockroach: %[6]s

  cli-env:   COCKROACH_INSECURE=false COCKROACH_CERTS_DIR=%[7]s COCKROACH_HOST=%s:%d`,
		node.Name(), "https://"+httpAddr.String(), localLogDir, node.Container.id[:5],
		base.DefaultHTTPPort, cmd, certsDir, httpAddr.IP, httpAddr.Port)
}

func (l *DockerCluster) RunInitCommand(ctx context.Context, nodeIdx int) {
	__antithesis_instrumentation__.Notify(274)
	containerConfig := container.Config{
		Image:      *cockroachImage,
		Entrypoint: cockroachEntrypoint(),
		Cmd: []string{
			"init",
			"--certs-dir=/certs/",
			"--host=" + l.Nodes[nodeIdx].nodeStr,
			"--log-dir=/logs/init-command",
			"--logtostderr=NONE",
		},
	}

	log.Infof(ctx, "trying to initialize via %v", containerConfig.Cmd)
	maybePanic(l.OneShot(ctx, defaultImage, types.ImagePullOptions{},
		containerConfig, container.HostConfig{}, platforms.DefaultSpec(), "init-command"))
	log.Info(ctx, "cluster successfully initialized")
}

func (l *DockerCluster) processEvent(ctx context.Context, event events.Message) bool {
	__antithesis_instrumentation__.Notify(275)
	l.mu.Lock()
	defer l.mu.Unlock()

	log.Infof(ctx, "processing event from Docker: %+v", event)

	if l.oneshot != nil && func() bool {
		__antithesis_instrumentation__.Notify(279)
		return event.ID == l.oneshot.id == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(280)
		return event.Status == eventDie == true
	}() == true {
		__antithesis_instrumentation__.Notify(281)
		return true
	} else {
		__antithesis_instrumentation__.Notify(282)
	}
	__antithesis_instrumentation__.Notify(276)

	for i, n := range l.Nodes {
		__antithesis_instrumentation__.Notify(283)
		if n != nil && func() bool {
			__antithesis_instrumentation__.Notify(284)
			return n.id == event.ID == true
		}() == true {
			__antithesis_instrumentation__.Notify(285)
			if log.V(1) {
				__antithesis_instrumentation__.Notify(288)
				log.Errorf(ctx, "node=%d status=%s", i, event.Status)
			} else {
				__antithesis_instrumentation__.Notify(289)
			}
			__antithesis_instrumentation__.Notify(286)
			select {
			case l.events <- Event{NodeIndex: i, Status: event.Status}:
				__antithesis_instrumentation__.Notify(290)
			default:
				__antithesis_instrumentation__.Notify(291)
				panic("events channel filled up")
			}
			__antithesis_instrumentation__.Notify(287)
			return true
		} else {
			__antithesis_instrumentation__.Notify(292)
		}
	}
	__antithesis_instrumentation__.Notify(277)

	log.Infof(ctx, "received docker event for unrecognized container: %+v",
		event)

	select {
	case <-l.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(293)
	case <-l.monitorCtx.Done():
		__antithesis_instrumentation__.Notify(294)
	default:
		__antithesis_instrumentation__.Notify(295)

		log.Errorf(ctx, "stopping due to unexpected event: %+v", event)
		if rc, err := l.client.ContainerLogs(context.Background(), event.Actor.ID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		}); err == nil {
			__antithesis_instrumentation__.Notify(296)
			defer rc.Close()
			if _, err := io.Copy(os.Stderr, rc); err != nil {
				__antithesis_instrumentation__.Notify(297)
				log.Infof(ctx, "error listing logs: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(298)
			}
		} else {
			__antithesis_instrumentation__.Notify(299)
		}
	}
	__antithesis_instrumentation__.Notify(278)
	return false
}

func (l *DockerCluster) monitor(ctx context.Context) {
	__antithesis_instrumentation__.Notify(300)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(303)
		log.Infof(ctx, "events monitor starts")
		defer log.Infof(ctx, "events monitor exits")
	} else {
		__antithesis_instrumentation__.Notify(304)
	}
	__antithesis_instrumentation__.Notify(301)
	longPoll := func() bool {
		__antithesis_instrumentation__.Notify(305)

		if l.monitorCtx.Err() != nil {
			__antithesis_instrumentation__.Notify(307)
			return false
		} else {
			__antithesis_instrumentation__.Notify(308)
		}
		__antithesis_instrumentation__.Notify(306)

		eventq, errq := l.client.Events(l.monitorCtx, types.EventsOptions{
			Filters: filters.NewArgs(
				filters.Arg("label", "Acceptance-cluster-id="+l.clusterID),
			),
		})
		for {
			__antithesis_instrumentation__.Notify(309)
			select {
			case err := <-errq:
				__antithesis_instrumentation__.Notify(310)
				log.Infof(ctx, "event stream done, resetting...: %s", err)

				return true
			case event := <-eventq:
				__antithesis_instrumentation__.Notify(311)

				switch event.Status {
				case eventDie, eventRestart:
					__antithesis_instrumentation__.Notify(312)
					if !l.processEvent(ctx, event) {
						__antithesis_instrumentation__.Notify(314)
						return false
					} else {
						__antithesis_instrumentation__.Notify(315)
					}
				default:
					__antithesis_instrumentation__.Notify(313)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(302)

	for longPoll() {
		__antithesis_instrumentation__.Notify(316)
	}
}

func (l *DockerCluster) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(317)
	defer l.stopOnPanic(ctx)

	l.mu.Lock()
	defer l.mu.Unlock()

	l.createNetwork(ctx)
	l.initCluster(ctx)
	log.Infof(ctx, "creating node certs (%dbit) in: %s", keyLen, certsDir)
	l.createNodeCerts()

	log.Infof(ctx, "starting %d nodes", len(l.Nodes))
	l.monitorCtx, l.monitorCtxCancelFunc = context.WithCancel(context.Background())
	go l.monitor(ctx)
	var wg sync.WaitGroup
	wg.Add(len(l.Nodes))
	for _, node := range l.Nodes {
		__antithesis_instrumentation__.Notify(319)
		go func(node *testNode) {
			__antithesis_instrumentation__.Notify(320)
			defer wg.Done()
			l.startNode(ctx, node)
		}(node)
	}
	__antithesis_instrumentation__.Notify(318)
	wg.Wait()

	if l.config.InitMode == INIT_COMMAND && func() bool {
		__antithesis_instrumentation__.Notify(321)
		return len(l.Nodes) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(322)
		l.RunInitCommand(ctx, 0)
	} else {
		__antithesis_instrumentation__.Notify(323)
	}
}

func (l *DockerCluster) Assert(ctx context.Context, t testing.TB) {
	__antithesis_instrumentation__.Notify(324)
	const almostZero = 50 * time.Millisecond
	filter := func(ch chan Event, wait time.Duration) *Event {
		__antithesis_instrumentation__.Notify(328)
		select {
		case act := <-ch:
			__antithesis_instrumentation__.Notify(330)
			return &act
		case <-time.After(wait):
			__antithesis_instrumentation__.Notify(331)
		}
		__antithesis_instrumentation__.Notify(329)
		return nil
	}
	__antithesis_instrumentation__.Notify(325)

	var events []Event
	for {
		__antithesis_instrumentation__.Notify(332)
		exp := filter(l.expectedEvents, almostZero)
		if exp == nil {
			__antithesis_instrumentation__.Notify(335)
			break
		} else {
			__antithesis_instrumentation__.Notify(336)
		}
		__antithesis_instrumentation__.Notify(333)
		act := filter(l.events, 15*time.Second)
		if act == nil || func() bool {
			__antithesis_instrumentation__.Notify(337)
			return *exp != *act == true
		}() == true {
			__antithesis_instrumentation__.Notify(338)
			t.Fatalf("expected event %v, got %v (after %v)", exp, act, events)
		} else {
			__antithesis_instrumentation__.Notify(339)
		}
		__antithesis_instrumentation__.Notify(334)
		events = append(events, *exp)
	}
	__antithesis_instrumentation__.Notify(326)
	if cur := filter(l.events, almostZero); cur != nil {
		__antithesis_instrumentation__.Notify(340)
		t.Fatalf("unexpected extra event %v (after %v)", cur, events)
	} else {
		__antithesis_instrumentation__.Notify(341)
	}
	__antithesis_instrumentation__.Notify(327)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(342)
		log.Infof(ctx, "asserted %v", events)
	} else {
		__antithesis_instrumentation__.Notify(343)
	}
}

func (l *DockerCluster) AssertAndStop(ctx context.Context, t testing.TB) {
	__antithesis_instrumentation__.Notify(344)
	defer l.stop(ctx)
	l.Assert(ctx, t)
}

func (l *DockerCluster) stop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(345)
	if *waitOnStop {
		__antithesis_instrumentation__.Notify(350)
		log.Infof(ctx, "waiting for interrupt")
		<-l.stopper.ShouldQuiesce()
	} else {
		__antithesis_instrumentation__.Notify(351)
	}
	__antithesis_instrumentation__.Notify(346)

	log.Infof(ctx, "stopping")

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.monitorCtxCancelFunc != nil {
		__antithesis_instrumentation__.Notify(352)
		l.monitorCtxCancelFunc()
		l.monitorCtxCancelFunc = nil
	} else {
		__antithesis_instrumentation__.Notify(353)
	}
	__antithesis_instrumentation__.Notify(347)

	if l.vols != nil {
		__antithesis_instrumentation__.Notify(354)
		maybePanic(l.vols.Kill(ctx))
		maybePanic(l.vols.Remove(ctx))
		l.vols = nil
	} else {
		__antithesis_instrumentation__.Notify(355)
	}
	__antithesis_instrumentation__.Notify(348)
	for i, n := range l.Nodes {
		__antithesis_instrumentation__.Notify(356)
		if n.Container == nil {
			__antithesis_instrumentation__.Notify(359)
			continue
		} else {
			__antithesis_instrumentation__.Notify(360)
		}
		__antithesis_instrumentation__.Notify(357)
		ci, err := n.Inspect(ctx)
		crashed := err != nil || func() bool {
			__antithesis_instrumentation__.Notify(361)
			return (!ci.State.Running && func() bool {
				__antithesis_instrumentation__.Notify(362)
				return ci.State.ExitCode != 0 == true
			}() == true) == true
		}() == true
		maybePanic(n.Kill(ctx))

		file := filepath.Join(l.volumesDir, "logs", nodeStr(l, i),
			fmt.Sprintf("stderr.%s.log", strings.Replace(
				timeutil.Now().Format(time.RFC3339), ":", "_", -1)))
		maybePanic(os.MkdirAll(filepath.Dir(file), 0755))
		w, err := os.Create(file)
		maybePanic(err)
		defer w.Close()
		maybePanic(n.Logs(ctx, w))
		log.Infof(ctx, "node %d: stderr at %s", i, file)
		if crashed {
			__antithesis_instrumentation__.Notify(363)
			log.Infof(ctx, "~~~ node %d CRASHED ~~~~", i)
		} else {
			__antithesis_instrumentation__.Notify(364)
		}
		__antithesis_instrumentation__.Notify(358)
		maybePanic(n.Remove(ctx))
	}
	__antithesis_instrumentation__.Notify(349)
	l.Nodes = nil

	if l.networkID != "" {
		__antithesis_instrumentation__.Notify(365)
		maybePanic(
			l.client.NetworkRemove(ctx, l.networkID))
		l.networkID = ""
		l.networkName = ""
	} else {
		__antithesis_instrumentation__.Notify(366)
	}
}

func (l *DockerCluster) NewDB(ctx context.Context, i int) (*gosql.DB, error) {
	__antithesis_instrumentation__.Notify(367)
	return gosql.Open("postgres", l.PGUrl(ctx, i))
}

func (l *DockerCluster) InternalIP(ctx context.Context, i int) net.IP {
	__antithesis_instrumentation__.Notify(368)
	c := l.Nodes[i]
	containerInfo, err := c.Inspect(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(370)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(371)
	}
	__antithesis_instrumentation__.Notify(369)
	return net.ParseIP(containerInfo.NetworkSettings.Networks[l.networkName].IPAddress)
}

func (l *DockerCluster) PGUrl(ctx context.Context, i int) string {
	__antithesis_instrumentation__.Notify(372)
	certUser := security.RootUser
	options := url.Values{}
	options.Add("sslmode", "verify-full")
	options.Add("sslcert", filepath.Join(certsDir, security.EmbeddedRootCert))
	options.Add("sslkey", filepath.Join(certsDir, security.EmbeddedRootKey))
	options.Add("sslrootcert", filepath.Join(certsDir, security.EmbeddedCACert))
	pgURL := url.URL{
		Scheme:   "postgres",
		User:     url.User(certUser),
		Host:     l.Nodes[i].Addr(ctx, DefaultTCP).String(),
		RawQuery: options.Encode(),
	}
	return pgURL.String()
}

func (l *DockerCluster) NumNodes() int {
	__antithesis_instrumentation__.Notify(373)
	return len(l.Nodes)
}

func (l *DockerCluster) Kill(ctx context.Context, i int) error {
	__antithesis_instrumentation__.Notify(374)
	if err := l.Nodes[i].Kill(ctx); err != nil {
		__antithesis_instrumentation__.Notify(376)
		return errors.Wrapf(err, "failed to kill node %d", i)
	} else {
		__antithesis_instrumentation__.Notify(377)
	}
	__antithesis_instrumentation__.Notify(375)
	return nil
}

func (l *DockerCluster) Restart(ctx context.Context, i int) error {
	__antithesis_instrumentation__.Notify(378)

	if err := l.Nodes[i].Restart(ctx, nil); err != nil {
		__antithesis_instrumentation__.Notify(380)
		return errors.Wrapf(err, "failed to restart node %d", i)
	} else {
		__antithesis_instrumentation__.Notify(381)
	}
	__antithesis_instrumentation__.Notify(379)
	return nil
}

func (l *DockerCluster) URL(ctx context.Context, i int) string {
	__antithesis_instrumentation__.Notify(382)
	return "https://" + l.Nodes[i].Addr(ctx, defaultHTTP).String()
}

func (l *DockerCluster) Addr(ctx context.Context, i int, port string) string {
	__antithesis_instrumentation__.Notify(383)
	return l.Nodes[i].Addr(ctx, nat.Port(port+"/tcp")).String()
}

func (l *DockerCluster) Hostname(i int) string {
	__antithesis_instrumentation__.Notify(384)
	return l.Nodes[i].nodeStr
}

func (l *DockerCluster) ExecCLI(ctx context.Context, i int, cmd []string) (string, string, error) {
	__antithesis_instrumentation__.Notify(385)
	cmd = append([]string{CockroachBinaryInContainer}, cmd...)
	cmd = append(cmd, "--host", l.Hostname(i), "--certs-dir=/certs")
	cfg := types.ExecConfig{
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}
	createResp, err := l.client.ContainerExecCreate(ctx, l.Nodes[i].Container.id, cfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(388)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(389)
	}
	__antithesis_instrumentation__.Notify(386)
	var outputStream, errorStream bytes.Buffer
	{
		__antithesis_instrumentation__.Notify(390)
		resp, err := l.client.ContainerExecAttach(ctx, createResp.ID, types.ExecStartCheck{})
		if err != nil {
			__antithesis_instrumentation__.Notify(393)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(394)
		}
		__antithesis_instrumentation__.Notify(391)
		defer resp.Close()
		ch := make(chan error)
		go func() {
			__antithesis_instrumentation__.Notify(395)
			_, err := stdcopy.StdCopy(&outputStream, &errorStream, resp.Reader)
			ch <- err
		}()
		__antithesis_instrumentation__.Notify(392)
		if err := <-ch; err != nil {
			__antithesis_instrumentation__.Notify(396)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(397)
		}
	}
	{
		__antithesis_instrumentation__.Notify(398)
		resp, err := l.client.ContainerExecInspect(ctx, createResp.ID)
		if err != nil {
			__antithesis_instrumentation__.Notify(401)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(402)
		}
		__antithesis_instrumentation__.Notify(399)
		if resp.Running {
			__antithesis_instrumentation__.Notify(403)
			return "", "", errors.Errorf("command still running")
		} else {
			__antithesis_instrumentation__.Notify(404)
		}
		__antithesis_instrumentation__.Notify(400)
		if resp.ExitCode != 0 {
			__antithesis_instrumentation__.Notify(405)
			o, e := outputStream.String(), errorStream.String()
			return o, e, fmt.Errorf("error executing %s:\n%s\n%s",
				cmd, o, e)
		} else {
			__antithesis_instrumentation__.Notify(406)
		}
	}
	__antithesis_instrumentation__.Notify(387)
	return outputStream.String(), errorStream.String(), nil
}

func (l *DockerCluster) Cleanup(ctx context.Context, preserveLogs bool) {
	__antithesis_instrumentation__.Notify(407)
	volumes, err := ioutil.ReadDir(l.volumesDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(409)
		log.Warningf(ctx, "%v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(410)
	}
	__antithesis_instrumentation__.Notify(408)
	for _, v := range volumes {
		__antithesis_instrumentation__.Notify(411)
		if preserveLogs && func() bool {
			__antithesis_instrumentation__.Notify(413)
			return v.Name() == "logs" == true
		}() == true {
			__antithesis_instrumentation__.Notify(414)
			continue
		} else {
			__antithesis_instrumentation__.Notify(415)
		}
		__antithesis_instrumentation__.Notify(412)
		if err := os.RemoveAll(filepath.Join(l.volumesDir, v.Name())); err != nil {
			__antithesis_instrumentation__.Notify(416)
			log.Warningf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(417)
		}
	}
}
