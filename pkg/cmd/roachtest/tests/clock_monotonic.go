package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

func runClockMonotonicity(
	ctx context.Context, t test.Test, c cluster.Cluster, tc clockMonotonicityTestCase,
) {
	__antithesis_instrumentation__.Notify(46666)

	if c.Spec().NodeCount != 1 {
		__antithesis_instrumentation__.Notify(46676)
		t.Fatalf("Expected num nodes to be 1, got: %d", c.Spec().NodeCount)
	} else {
		__antithesis_instrumentation__.Notify(46677)
	}
	__antithesis_instrumentation__.Notify(46667)

	t.Status("deploying offset injector")
	offsetInjector := newOffsetInjector(t, c)
	if err := offsetInjector.deploy(ctx); err != nil {
		__antithesis_instrumentation__.Notify(46678)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46679)
	}
	__antithesis_instrumentation__.Notify(46668)

	if err := c.RunE(ctx, c.Node(1), "test -x ./cockroach"); err != nil {
		__antithesis_instrumentation__.Notify(46680)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	} else {
		__antithesis_instrumentation__.Notify(46681)
	}
	__antithesis_instrumentation__.Notify(46669)
	c.Wipe(ctx)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	db := c.Conn(ctx, t.L(), c.Spec().NodeCount)
	defer db.Close()
	if _, err := db.Exec(
		fmt.Sprintf(`SET CLUSTER SETTING server.clock.persist_upper_bound_interval = '%v'`,
			tc.persistWallTimeInterval)); err != nil {
		__antithesis_instrumentation__.Notify(46682)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46683)
	}
	__antithesis_instrumentation__.Notify(46670)

	time.Sleep(10 * time.Second)

	if !isAlive(db, t.L()) {
		__antithesis_instrumentation__.Notify(46684)
		t.Fatal("Node unexpectedly crashed")
	} else {
		__antithesis_instrumentation__.Notify(46685)
	}
	__antithesis_instrumentation__.Notify(46671)

	preRestartTime, err := dbUnixEpoch(db)
	if err != nil {
		__antithesis_instrumentation__.Notify(46686)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46687)
	}
	__antithesis_instrumentation__.Notify(46672)

	defer func() {
		__antithesis_instrumentation__.Notify(46688)
		if !isAlive(db, t.L()) {
			__antithesis_instrumentation__.Notify(46690)
			t.Fatal("Node unexpectedly crashed")
		} else {
			__antithesis_instrumentation__.Notify(46691)
		}
		__antithesis_instrumentation__.Notify(46689)

		c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(c.Spec().NodeCount))
		t.L().Printf("recovering from injected clock offset")

		offsetInjector.recover(ctx, c.Spec().NodeCount)

		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(c.Spec().NodeCount))
		if !isAlive(db, t.L()) {
			__antithesis_instrumentation__.Notify(46692)
			t.Fatal("Node unexpectedly crashed")
		} else {
			__antithesis_instrumentation__.Notify(46693)
		}
	}()
	__antithesis_instrumentation__.Notify(46673)

	t.Status("stopping cockroach")
	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(c.Spec().NodeCount))
	t.Status("injecting offset")
	offsetInjector.offset(ctx, c.Spec().NodeCount, tc.offset)
	t.Status("starting cockroach post offset")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(c.Spec().NodeCount))

	if !isAlive(db, t.L()) {
		__antithesis_instrumentation__.Notify(46694)
		t.Fatal("Node unexpectedly crashed")
	} else {
		__antithesis_instrumentation__.Notify(46695)
	}
	__antithesis_instrumentation__.Notify(46674)

	postRestartTime, err := dbUnixEpoch(db)
	if err != nil {
		__antithesis_instrumentation__.Notify(46696)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46697)
	}
	__antithesis_instrumentation__.Notify(46675)

	t.Status("validating clock monotonicity")
	t.L().Printf("pre-restart time:  %f\n", preRestartTime)
	t.L().Printf("post-restart time: %f\n", postRestartTime)
	difference := postRestartTime - preRestartTime
	t.L().Printf("time-difference: %v\n", time.Duration(difference*float64(time.Second)))

	if tc.expectIncreasingWallTime {
		__antithesis_instrumentation__.Notify(46698)
		if preRestartTime > postRestartTime {
			__antithesis_instrumentation__.Notify(46699)
			t.Fatalf("Expected pre-restart time %f < post-restart time %f", preRestartTime, postRestartTime)
		} else {
			__antithesis_instrumentation__.Notify(46700)
		}
	} else {
		__antithesis_instrumentation__.Notify(46701)
		if preRestartTime < postRestartTime {
			__antithesis_instrumentation__.Notify(46702)
			t.Fatalf("Expected pre-restart time %f > post-restart time %f", preRestartTime, postRestartTime)
		} else {
			__antithesis_instrumentation__.Notify(46703)
		}
	}
}

type clockMonotonicityTestCase struct {
	name                     string
	persistWallTimeInterval  time.Duration
	offset                   time.Duration
	expectIncreasingWallTime bool
}

func registerClockMonotonicTests(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46704)
	testCases := []clockMonotonicityTestCase{
		{
			name:                     "persistent",
			offset:                   -60 * time.Second,
			persistWallTimeInterval:  500 * time.Millisecond,
			expectIncreasingWallTime: true,
		},
	}

	for i := range testCases {
		__antithesis_instrumentation__.Notify(46705)
		tc := testCases[i]
		s := registry.TestSpec{
			Name:  "clock/monotonic/" + tc.name,
			Owner: registry.OwnerKV,

			Cluster: r.MakeClusterSpec(1, spec.ReuseTagged("offset-injector")),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46707)
				runClockMonotonicity(ctx, t, c, tc)
			},
		}
		__antithesis_instrumentation__.Notify(46706)
		r.Add(s)
	}
}
