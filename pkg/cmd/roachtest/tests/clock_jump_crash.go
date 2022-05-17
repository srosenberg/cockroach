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

func runClockJump(ctx context.Context, t test.Test, c cluster.Cluster, tc clockJumpTestCase) {
	__antithesis_instrumentation__.Notify(46640)

	if c.Spec().NodeCount != 1 {
		__antithesis_instrumentation__.Notify(46647)
		t.Fatalf("Expected num nodes to be 1, got: %d", c.Spec().NodeCount)
	} else {
		__antithesis_instrumentation__.Notify(46648)
	}
	__antithesis_instrumentation__.Notify(46641)

	t.Status("deploying offset injector")
	offsetInjector := newOffsetInjector(t, c)
	if err := offsetInjector.deploy(ctx); err != nil {
		__antithesis_instrumentation__.Notify(46649)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46650)
	}
	__antithesis_instrumentation__.Notify(46642)

	if err := c.RunE(ctx, c.Node(1), "test -x ./cockroach"); err != nil {
		__antithesis_instrumentation__.Notify(46651)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	} else {
		__antithesis_instrumentation__.Notify(46652)
	}
	__antithesis_instrumentation__.Notify(46643)
	c.Wipe(ctx)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	db := c.Conn(ctx, t.L(), c.Spec().NodeCount)
	defer db.Close()
	if _, err := db.Exec(
		fmt.Sprintf(
			`SET CLUSTER SETTING server.clock.forward_jump_check_enabled= %v`,
			tc.jumpCheckEnabled)); err != nil {
		__antithesis_instrumentation__.Notify(46653)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46654)
	}
	__antithesis_instrumentation__.Notify(46644)

	time.Sleep(10 * time.Second)

	if !isAlive(db, t.L()) {
		__antithesis_instrumentation__.Notify(46655)
		t.Fatal("Node unexpectedly crashed")
	} else {
		__antithesis_instrumentation__.Notify(46656)
	}
	__antithesis_instrumentation__.Notify(46645)

	t.Status("injecting offset")

	var aliveAfterOffset bool
	defer func() {
		__antithesis_instrumentation__.Notify(46657)
		offsetInjector.recover(ctx, c.Spec().NodeCount)

		time.Sleep(3 * time.Second)
		if !isAlive(db, t.L()) {
			__antithesis_instrumentation__.Notify(46658)
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(1))
		} else {
			__antithesis_instrumentation__.Notify(46659)
		}
	}()
	__antithesis_instrumentation__.Notify(46646)
	defer offsetInjector.recover(ctx, c.Spec().NodeCount)
	offsetInjector.offset(ctx, c.Spec().NodeCount, tc.offset)

	time.Sleep(3 * time.Second)

	t.Status("validating health")
	aliveAfterOffset = isAlive(db, t.L())
	if aliveAfterOffset != tc.aliveAfterOffset {
		__antithesis_instrumentation__.Notify(46660)
		t.Fatalf("Expected node health %v, got %v", tc.aliveAfterOffset, aliveAfterOffset)
	} else {
		__antithesis_instrumentation__.Notify(46661)
	}
}

type clockJumpTestCase struct {
	name             string
	jumpCheckEnabled bool
	offset           time.Duration
	aliveAfterOffset bool
}

func registerClockJumpTests(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46662)
	testCases := []clockJumpTestCase{
		{
			name:             "large_forward_enabled",
			offset:           500 * time.Millisecond,
			jumpCheckEnabled: true,
			aliveAfterOffset: false,
		},
		{
			name: "small_forward_enabled",

			offset:           100 * time.Millisecond,
			jumpCheckEnabled: true,
			aliveAfterOffset: true,
		},
		{
			name:             "large_backward_enabled",
			offset:           -500 * time.Millisecond,
			jumpCheckEnabled: true,
			aliveAfterOffset: true,
		},
		{
			name:             "large_forward_disabled",
			offset:           500 * time.Millisecond,
			jumpCheckEnabled: false,
			aliveAfterOffset: true,
		},
		{
			name:             "large_backward_disabled",
			offset:           -500 * time.Millisecond,
			jumpCheckEnabled: false,
			aliveAfterOffset: true,
		},
	}

	for i := range testCases {
		__antithesis_instrumentation__.Notify(46663)
		tc := testCases[i]
		s := registry.TestSpec{
			Name:  "clock/jump/" + tc.name,
			Owner: registry.OwnerKV,

			Cluster: r.MakeClusterSpec(1, spec.ReuseTagged("offset-injector")),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46665)
				runClockJump(ctx, t, c, tc)
			},
		}
		__antithesis_instrumentation__.Notify(46664)
		r.Add(s)
	}
}
