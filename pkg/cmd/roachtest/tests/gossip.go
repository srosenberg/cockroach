package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func registerGossip(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48130)
	runGossipChaos := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(48132)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--vmodule=*=1")
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.All())
		err := WaitFor3XReplication(ctx, t, c.Conn(ctx, t.L(), 1))
		require.NoError(t, err)

		gossipNetworkAccordingTo := func(node int) (network string) {
			__antithesis_instrumentation__.Notify(48137)
			const query = `
SELECT string_agg(source_id::TEXT || ':' || target_id::TEXT, ',')
  FROM (SELECT * FROM crdb_internal.gossip_network ORDER BY source_id, target_id)
`

			db := c.Conn(ctx, t.L(), node)
			defer db.Close()
			var s gosql.NullString
			if err = db.QueryRow(query).Scan(&s); err != nil {
				__antithesis_instrumentation__.Notify(48140)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48141)
			}
			__antithesis_instrumentation__.Notify(48138)
			if s.Valid {
				__antithesis_instrumentation__.Notify(48142)
				return s.String
			} else {
				__antithesis_instrumentation__.Notify(48143)
			}
			__antithesis_instrumentation__.Notify(48139)
			return ""
		}
		__antithesis_instrumentation__.Notify(48133)

		nodesInNetworkAccordingTo := func(node int) (nodes []int, network string) {
			__antithesis_instrumentation__.Notify(48144)
			split := func(c rune) bool {
				__antithesis_instrumentation__.Notify(48148)
				return !unicode.IsNumber(c)
			}
			__antithesis_instrumentation__.Notify(48145)

			uniqueNodes := make(map[int]struct{})
			network = gossipNetworkAccordingTo(node)
			for _, idStr := range strings.FieldsFunc(network, split) {
				__antithesis_instrumentation__.Notify(48149)
				nodeID, err := strconv.Atoi(idStr)
				if err != nil {
					__antithesis_instrumentation__.Notify(48151)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(48152)
				}
				__antithesis_instrumentation__.Notify(48150)
				uniqueNodes[nodeID] = struct{}{}
			}
			__antithesis_instrumentation__.Notify(48146)
			for node := range uniqueNodes {
				__antithesis_instrumentation__.Notify(48153)
				nodes = append(nodes, node)
			}
			__antithesis_instrumentation__.Notify(48147)
			sort.Ints(nodes)
			return nodes, network
		}
		__antithesis_instrumentation__.Notify(48134)

		gossipOK := func(start time.Time, deadNode int) bool {
			__antithesis_instrumentation__.Notify(48154)
			var expLiveNodes []int
			var expGossipNetwork string

			for i := 1; i <= c.Spec().NodeCount; i++ {
				__antithesis_instrumentation__.Notify(48156)
				if elapsed := timeutil.Since(start); elapsed >= 20*time.Second {
					__antithesis_instrumentation__.Notify(48162)
					t.Fatalf("gossip did not stabilize (dead node %d) in %.1fs", deadNode, elapsed.Seconds())
				} else {
					__antithesis_instrumentation__.Notify(48163)
				}
				__antithesis_instrumentation__.Notify(48157)

				if i == deadNode {
					__antithesis_instrumentation__.Notify(48164)
					continue
				} else {
					__antithesis_instrumentation__.Notify(48165)
				}
				__antithesis_instrumentation__.Notify(48158)

				t.L().Printf("%d: checking gossip\n", i)
				liveNodes, gossipNetwork := nodesInNetworkAccordingTo(i)
				for _, id := range liveNodes {
					__antithesis_instrumentation__.Notify(48166)
					if id == deadNode {
						__antithesis_instrumentation__.Notify(48167)
						t.L().Printf("%d: gossip not ok (dead node %d present): %s (%.0fs)\n",
							i, deadNode, gossipNetwork, timeutil.Since(start).Seconds())
						return false
					} else {
						__antithesis_instrumentation__.Notify(48168)
					}
				}
				__antithesis_instrumentation__.Notify(48159)

				if len(expLiveNodes) == 0 {
					__antithesis_instrumentation__.Notify(48169)
					expLiveNodes = liveNodes
					expGossipNetwork = gossipNetwork
					continue
				} else {
					__antithesis_instrumentation__.Notify(48170)
				}
				__antithesis_instrumentation__.Notify(48160)

				if len(liveNodes) != len(expLiveNodes) {
					__antithesis_instrumentation__.Notify(48171)
					t.L().Printf("%d: gossip not ok (mismatched size of network: %s); expected %d, got %d (%.0fs)\n",
						i, gossipNetwork, len(expLiveNodes), len(liveNodes), timeutil.Since(start).Seconds())
					return false
				} else {
					__antithesis_instrumentation__.Notify(48172)
				}
				__antithesis_instrumentation__.Notify(48161)

				for i := range liveNodes {
					__antithesis_instrumentation__.Notify(48173)
					if liveNodes[i] != expLiveNodes[i] {
						__antithesis_instrumentation__.Notify(48174)
						t.L().Printf("%d: gossip not ok (mismatched view of live nodes); expected %s, got %s (%.0fs)\n",
							i, gossipNetwork, expLiveNodes, liveNodes, timeutil.Since(start).Seconds())
						return false
					} else {
						__antithesis_instrumentation__.Notify(48175)
					}
				}
			}
			__antithesis_instrumentation__.Notify(48155)
			t.L().Printf("gossip ok: %s (size: %d) (%0.0fs)\n", expGossipNetwork, len(expLiveNodes), timeutil.Since(start).Seconds())
			return true
		}
		__antithesis_instrumentation__.Notify(48135)

		waitForGossip := func(deadNode int) {
			__antithesis_instrumentation__.Notify(48176)
			t.Status("waiting for gossip to exclude dead node")
			start := timeutil.Now()
			for {
				__antithesis_instrumentation__.Notify(48177)
				if gossipOK(start, deadNode) {
					__antithesis_instrumentation__.Notify(48179)
					return
				} else {
					__antithesis_instrumentation__.Notify(48180)
				}
				__antithesis_instrumentation__.Notify(48178)
				time.Sleep(time.Second)
			}
		}
		__antithesis_instrumentation__.Notify(48136)

		waitForGossip(0)
		nodes := c.All()
		for j := 0; j < 10; j++ {
			__antithesis_instrumentation__.Notify(48181)
			deadNode := nodes.RandNode()[0]
			c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(deadNode))
			waitForGossip(deadNode)
			c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(deadNode))
		}
	}
	__antithesis_instrumentation__.Notify(48131)

	r.Add(registry.TestSpec{
		Name:    "gossip/chaos/nodes=9",
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(9),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48182)
			runGossipChaos(ctx, t, c)
		},
	})
}

type gossipUtil struct {
	waitTime time.Duration
	urlMap   map[int]string
	conn     func(ctx context.Context, l *logger.Logger, i int) *gosql.DB
}

func newGossipUtil(ctx context.Context, t test.Test, c cluster.Cluster) *gossipUtil {
	__antithesis_instrumentation__.Notify(48183)
	urlMap := make(map[int]string)
	adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.All())
	if err != nil {
		__antithesis_instrumentation__.Notify(48186)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48187)
	}
	__antithesis_instrumentation__.Notify(48184)
	for i, addr := range adminUIAddrs {
		__antithesis_instrumentation__.Notify(48188)
		urlMap[i+1] = `http://` + addr
	}
	__antithesis_instrumentation__.Notify(48185)
	return &gossipUtil{
		waitTime: 30 * time.Second,
		urlMap:   urlMap,
		conn:     c.Conn,
	}
}

type checkGossipFunc func(map[string]gossip.Info) error

func (g *gossipUtil) check(ctx context.Context, c cluster.Cluster, f checkGossipFunc) error {
	__antithesis_instrumentation__.Notify(48189)
	return retry.ForDuration(g.waitTime, func() error {
		__antithesis_instrumentation__.Notify(48190)
		var infoStatus gossip.InfoStatus
		for i := 1; i <= c.Spec().NodeCount; i++ {
			__antithesis_instrumentation__.Notify(48192)
			url := g.urlMap[i] + `/_status/gossip/local`
			if err := httputil.GetJSON(http.Client{}, url, &infoStatus); err != nil {
				__antithesis_instrumentation__.Notify(48194)
				return errors.Wrapf(err, "failed to get gossip status from node %d", i)
			} else {
				__antithesis_instrumentation__.Notify(48195)
			}
			__antithesis_instrumentation__.Notify(48193)
			if err := f(infoStatus.Infos); err != nil {
				__antithesis_instrumentation__.Notify(48196)
				return errors.Wrapf(err, "node %d", i)
			} else {
				__antithesis_instrumentation__.Notify(48197)
			}
		}
		__antithesis_instrumentation__.Notify(48191)

		return nil
	})
}

func (gossipUtil) hasPeers(expected int) checkGossipFunc {
	__antithesis_instrumentation__.Notify(48198)
	return func(infos map[string]gossip.Info) error {
		__antithesis_instrumentation__.Notify(48199)
		count := 0
		for k := range infos {
			__antithesis_instrumentation__.Notify(48202)
			if strings.HasPrefix(k, gossip.KeyNodeIDPrefix) {
				__antithesis_instrumentation__.Notify(48203)
				count++
			} else {
				__antithesis_instrumentation__.Notify(48204)
			}
		}
		__antithesis_instrumentation__.Notify(48200)
		if count != expected {
			__antithesis_instrumentation__.Notify(48205)
			return errors.Errorf("expected %d peers, found %d", expected, count)
		} else {
			__antithesis_instrumentation__.Notify(48206)
		}
		__antithesis_instrumentation__.Notify(48201)
		return nil
	}
}

func (gossipUtil) hasSentinel(infos map[string]gossip.Info) error {
	__antithesis_instrumentation__.Notify(48207)
	if _, ok := infos[gossip.KeySentinel]; !ok {
		__antithesis_instrumentation__.Notify(48209)
		return errors.Errorf("sentinel not found")
	} else {
		__antithesis_instrumentation__.Notify(48210)
	}
	__antithesis_instrumentation__.Notify(48208)
	return nil
}

func (gossipUtil) hasClusterID(infos map[string]gossip.Info) error {
	__antithesis_instrumentation__.Notify(48211)
	if _, ok := infos[gossip.KeyClusterID]; !ok {
		__antithesis_instrumentation__.Notify(48213)
		return errors.Errorf("cluster ID not found")
	} else {
		__antithesis_instrumentation__.Notify(48214)
	}
	__antithesis_instrumentation__.Notify(48212)
	return nil
}

func (g *gossipUtil) checkConnectedAndFunctional(
	ctx context.Context, t test.Test, c cluster.Cluster,
) {
	__antithesis_instrumentation__.Notify(48215)
	t.L().Printf("waiting for gossip to be connected\n")
	if err := g.check(ctx, c, g.hasPeers(c.Spec().NodeCount)); err != nil {
		__antithesis_instrumentation__.Notify(48219)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48220)
	}
	__antithesis_instrumentation__.Notify(48216)
	if err := g.check(ctx, c, g.hasClusterID); err != nil {
		__antithesis_instrumentation__.Notify(48221)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48222)
	}
	__antithesis_instrumentation__.Notify(48217)
	if err := g.check(ctx, c, g.hasSentinel); err != nil {
		__antithesis_instrumentation__.Notify(48223)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48224)
	}
	__antithesis_instrumentation__.Notify(48218)

	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(48225)
		db := g.conn(ctx, t.L(), i)
		defer db.Close()
		if i == 1 {
			__antithesis_instrumentation__.Notify(48228)
			if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS test"); err != nil {
				__antithesis_instrumentation__.Notify(48231)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48232)
			}
			__antithesis_instrumentation__.Notify(48229)
			if _, err := db.Exec("CREATE TABLE IF NOT EXISTS test.kv (k INT PRIMARY KEY, v INT)"); err != nil {
				__antithesis_instrumentation__.Notify(48233)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48234)
			}
			__antithesis_instrumentation__.Notify(48230)
			if _, err := db.Exec(`UPSERT INTO test.kv (k, v) VALUES (1, 0)`); err != nil {
				__antithesis_instrumentation__.Notify(48235)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48236)
			}
		} else {
			__antithesis_instrumentation__.Notify(48237)
		}
		__antithesis_instrumentation__.Notify(48226)
		rows, err := db.Query(`UPDATE test.kv SET v=v+1 WHERE k=1 RETURNING v`)
		if err != nil {
			__antithesis_instrumentation__.Notify(48238)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48239)
		}
		__antithesis_instrumentation__.Notify(48227)
		defer rows.Close()
		var count int
		if rows.Next() {
			__antithesis_instrumentation__.Notify(48240)
			if err := rows.Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(48242)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48243)
			}
			__antithesis_instrumentation__.Notify(48241)
			if count != i {
				__antithesis_instrumentation__.Notify(48244)
				t.Fatalf("unexpected value %d for write #%d (expected %d)", count, i, i)
			} else {
				__antithesis_instrumentation__.Notify(48245)
			}
		} else {
			__antithesis_instrumentation__.Notify(48246)
			t.Fatalf("no results found from update")
		}
	}
}

func runGossipPeerings(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(48247)
	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	g := newGossipUtil(ctx, t, c)
	deadline := timeutil.Now().Add(time.Minute)

	for i := 1; timeutil.Now().Before(deadline); i++ {
		__antithesis_instrumentation__.Notify(48248)
		if err := g.check(ctx, c, g.hasPeers(c.Spec().NodeCount)); err != nil {
			__antithesis_instrumentation__.Notify(48252)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48253)
		}
		__antithesis_instrumentation__.Notify(48249)
		if err := g.check(ctx, c, g.hasClusterID); err != nil {
			__antithesis_instrumentation__.Notify(48254)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48255)
		}
		__antithesis_instrumentation__.Notify(48250)
		if err := g.check(ctx, c, g.hasSentinel); err != nil {
			__antithesis_instrumentation__.Notify(48256)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48257)
		}
		__antithesis_instrumentation__.Notify(48251)
		t.L().Printf("%d: OK\n", i)

		node := c.All().RandNode()
		t.L().Printf("%d: restarting node %d\n", i, node[0])
		c.Stop(ctx, t.L(), option.DefaultStopOpts(), node)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), node)

		time.Sleep(3 * time.Second)
	}
}

func runGossipRestart(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(48258)
	t.Skip("skipping flaky acceptance/gossip/restart", "https://github.com/cockroachdb/cockroach/issues/48423")

	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	g := newGossipUtil(ctx, t, c)
	deadline := timeutil.Now().Add(time.Minute)

	for i := 1; timeutil.Now().Before(deadline); i++ {
		__antithesis_instrumentation__.Notify(48259)
		g.checkConnectedAndFunctional(ctx, t, c)
		t.L().Printf("%d: OK\n", i)

		t.L().Printf("%d: killing all nodes\n", i)
		c.Stop(ctx, t.L(), option.DefaultStopOpts())

		t.L().Printf("%d: restarting all nodes\n", i)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
	}
}

func runGossipRestartNodeOne(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(48260)
	c.Put(ctx, t.Cockroach(), "./cockroach")

	settings := install.MakeClusterSettings(install.NumRacksOption(c.Spec().NodeCount))
	settings.Env = append(settings.Env, "COCKROACH_SCAN_MAX_IDLE_TIME=5ms")

	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.EncryptedStores = false
	c.Start(ctx, t.L(), startOpts, settings)

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	run := func(stmtStr string) {
		__antithesis_instrumentation__.Notify(48268)
		stmt := fmt.Sprintf(stmtStr, "", "=")
		t.L().Printf("%s\n", stmt)
		_, err := db.ExecContext(ctx, stmt)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(48270)
			return strings.Contains(err.Error(), "syntax error") == true
		}() == true {
			__antithesis_instrumentation__.Notify(48271)

			stmt = fmt.Sprintf(stmtStr, "EXPERIMENTAL", "")
			t.L().Printf("%s\n", stmt)
			_, err = db.ExecContext(ctx, stmt)
		} else {
			__antithesis_instrumentation__.Notify(48272)
		}
		__antithesis_instrumentation__.Notify(48269)
		if err != nil {
			__antithesis_instrumentation__.Notify(48273)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48274)
		}
	}
	__antithesis_instrumentation__.Notify(48261)

	var lastNodeCount int
	if err := retry.ForDuration(30*time.Second, func() error {
		__antithesis_instrumentation__.Notify(48275)
		const query = `SELECT count(*) FROM crdb_internal.gossip_nodes`
		var count int
		if err := db.QueryRow(query).Scan(&count); err != nil {
			__antithesis_instrumentation__.Notify(48278)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48279)
		}
		__antithesis_instrumentation__.Notify(48276)
		if count <= 1 {
			__antithesis_instrumentation__.Notify(48280)
			err := errors.Errorf("node 1 still only knows about %d node%s",
				count, util.Pluralize(int64(count)))
			if count != lastNodeCount {
				__antithesis_instrumentation__.Notify(48282)
				lastNodeCount = count
				t.L().Printf("%s\n", err)
			} else {
				__antithesis_instrumentation__.Notify(48283)
			}
			__antithesis_instrumentation__.Notify(48281)
			return err
		} else {
			__antithesis_instrumentation__.Notify(48284)
		}
		__antithesis_instrumentation__.Notify(48277)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(48285)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48286)
	}
	__antithesis_instrumentation__.Notify(48262)

	run(`ALTER RANGE default %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
	run(`ALTER RANGE system %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
	run(`ALTER DATABASE system %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
	run(`ALTER RANGE meta %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
	run(`ALTER RANGE liveness %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)

	run(`ALTER TABLE system.jobs %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
	if t.IsBuildVersion("v19.2.0") {
		__antithesis_instrumentation__.Notify(48287)
		run(`ALTER TABLE system.replication_stats %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
		run(`ALTER TABLE system.replication_constraint_stats %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
	} else {
		__antithesis_instrumentation__.Notify(48288)
	}
	__antithesis_instrumentation__.Notify(48263)
	if t.IsBuildVersion("v21.2.0") {
		__antithesis_instrumentation__.Notify(48289)
		run(`ALTER TABLE system.tenant_usage %[1]s CONFIGURE ZONE %[2]s 'constraints: {"-rack=0"}'`)
	} else {
		__antithesis_instrumentation__.Notify(48290)
	}
	__antithesis_instrumentation__.Notify(48264)

	var lastReplCount int
	if err := retry.ForDuration(2*time.Minute, func() error {
		__antithesis_instrumentation__.Notify(48291)
		const query = `
SELECT count(replicas)
  FROM crdb_internal.ranges
 WHERE array_position(replicas, 1) IS NOT NULL
`
		var count int
		if err := db.QueryRow(query).Scan(&count); err != nil {
			__antithesis_instrumentation__.Notify(48294)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48295)
		}
		__antithesis_instrumentation__.Notify(48292)
		if count > 0 {
			__antithesis_instrumentation__.Notify(48296)
			err := errors.Errorf("node 1 still has %d replicas", count)
			if count != lastReplCount {
				__antithesis_instrumentation__.Notify(48298)
				lastReplCount = count
				t.L().Printf("%s\n", err)
			} else {
				__antithesis_instrumentation__.Notify(48299)
			}
			__antithesis_instrumentation__.Notify(48297)
			return err
		} else {
			__antithesis_instrumentation__.Notify(48300)
		}
		__antithesis_instrumentation__.Notify(48293)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(48301)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48302)
	}
	__antithesis_instrumentation__.Notify(48265)

	t.L().Printf("killing all nodes\n")
	c.Stop(ctx, t.L(), option.DefaultStopOpts())

	err := c.RunE(ctx, c.Node(1),
		`./cockroach start --insecure --background --store={store-dir} `+
			`--log-dir={log-dir} --cache=10% --max-sql-memory=10% `+
			`--listen-addr=:$[{pgport:1}+10000] --http-port=$[{pgport:1}+1] `+
			`--join={pghost:1}:{pgport:1}`+
			`> {log-dir}/cockroach.stdout 2> {log-dir}/cockroach.stderr`)
	if err != nil {
		__antithesis_instrumentation__.Notify(48303)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48304)
	}
	__antithesis_instrumentation__.Notify(48266)

	settings = install.MakeClusterSettings()
	settings.Env = append(settings.Env, "COCKROACH_SCAN_MAX_IDLE_TIME=5ms")
	c.Start(ctx, t.L(), startOpts, settings, c.Range(2, c.Spec().NodeCount))

	g := newGossipUtil(ctx, t, c)
	g.conn = func(ctx context.Context, l *logger.Logger, i int) *gosql.DB {
		__antithesis_instrumentation__.Notify(48305)
		if i != 1 {
			__antithesis_instrumentation__.Notify(48312)
			return c.Conn(ctx, l, i)
		} else {
			__antithesis_instrumentation__.Notify(48313)
		}
		__antithesis_instrumentation__.Notify(48306)
		urls, err := c.ExternalPGUrl(ctx, l, c.Node(1))
		if err != nil {
			__antithesis_instrumentation__.Notify(48314)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48315)
		}
		__antithesis_instrumentation__.Notify(48307)
		url, err := url.Parse(urls[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(48316)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48317)
		}
		__antithesis_instrumentation__.Notify(48308)
		host, port, err := net.SplitHostPort(url.Host)
		if err != nil {
			__antithesis_instrumentation__.Notify(48318)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48319)
		}
		__antithesis_instrumentation__.Notify(48309)
		v, err := strconv.Atoi(port)
		if err != nil {
			__antithesis_instrumentation__.Notify(48320)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48321)
		}
		__antithesis_instrumentation__.Notify(48310)
		url.Host = fmt.Sprintf("%s:%d", host, v+10000)
		db, err := gosql.Open("postgres", url.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(48322)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48323)
		}
		__antithesis_instrumentation__.Notify(48311)
		return db
	}
	__antithesis_instrumentation__.Notify(48267)

	g.checkConnectedAndFunctional(ctx, t, c)

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(1))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(1))
}

func runCheckLocalityIPAddress(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(48324)
	c.Put(ctx, t.Cockroach(), "./cockroach")

	externalIP, err := c.ExternalIP(ctx, t.L(), c.Range(1, c.Spec().NodeCount))
	if err != nil {
		__antithesis_instrumentation__.Notify(48328)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48329)
	}
	__antithesis_instrumentation__.Notify(48325)

	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(48330)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(48332)
			externalIP[i-1] = "localhost"
		} else {
			__antithesis_instrumentation__.Notify(48333)
		}
		__antithesis_instrumentation__.Notify(48331)
		extAddr := externalIP[i-1]

		settings := install.MakeClusterSettings(install.NumRacksOption(1))
		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, fmt.Sprintf("--locality-advertise-addr=rack=0@%s", extAddr))
		c.Start(ctx, t.L(), startOpts, settings, c.Node(i))
	}
	__antithesis_instrumentation__.Notify(48326)

	rowCount := 0

	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(48334)
		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()

		rows, err := db.Query(
			`SELECT node_id, advertise_address FROM crdb_internal.gossip_nodes`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(48336)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48337)
		}
		__antithesis_instrumentation__.Notify(48335)

		for rows.Next() {
			__antithesis_instrumentation__.Notify(48338)
			rowCount++
			var nodeID int
			var advertiseAddress string
			if err := rows.Scan(&nodeID, &advertiseAddress); err != nil {
				__antithesis_instrumentation__.Notify(48340)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48341)
			}
			__antithesis_instrumentation__.Notify(48339)

			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(48342)
				if !strings.Contains(advertiseAddress, "localhost") {
					__antithesis_instrumentation__.Notify(48343)
					t.Fatal("Expected connect address to contain localhost")
				} else {
					__antithesis_instrumentation__.Notify(48344)
				}
			} else {
				__antithesis_instrumentation__.Notify(48345)
				exps, err := c.ExternalAddr(ctx, t.L(), c.Node(nodeID))
				if err != nil {
					__antithesis_instrumentation__.Notify(48347)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(48348)
				}
				__antithesis_instrumentation__.Notify(48346)
				if exps[0] != advertiseAddress {
					__antithesis_instrumentation__.Notify(48349)
					t.Fatalf("Connection address is %s but expected %s", advertiseAddress, exps[0])
				} else {
					__antithesis_instrumentation__.Notify(48350)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(48327)
	if rowCount <= 0 {
		__antithesis_instrumentation__.Notify(48351)
		t.Fatal("No results for " +
			"SELECT node_id, advertise_address FROM crdb_internal.gossip_nodes")
	} else {
		__antithesis_instrumentation__.Notify(48352)
	}
}
