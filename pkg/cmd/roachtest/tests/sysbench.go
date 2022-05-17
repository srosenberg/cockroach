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
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

type sysbenchWorkload int

const (
	oltpDelete sysbenchWorkload = iota
	oltpInsert
	oltpPointSelect
	oltpUpdateIndex
	oltpUpdateNonIndex
	oltpReadOnly
	oltpReadWrite
	oltpWriteOnly

	numSysbenchWorkloads
)

var sysbenchWorkloadName = map[sysbenchWorkload]string{
	oltpDelete:         "oltp_delete",
	oltpInsert:         "oltp_insert",
	oltpPointSelect:    "oltp_point_select",
	oltpUpdateIndex:    "oltp_update_index",
	oltpUpdateNonIndex: "oltp_update_non_index",
	oltpReadOnly:       "oltp_read_only",
	oltpReadWrite:      "oltp_read_write",
	oltpWriteOnly:      "oltp_write_only",
}

func (w sysbenchWorkload) String() string {
	__antithesis_instrumentation__.Notify(51254)
	return sysbenchWorkloadName[w]
}

type sysbenchOptions struct {
	workload     sysbenchWorkload
	duration     time.Duration
	concurrency  int
	tables       int
	rowsPerTable int
}

func (o *sysbenchOptions) cmd(haproxy bool) string {
	__antithesis_instrumentation__.Notify(51255)
	pghost := "{pghost:1}"
	if haproxy {
		__antithesis_instrumentation__.Notify(51257)
		pghost = "127.0.0.1"
	} else {
		__antithesis_instrumentation__.Notify(51258)
	}
	__antithesis_instrumentation__.Notify(51256)
	return fmt.Sprintf(`sysbench \
		--db-driver=pgsql \
		--pgsql-host=%s \
		--pgsql-port=26257 \
		--pgsql-user=root \
		--pgsql-password= \
		--pgsql-db=sysbench \
		--report-interval=1 \
		--time=%d \
		--threads=%d \
		--tables=%d \
		--table_size=%d \
		--auto_inc=false \
		%s`,
		pghost,
		int(o.duration.Seconds()),
		o.concurrency,
		o.tables,
		o.rowsPerTable,
		o.workload,
	)
}

func runSysbench(ctx context.Context, t test.Test, c cluster.Cluster, opts sysbenchOptions) {
	__antithesis_instrumentation__.Notify(51259)
	allNodes := c.Range(1, c.Spec().NodeCount)
	roachNodes := c.Range(1, c.Spec().NodeCount-1)
	loadNode := c.Node(c.Spec().NodeCount)

	t.Status("installing cockroach")
	c.Put(ctx, t.Cockroach(), "./cockroach", allNodes)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)
	err := WaitFor3XReplication(ctx, t, c.Conn(ctx, t.L(), allNodes[0]))
	require.NoError(t, err)

	t.Status("installing haproxy")
	if err = c.Install(ctx, t.L(), loadNode, "haproxy"); err != nil {
		__antithesis_instrumentation__.Notify(51263)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51264)
	}
	__antithesis_instrumentation__.Notify(51260)
	c.Run(ctx, loadNode, "./cockroach gen haproxy --insecure --url {pgurl:1}")
	c.Run(ctx, loadNode, "haproxy -f haproxy.cfg -D")

	t.Status("installing sysbench")
	if err := c.Install(ctx, t.L(), loadNode, "sysbench"); err != nil {
		__antithesis_instrumentation__.Notify(51265)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51266)
	}
	__antithesis_instrumentation__.Notify(51261)

	m := c.NewMonitor(ctx, roachNodes)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(51267)
		t.Status("preparing workload")
		c.Run(ctx, c.Node(1), `./cockroach sql --insecure -e "CREATE DATABASE sysbench"`)
		c.Run(ctx, loadNode, opts.cmd(false)+" prepare")

		t.Status("running workload")
		err := c.RunE(ctx, loadNode, opts.cmd(true)+" run")

		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(51269)
			return !strings.Contains(err.Error(), "Segmentation fault") == true
		}() == true {
			__antithesis_instrumentation__.Notify(51270)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51271)
		}
		__antithesis_instrumentation__.Notify(51268)
		t.L().Printf("sysbench segfaulted; passing test anyway")
		return nil
	})
	__antithesis_instrumentation__.Notify(51262)
	m.Wait()
}

func registerSysbench(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51272)
	for w := sysbenchWorkload(0); w < numSysbenchWorkloads; w++ {
		__antithesis_instrumentation__.Notify(51273)
		const n = 3
		const cpus = 32
		const conc = 4 * cpus
		opts := sysbenchOptions{
			workload:     w,
			duration:     10 * time.Minute,
			concurrency:  conc,
			tables:       10,
			rowsPerTable: 10000000,
		}

		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("sysbench/%s/nodes=%d/cpu=%d/conc=%d", w, n, cpus, conc),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(n+1, spec.CPU(cpus)),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(51274)
				runSysbench(ctx, t, c, opts)
			},
		})
	}
}
