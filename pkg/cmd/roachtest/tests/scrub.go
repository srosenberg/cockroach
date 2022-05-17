package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func registerScrubIndexOnlyTPCC(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50746)

	r.Add(makeScrubTPCCTest(r, 5, 100, 30*time.Minute, "index-only", 20))
}

func registerScrubAllChecksTPCC(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50747)

	r.Add(makeScrubTPCCTest(r, 5, 100, 30*time.Minute, "all-checks", 10))
}

func makeScrubTPCCTest(
	r registry.Registry,
	numNodes, warehouses int,
	length time.Duration,
	optionName string,
	numScrubRuns int,
) registry.TestSpec {
	__antithesis_instrumentation__.Notify(50748)
	var stmtOptions string

	switch optionName {
	case "index-only":
		__antithesis_instrumentation__.Notify(50750)
		stmtOptions = `AS OF SYSTEM TIME '-1m' WITH OPTIONS INDEX ALL`
	case "all-checks":
		__antithesis_instrumentation__.Notify(50751)
		stmtOptions = `AS OF SYSTEM TIME '-1m'`
	default:
		__antithesis_instrumentation__.Notify(50752)
		panic(fmt.Sprintf("Not a valid option: %s", optionName))
	}
	__antithesis_instrumentation__.Notify(50749)

	return registry.TestSpec{
		Name:    fmt.Sprintf("scrub/%s/tpcc/w=%d", optionName, warehouses),
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50753)
			runTPCC(ctx, t, c, tpccOptions{
				Warehouses:   warehouses,
				ExtraRunArgs: "--wait=false --tolerate-errors",
				During: func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(50754)
					if !c.IsLocal() {
						__antithesis_instrumentation__.Notify(50757)

						sleepInterval := time.Minute * 10
						maxSleep := length / 2
						if sleepInterval > maxSleep {
							__antithesis_instrumentation__.Notify(50759)
							sleepInterval = maxSleep
						} else {
							__antithesis_instrumentation__.Notify(50760)
						}
						__antithesis_instrumentation__.Notify(50758)
						time.Sleep(sleepInterval)
					} else {
						__antithesis_instrumentation__.Notify(50761)
					}
					__antithesis_instrumentation__.Notify(50755)

					conn := c.Conn(ctx, t.L(), 1)
					defer conn.Close()

					t.L().Printf("Starting %d SCRUB checks", numScrubRuns)
					for i := 0; i < numScrubRuns; i++ {
						__antithesis_instrumentation__.Notify(50762)
						t.L().Printf("Running SCRUB check %d\n", i+1)
						before := timeutil.Now()
						err := sqlutils.RunScrubWithOptions(conn, "tpcc", "order", stmtOptions)
						t.L().Printf("SCRUB check %d took %v\n", i+1, timeutil.Since(before))

						if err != nil {
							__antithesis_instrumentation__.Notify(50763)
							t.Fatal(err)
						} else {
							__antithesis_instrumentation__.Notify(50764)
						}
					}
					__antithesis_instrumentation__.Notify(50756)
					return nil
				},
				DisablePrometheus: true,
				Duration:          length,
				SetupType:         usingImport,
			})
		},
	}
}
