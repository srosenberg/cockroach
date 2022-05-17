package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

var flowableReleaseTagRegex = regexp.MustCompile(`^flowable-(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)

func registerFlowable(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47700)
	runFlowable := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(47702)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(47710)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(47711)
		}
		__antithesis_instrumentation__.Notify(47703)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		t.Status("cloning flowable and installing prerequisites")
		latestTag, err := repeatGetLatestTag(
			ctx, t, "flowable", "flowable-engine", flowableReleaseTagRegex,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47712)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47713)
		}
		__antithesis_instrumentation__.Notify(47704)
		t.L().Printf("Latest Flowable release is %s.", latestTag)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47714)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47715)
		}
		__antithesis_instrumentation__.Notify(47705)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install default-jre openjdk-8-jdk-headless gradle maven`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47716)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47717)
		}
		__antithesis_instrumentation__.Notify(47706)

		if err := repeatRunE(
			ctx, t, c, node, "remove old Flowable", `rm -rf /mnt/data1/flowable-engine`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47718)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47719)
		}
		__antithesis_instrumentation__.Notify(47707)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/flowable/flowable-engine.git",
			"/mnt/data1/flowable-engine",
			latestTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(47720)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47721)
		}
		__antithesis_instrumentation__.Notify(47708)

		t.Status("building Flowable")
		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"building Flowable",
			`cd /mnt/data1/flowable-engine/ && mvn clean install -DskipTests`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47722)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47723)
		}
		__antithesis_instrumentation__.Notify(47709)

		if err := c.RunE(ctx, node,
			`cd /mnt/data1/flowable-engine/ && mvn clean test -Dtest=Flowable6Test#testLongServiceTaskLoop -Ddb=crdb`,
		); err != nil {
			__antithesis_instrumentation__.Notify(47724)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47725)
		}
	}
	__antithesis_instrumentation__.Notify(47701)

	r.Add(registry.TestSpec{
		Name:    "flowable",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(47726)
			runFlowable(ctx, t, c)
		},
	})
}
