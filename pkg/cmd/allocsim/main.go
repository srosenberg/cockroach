package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cockroachdb/cockroach/pkg/acceptance/localcluster"
	"github.com/cockroachdb/cockroach/pkg/acceptance/localcluster/tc"
	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var workers = flag.Int("w", 1, "number of workers; the i'th worker talks to node i%numNodes")
var numNodes = flag.Int("n", 4, "number of nodes")
var duration = flag.Duration("duration", math.MaxInt64, "how long to run the simulation for")
var blockSize = flag.Int("b", 1000, "block size")
var configFile = flag.String("f", "", "config file that specifies an allocsim workload (overrides -n)")

type Configuration struct {
	NumWorkers int        `json:"NumWorkers"`
	Localities []Locality `json:"Localities"`
}

type Locality struct {
	Name              string `json:"Name"`
	LocalityStr       string `json:"LocalityStr"`
	NumNodes          int    `json:"NumNodes"`
	NumWorkers        int    `json:"NumWorkers"`
	OutgoingLatencies []*struct {
		Name    string       `json:"Name"`
		Latency jsonDuration `json:"Latency"`
	} `json:"OutgoingLatencies"`
}

type jsonDuration time.Duration

func (j *jsonDuration) UnmarshalJSON(b []byte) error {
	__antithesis_instrumentation__.Notify(37350)
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		__antithesis_instrumentation__.Notify(37353)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37354)
	}
	__antithesis_instrumentation__.Notify(37351)
	dur, err := time.ParseDuration(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(37355)
		return err
	} else {
		__antithesis_instrumentation__.Notify(37356)
	}
	__antithesis_instrumentation__.Notify(37352)
	*j = jsonDuration(dur)
	return nil
}

func loadConfig(file string) (Configuration, error) {
	__antithesis_instrumentation__.Notify(37357)
	fileHandle, err := os.Open(file)
	if err != nil {
		__antithesis_instrumentation__.Notify(37361)
		return Configuration{}, errors.Wrapf(err, "failed to open config file %q", file)
	} else {
		__antithesis_instrumentation__.Notify(37362)
	}
	__antithesis_instrumentation__.Notify(37358)
	defer fileHandle.Close()

	var config Configuration
	jsonParser := json.NewDecoder(fileHandle)
	if err := jsonParser.Decode(&config); err != nil {
		__antithesis_instrumentation__.Notify(37363)
		return Configuration{}, errors.Wrapf(err, "failed to decode %q as json", file)
	} else {
		__antithesis_instrumentation__.Notify(37364)
	}
	__antithesis_instrumentation__.Notify(37359)

	*numNodes = 0
	*workers = config.NumWorkers
	for _, locality := range config.Localities {
		__antithesis_instrumentation__.Notify(37365)
		*numNodes += locality.NumNodes
		*workers += locality.NumWorkers
	}
	__antithesis_instrumentation__.Notify(37360)
	return config, nil
}

type allocSim struct {
	*localcluster.Cluster
	stats struct {
		ops               uint64
		totalLatencyNanos uint64
		errors            uint64
	}
	ranges struct {
		syncutil.Mutex
		stats allocStats
	}
	localities []Locality
}

type allocStats struct {
	count          int
	replicas       []int
	leases         []int
	replicaAdds    []int
	leaseTransfers []int
}

func newAllocSim(c *localcluster.Cluster) *allocSim {
	__antithesis_instrumentation__.Notify(37366)
	return &allocSim{
		Cluster: c,
	}
}

func (a *allocSim) run(workers int) {
	__antithesis_instrumentation__.Notify(37367)
	a.setup()
	for i := 0; i < workers; i++ {
		__antithesis_instrumentation__.Notify(37369)
		go a.roundRobinWorker(i, workers)
	}
	__antithesis_instrumentation__.Notify(37368)
	go a.rangeStats(time.Second)
	a.monitor(time.Second)
}

func (a *allocSim) runWithConfig(config Configuration) {
	__antithesis_instrumentation__.Notify(37370)
	a.setup()

	numWorkers := config.NumWorkers
	for _, locality := range config.Localities {
		__antithesis_instrumentation__.Notify(37374)
		numWorkers += locality.NumWorkers
	}
	__antithesis_instrumentation__.Notify(37371)

	firstNodeInLocality := 0
	for _, locality := range config.Localities {
		__antithesis_instrumentation__.Notify(37375)
		for i := 0; i < locality.NumWorkers; i++ {
			__antithesis_instrumentation__.Notify(37377)
			node := firstNodeInLocality + (i % locality.NumNodes)
			startNum := firstNodeInLocality + i
			go a.worker(node, startNum, numWorkers)
		}
		__antithesis_instrumentation__.Notify(37376)
		firstNodeInLocality += locality.NumNodes
	}
	__antithesis_instrumentation__.Notify(37372)
	for i := 0; i < config.NumWorkers; i++ {
		__antithesis_instrumentation__.Notify(37378)
		go a.roundRobinWorker(firstNodeInLocality+i, numWorkers)
	}
	__antithesis_instrumentation__.Notify(37373)

	go a.rangeStats(time.Second)
	a.monitor(time.Second)
}

func (a *allocSim) setup() {
	__antithesis_instrumentation__.Notify(37379)
	db := a.Nodes[0].DB()
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS allocsim"); err != nil {
		__antithesis_instrumentation__.Notify(37381)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(37382)
	}
	__antithesis_instrumentation__.Notify(37380)

	blocks := `
CREATE TABLE IF NOT EXISTS blocks (
  id INT NOT NULL,
  num INT NOT NULL,
  data BYTES NOT NULL,
  PRIMARY KEY (id, num)
)
`
	if _, err := db.Exec(blocks); err != nil {
		__antithesis_instrumentation__.Notify(37383)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(37384)
	}
}

func (a *allocSim) maybeLogError(err error) {
	__antithesis_instrumentation__.Notify(37385)
	if localcluster.IsUnavailableError(err) {
		__antithesis_instrumentation__.Notify(37387)
		return
	} else {
		__antithesis_instrumentation__.Notify(37388)
	}
	__antithesis_instrumentation__.Notify(37386)
	log.Errorf(context.Background(), "%v", err)
	atomic.AddUint64(&a.stats.errors, 1)
}

const insertStmt = `INSERT INTO allocsim.blocks (id, num, data) VALUES ($1, $2, repeat('a', $3)::bytes)`

func (a *allocSim) worker(dbIdx, startNum, workers int) {
	__antithesis_instrumentation__.Notify(37389)
	r, _ := randutil.NewPseudoRand()
	db := a.Nodes[dbIdx%len(a.Nodes)].DB()
	for num := startNum; true; num += workers {
		__antithesis_instrumentation__.Notify(37390)
		now := timeutil.Now()
		if _, err := db.Exec(insertStmt, r.Int63(), num, *blockSize); err != nil {
			__antithesis_instrumentation__.Notify(37391)
			a.maybeLogError(err)
		} else {
			__antithesis_instrumentation__.Notify(37392)
			atomic.AddUint64(&a.stats.ops, 1)
			atomic.AddUint64(&a.stats.totalLatencyNanos, uint64(timeutil.Since(now).Nanoseconds()))
		}
	}
}

func (a *allocSim) roundRobinWorker(startNum, workers int) {
	__antithesis_instrumentation__.Notify(37393)
	r, _ := randutil.NewPseudoRand()
	for i := 0; ; i++ {
		__antithesis_instrumentation__.Notify(37394)
		now := timeutil.Now()
		db := a.Nodes[i%len(a.Nodes)].DB()
		if db == nil {
			__antithesis_instrumentation__.Notify(37396)
			continue
		} else {
			__antithesis_instrumentation__.Notify(37397)
		}
		__antithesis_instrumentation__.Notify(37395)
		if _, err := db.Exec(insertStmt, r.Int63(), startNum+i*workers, *blockSize); err != nil {
			__antithesis_instrumentation__.Notify(37398)
			a.maybeLogError(err)
		} else {
			__antithesis_instrumentation__.Notify(37399)
			atomic.AddUint64(&a.stats.ops, 1)
			atomic.AddUint64(&a.stats.totalLatencyNanos, uint64(timeutil.Since(now).Nanoseconds()))
		}
	}
}

func (a *allocSim) rangeInfo() allocStats {
	__antithesis_instrumentation__.Notify(37400)
	stats := allocStats{
		replicas:       make([]int, len(a.Nodes)),
		replicaAdds:    make([]int, len(a.Nodes)),
		leases:         make([]int, len(a.Nodes)),
		leaseTransfers: make([]int, len(a.Nodes)),
	}

	var wg sync.WaitGroup
	wg.Add(len(a.Nodes))
	for i := 0; i < len(a.Nodes); i++ {
		__antithesis_instrumentation__.Notify(37403)
		go func(i int) {
			__antithesis_instrumentation__.Notify(37404)
			defer wg.Done()
			status := a.Nodes[i].StatusClient()
			if status == nil {
				__antithesis_instrumentation__.Notify(37408)

				return
			} else {
				__antithesis_instrumentation__.Notify(37409)
			}
			__antithesis_instrumentation__.Notify(37405)
			resp, err := status.Metrics(context.Background(), &serverpb.MetricsRequest{
				NodeId: "local",
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(37410)
				log.Fatalf(context.Background(), "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(37411)
			}
			__antithesis_instrumentation__.Notify(37406)
			var metrics map[string]interface{}
			if err := json.Unmarshal(resp.Data, &metrics); err != nil {
				__antithesis_instrumentation__.Notify(37412)
				log.Fatalf(context.Background(), "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(37413)
			}
			__antithesis_instrumentation__.Notify(37407)
			stores := metrics["stores"].(map[string]interface{})
			for _, v := range stores {
				__antithesis_instrumentation__.Notify(37414)
				storeMetrics := v.(map[string]interface{})
				if v, ok := storeMetrics["replicas"]; ok {
					__antithesis_instrumentation__.Notify(37418)
					stats.replicas[i] += int(v.(float64))
				} else {
					__antithesis_instrumentation__.Notify(37419)
				}
				__antithesis_instrumentation__.Notify(37415)
				if v, ok := storeMetrics["replicas.leaseholders"]; ok {
					__antithesis_instrumentation__.Notify(37420)
					stats.leases[i] += int(v.(float64))
				} else {
					__antithesis_instrumentation__.Notify(37421)
				}
				__antithesis_instrumentation__.Notify(37416)
				if v, ok := storeMetrics["range.adds"]; ok {
					__antithesis_instrumentation__.Notify(37422)
					stats.replicaAdds[i] += int(v.(float64))
				} else {
					__antithesis_instrumentation__.Notify(37423)
				}
				__antithesis_instrumentation__.Notify(37417)
				if v, ok := storeMetrics["leases.transfers.success"]; ok {
					__antithesis_instrumentation__.Notify(37424)
					stats.leaseTransfers[i] += int(v.(float64))
				} else {
					__antithesis_instrumentation__.Notify(37425)
				}
			}
		}(i)
	}
	__antithesis_instrumentation__.Notify(37401)
	wg.Wait()

	for _, v := range stats.replicas {
		__antithesis_instrumentation__.Notify(37426)
		stats.count += v
	}
	__antithesis_instrumentation__.Notify(37402)
	return stats
}

func (a *allocSim) rangeStats(d time.Duration) {
	__antithesis_instrumentation__.Notify(37427)
	for {
		__antithesis_instrumentation__.Notify(37428)
		stats := a.rangeInfo()
		a.ranges.Lock()
		a.ranges.stats = stats
		a.ranges.Unlock()

		time.Sleep(d)
	}
}

const padding = "__________________"

func formatHeader(header string, numberNodes int, localities []Locality) string {
	__antithesis_instrumentation__.Notify(37429)
	var buf bytes.Buffer
	_, _ = buf.WriteString(header)
	for i := 1; i <= numberNodes; i++ {
		__antithesis_instrumentation__.Notify(37431)
		node := fmt.Sprintf("%d", i)
		if localities != nil {
			__antithesis_instrumentation__.Notify(37433)
			node += fmt.Sprintf(":%s", localities[i-1].Name)
		} else {
			__antithesis_instrumentation__.Notify(37434)
		}
		__antithesis_instrumentation__.Notify(37432)
		fmt.Fprintf(&buf, "%s%s", padding[:len(padding)-len(node)], node)
	}
	__antithesis_instrumentation__.Notify(37430)
	return buf.String()
}

func (a *allocSim) monitor(d time.Duration) {
	__antithesis_instrumentation__.Notify(37435)
	formatNodes := func(stats allocStats) string {
		__antithesis_instrumentation__.Notify(37437)
		var buf bytes.Buffer
		for i := range stats.replicas {
			__antithesis_instrumentation__.Notify(37439)
			alive := a.Nodes[i].Alive()
			if !alive {
				__antithesis_instrumentation__.Notify(37441)
				_, _ = buf.WriteString("\033[0;31;49m")
			} else {
				__antithesis_instrumentation__.Notify(37442)
			}
			__antithesis_instrumentation__.Notify(37440)
			fmt.Fprintf(&buf, "%*s", len(padding), fmt.Sprintf("%d/%d/%d/%d",
				stats.replicas[i], stats.leases[i], stats.replicaAdds[i], stats.leaseTransfers[i]))
			if !alive {
				__antithesis_instrumentation__.Notify(37443)
				_, _ = buf.WriteString("\033[0m")
			} else {
				__antithesis_instrumentation__.Notify(37444)
			}
		}
		__antithesis_instrumentation__.Notify(37438)
		return buf.String()
	}
	__antithesis_instrumentation__.Notify(37436)

	start := timeutil.Now()
	lastTime := start
	var numReplicas int
	var lastOps uint64

	for ticks := 0; true; ticks++ {
		__antithesis_instrumentation__.Notify(37445)
		time.Sleep(d)

		now := timeutil.Now()
		elapsed := now.Sub(lastTime).Seconds()
		ops := atomic.LoadUint64(&a.stats.ops)
		totalLatencyNanos := atomic.LoadUint64(&a.stats.totalLatencyNanos)

		a.ranges.Lock()
		rangeStats := a.ranges.stats
		a.ranges.Unlock()

		if ticks%20 == 0 || func() bool {
			__antithesis_instrumentation__.Notify(37448)
			return numReplicas != len(rangeStats.replicas) == true
		}() == true {
			__antithesis_instrumentation__.Notify(37449)
			numReplicas = len(rangeStats.replicas)
			fmt.Println(formatHeader("_elapsed__ops/sec__average__latency___errors_replicas", numReplicas, a.localities))
		} else {
			__antithesis_instrumentation__.Notify(37450)
		}
		__antithesis_instrumentation__.Notify(37446)

		var avgLatency float64
		if ops > 0 {
			__antithesis_instrumentation__.Notify(37451)
			avgLatency = float64(totalLatencyNanos/ops) / float64(time.Millisecond)
		} else {
			__antithesis_instrumentation__.Notify(37452)
		}
		__antithesis_instrumentation__.Notify(37447)
		fmt.Printf("%8s %8.1f %8.1f %6.1fms %8d %8d%s\n",
			time.Duration(now.Sub(start).Seconds()+0.5)*time.Second,
			float64(ops-lastOps)/elapsed, float64(ops)/now.Sub(start).Seconds(), avgLatency,
			atomic.LoadUint64(&a.stats.errors), rangeStats.count, formatNodes(rangeStats))
		lastTime = now
		lastOps = ops
	}
}

func (a *allocSim) finalStatus() {
	__antithesis_instrumentation__.Notify(37453)
	a.ranges.Lock()
	defer a.ranges.Unlock()

	fmt.Println(formatHeader("___stats___________________________", len(a.ranges.stats.replicas), a.localities))

	genStats := func(name string, counts []int) {
		__antithesis_instrumentation__.Notify(37455)
		var total float64
		for _, count := range counts {
			__antithesis_instrumentation__.Notify(37458)
			total += float64(count)
		}
		__antithesis_instrumentation__.Notify(37456)
		mean := total / float64(len(counts))
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "%8s  (total%% / diff%%)         ", name)
		for _, count := range counts {
			__antithesis_instrumentation__.Notify(37459)
			var percent, fromMean float64
			if total != 0 {
				__antithesis_instrumentation__.Notify(37461)
				percent = float64(count) / total * 100
				fromMean = (float64(count) - mean) / total * 100
			} else {
				__antithesis_instrumentation__.Notify(37462)
			}
			__antithesis_instrumentation__.Notify(37460)
			fmt.Fprintf(&buf, " %9.9s", fmt.Sprintf("%.0f/%.0f", percent, fromMean))
		}
		__antithesis_instrumentation__.Notify(37457)
		fmt.Println(buf.String())
	}
	__antithesis_instrumentation__.Notify(37454)
	genStats("replicas", a.ranges.stats.replicas)
	genStats("leases", a.ranges.stats.leases)
}

func handleStart() bool {
	__antithesis_instrumentation__.Notify(37463)
	if len(os.Args) < 2 || func() bool {
		__antithesis_instrumentation__.Notify(37465)
		return os.Args[1] != "start" == true
	}() == true {
		__antithesis_instrumentation__.Notify(37466)
		return false
	} else {
		__antithesis_instrumentation__.Notify(37467)
	}
	__antithesis_instrumentation__.Notify(37464)

	kvserver.MinLeaseTransferStatsDuration = 10 * time.Second

	cli.Main()
	return true
}

func main() {
	__antithesis_instrumentation__.Notify(37468)
	if handleStart() {
		__antithesis_instrumentation__.Notify(37477)
		return
	} else {
		__antithesis_instrumentation__.Notify(37478)
	}
	__antithesis_instrumentation__.Notify(37469)

	flag.Parse()

	var config Configuration
	if *configFile != "" {
		__antithesis_instrumentation__.Notify(37479)
		var err error
		config, err = loadConfig(*configFile)
		if err != nil {
			__antithesis_instrumentation__.Notify(37480)
			log.Fatalf(context.Background(), "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(37481)
		}
	} else {
		__antithesis_instrumentation__.Notify(37482)
	}
	__antithesis_instrumentation__.Notify(37470)

	perNodeCfg := localcluster.MakePerNodeFixedPortsCfg(*numNodes)

	var separateAddrs bool
	for _, locality := range config.Localities {
		__antithesis_instrumentation__.Notify(37483)
		if len(locality.OutgoingLatencies) != 0 {
			__antithesis_instrumentation__.Notify(37484)
			separateAddrs = true
			if runtime.GOOS != "linux" {
				__antithesis_instrumentation__.Notify(37486)
				log.Fatal(context.Background(),
					"configs that set per-locality outgoing latencies are only supported on linux")
			} else {
				__antithesis_instrumentation__.Notify(37487)
			}
			__antithesis_instrumentation__.Notify(37485)
			break
		} else {
			__antithesis_instrumentation__.Notify(37488)
		}
	}
	__antithesis_instrumentation__.Notify(37471)

	if separateAddrs {
		__antithesis_instrumentation__.Notify(37489)
		for i := range perNodeCfg {
			__antithesis_instrumentation__.Notify(37490)
			s := perNodeCfg[i]
			s.Addr = fmt.Sprintf("127.0.0.%d", i)
			perNodeCfg[i] = s
		}
	} else {
		__antithesis_instrumentation__.Notify(37491)
	}
	__antithesis_instrumentation__.Notify(37472)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	localities := make([]Locality, *numNodes)
	if len(config.Localities) != 0 {
		__antithesis_instrumentation__.Notify(37492)
		nodesPerLocality := make(map[string][]int)
		var nodeIdx int
		for _, locality := range config.Localities {
			__antithesis_instrumentation__.Notify(37495)
			for i := 0; i < locality.NumNodes; i++ {
				__antithesis_instrumentation__.Notify(37496)
				s := perNodeCfg[nodeIdx]
				if locality.LocalityStr != "" {
					__antithesis_instrumentation__.Notify(37499)
					s.ExtraArgs = []string{fmt.Sprintf("--locality=%s", locality.LocalityStr)}
				} else {
					__antithesis_instrumentation__.Notify(37500)
					s.ExtraArgs = []string{fmt.Sprintf("--locality=l=%s", locality.Name)}
				}
				__antithesis_instrumentation__.Notify(37497)
				if separateAddrs {
					__antithesis_instrumentation__.Notify(37501)
					s.ExtraEnv = []string{fmt.Sprintf("COCKROACH_SOURCE_IP_ADDRESS=%s", s.Addr)}
				} else {
					__antithesis_instrumentation__.Notify(37502)
				}
				__antithesis_instrumentation__.Notify(37498)
				localities[nodeIdx] = locality
				nodesPerLocality[locality.Name] = append(nodesPerLocality[locality.Name], nodeIdx)

				perNodeCfg[nodeIdx] = s
				nodeIdx++
			}
		}
		__antithesis_instrumentation__.Notify(37493)
		var tcController *tc.Controller
		if separateAddrs {
			__antithesis_instrumentation__.Notify(37503)

			tcController = tc.NewController("lo")
			if err := tcController.Init(); err != nil {
				__antithesis_instrumentation__.Notify(37505)
				log.Fatalf(context.Background(), "%v", err)
			} else {
				__antithesis_instrumentation__.Notify(37506)
			}
			__antithesis_instrumentation__.Notify(37504)
			defer func() {
				__antithesis_instrumentation__.Notify(37507)
				if err := tcController.CleanUp(); err != nil {
					__antithesis_instrumentation__.Notify(37508)
					log.Errorf(context.Background(), "%v", err)
				} else {
					__antithesis_instrumentation__.Notify(37509)
				}
			}()
		} else {
			__antithesis_instrumentation__.Notify(37510)
		}
		__antithesis_instrumentation__.Notify(37494)
		for _, locality := range localities {
			__antithesis_instrumentation__.Notify(37511)
			for _, outgoing := range locality.OutgoingLatencies {
				__antithesis_instrumentation__.Notify(37512)
				if outgoing.Latency > 0 {
					__antithesis_instrumentation__.Notify(37513)
					for _, srcNodeIdx := range nodesPerLocality[locality.Name] {
						__antithesis_instrumentation__.Notify(37514)
						for _, dstNodeIdx := range nodesPerLocality[outgoing.Name] {
							__antithesis_instrumentation__.Notify(37515)
							if err := tcController.AddLatency(
								perNodeCfg[srcNodeIdx].Addr, perNodeCfg[dstNodeIdx].Addr, time.Duration(outgoing.Latency/2),
							); err != nil {
								__antithesis_instrumentation__.Notify(37516)
								log.Fatalf(context.Background(), "%v", err)
							} else {
								__antithesis_instrumentation__.Notify(37517)
							}
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(37518)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(37519)
	}
	__antithesis_instrumentation__.Notify(37473)

	cfg := localcluster.ClusterConfig{
		AllNodeArgs: append(flag.Args(), "--vmodule=allocator=3,allocator_scorer=3,replicate_queue=3"),
		Binary:      os.Args[0],
		NumNodes:    *numNodes,
		DB:          "allocsim",
		NumWorkers:  *workers,
		PerNodeCfg:  perNodeCfg,
		DataDir:     "cockroach-data-allocsim",
	}

	c := localcluster.New(cfg)
	a := newAllocSim(c)
	a.localities = localities

	log.SetExitFunc(false, func(code exit.Code) {
		__antithesis_instrumentation__.Notify(37520)
		c.Close()
		exit.WithCode(code)
	})
	__antithesis_instrumentation__.Notify(37474)

	go func() {
		__antithesis_instrumentation__.Notify(37521)
		var exitStatus int
		select {
		case s := <-signalCh:
			__antithesis_instrumentation__.Notify(37523)
			log.Infof(context.Background(), "signal received: %v", s)
			exitStatus = 1
		case <-time.After(*duration):
			__antithesis_instrumentation__.Notify(37524)
			log.Infof(context.Background(), "finished run of: %s", *duration)
		}
		__antithesis_instrumentation__.Notify(37522)
		c.Close()
		a.finalStatus()
		os.Exit(exitStatus)
	}()
	__antithesis_instrumentation__.Notify(37475)

	c.Start(context.Background())
	defer c.Close()
	c.UpdateZoneConfig(1, 1<<20)
	_, err := c.Nodes[0].DB().Exec("SET CLUSTER SETTING kv.raft_log.disable_synchronization_unsafe = true")
	if err != nil {
		__antithesis_instrumentation__.Notify(37525)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(37526)
	}
	__antithesis_instrumentation__.Notify(37476)
	if len(config.Localities) != 0 {
		__antithesis_instrumentation__.Notify(37527)
		a.runWithConfig(config)
	} else {
		__antithesis_instrumentation__.Notify(37528)
		a.run(*workers)
	}
}
