package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/cmpconn"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload/tpcds"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func registerTPCDSVec(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51805)
	const (
		timeout                         = 5 * time.Minute
		withStatsSlowerWarningThreshold = 1.25
	)

	queriesToSkip := map[int]bool{

		1:  true,
		64: true,

		5:  true,
		14: true,
		18: true,
		22: true,
		67: true,
		77: true,
		80: true,
	}

	tpcdsTables := []string{
		`call_center`, `catalog_page`, `catalog_returns`, `catalog_sales`,
		`customer`, `customer_address`, `customer_demographics`, `date_dim`,
		`dbgen_version`, `household_demographics`, `income_band`, `inventory`,
		`item`, `promotion`, `reason`, `ship_mode`, `store`, `store_returns`,
		`store_sales`, `time_dim`, `warehouse`, `web_page`, `web_returns`,
		`web_sales`, `web_site`,
	}

	runTPCDSVec := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(51807)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

		clusterConn := c.Conn(ctx, t.L(), 1)
		disableAutoStats(t, clusterConn)
		t.Status("restoring TPCDS dataset for Scale Factor 1")
		if _, err := clusterConn.Exec(
			`RESTORE DATABASE tpcds FROM 'gs://cockroach-fixtures/workload/tpcds/scalefactor=1/backup?AUTH=implicit';`,
		); err != nil {
			__antithesis_instrumentation__.Notify(51813)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51814)
		}
		__antithesis_instrumentation__.Notify(51808)

		if _, err := clusterConn.Exec("USE tpcds;"); err != nil {
			__antithesis_instrumentation__.Notify(51815)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51816)
		}
		__antithesis_instrumentation__.Notify(51809)
		scatterTables(t, clusterConn, tpcdsTables)
		t.Status("waiting for full replication")
		err := WaitFor3XReplication(ctx, t, clusterConn)
		require.NoError(t, err)

		setStmtTimeout := fmt.Sprintf("SET statement_timeout='%s';", timeout)
		firstNode := c.Node(1)
		urls, err := c.ExternalPGUrl(ctx, t.L(), firstNode)
		if err != nil {
			__antithesis_instrumentation__.Notify(51817)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51818)
		}
		__antithesis_instrumentation__.Notify(51810)
		firstNodeURL := urls[0]
		openNewConnections := func() (map[string]cmpconn.Conn, func()) {
			__antithesis_instrumentation__.Notify(51819)
			conns := map[string]cmpconn.Conn{}
			vecOffConn, err := cmpconn.NewConn(
				ctx, firstNodeURL, setStmtTimeout+"SET vectorize=off; USE tpcds;",
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(51823)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51824)
			}
			__antithesis_instrumentation__.Notify(51820)
			conns["vectorize=OFF"] = vecOffConn
			vecOnConn, err := cmpconn.NewConn(
				ctx, firstNodeURL, setStmtTimeout+"SET vectorize=on; USE tpcds;",
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(51825)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51826)
			}
			__antithesis_instrumentation__.Notify(51821)
			conns["vectorize=ON"] = vecOnConn

			if _, err := cmpconn.CompareConns(
				ctx, timeout, conns, "", "SHOW vectorize;", false,
			); err == nil {
				__antithesis_instrumentation__.Notify(51827)
				t.Fatal("unexpectedly SHOW vectorize didn't trigger an error on comparison")
			} else {
				__antithesis_instrumentation__.Notify(51828)
			}
			__antithesis_instrumentation__.Notify(51822)
			return conns, func() {
				__antithesis_instrumentation__.Notify(51829)
				vecOffConn.Close(ctx)
				vecOnConn.Close(ctx)
			}
		}
		__antithesis_instrumentation__.Notify(51811)

		noStatsRunTimes := make(map[int]float64)
		var errToReport error

		for _, haveStats := range []bool{false, true} {
			__antithesis_instrumentation__.Notify(51830)
			for queryNum := 1; queryNum <= tpcds.NumQueries; queryNum++ {
				__antithesis_instrumentation__.Notify(51832)
				if _, toSkip := queriesToSkip[queryNum]; toSkip {
					__antithesis_instrumentation__.Notify(51836)
					continue
				} else {
					__antithesis_instrumentation__.Notify(51837)
				}
				__antithesis_instrumentation__.Notify(51833)
				query, ok := tpcds.QueriesByNumber[queryNum]
				if !ok {
					__antithesis_instrumentation__.Notify(51838)
					continue
				} else {
					__antithesis_instrumentation__.Notify(51839)
				}
				__antithesis_instrumentation__.Notify(51834)
				t.Status(fmt.Sprintf("running query %d\n", queryNum))

				conns, cleanup := openNewConnections()
				start := timeutil.Now()
				if _, err := cmpconn.CompareConns(
					ctx, 3*timeout, conns, "", query, false,
				); err != nil {
					__antithesis_instrumentation__.Notify(51840)
					t.Status(fmt.Sprintf("encountered an error: %s\n", err))
					errToReport = errors.CombineErrors(errToReport, err)
				} else {
					__antithesis_instrumentation__.Notify(51841)
					runTimeInSeconds := timeutil.Since(start).Seconds()
					t.Status(
						fmt.Sprintf("[q%d] took about %.2fs to run on both configs",
							queryNum, runTimeInSeconds),
					)
					if haveStats {
						__antithesis_instrumentation__.Notify(51842)
						noStatsRunTime, ok := noStatsRunTimes[queryNum]
						if ok && func() bool {
							__antithesis_instrumentation__.Notify(51843)
							return noStatsRunTime*withStatsSlowerWarningThreshold < runTimeInSeconds == true
						}() == true {
							__antithesis_instrumentation__.Notify(51844)
							t.Status(fmt.Sprintf("WARNING: suboptimal plan when stats are present\n"+
								"no stats: %.2fs\twith stats: %.2fs", noStatsRunTime, runTimeInSeconds))
						} else {
							__antithesis_instrumentation__.Notify(51845)
						}
					} else {
						__antithesis_instrumentation__.Notify(51846)
						noStatsRunTimes[queryNum] = runTimeInSeconds
					}
				}
				__antithesis_instrumentation__.Notify(51835)
				cleanup()
			}
			__antithesis_instrumentation__.Notify(51831)

			if !haveStats {
				__antithesis_instrumentation__.Notify(51847)
				createStatsFromTables(t, clusterConn, tpcdsTables)
			} else {
				__antithesis_instrumentation__.Notify(51848)
			}
		}
		__antithesis_instrumentation__.Notify(51812)
		if errToReport != nil {
			__antithesis_instrumentation__.Notify(51849)
			t.Fatal(errToReport)
		} else {
			__antithesis_instrumentation__.Notify(51850)
		}
	}
	__antithesis_instrumentation__.Notify(51806)

	r.Add(registry.TestSpec{
		Name:    "tpcdsvec",
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51851)
			runTPCDSVec(ctx, t, c)
		},
	})
}
