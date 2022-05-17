package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/util/version"
)

func registerSchemaChangeMixedVersions(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49393)
	r.Add(registry.TestSpec{
		Name:  "schemachange/mixed-versions",
		Owner: registry.OwnerSQLSchema,

		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49394)
			maxOps := 100
			concurrency := 5
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(49396)
				maxOps = 10
				concurrency = 2
			} else {
				__antithesis_instrumentation__.Notify(49397)
			}
			__antithesis_instrumentation__.Notify(49395)
			runSchemaChangeMixedVersions(ctx, t, c, maxOps, concurrency, *t.BuildVersion())
		},
	})
}

func uploadAndInitSchemaChangeWorkload() versionStep {
	__antithesis_instrumentation__.Notify(49398)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49399)

		u.c.Put(ctx, t.DeprecatedWorkload(), "./workload", u.c.All())
		u.c.Run(ctx, u.c.All(), "./workload init schemachange")
	}
}

func runSchemaChangeWorkloadStep(loadNode, maxOps, concurrency int) versionStep {
	__antithesis_instrumentation__.Notify(49400)
	var numFeatureRuns int
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49401)
		numFeatureRuns++
		t.L().Printf("Workload step run: %d", numFeatureRuns)
		runCmd := []string{
			"./workload run schemachange --verbose=1",

			"--tolerate-errors=true",
			fmt.Sprintf("--max-ops %d", maxOps),
			fmt.Sprintf("--concurrency %d", concurrency),
			fmt.Sprintf("{pgurl:1-%d}", u.c.Spec().NodeCount),
		}
		u.c.Run(ctx, u.c.Node(loadNode), runCmd...)
	}
}

func runSchemaChangeMixedVersions(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	maxOps int,
	concurrency int,
	buildVersion version.Version,
) {
	__antithesis_instrumentation__.Notify(49402)
	predecessorVersion, err := PredecessorVersion(buildVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(49405)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49406)
	}
	__antithesis_instrumentation__.Notify(49403)

	const mainVersion = ""
	schemaChangeStep := runSchemaChangeWorkloadStep(c.All().RandNode()[0], maxOps, concurrency)
	if buildVersion.Major() < 20 {
		__antithesis_instrumentation__.Notify(49407)

		schemaChangeStep = nil
	} else {
		__antithesis_instrumentation__.Notify(49408)
	}
	__antithesis_instrumentation__.Notify(49404)

	u := newVersionUpgradeTest(c,
		uploadAndStartFromCheckpointFixture(c.All(), predecessorVersion),
		uploadAndInitSchemaChangeWorkload(),
		waitForUpgradeStep(c.All()),

		preventAutoUpgradeStep(1),
		schemaChangeStep,

		binaryUpgradeStep(c.Node(3), mainVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(2), mainVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(1), mainVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(4), mainVersion),
		schemaChangeStep,

		binaryUpgradeStep(c.Node(2), predecessorVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(4), predecessorVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(3), predecessorVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(1), predecessorVersion),
		schemaChangeStep,

		binaryUpgradeStep(c.Node(4), mainVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(3), mainVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(1), mainVersion),
		schemaChangeStep,
		binaryUpgradeStep(c.Node(2), mainVersion),
		schemaChangeStep,

		allowAutoUpgradeStep(1),
		waitForUpgradeStep(c.All()),
		schemaChangeStep,
	)

	u.run(ctx, t)
}
