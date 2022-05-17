package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

var popReleaseTag = regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var popSupportedTag = "v5.3.3"

func registerPop(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49801)
	runPop := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(49803)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(49810)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(49811)
		}
		__antithesis_instrumentation__.Notify(49804)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())
		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49812)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49813)
		}
		__antithesis_instrumentation__.Notify(49805)
		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(49814)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49815)
		}
		__antithesis_instrumentation__.Notify(49806)

		t.Status("cloning pop and installing prerequisites")
		latestTag, err := repeatGetLatestTag(
			ctx, t, "gobuffalo", "pop", popReleaseTag)
		if err != nil {
			__antithesis_instrumentation__.Notify(49816)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49817)
		}
		__antithesis_instrumentation__.Notify(49807)
		t.L().Printf("Latest pop release is %s.", latestTag)
		t.L().Printf("Supported pop release is %s.", popSupportedTag)

		installGolang(ctx, t, c, node)

		const (
			popPath = "/mnt/data1/pop/"
		)

		if err := repeatRunE(
			ctx, t, c, node, "remove old pop", fmt.Sprintf("rm -rf %s", popPath),
		); err != nil {
			__antithesis_instrumentation__.Notify(49818)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49819)
		}
		__antithesis_instrumentation__.Notify(49808)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/gobuffalo/pop.git",
			popPath,
			popSupportedTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(49820)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49821)
		}
		__antithesis_instrumentation__.Notify(49809)

		t.Status("building and setting up tests")

		err = c.RunE(ctx, node, fmt.Sprintf(`cd %s && go build -v -tags sqlite -o tsoda ./soda`, popPath))
		require.NoError(t, err)

		err = c.RunE(ctx, node, fmt.Sprintf(`cd %s && ./tsoda drop -e cockroach -c ./database.yml -p ./testdata/migrations`, popPath))
		require.NoError(t, err)

		err = c.RunE(ctx, node, fmt.Sprintf(`cd %s && ./tsoda create -e cockroach -c ./database.yml -p ./testdata/migrations`, popPath))
		require.NoError(t, err)

		err = c.RunE(ctx, node, fmt.Sprintf(`cd %s && ./tsoda migrate -e cockroach -c ./database.yml -p ./testdata/migrations`, popPath))
		require.NoError(t, err)

		t.Status("running pop test suite")

		err = c.RunE(ctx, node, fmt.Sprintf(`cd %s && SODA_DIALECT=cockroach go test -race -tags sqlite -v ./... -count=1`, popPath))
		require.NoError(t, err, "error while running pop tests")
	}
	__antithesis_instrumentation__.Notify(49802)

	r.Add(registry.TestSpec{
		Name:    "pop",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run:     runPop,
	})
}
