package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func registerSchemaChangeInvertedIndex(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48592)
	r.Add(registry.TestSpec{
		Name:    "schemachange/invertedindex",
		Owner:   registry.OwnerSQLSchema,
		Cluster: r.MakeClusterSpec(5),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48593)
			runSchemaChangeInvertedIndex(ctx, t, c)
		},
	})
}

func runSchemaChangeInvertedIndex(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(48594)
	crdbNodes := c.Range(1, c.Spec().NodeCount-1)
	workloadNode := c.Node(c.Spec().NodeCount)

	c.Put(ctx, t.Cockroach(), "./cockroach", crdbNodes)
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", workloadNode)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes)

	cmdInit := "./workload init json {pgurl:1}"
	c.Run(ctx, workloadNode, cmdInit)

	initialDataDuration := time.Minute * 20
	indexDuration := time.Hour
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(48599)
		initialDataDuration = time.Minute
		indexDuration = time.Minute
	} else {
		__antithesis_instrumentation__.Notify(48600)
	}
	__antithesis_instrumentation__.Notify(48595)

	m := c.NewMonitor(ctx, crdbNodes)

	cmdWrite := fmt.Sprintf(
		"./workload run json --read-percent=0 --duration %s {pgurl:1-%d} --batch 1000 --sequential",
		initialDataDuration.String(), c.Spec().NodeCount-1,
	)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(48601)
		c.Run(ctx, workloadNode, cmdWrite)

		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()

		var count int
		if err := db.QueryRow(`SELECT count(*) FROM json.j`).Scan(&count); err != nil {
			__antithesis_instrumentation__.Notify(48603)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48604)
		}
		__antithesis_instrumentation__.Notify(48602)
		t.L().Printf("finished writing %d rows to table", count)

		return nil
	})
	__antithesis_instrumentation__.Notify(48596)

	m.Wait()

	m = c.NewMonitor(ctx, crdbNodes)

	cmdWriteAndRead := fmt.Sprintf(
		"./workload run json --read-percent=50 --duration %s {pgurl:1-%d} --sequential",
		indexDuration.String(), c.Spec().NodeCount-1,
	)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(48605)
		c.Run(ctx, workloadNode, cmdWriteAndRead)
		return nil
	})
	__antithesis_instrumentation__.Notify(48597)

	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(48606)
		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()

		t.L().Printf("creating index")
		start := timeutil.Now()
		if _, err := db.Exec(`CREATE INVERTED INDEX ON json.j (v)`); err != nil {
			__antithesis_instrumentation__.Notify(48608)
			return err
		} else {
			__antithesis_instrumentation__.Notify(48609)
		}
		__antithesis_instrumentation__.Notify(48607)
		t.L().Printf("index was created, took %v", timeutil.Since(start))

		return nil
	})
	__antithesis_instrumentation__.Notify(48598)

	m.Wait()
}
