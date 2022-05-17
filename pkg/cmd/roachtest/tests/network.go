package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	toxiproxy "github.com/Shopify/toxiproxy/client"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func runNetworkSanity(ctx context.Context, t test.Test, origC cluster.Cluster, nodes int) {
	__antithesis_instrumentation__.Notify(49460)
	origC.Put(ctx, t.Cockroach(), "./cockroach", origC.All())
	c, err := Toxify(ctx, t, origC, origC.All())
	if err != nil {
		__antithesis_instrumentation__.Notify(49464)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49465)
	}
	__antithesis_instrumentation__.Notify(49461)

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()
	err = WaitFor3XReplication(ctx, t, db)
	require.NoError(t, err)

	const latency = 300 * time.Millisecond
	for i := 1; i <= nodes; i++ {
		__antithesis_instrumentation__.Notify(49466)

		proxy := c.Proxy(i)
		if _, err := proxy.AddToxic("", "latency", "downstream", 1, toxiproxy.Attributes{
			"latency": latency / (2 * time.Millisecond),
		}); err != nil {
			__antithesis_instrumentation__.Notify(49468)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49469)
		}
		__antithesis_instrumentation__.Notify(49467)
		if _, err := proxy.AddToxic("", "latency", "upstream", 1, toxiproxy.Attributes{
			"latency": latency / (2 * time.Millisecond),
		}); err != nil {
			__antithesis_instrumentation__.Notify(49470)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49471)
		}
	}
	__antithesis_instrumentation__.Notify(49462)

	m := c.Cluster.NewMonitor(ctx, c.All())
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(49472)
		c.Measure(ctx, 1, `SET CLUSTER SETTING trace.debug.enable = true`)
		c.Measure(ctx, 1, "CREATE DATABASE test")
		c.Measure(ctx, 1, `CREATE TABLE test.commit (a INT, b INT, v INT, PRIMARY KEY (a, b))`)

		for i := 0; i < 10; i++ {
			__antithesis_instrumentation__.Notify(49475)
			duration := c.Measure(ctx, 1, fmt.Sprintf(
				"BEGIN; INSERT INTO test.commit VALUES (2, %[1]d), (1, %[1]d), (3, %[1]d); COMMIT",
				i,
			))
			t.L().Printf("%s\n", duration)
		}
		__antithesis_instrumentation__.Notify(49473)

		c.Measure(ctx, 1, `
set tracing=on;
insert into test.commit values(3,1000), (1,1000), (2,1000);
select age, message from [ show trace for session ];
`)

		for i := 1; i <= origC.Spec().NodeCount; i++ {
			__antithesis_instrumentation__.Notify(49476)
			if dur := c.Measure(ctx, i, `SELECT 1`); dur > latency {
				__antithesis_instrumentation__.Notify(49477)
				t.Fatalf("node %d unexpectedly affected by latency: select 1 took %.2fs", i, dur.Seconds())
			} else {
				__antithesis_instrumentation__.Notify(49478)
			}
		}
		__antithesis_instrumentation__.Notify(49474)

		return nil
	})
	__antithesis_instrumentation__.Notify(49463)

	m.Wait()
}

func runNetworkAuthentication(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(49479)
	n := c.Spec().NodeCount
	serverNodes, clientNode := c.Range(1, n-1), c.Node(n)

	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())

	t.L().Printf("starting nodes to initialize TLS certs...")

	settings := install.MakeClusterSettings(install.SecureOption(true))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, serverNodes)
	require.NoError(t, c.StopE(ctx, t.L(), option.DefaultStopOpts(), serverNodes))

	t.L().Printf("restarting nodes...")

	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--locality=node=1", "--accept-sql-without-tls")
	c.Start(ctx, t.L(), startOpts, settings, c.Node(1))

	startOpts = option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--locality=node=other", "--accept-sql-without-tls")
	c.Start(ctx, t.L(), startOpts, settings, c.Range(2, n-1))

	t.L().Printf("retrieving server addresses...")
	serverAddrs, err := c.InternalAddr(ctx, t.L(), serverNodes)
	require.NoError(t, err)

	t.L().Printf("fetching certs...")
	certsDir := "/home/ubuntu/certs"
	localCertsDir, err := filepath.Abs("./network-certs")
	require.NoError(t, err)
	require.NoError(t, os.RemoveAll(localCertsDir))
	require.NoError(t, c.Get(ctx, t.L(), certsDir, localCertsDir, c.Node(1)))
	require.NoError(t, filepath.Walk(localCertsDir, func(path string, info os.FileInfo, err error) error {
		__antithesis_instrumentation__.Notify(49485)

		if path == localCertsDir {
			__antithesis_instrumentation__.Notify(49488)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(49489)
		}
		__antithesis_instrumentation__.Notify(49486)
		if err != nil {
			__antithesis_instrumentation__.Notify(49490)
			return err
		} else {
			__antithesis_instrumentation__.Notify(49491)
		}
		__antithesis_instrumentation__.Notify(49487)
		return os.Chmod(path, os.FileMode(0600))
	}))
	__antithesis_instrumentation__.Notify(49480)

	t.L().Printf("connecting to cluster from roachtest...")
	db, err := c.ConnE(ctx, t.L(), 1)
	require.NoError(t, err)
	defer db.Close()

	err = WaitFor3XReplication(ctx, t, db)
	require.NoError(t, err)

	t.L().Printf("creating test user...")
	_, err = db.Exec(`CREATE USER testuser WITH PASSWORD 'password' VALID UNTIL '2060-01-01'`)
	require.NoError(t, err)
	_, err = db.Exec(`GRANT admin TO testuser`)
	require.NoError(t, err)

	const expectedLeaseholder = 1
	lh := fmt.Sprintf("%d", expectedLeaseholder)

	t.L().Printf("configuring zones to move ranges to node 1...")
	for _, zone := range []string{
		`RANGE liveness`,
		`RANGE meta`,
		`RANGE system`,
		`RANGE default`,
		`DATABASE system`,
	} {
		__antithesis_instrumentation__.Notify(49492)
		zoneCmd := `ALTER ` + zone + ` CONFIGURE ZONE USING lease_preferences = '[[+node=` + lh + `]]', constraints = '{"+node=` + lh + `": 1}'`
		t.L().Printf("SQL: %s", zoneCmd)
		_, err = db.Exec(zoneCmd)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(49481)

	t.L().Printf("waiting for leases to move...")
	{
		__antithesis_instrumentation__.Notify(49493)
		tStart := timeutil.Now()
		for ok := false; !ok; time.Sleep(time.Second) {
			__antithesis_instrumentation__.Notify(49494)
			if timeutil.Since(tStart) > 30*time.Second {
				__antithesis_instrumentation__.Notify(49496)
				t.L().Printf("still waiting for leases to move")

				dumpRangesCmd := fmt.Sprintf(`./cockroach sql --certs-dir %s -e 'TABLE crdb_internal.ranges'`, certsDir)
				t.L().Printf("SQL: %s", dumpRangesCmd)
				err := c.RunE(ctx, c.Node(1), dumpRangesCmd)
				require.NoError(t, err)
			} else {
				__antithesis_instrumentation__.Notify(49497)
			}
			__antithesis_instrumentation__.Notify(49495)

			const waitLeases = `
SELECT $1::INT = ALL (
    SELECT lease_holder
    FROM   crdb_internal.ranges
    WHERE  (start_pretty = '/System/NodeLiveness' AND end_pretty = '/System/NodeLivenessMax')
       OR  (database_name = 'system' AND table_name IN ('users', 'role_members', 'role_options'))
)`
			t.L().Printf("SQL: %s", waitLeases)
			require.NoError(t, db.QueryRow(waitLeases, expectedLeaseholder).Scan(&ok))
		}
	}
	__antithesis_instrumentation__.Notify(49482)

	cancelTestCtx, cancelTest := context.WithCancel(ctx)

	woopsCh := make(chan struct{}, len(serverNodes)-1)

	m := c.NewMonitor(ctx)

	var numConns uint32

	for i := 1; i <= c.Spec().NodeCount-1; i++ {
		__antithesis_instrumentation__.Notify(49498)
		if i == expectedLeaseholder {
			__antithesis_instrumentation__.Notify(49500)
			continue
		} else {
			__antithesis_instrumentation__.Notify(49501)
		}
		__antithesis_instrumentation__.Notify(49499)

		server := i

		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(49502)
			errCount := 0
			for attempt := 0; ; attempt++ {
				__antithesis_instrumentation__.Notify(49503)
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(49505)

					t.L().Printf("server %d: stopping connections due to error", server)

					woopsCh <- struct{}{}

					return ctx.Err()

				case <-cancelTestCtx.Done():
					__antithesis_instrumentation__.Notify(49506)

					t.L().Printf("server %d: stopping connections due to end of test", server)
					return nil

				case <-time.After(500 * time.Millisecond):
					__antithesis_instrumentation__.Notify(49507)

				}
				__antithesis_instrumentation__.Notify(49504)

				url := fmt.Sprintf("postgres://testuser:password@%s/defaultdb?sslmode=require", serverAddrs[server-1])

				t.L().Printf("server %d, attempt %d; url: %s\n", server, attempt, url)

				b, err := c.RunWithDetailsSingleNode(ctx, t.L(), clientNode, "time", "-p", "./cockroach", "sql",
					"--url", url, "--certs-dir", certsDir, "-e", "'SELECT 1'")

				t.L().Printf("server %d, attempt %d, result:\n%s\n", server, attempt, b)

				atomic.AddUint32(&numConns, 1)

				if err != nil {
					__antithesis_instrumentation__.Notify(49508)
					if errCount == 0 {
						__antithesis_instrumentation__.Notify(49510)

						t.L().Printf("server %d, attempt %d (1st ERROR, TOLERATE): %v", server, attempt, err)
						errCount++
						continue
					} else {
						__antithesis_instrumentation__.Notify(49511)
					}
					__antithesis_instrumentation__.Notify(49509)

					t.L().Printf("server %d, attempt %d (2nd ERROR, BAD): %v", server, attempt, err)

					woopsCh <- struct{}{}
					return err
				} else {
					__antithesis_instrumentation__.Notify(49512)
				}
			}
		})
	}
	__antithesis_instrumentation__.Notify(49483)

	func() {
		__antithesis_instrumentation__.Notify(49513)
		t.L().Printf("waiting for clients to start connecting...")
		testutils.SucceedsSoon(t, func() error {
			__antithesis_instrumentation__.Notify(49517)
			select {
			case <-woopsCh:
				__antithesis_instrumentation__.Notify(49520)
				t.Fatal("connection error before network partition")
			default:
				__antithesis_instrumentation__.Notify(49521)
			}
			__antithesis_instrumentation__.Notify(49518)
			if atomic.LoadUint32(&numConns) == 0 {
				__antithesis_instrumentation__.Notify(49522)
				return errors.New("no connection yet")
			} else {
				__antithesis_instrumentation__.Notify(49523)
			}
			__antithesis_instrumentation__.Notify(49519)
			return nil
		})
		__antithesis_instrumentation__.Notify(49514)

		t.L().Printf("blocking networking on node 1...")
		const netConfigCmd = `
# ensure any failure fails the entire script.
set -e;

# Setting default filter policy
sudo iptables -P INPUT ACCEPT;
sudo iptables -P OUTPUT ACCEPT;

# Drop any node-to-node crdb traffic.
sudo iptables -A INPUT -p tcp --dport 26257 -j DROP;
sudo iptables -A OUTPUT -p tcp --dport 26257 -j DROP;

sudo iptables-save
`
		t.L().Printf("partitioning using iptables; config cmd:\n%s", netConfigCmd)
		require.NoError(t, c.RunE(ctx, c.Node(expectedLeaseholder), netConfigCmd))

		defer func() {
			__antithesis_instrumentation__.Notify(49524)
			const restoreNet = `
set -e;
sudo iptables -D INPUT -p tcp --dport 26257 -j DROP;
sudo iptables -D OUTPUT -p tcp --dport 26257 -j DROP;
sudo iptables-save
`
			t.L().Printf("restoring iptables; config cmd:\n%s", restoreNet)
			require.NoError(t, c.RunE(ctx, c.Node(expectedLeaseholder), restoreNet))
		}()
		__antithesis_instrumentation__.Notify(49515)

		t.L().Printf("waiting while clients attempt to connect...")
		select {
		case <-time.After(20 * time.Second):
			__antithesis_instrumentation__.Notify(49525)
		case <-woopsCh:
			__antithesis_instrumentation__.Notify(49526)
		}
		__antithesis_instrumentation__.Notify(49516)

		cancelTest()
	}()
	__antithesis_instrumentation__.Notify(49484)

	m.Wait()
}

func runNetworkTPCC(ctx context.Context, t test.Test, origC cluster.Cluster, nodes int) {
	__antithesis_instrumentation__.Notify(49527)
	n := origC.Spec().NodeCount
	serverNodes, workerNode := origC.Range(1, n-1), origC.Node(n)
	origC.Put(ctx, t.Cockroach(), "./cockroach", origC.All())
	origC.Put(ctx, t.DeprecatedWorkload(), "./workload", origC.All())

	c, err := Toxify(ctx, t, origC, serverNodes)
	if err != nil {
		__antithesis_instrumentation__.Notify(49533)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49534)
	}
	__antithesis_instrumentation__.Notify(49528)

	const warehouses = 1
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), serverNodes)
	c.Run(ctx, c.Node(1), tpccImportCmd(warehouses))

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()
	err = WaitFor3XReplication(ctx, t, db)
	require.NoError(t, err)

	duration := time.Hour
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(49535)

		duration = 5 * time.Minute
	} else {
		__antithesis_instrumentation__.Notify(49536)
	}
	__antithesis_instrumentation__.Notify(49529)

	m := c.NewMonitor(ctx, serverNodes)

	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(49537)
		t.WorkerStatus("running tpcc")

		cmd := fmt.Sprintf(
			"./workload run tpcc --warehouses=%d --wait=false"+
				" --histograms="+t.PerfArtifactsDir()+"/stats.json"+
				" --duration=%s {pgurl:2-%d}",
			warehouses, duration, c.Spec().NodeCount-1)
		return c.RunE(ctx, workerNode, cmd)
	})
	__antithesis_instrumentation__.Notify(49530)

	checkGoroutines := func(ctx context.Context) int {
		__antithesis_instrumentation__.Notify(49538)

		const thresh = 350

		uiAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), serverNodes)
		if err != nil {
			__antithesis_instrumentation__.Notify(49541)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49542)
		}
		__antithesis_instrumentation__.Notify(49539)
		var maxSeen int

		httpClient := httputil.NewClientWithTimeout(15 * time.Second)
		for _, addr := range uiAddrs {
			__antithesis_instrumentation__.Notify(49543)
			url := "http://" + addr + "/debug/pprof/goroutine?debug=2"
			resp, err := httpClient.Get(ctx, url)
			if err != nil {
				__antithesis_instrumentation__.Notify(49547)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49548)
			}
			__antithesis_instrumentation__.Notify(49544)
			content, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				__antithesis_instrumentation__.Notify(49549)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49550)
			}
			__antithesis_instrumentation__.Notify(49545)
			numGoroutines := bytes.Count(content, []byte("goroutine "))
			if numGoroutines >= thresh {
				__antithesis_instrumentation__.Notify(49551)
				t.Fatalf("%s shows %d goroutines (expected <%d)", url, numGoroutines, thresh)
			} else {
				__antithesis_instrumentation__.Notify(49552)
			}
			__antithesis_instrumentation__.Notify(49546)
			if maxSeen < numGoroutines {
				__antithesis_instrumentation__.Notify(49553)
				maxSeen = numGoroutines
			} else {
				__antithesis_instrumentation__.Notify(49554)
			}
		}
		__antithesis_instrumentation__.Notify(49540)
		return maxSeen
	}
	__antithesis_instrumentation__.Notify(49531)

	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(49555)
		time.Sleep(10 * time.Second)

		proxy := c.Proxy(1)
		t.L().Printf("letting inbound traffic to first node time out")
		for _, direction := range []string{"upstream", "downstream"} {
			__antithesis_instrumentation__.Notify(49557)
			if _, err := proxy.AddToxic("", "timeout", direction, 1, toxiproxy.Attributes{
				"timeout": 0,
			}); err != nil {
				__antithesis_instrumentation__.Notify(49558)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49559)
			}
		}
		__antithesis_instrumentation__.Notify(49556)

		t.WorkerStatus("checking goroutines")
		done := time.After(duration)
		var maxSeen int
		for {
			__antithesis_instrumentation__.Notify(49560)
			cur := checkGoroutines(ctx)
			if maxSeen < cur {
				__antithesis_instrumentation__.Notify(49562)
				t.L().Printf("new goroutine peak: %d", cur)
				maxSeen = cur
			} else {
				__antithesis_instrumentation__.Notify(49563)
			}
			__antithesis_instrumentation__.Notify(49561)

			select {
			case <-done:
				__antithesis_instrumentation__.Notify(49564)
				t.L().Printf("done checking goroutines, repairing network")

				toxics, err := proxy.Toxics()
				if err != nil {
					__antithesis_instrumentation__.Notify(49569)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(49570)
				}
				__antithesis_instrumentation__.Notify(49565)
				for _, toxic := range toxics {
					__antithesis_instrumentation__.Notify(49571)
					if err := proxy.RemoveToxic(toxic.Name); err != nil {
						__antithesis_instrumentation__.Notify(49572)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(49573)
					}
				}
				__antithesis_instrumentation__.Notify(49566)
				t.L().Printf("network is repaired")

				for i := 0; i < 20; i++ {
					__antithesis_instrumentation__.Notify(49574)
					nowGoroutines := checkGoroutines(ctx)
					t.L().Printf("currently at most %d goroutines per node", nowGoroutines)
					time.Sleep(time.Second)
				}
				__antithesis_instrumentation__.Notify(49567)

				return nil
			default:
				__antithesis_instrumentation__.Notify(49568)
				time.Sleep(3 * time.Second)
			}
		}
	})
	__antithesis_instrumentation__.Notify(49532)

	m.Wait()
}

func registerNetwork(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49575)
	const numNodes = 4

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("network/sanity/nodes=%d", numNodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Skip:    "deleted in 22.2",
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49578)
			runNetworkSanity(ctx, t, c, numNodes)
		},
	})
	__antithesis_instrumentation__.Notify(49576)
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("network/authentication/nodes=%d", numNodes),
		Owner:   registry.OwnerServer,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49579)
			runNetworkAuthentication(ctx, t, c)
		},
	})
	__antithesis_instrumentation__.Notify(49577)
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("network/tpcc/nodes=%d", numNodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Skip:    "https://github.com/cockroachdb/cockroach/issues/49901#issuecomment-640666646",
		SkipDetails: `The ordering of steps in the test is:

- install toxiproxy
- start cluster, wait for up-replication
- launch the goroutine that starts the tpcc client command, but do not wait on
it starting
- immediately, cause a network partition
- only then, the goroutine meant to start the tpcc client goes to fetch the
pg URLs and start workload, but of course this fails because network
partition
- tpcc fails to start, so the test tears down before it resolves the network partition
- test tear-down and debug zip fail because the network partition is still active

There are two problems here:

the tpcc client is not actually started yet when the test sets up the
network partition. This is a race condition. there should be a defer in
there to resolve the partition when the test aborts prematurely. (And the
command to resolve the partition should not be sensitive to the test
context's Done() channel, because during a tear-down that is closed already)
`,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49580)
			runNetworkTPCC(ctx, t, c, numNodes)
		},
	})
}
