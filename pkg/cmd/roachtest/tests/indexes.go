package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
)

func registerNIndexes(r registry.Registry, secondaryIndexes int) {
	__antithesis_instrumentation__.Notify(48560)
	const nodes = 6
	geoZones := []string{"us-east1-b", "us-west1-b", "europe-west2-b"}
	if r.MakeClusterSpec(1).Cloud == spec.AWS {
		__antithesis_instrumentation__.Notify(48562)
		geoZones = []string{"us-east-2b", "us-west-1a", "eu-west-1a"}
	} else {
		__antithesis_instrumentation__.Notify(48563)
	}
	__antithesis_instrumentation__.Notify(48561)
	geoZonesStr := strings.Join(geoZones, ",")
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("indexes/%d/nodes=%d/multi-region", secondaryIndexes, nodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(nodes+1, spec.CPU(16), spec.Geo(), spec.Zones(geoZonesStr)),

		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48564)
			firstAZ := geoZones[0]
			roachNodes := c.Range(1, nodes)
			gatewayNodes := c.Range(1, nodes/3)
			loadNode := c.Node(nodes + 1)

			c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)
			c.Put(ctx, t.DeprecatedWorkload(), "./workload", loadNode)
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)
			conn := c.Conn(ctx, t.L(), 1)

			t.Status("running workload")
			m := c.NewMonitor(ctx, roachNodes)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(48566)
				secondary := " --secondary-indexes=" + strconv.Itoa(secondaryIndexes)
				initCmd := "./workload init indexes" + secondary + " {pgurl:1}"
				c.Run(ctx, loadNode, initCmd)

				if !c.IsLocal() {
					__antithesis_instrumentation__.Notify(48569)
					t.L().Printf("setting lease preferences")
					if _, err := conn.ExecContext(ctx, fmt.Sprintf(`
						ALTER TABLE indexes.indexes
						CONFIGURE ZONE USING
						constraints = COPY FROM PARENT,
						lease_preferences = '[[+zone=%s]]'`,
						firstAZ,
					)); err != nil {
						__antithesis_instrumentation__.Notify(48572)
						return err
					} else {
						__antithesis_instrumentation__.Notify(48573)
					}
					__antithesis_instrumentation__.Notify(48570)

					t.L().Printf("checking replica balance")
					retryOpts := retry.Options{MaxBackoff: 15 * time.Second}
					for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
						__antithesis_instrumentation__.Notify(48574)
						WaitForUpdatedReplicationReport(ctx, t, conn)

						var ok bool
						if err := conn.QueryRowContext(ctx, `
							SELECT count(*) = 0
							FROM system.replication_critical_localities
							WHERE at_risk_ranges > 0
							AND locality LIKE '%region%'`,
						).Scan(&ok); err != nil {
							__antithesis_instrumentation__.Notify(48576)
							return err
						} else {
							__antithesis_instrumentation__.Notify(48577)
							if ok {
								__antithesis_instrumentation__.Notify(48578)
								break
							} else {
								__antithesis_instrumentation__.Notify(48579)
							}
						}
						__antithesis_instrumentation__.Notify(48575)

						t.L().Printf("replicas still rebalancing...")
					}
					__antithesis_instrumentation__.Notify(48571)

					t.L().Printf("checking lease preferences")
					for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
						__antithesis_instrumentation__.Notify(48580)
						var ok bool
						if err := conn.QueryRowContext(ctx, `
							SELECT lease_holder <= $1
							FROM crdb_internal.ranges
							WHERE table_name = 'indexes'`,
							nodes/3,
						).Scan(&ok); err != nil {
							__antithesis_instrumentation__.Notify(48582)
							return err
						} else {
							__antithesis_instrumentation__.Notify(48583)
							if ok {
								__antithesis_instrumentation__.Notify(48584)
								break
							} else {
								__antithesis_instrumentation__.Notify(48585)
							}
						}
						__antithesis_instrumentation__.Notify(48581)

						t.L().Printf("leases still rebalancing...")
					}
				} else {
					__antithesis_instrumentation__.Notify(48586)
				}
				__antithesis_instrumentation__.Notify(48567)

				conc := 16 * len(gatewayNodes)
				parallelWrites := (secondaryIndexes + 1) * conc
				distSenderConc := 2 * parallelWrites
				if _, err := conn.ExecContext(ctx, `
					SET CLUSTER SETTING kv.dist_sender.concurrency_limit = $1`,
					distSenderConc,
				); err != nil {
					__antithesis_instrumentation__.Notify(48587)
					return err
				} else {
					__antithesis_instrumentation__.Notify(48588)
				}
				__antithesis_instrumentation__.Notify(48568)

				payload := " --payload=64"
				concurrency := ifLocal(c, "", " --concurrency="+strconv.Itoa(conc))
				duration := " --duration=" + ifLocal(c, "10s", "10m")
				runCmd := fmt.Sprintf("./workload run indexes --histograms="+t.PerfArtifactsDir()+"/stats.json"+
					payload+concurrency+duration+" {pgurl%s}", gatewayNodes)
				c.Run(ctx, loadNode, runCmd)
				return nil
			})
			__antithesis_instrumentation__.Notify(48565)
			m.Wait()
		},
	})
}

func registerIndexes(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48589)
	registerNIndexes(r, 2)
}

func registerIndexesBench(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48590)
	for i := 0; i <= 100; i++ {
		__antithesis_instrumentation__.Notify(48591)
		registerNIndexes(r, i)
	}
}
