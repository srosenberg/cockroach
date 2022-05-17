package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach/pkg/acceptance/cluster"
	"github.com/cockroachdb/cockroach/pkg/acceptance/localcluster"
	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors/oserror"
)

var workers = flag.Int("w", 2*runtime.GOMAXPROCS(0), "number of workers")
var monkeys = flag.Int("m", 3, "number of monkeys")
var numNodes = flag.Int("n", 4, "number of nodes")
var numAccounts = flag.Int("a", 1e5, "number of accounts")
var chaosType = flag.String("c", "simple", "chaos type [none|simple|flappy]")
var verify = flag.Bool("verify", true, "verify range and account consistency")

func newRand() *rand.Rand {
	__antithesis_instrumentation__.Notify(53394)
	return rand.New(rand.NewSource(timeutil.Now().UnixNano()))
}

type zeroSum struct {
	*localcluster.LocalCluster
	numAccounts int
	chaosType   string
	accounts    struct {
		syncutil.Mutex
		m map[uint64]struct{}
	}
	stats struct {
		ops       uint64
		errors    uint64
		splits    uint64
		transfers uint64
	}
	ranges struct {
		syncutil.Mutex
		count    int
		replicas []int
	}
}

func newZeroSum(c *localcluster.LocalCluster, numAccounts int, chaosType string) *zeroSum {
	__antithesis_instrumentation__.Notify(53395)
	z := &zeroSum{
		LocalCluster: c,
		numAccounts:  numAccounts,
		chaosType:    chaosType,
	}
	z.accounts.m = make(map[uint64]struct{})
	return z
}

func (z *zeroSum) run(workers, monkeys int) {
	__antithesis_instrumentation__.Notify(53396)
	tableID := z.setup()
	for i := 0; i < workers; i++ {
		__antithesis_instrumentation__.Notify(53400)
		go z.worker()
	}
	__antithesis_instrumentation__.Notify(53397)
	for i := 0; i < monkeys; i++ {
		__antithesis_instrumentation__.Notify(53401)
		go z.monkey(tableID, 2*time.Second)
	}
	__antithesis_instrumentation__.Notify(53398)
	if workers > 0 || func() bool {
		__antithesis_instrumentation__.Notify(53402)
		return monkeys > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(53403)
		z.chaos()
		if *verify {
			__antithesis_instrumentation__.Notify(53404)
			go z.check(20 * time.Second)
			go z.verify(10 * time.Second)
		} else {
			__antithesis_instrumentation__.Notify(53405)
		}
	} else {
		__antithesis_instrumentation__.Notify(53406)
	}
	__antithesis_instrumentation__.Notify(53399)
	go z.rangeStats(time.Second)
	z.monitor(time.Second)
}

func (z *zeroSum) setup() uint32 {
	__antithesis_instrumentation__.Notify(53407)
	db := z.Nodes[0].DB()
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS zerosum"); err != nil {
		__antithesis_instrumentation__.Notify(53411)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(53412)
	}
	__antithesis_instrumentation__.Notify(53408)

	accounts := `
CREATE TABLE IF NOT EXISTS accounts (
  id INT PRIMARY KEY,
  balance INT NOT NULL
)
`
	if _, err := db.Exec(accounts); err != nil {
		__antithesis_instrumentation__.Notify(53413)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(53414)
	}
	__antithesis_instrumentation__.Notify(53409)

	tableIDQuery := `
SELECT tables.id FROM system.namespace tables
  JOIN system.namespace dbs ON dbs.id = tables."parentID"
  WHERE dbs.name = $1 AND tables.name = $2
`
	var tableID uint32
	if err := db.QueryRow(tableIDQuery, "zerosum", "accounts").Scan(&tableID); err != nil {
		__antithesis_instrumentation__.Notify(53415)
		log.Fatalf(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(53416)
	}
	__antithesis_instrumentation__.Notify(53410)
	return tableID
}

func (z *zeroSum) accountDistribution(r *rand.Rand) *rand.Zipf {
	__antithesis_instrumentation__.Notify(53417)

	return rand.NewZipf(r, 1.1, float64(z.numAccounts/10), uint64(z.numAccounts-1))
}

func (z *zeroSum) accountsLen() int {
	__antithesis_instrumentation__.Notify(53418)
	z.accounts.Lock()
	defer z.accounts.Unlock()
	return len(z.accounts.m)
}

func (z *zeroSum) maybeLogError(err error) {
	__antithesis_instrumentation__.Notify(53419)
	if localcluster.IsUnavailableError(err) || func() bool {
		__antithesis_instrumentation__.Notify(53421)
		return strings.Contains(err.Error(), "range is frozen") == true
	}() == true {
		__antithesis_instrumentation__.Notify(53422)
		return
	} else {
		__antithesis_instrumentation__.Notify(53423)
	}
	__antithesis_instrumentation__.Notify(53420)
	log.Errorf(context.Background(), "%v", err)
	atomic.AddUint64(&z.stats.errors, 1)
}

func (z *zeroSum) worker() {
	__antithesis_instrumentation__.Notify(53424)
	r := newRand()
	zipf := z.accountDistribution(r)

	for {
		__antithesis_instrumentation__.Notify(53425)
		from := zipf.Uint64()
		to := zipf.Uint64()
		if from == to {
			__antithesis_instrumentation__.Notify(53429)
			continue
		} else {
			__antithesis_instrumentation__.Notify(53430)
		}
		__antithesis_instrumentation__.Notify(53426)

		db := z.Nodes[z.RandNode(r.Intn)].DB()
		if db == nil {
			__antithesis_instrumentation__.Notify(53431)

			continue
		} else {
			__antithesis_instrumentation__.Notify(53432)
		}
		__antithesis_instrumentation__.Notify(53427)
		err := crdb.ExecuteTx(context.Background(), db, nil, func(tx *gosql.Tx) error {
			__antithesis_instrumentation__.Notify(53433)
			rows, err := tx.Query(`SELECT id, balance FROM accounts WHERE id IN ($1, $2)`, from, to)
			if err != nil {
				__antithesis_instrumentation__.Notify(53436)
				return err
			} else {
				__antithesis_instrumentation__.Notify(53437)
			}
			__antithesis_instrumentation__.Notify(53434)

			var fromBalance, toBalance int64
			for rows.Next() {
				__antithesis_instrumentation__.Notify(53438)
				var id uint64
				var balance int64
				if err = rows.Scan(&id, &balance); err != nil {
					__antithesis_instrumentation__.Notify(53440)
					log.Fatalf(context.Background(), "%v", err)
				} else {
					__antithesis_instrumentation__.Notify(53441)
				}
				__antithesis_instrumentation__.Notify(53439)
				switch id {
				case from:
					__antithesis_instrumentation__.Notify(53442)
					fromBalance = balance
				case to:
					__antithesis_instrumentation__.Notify(53443)
					toBalance = balance
				default:
					__antithesis_instrumentation__.Notify(53444)
					panic(fmt.Sprintf("got unexpected account %d", id))
				}
			}
			__antithesis_instrumentation__.Notify(53435)

			upsert := `UPSERT INTO accounts VALUES ($1, $3), ($2, $4)`
			_, err = tx.Exec(upsert, to, from, toBalance+1, fromBalance-1)
			return err
		})
		__antithesis_instrumentation__.Notify(53428)
		if err != nil {
			__antithesis_instrumentation__.Notify(53445)
			z.maybeLogError(err)
		} else {
			__antithesis_instrumentation__.Notify(53446)
			atomic.AddUint64(&z.stats.ops, 1)
			z.accounts.Lock()
			z.accounts.m[from] = struct{}{}
			z.accounts.m[to] = struct{}{}
			z.accounts.Unlock()
		}
	}
}

func (z *zeroSum) monkey(tableID uint32, d time.Duration) {
	__antithesis_instrumentation__.Notify(53447)
	r := newRand()
	zipf := z.accountDistribution(r)

	for {
		__antithesis_instrumentation__.Notify(53448)
		time.Sleep(time.Duration(rand.Float64() * float64(d)))

		key := keys.SystemSQLCodec.TablePrefix(tableID)
		key = encoding.EncodeVarintAscending(key, int64(zipf.Uint64()))

		switch r.Intn(2) {
		case 0:
			__antithesis_instrumentation__.Notify(53449)
			if err := z.Split(z.RandNode(r.Intn), key); err != nil {
				__antithesis_instrumentation__.Notify(53452)
				z.maybeLogError(err)
			} else {
				__antithesis_instrumentation__.Notify(53453)
				atomic.AddUint64(&z.stats.splits, 1)
			}
		case 1:
			__antithesis_instrumentation__.Notify(53450)
			if transferred, err := z.TransferLease(z.RandNode(r.Intn), r, key); err != nil {
				__antithesis_instrumentation__.Notify(53454)
				z.maybeLogError(err)
			} else {
				__antithesis_instrumentation__.Notify(53455)
				if transferred {
					__antithesis_instrumentation__.Notify(53456)
					atomic.AddUint64(&z.stats.transfers, 1)
				} else {
					__antithesis_instrumentation__.Notify(53457)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(53451)
		}
	}
}

func (z *zeroSum) chaosSimple() {
	__antithesis_instrumentation__.Notify(53458)
	d := 15 * time.Second
	fmt.Printf("chaos(simple): first event in %s\n", d)
	time.Sleep(d)

	nodeIdx := 0
	node := z.Nodes[nodeIdx]
	d = 20 * time.Second
	fmt.Printf("chaos: killing node %d for %s\n", nodeIdx+1, d)
	node.Kill()

	time.Sleep(d)
	fmt.Printf("chaos: starting node %d\n", nodeIdx+1)
	node.Start(context.Background())
}

func (z *zeroSum) chaosFlappy() {
	__antithesis_instrumentation__.Notify(53459)
	r := newRand()
	d := time.Duration(15+r.Intn(30)) * time.Second
	fmt.Printf("chaos(flappy): first event in %s\n", d)

	for i := 1; true; i++ {
		__antithesis_instrumentation__.Notify(53460)
		time.Sleep(d)

		nodeIdx := z.RandNode(r.Intn)
		node := z.Nodes[nodeIdx]
		d = time.Duration(15+r.Intn(30)) * time.Second
		fmt.Printf("chaos %d: killing node %d for %s\n", i, nodeIdx+1, d)
		node.Kill()

		time.Sleep(d)

		d = time.Duration(15+r.Intn(30)) * time.Second
		fmt.Printf("chaos %d: starting node %d, next event in %s\n", i, nodeIdx+1, d)
		node.Start(context.Background())
	}
}

func (z *zeroSum) chaos() {
	__antithesis_instrumentation__.Notify(53461)
	switch z.chaosType {
	case "none":
		__antithesis_instrumentation__.Notify(53462)

	case "simple":
		__antithesis_instrumentation__.Notify(53463)
		go z.chaosSimple()
	case "flappy":
		__antithesis_instrumentation__.Notify(53464)
		go z.chaosFlappy()
	default:
		__antithesis_instrumentation__.Notify(53465)
		log.Fatalf(context.Background(), "unknown chaos type: %s", z.chaosType)
	}
}

func (z *zeroSum) check(d time.Duration) {
	__antithesis_instrumentation__.Notify(53466)
	for {
		__antithesis_instrumentation__.Notify(53467)
		time.Sleep(d)
		if err := cluster.Consistent(context.Background(), z.LocalCluster, z.RandNode(rand.Intn)); err != nil {
			__antithesis_instrumentation__.Notify(53468)
			z.maybeLogError(err)
		} else {
			__antithesis_instrumentation__.Notify(53469)
		}
	}
}

func (z *zeroSum) verify(d time.Duration) {
	__antithesis_instrumentation__.Notify(53470)
	for {
		__antithesis_instrumentation__.Notify(53471)
		time.Sleep(d)

		committedAccounts := uint64(z.accountsLen())

		q := `SELECT count(*), sum(balance) FROM accounts`
		var accounts uint64
		var total int64
		db := z.Nodes[z.RandNode(rand.Intn)].DB()
		if err := db.QueryRow(q).Scan(&accounts, &total); err != nil {
			__antithesis_instrumentation__.Notify(53474)
			z.maybeLogError(err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(53475)
		}
		__antithesis_instrumentation__.Notify(53472)
		if total != 0 {
			__antithesis_instrumentation__.Notify(53476)
			log.Fatalf(context.Background(), "unexpected total balance %d", total)
		} else {
			__antithesis_instrumentation__.Notify(53477)
		}
		__antithesis_instrumentation__.Notify(53473)
		if accounts < committedAccounts {
			__antithesis_instrumentation__.Notify(53478)
			log.Fatalf(context.Background(), "expected at least %d accounts, but found %d",
				committedAccounts, accounts)
		} else {
			__antithesis_instrumentation__.Notify(53479)
		}
	}
}

func (z *zeroSum) rangeInfo() (int, []int) {
	__antithesis_instrumentation__.Notify(53480)
	replicas := make([]int, len(z.Nodes))
	db, err := z.NewDB(context.Background(), z.RandNode(rand.Intn))
	if err != nil {
		__antithesis_instrumentation__.Notify(53484)
		z.maybeLogError(err)
		return -1, replicas
	} else {
		__antithesis_instrumentation__.Notify(53485)
	}
	__antithesis_instrumentation__.Notify(53481)
	rows, err := db.Query(`SELECT array_length(replicas, 1) FROM crdb_internal.ranges`)
	if err != nil {
		__antithesis_instrumentation__.Notify(53486)
		z.maybeLogError(err)
		return -1, replicas
	} else {
		__antithesis_instrumentation__.Notify(53487)
	}
	__antithesis_instrumentation__.Notify(53482)
	defer rows.Close()

	var count int
	for rows.Next() {
		__antithesis_instrumentation__.Notify(53488)
		var numReplicas int
		if err := rows.Scan(&numReplicas); err != nil {
			__antithesis_instrumentation__.Notify(53491)
			z.maybeLogError(err)
			return -1, replicas
		} else {
			__antithesis_instrumentation__.Notify(53492)
		}
		__antithesis_instrumentation__.Notify(53489)
		for i := 0; i < numReplicas; i++ {
			__antithesis_instrumentation__.Notify(53493)
			replicas[i]++
		}
		__antithesis_instrumentation__.Notify(53490)
		count++
	}
	__antithesis_instrumentation__.Notify(53483)

	return count, replicas
}

func (z *zeroSum) rangeStats(d time.Duration) {
	__antithesis_instrumentation__.Notify(53494)
	for {
		__antithesis_instrumentation__.Notify(53495)
		count, replicas := z.rangeInfo()
		z.ranges.Lock()
		z.ranges.count, z.ranges.replicas = count, replicas
		z.ranges.Unlock()

		time.Sleep(d)
	}
}

func (z *zeroSum) formatReplicas(replicas []int) string {
	__antithesis_instrumentation__.Notify(53496)
	var buf bytes.Buffer
	for i := range replicas {
		__antithesis_instrumentation__.Notify(53498)
		if i > 0 {
			__antithesis_instrumentation__.Notify(53500)
			_, _ = buf.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(53501)
		}
		__antithesis_instrumentation__.Notify(53499)
		fmt.Fprintf(&buf, "%d", replicas[i])
		if !z.Nodes[i].Alive() {
			__antithesis_instrumentation__.Notify(53502)
			_, _ = buf.WriteString("*")
		} else {
			__antithesis_instrumentation__.Notify(53503)
		}
	}
	__antithesis_instrumentation__.Notify(53497)
	return buf.String()
}

func (z *zeroSum) monitor(d time.Duration) {
	__antithesis_instrumentation__.Notify(53504)
	start := timeutil.Now()
	lastTime := start
	var lastOps uint64

	for ticks := 0; true; ticks++ {
		__antithesis_instrumentation__.Notify(53505)
		time.Sleep(d)

		if ticks%20 == 0 {
			__antithesis_instrumentation__.Notify(53507)
			fmt.Printf("_elapsed__accounts_________ops__ops/sec___errors___splits____xfers___ranges_____________replicas\n")
		} else {
			__antithesis_instrumentation__.Notify(53508)
		}
		__antithesis_instrumentation__.Notify(53506)

		now := timeutil.Now()
		elapsed := now.Sub(lastTime).Seconds()
		ops := atomic.LoadUint64(&z.stats.ops)

		z.ranges.Lock()
		ranges, replicas := z.ranges.count, z.ranges.replicas
		z.ranges.Unlock()

		fmt.Printf("%8s %9d %11d %8.1f %8d %8d %8d %8d %20s\n",
			time.Duration(now.Sub(start).Seconds()+0.5)*time.Second,
			z.accountsLen(), ops, float64(ops-lastOps)/elapsed,
			atomic.LoadUint64(&z.stats.errors),
			atomic.LoadUint64(&z.stats.splits),
			atomic.LoadUint64(&z.stats.transfers),
			ranges, z.formatReplicas(replicas))
		lastTime = now
		lastOps = ops
	}
}

func main() {
	__antithesis_instrumentation__.Notify(53509)
	flag.Parse()

	cockroachBin := func() string {
		__antithesis_instrumentation__.Notify(53513)
		bin := "./cockroach"
		if _, err := os.Stat(bin); oserror.IsNotExist(err) {
			__antithesis_instrumentation__.Notify(53515)
			bin = "cockroach"
		} else {
			__antithesis_instrumentation__.Notify(53516)
			if err != nil {
				__antithesis_instrumentation__.Notify(53517)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(53518)
			}
		}
		__antithesis_instrumentation__.Notify(53514)
		return bin
	}()
	__antithesis_instrumentation__.Notify(53510)

	perNodeCfg := localcluster.MakePerNodeFixedPortsCfg(*numNodes)

	cfg := localcluster.ClusterConfig{
		DataDir:     "cockroach-data-zerosum",
		Binary:      cockroachBin,
		NumNodes:    *numNodes,
		NumWorkers:  *workers,
		AllNodeArgs: flag.Args(),
		DB:          "zerosum",
		PerNodeCfg:  perNodeCfg,
	}

	c := &localcluster.LocalCluster{Cluster: localcluster.New(cfg)}
	defer c.Close()

	log.SetExitFunc(false, func(code exit.Code) {
		__antithesis_instrumentation__.Notify(53519)
		c.Close()
		exit.WithCode(code)
	})
	__antithesis_instrumentation__.Notify(53511)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		__antithesis_instrumentation__.Notify(53520)
		s := <-signalCh
		log.Infof(context.Background(), "signal received: %v", s)
		c.Close()
		os.Exit(1)
	}()
	__antithesis_instrumentation__.Notify(53512)

	c.Start(context.Background())

	z := newZeroSum(c, *numAccounts, *chaosType)
	z.run(*workers, *monkeys)
}
