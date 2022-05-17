package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload/tpch"
	"github.com/cockroachdb/errors"
)

func registerCancel(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46015)
	runCancel := func(ctx context.Context, t test.Test, c cluster.Cluster, tpchQueriesToRun []int, useDistsql bool) {
		__antithesis_instrumentation__.Notify(46019)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		m := c.NewMonitor(ctx, c.All())
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(46021)
			t.Status("restoring TPCH dataset for Scale Factor 1")
			if err := loadTPCHDataset(ctx, t, c, 1, c.NewMonitor(ctx), c.All()); err != nil {
				__antithesis_instrumentation__.Notify(46025)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(46026)
			}
			__antithesis_instrumentation__.Notify(46022)

			conn := c.Conn(ctx, t.L(), 1)
			defer conn.Close()

			queryPrefix := "USE tpch; "
			if !useDistsql {
				__antithesis_instrumentation__.Notify(46027)
				queryPrefix += "SET distsql = off; "
			} else {
				__antithesis_instrumentation__.Notify(46028)
			}
			__antithesis_instrumentation__.Notify(46023)

			t.Status("running queries to cancel")
			for _, queryNum := range tpchQueriesToRun {
				__antithesis_instrumentation__.Notify(46029)

				sem := make(chan struct{})

				errCh := make(chan error, 1)
				go func(query string) {
					__antithesis_instrumentation__.Notify(46031)
					defer close(errCh)
					t.L().Printf("executing q%d\n", queryNum)
					sem <- struct{}{}
					close(sem)
					_, err := conn.Exec(queryPrefix + query)
					if err == nil {
						__antithesis_instrumentation__.Notify(46032)
						errCh <- errors.New("query completed before it could be canceled")
					} else {
						__antithesis_instrumentation__.Notify(46033)
						fmt.Printf("query failed with error: %s\n", err)

						if !strings.Contains(err.Error(), cancelchecker.QueryCanceledError.Error()) {
							__antithesis_instrumentation__.Notify(46034)
							errCh <- errors.Wrap(err, "unexpected error")
						} else {
							__antithesis_instrumentation__.Notify(46035)
						}
					}
				}(tpch.QueriesByNumber[queryNum])
				__antithesis_instrumentation__.Notify(46030)

				<-sem

				time.Sleep(250 * time.Millisecond)

				const cancelQuery = `CANCEL QUERIES
	SELECT query_id FROM [SHOW CLUSTER QUERIES] WHERE query not like '%SHOW CLUSTER QUERIES%'`
				c.Run(ctx, c.Node(1), `./cockroach sql --insecure -e "`+cancelQuery+`"`)
				cancelStartTime := timeutil.Now()

				select {
				case err, ok := <-errCh:
					__antithesis_instrumentation__.Notify(46036)
					if ok {
						__antithesis_instrumentation__.Notify(46039)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(46040)
					}
					__antithesis_instrumentation__.Notify(46037)

					timeToCancel := timeutil.Since(cancelStartTime)
					fmt.Printf("canceling q%d took %s\n", queryNum, timeToCancel)

				case <-time.After(5 * time.Second):
					__antithesis_instrumentation__.Notify(46038)
					t.Fatal("query took too long to respond to cancellation")
				}
			}
			__antithesis_instrumentation__.Notify(46024)

			return nil
		})
		__antithesis_instrumentation__.Notify(46020)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(46016)

	const numNodes = 3

	tpchQueriesToRun := []int{7, 9, 20, 21}
	var queries string
	for i, q := range tpchQueriesToRun {
		__antithesis_instrumentation__.Notify(46041)
		if i > 0 {
			__antithesis_instrumentation__.Notify(46043)
			queries += ","
		} else {
			__antithesis_instrumentation__.Notify(46044)
		}
		__antithesis_instrumentation__.Notify(46042)
		queries += fmt.Sprintf("%d", q)
	}
	__antithesis_instrumentation__.Notify(46017)

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("cancel/tpch/distsql/queries=%s,nodes=%d", queries, numNodes),
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46045)
			runCancel(ctx, t, c, tpchQueriesToRun, true)
		},
	})
	__antithesis_instrumentation__.Notify(46018)

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("cancel/tpch/local/queries=%s,nodes=%d", queries, numNodes),
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46046)
			runCancel(ctx, t, c, tpchQueriesToRun, false)
		},
	})
}
