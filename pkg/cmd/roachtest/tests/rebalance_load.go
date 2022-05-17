package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

func registerRebalanceLoad(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50086)

	rebalanceLoadRun := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
		rebalanceMode string,
		maxDuration time.Duration,
		concurrency int,
	) {
		__antithesis_instrumentation__.Notify(50089)
		roachNodes := c.Range(1, c.Spec().NodeCount-1)
		appNode := c.Node(c.Spec().NodeCount)
		splits := len(roachNodes) - 1

		c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)
		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--vmodule=store_rebalancer=5,allocator=5,allocator_scorer=5,replicate_queue=5")
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), roachNodes)

		c.Put(ctx, t.DeprecatedWorkload(), "./workload", appNode)
		c.Run(ctx, appNode, fmt.Sprintf("./workload init kv --drop --splits=%d {pgurl:1}", splits))

		var m *errgroup.Group
		m, ctx = errgroup.WithContext(ctx)

		ctx, cancel := context.WithCancel(ctx)

		m.Go(func() error {
			__antithesis_instrumentation__.Notify(50092)
			t.L().Printf("starting load generator\n")

			err := c.RunE(ctx, appNode, fmt.Sprintf(
				"./workload run kv --read-percent=95 --tolerate-errors --concurrency=%d "+
					"--duration=%v {pgurl:1-%d}",
				concurrency, maxDuration, len(roachNodes)))
			if errors.Is(ctx.Err(), context.Canceled) {
				__antithesis_instrumentation__.Notify(50094)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(50095)
			}
			__antithesis_instrumentation__.Notify(50093)
			return err
		})
		__antithesis_instrumentation__.Notify(50090)

		m.Go(func() error {
			__antithesis_instrumentation__.Notify(50096)
			t.Status("checking for lease balance")

			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			t.Status("disable load based splitting")
			if err := disableLoadBasedSplitting(ctx, db); err != nil {
				__antithesis_instrumentation__.Notify(50100)
				return err
			} else {
				__antithesis_instrumentation__.Notify(50101)
			}
			__antithesis_instrumentation__.Notify(50097)

			if _, err := db.ExecContext(
				ctx, `SET CLUSTER SETTING kv.allocator.load_based_rebalancing=$1::string`, rebalanceMode,
			); err != nil {
				__antithesis_instrumentation__.Notify(50102)
				return err
			} else {
				__antithesis_instrumentation__.Notify(50103)
			}
			__antithesis_instrumentation__.Notify(50098)

			for tBegin := timeutil.Now(); timeutil.Since(tBegin) <= maxDuration; {
				__antithesis_instrumentation__.Notify(50104)
				if done, err := isLoadEvenlyDistributed(t.L(), db, len(roachNodes)); err != nil {
					__antithesis_instrumentation__.Notify(50106)
					return err
				} else {
					__antithesis_instrumentation__.Notify(50107)
					if done {
						__antithesis_instrumentation__.Notify(50108)
						t.Status("successfully achieved lease balance; waiting for kv to finish running")
						cancel()
						return nil
					} else {
						__antithesis_instrumentation__.Notify(50109)
					}
				}
				__antithesis_instrumentation__.Notify(50105)

				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(50110)
					return ctx.Err()
				case <-time.After(5 * time.Second):
					__antithesis_instrumentation__.Notify(50111)
				}
			}
			__antithesis_instrumentation__.Notify(50099)

			return fmt.Errorf("timed out before leases were evenly spread")
		})
		__antithesis_instrumentation__.Notify(50091)
		if err := m.Wait(); err != nil {
			__antithesis_instrumentation__.Notify(50112)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50113)
		}
	}
	__antithesis_instrumentation__.Notify(50087)

	concurrency := 128

	r.Add(registry.TestSpec{
		Name:    `rebalance/by-load/leases`,
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50114)
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(50116)
				concurrency = 32
				fmt.Printf("lowering concurrency to %d in local testing\n", concurrency)
			} else {
				__antithesis_instrumentation__.Notify(50117)
			}
			__antithesis_instrumentation__.Notify(50115)
			rebalanceLoadRun(ctx, t, c, "leases", 3*time.Minute, concurrency)
		},
	})
	__antithesis_instrumentation__.Notify(50088)
	r.Add(registry.TestSpec{
		Name:    `rebalance/by-load/replicas`,
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(7),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50118)
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(50120)
				concurrency = 32
				fmt.Printf("lowering concurrency to %d in local testing\n", concurrency)
			} else {
				__antithesis_instrumentation__.Notify(50121)
			}
			__antithesis_instrumentation__.Notify(50119)
			rebalanceLoadRun(ctx, t, c, "leases and replicas", 5*time.Minute, concurrency)
		},
	})
}

func isLoadEvenlyDistributed(l *logger.Logger, db *gosql.DB, numNodes int) (bool, error) {
	__antithesis_instrumentation__.Notify(50122)
	rows, err := db.Query(
		`select lease_holder, count(*) ` +
			`from [show ranges from table kv.kv] ` +
			`group by lease_holder;`)
	if err != nil {
		__antithesis_instrumentation__.Notify(50130)

		if strings.Contains(err.Error(), "syntax error at or near \"ranges\"") {
			__antithesis_instrumentation__.Notify(50131)
			rows, err = db.Query(
				`select lease_holder, count(*) ` +
					`from [show experimental_ranges from table kv.kv] ` +
					`group by lease_holder;`)
		} else {
			__antithesis_instrumentation__.Notify(50132)
		}
	} else {
		__antithesis_instrumentation__.Notify(50133)
	}
	__antithesis_instrumentation__.Notify(50123)
	if err != nil {
		__antithesis_instrumentation__.Notify(50134)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(50135)
	}
	__antithesis_instrumentation__.Notify(50124)
	defer rows.Close()
	leaseCounts := make(map[int]int)
	var rangeCount int
	for rows.Next() {
		__antithesis_instrumentation__.Notify(50136)
		var storeID, leaseCount int
		if err := rows.Scan(&storeID, &leaseCount); err != nil {
			__antithesis_instrumentation__.Notify(50138)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(50139)
		}
		__antithesis_instrumentation__.Notify(50137)
		leaseCounts[storeID] = leaseCount
		rangeCount += leaseCount
	}
	__antithesis_instrumentation__.Notify(50125)

	if len(leaseCounts) < numNodes {
		__antithesis_instrumentation__.Notify(50140)
		l.Printf("not all nodes have a lease yet: %v\n", formatLeaseCounts(leaseCounts))
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(50141)
	}
	__antithesis_instrumentation__.Notify(50126)

	if rangeCount == numNodes {
		__antithesis_instrumentation__.Notify(50142)
		for _, leaseCount := range leaseCounts {
			__antithesis_instrumentation__.Notify(50144)
			if leaseCount != 1 {
				__antithesis_instrumentation__.Notify(50145)
				l.Printf("uneven lease distribution: %s\n", formatLeaseCounts(leaseCounts))
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(50146)
			}
		}
		__antithesis_instrumentation__.Notify(50143)
		l.Printf("leases successfully distributed: %s\n", formatLeaseCounts(leaseCounts))
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(50147)
	}
	__antithesis_instrumentation__.Notify(50127)

	leases := make([]int, 0, numNodes)
	for _, leaseCount := range leaseCounts {
		__antithesis_instrumentation__.Notify(50148)
		leases = append(leases, leaseCount)
	}
	__antithesis_instrumentation__.Notify(50128)
	sort.Ints(leases)
	if leases[0]+1 < leases[len(leases)-1] {
		__antithesis_instrumentation__.Notify(50149)
		l.Printf("leases per store differ by more than one: %s\n", formatLeaseCounts(leaseCounts))
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(50150)
	}
	__antithesis_instrumentation__.Notify(50129)

	l.Printf("leases successfully distributed: %s\n", formatLeaseCounts(leaseCounts))
	return true, nil
}

func formatLeaseCounts(counts map[int]int) string {
	__antithesis_instrumentation__.Notify(50151)
	storeIDs := make([]int, 0, len(counts))
	for storeID := range counts {
		__antithesis_instrumentation__.Notify(50154)
		storeIDs = append(storeIDs, storeID)
	}
	__antithesis_instrumentation__.Notify(50152)
	sort.Ints(storeIDs)
	strs := make([]string, 0, len(counts))
	for _, storeID := range storeIDs {
		__antithesis_instrumentation__.Notify(50155)
		strs = append(strs, fmt.Sprintf("s%d: %d", storeID, counts[storeID]))
	}
	__antithesis_instrumentation__.Notify(50153)
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
}
