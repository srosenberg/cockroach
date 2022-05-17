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
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/errors"
)

func registerAutoUpgrade(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45501)
	runAutoUpgrade := func(ctx context.Context, t test.Test, c cluster.Cluster, oldVersion string) {
		__antithesis_instrumentation__.Notify(45503)
		nodes := c.Spec().NodeCount

		if err := c.Stage(ctx, t.L(), "release", "v"+oldVersion, "", c.Range(1, nodes)); err != nil {
			__antithesis_instrumentation__.Notify(45532)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45533)
		}
		__antithesis_instrumentation__.Notify(45504)

		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, nodes))

		const stageDuration = 30 * time.Second
		const timeUntilStoreDead = 90 * time.Second
		const buff = 10 * time.Second

		sleep := func(ts time.Duration) error {
			__antithesis_instrumentation__.Notify(45534)
			t.WorkerStatus("sleeping")
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(45535)
				return ctx.Err()
			case <-time.After(ts):
				__antithesis_instrumentation__.Notify(45536)
				return nil
			}
		}
		__antithesis_instrumentation__.Notify(45505)

		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()

		if _, err := db.ExecContext(ctx,
			"SET CLUSTER SETTING server.time_until_store_dead = $1", timeUntilStoreDead.String(),
		); err != nil {
			__antithesis_instrumentation__.Notify(45537)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45538)
		}
		__antithesis_instrumentation__.Notify(45506)

		if err := sleep(stageDuration); err != nil {
			__antithesis_instrumentation__.Notify(45539)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45540)
		}
		__antithesis_instrumentation__.Notify(45507)

		decommissionAndStop := func(node int) error {
			__antithesis_instrumentation__.Notify(45541)
			t.WorkerStatus("decommission")
			port := fmt.Sprintf("{pgport:%d}", node)
			if err := c.RunE(ctx, c.Node(node),
				fmt.Sprintf("./cockroach node decommission %d --insecure --port=%s", node, port)); err != nil {
				__antithesis_instrumentation__.Notify(45543)
				return err
			} else {
				__antithesis_instrumentation__.Notify(45544)
			}
			__antithesis_instrumentation__.Notify(45542)
			t.WorkerStatus("stop")
			c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(node))
			return nil
		}
		__antithesis_instrumentation__.Notify(45508)

		clusterVersion := func() (string, error) {
			__antithesis_instrumentation__.Notify(45545)
			var version string
			if err := db.QueryRowContext(ctx, `SHOW CLUSTER SETTING version`).Scan(&version); err != nil {
				__antithesis_instrumentation__.Notify(45547)
				return "", errors.Wrap(err, "determining cluster version")
			} else {
				__antithesis_instrumentation__.Notify(45548)
			}
			__antithesis_instrumentation__.Notify(45546)
			return version, nil
		}
		__antithesis_instrumentation__.Notify(45509)

		oldVersion, err := clusterVersion()
		if err != nil {
			__antithesis_instrumentation__.Notify(45549)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45550)
		}
		__antithesis_instrumentation__.Notify(45510)

		checkUpgraded := func() (bool, error) {
			__antithesis_instrumentation__.Notify(45551)
			upgradedVersion, err := clusterVersion()
			if err != nil {
				__antithesis_instrumentation__.Notify(45553)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(45554)
			}
			__antithesis_instrumentation__.Notify(45552)
			return upgradedVersion != oldVersion, nil
		}
		__antithesis_instrumentation__.Notify(45511)

		checkDowngradeOption := func(version string) error {
			__antithesis_instrumentation__.Notify(45555)
			if _, err := db.ExecContext(ctx,
				"SET CLUSTER SETTING cluster.preserve_downgrade_option = $1;", version,
			); err == nil {
				__antithesis_instrumentation__.Notify(45557)
				return fmt.Errorf("cluster.preserve_downgrade_option shouldn't be set to any other values besides current cluster version; was able to set it to %s", version)
			} else {
				__antithesis_instrumentation__.Notify(45558)
				if !testutils.IsError(err, "cannot set cluster.preserve_downgrade_option") {
					__antithesis_instrumentation__.Notify(45559)
					return err
				} else {
					__antithesis_instrumentation__.Notify(45560)
				}
			}
			__antithesis_instrumentation__.Notify(45556)
			return nil
		}
		__antithesis_instrumentation__.Notify(45512)

		for i := 1; i < nodes; i++ {
			__antithesis_instrumentation__.Notify(45561)
			t.WorkerStatus("upgrading ", i)
			if err := c.StopCockroachGracefullyOnNode(ctx, t.L(), i); err != nil {
				__antithesis_instrumentation__.Notify(45564)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45565)
			}
			__antithesis_instrumentation__.Notify(45562)
			c.Put(ctx, t.Cockroach(), "./cockroach", c.Node(i))
			startOpts := option.DefaultStartOpts()
			c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(i))
			if err := sleep(stageDuration); err != nil {
				__antithesis_instrumentation__.Notify(45566)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45567)
			}
			__antithesis_instrumentation__.Notify(45563)

			if upgraded, err := checkUpgraded(); err != nil {
				__antithesis_instrumentation__.Notify(45568)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45569)
				if upgraded {
					__antithesis_instrumentation__.Notify(45570)
					t.Fatal("cluster setting version shouldn't be upgraded before all nodes are running the new version")
				} else {
					__antithesis_instrumentation__.Notify(45571)
				}
			}
		}
		__antithesis_instrumentation__.Notify(45513)

		if err := c.StopCockroachGracefullyOnNode(ctx, t.L(), nodes-1); err != nil {
			__antithesis_instrumentation__.Notify(45572)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45573)
		}
		__antithesis_instrumentation__.Notify(45514)
		if err := c.StopCockroachGracefullyOnNode(ctx, t.L(), nodes); err != nil {
			__antithesis_instrumentation__.Notify(45574)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45575)
		}
		__antithesis_instrumentation__.Notify(45515)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.Node(nodes))
		startOpts := option.DefaultStartOpts()
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(nodes))
		if err := sleep(stageDuration); err != nil {
			__antithesis_instrumentation__.Notify(45576)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45577)
		}
		__antithesis_instrumentation__.Notify(45516)

		if upgraded, err := checkUpgraded(); err != nil {
			__antithesis_instrumentation__.Notify(45578)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45579)
			if upgraded {
				__antithesis_instrumentation__.Notify(45580)
				t.Fatal("cluster setting version shouldn't be upgraded before all non-decommissioned nodes are alive")
			} else {
				__antithesis_instrumentation__.Notify(45581)
			}
		}
		__antithesis_instrumentation__.Notify(45517)

		nodeDecommissioned := nodes - 2
		if err := decommissionAndStop(nodeDecommissioned); err != nil {
			__antithesis_instrumentation__.Notify(45582)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45583)
		}
		__antithesis_instrumentation__.Notify(45518)
		if err := sleep(timeUntilStoreDead + buff); err != nil {
			__antithesis_instrumentation__.Notify(45584)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45585)
		}
		__antithesis_instrumentation__.Notify(45519)

		if err := checkDowngradeOption("1.9"); err != nil {
			__antithesis_instrumentation__.Notify(45586)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45587)
		}
		__antithesis_instrumentation__.Notify(45520)
		if err := checkDowngradeOption("99.9"); err != nil {
			__antithesis_instrumentation__.Notify(45588)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45589)
		}
		__antithesis_instrumentation__.Notify(45521)

		if _, err := db.ExecContext(ctx,
			"SET CLUSTER SETTING cluster.preserve_downgrade_option = $1;", oldVersion,
		); err != nil {
			__antithesis_instrumentation__.Notify(45590)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45591)
		}
		__antithesis_instrumentation__.Notify(45522)
		if err := sleep(stageDuration); err != nil {
			__antithesis_instrumentation__.Notify(45592)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45593)
		}
		__antithesis_instrumentation__.Notify(45523)

		startOpts = option.DefaultStartOpts()
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(nodes-1))
		if err := sleep(stageDuration); err != nil {
			__antithesis_instrumentation__.Notify(45594)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45595)
		}
		__antithesis_instrumentation__.Notify(45524)

		t.WorkerStatus("check cluster version has not been upgraded")
		if upgraded, err := checkUpgraded(); err != nil {
			__antithesis_instrumentation__.Notify(45596)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45597)
			if upgraded {
				__antithesis_instrumentation__.Notify(45598)
				t.Fatal("cluster setting version shouldn't be upgraded because cluster.preserve_downgrade_option is set properly")
			} else {
				__antithesis_instrumentation__.Notify(45599)
			}
		}
		__antithesis_instrumentation__.Notify(45525)

		if _, err := db.ExecContext(ctx,
			"SET CLUSTER SETTING version = crdb_internal.node_executable_version();",
		); err == nil {
			__antithesis_instrumentation__.Notify(45600)
			t.Fatal("should not be able to set cluster setting version before resetting cluster.preserve_downgrade_option")
		} else {
			__antithesis_instrumentation__.Notify(45601)
			if !testutils.IsError(err, "cluster.preserve_downgrade_option is set to") {
				__antithesis_instrumentation__.Notify(45602)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45603)
			}
		}
		__antithesis_instrumentation__.Notify(45526)

		if _, err := db.ExecContext(ctx,
			"RESET CLUSTER SETTING cluster.preserve_downgrade_option;",
		); err != nil {
			__antithesis_instrumentation__.Notify(45604)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45605)
		}
		__antithesis_instrumentation__.Notify(45527)
		if err := sleep(stageDuration); err != nil {
			__antithesis_instrumentation__.Notify(45606)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45607)
		}
		__antithesis_instrumentation__.Notify(45528)

		t.WorkerStatus("check cluster version has been upgraded")
		if upgraded, err := checkUpgraded(); err != nil {
			__antithesis_instrumentation__.Notify(45608)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45609)
			if !upgraded {
				__antithesis_instrumentation__.Notify(45610)
				t.Fatalf("cluster setting version is not upgraded, still %s", oldVersion)
			} else {
				__antithesis_instrumentation__.Notify(45611)
			}
		}
		__antithesis_instrumentation__.Notify(45529)

		t.WorkerStatus("check cluster setting cluster.preserve_downgrade_option has been set to an empty string")
		var downgradeVersion string
		if err := db.QueryRowContext(ctx,
			"SHOW CLUSTER SETTING cluster.preserve_downgrade_option",
		).Scan(&downgradeVersion); err != nil {
			__antithesis_instrumentation__.Notify(45612)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45613)
		}
		__antithesis_instrumentation__.Notify(45530)
		if downgradeVersion != "" {
			__antithesis_instrumentation__.Notify(45614)
			t.Fatalf("cluster setting cluster.preserve_downgrade_option is %s, should be an empty string", downgradeVersion)
		} else {
			__antithesis_instrumentation__.Notify(45615)
		}
		__antithesis_instrumentation__.Notify(45531)

		c.Wipe(ctx, c.Node(nodeDecommissioned))
	}
	__antithesis_instrumentation__.Notify(45502)

	r.Add(registry.TestSpec{
		Name:    `autoupgrade`,
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(5),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45616)
			pred, err := PredecessorVersion(*t.BuildVersion())
			if err != nil {
				__antithesis_instrumentation__.Notify(45618)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45619)
			}
			__antithesis_instrumentation__.Notify(45617)
			runAutoUpgrade(ctx, t, c, pred)
		},
	})
}
