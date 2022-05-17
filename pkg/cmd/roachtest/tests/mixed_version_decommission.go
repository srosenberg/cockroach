package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
)

func runDecommissionMixedVersions(
	ctx context.Context, t test.Test, c cluster.Cluster, buildVersion version.Version,
) {
	__antithesis_instrumentation__.Notify(49254)
	predecessorVersion, err := PredecessorVersion(buildVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(49256)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49257)
	}
	__antithesis_instrumentation__.Notify(49255)

	h := newDecommTestHelper(t, c)

	pinnedUpgrade := h.getRandNode()
	t.L().Printf("pinned n%d for upgrade", pinnedUpgrade)

	const mainVersion = ""
	allNodes := c.All()
	u := newVersionUpgradeTest(c,

		uploadVersionStep(allNodes, predecessorVersion),
		uploadVersionStep(allNodes, mainVersion),

		startVersion(allNodes, predecessorVersion),
		waitForUpgradeStep(allNodes),
		preventAutoUpgradeStep(h.nodeIDs[0]),

		binaryUpgradeStep(c.Node(pinnedUpgrade), mainVersion),
		binaryUpgradeStep(c.Node(h.getRandNodeOtherThan(pinnedUpgrade)), mainVersion),
		checkAllMembership(pinnedUpgrade, "active"),

		partialDecommissionStep(h.getRandNode(), h.getRandNode(), predecessorVersion),
		checkOneDecommissioning(h.getRandNode()),
		checkOneMembership(pinnedUpgrade, "decommissioning"),

		recommissionAllStep(h.getRandNode(), predecessorVersion),
		checkNoDecommissioning(h.getRandNode()),
		checkAllMembership(pinnedUpgrade, "active"),

		binaryUpgradeStep(allNodes, predecessorVersion),

		binaryUpgradeStep(allNodes, mainVersion),
		allowAutoUpgradeStep(1),
		waitForUpgradeStep(allNodes),

		fullyDecommissionStep(h.getRandNodeOtherThan(pinnedUpgrade), h.getRandNode(), ""),
		checkOneMembership(pinnedUpgrade, "decommissioned"),
	)

	u.run(ctx, t)
}

func cockroachBinaryPath(version string) string {
	__antithesis_instrumentation__.Notify(49258)
	if version == "" {
		__antithesis_instrumentation__.Notify(49260)
		return "./cockroach"
	} else {
		__antithesis_instrumentation__.Notify(49261)
	}
	__antithesis_instrumentation__.Notify(49259)
	return fmt.Sprintf("./v%s/cockroach", version)
}

func partialDecommissionStep(target, from int, binaryVersion string) versionStep {
	__antithesis_instrumentation__.Notify(49262)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49263)
		c := u.c
		c.Run(ctx, c.Node(from), cockroachBinaryPath(binaryVersion), "node", "decommission",
			"--wait=none", "--insecure", strconv.Itoa(target))
	}
}

func recommissionAllStep(from int, binaryVersion string) versionStep {
	__antithesis_instrumentation__.Notify(49264)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49265)
		c := u.c
		c.Run(ctx, c.Node(from), cockroachBinaryPath(binaryVersion), "node", "recommission",
			"--insecure", c.All().NodeIDsString())
	}
}

func fullyDecommissionStep(target, from int, binaryVersion string) versionStep {
	__antithesis_instrumentation__.Notify(49266)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49267)
		c := u.c
		c.Run(ctx, c.Node(from), cockroachBinaryPath(binaryVersion), "node", "decommission",
			"--wait=all", "--insecure", strconv.Itoa(target))
	}
}

func checkOneDecommissioning(from int) versionStep {
	__antithesis_instrumentation__.Notify(49268)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49269)

		if err := retry.ForDuration(testutils.DefaultSucceedsSoonDuration, func() error {
			__antithesis_instrumentation__.Notify(49270)
			db := u.conn(ctx, t, from)
			var count int
			if err := db.QueryRow(
				`select count(*) from crdb_internal.gossip_liveness where decommissioning = true;`).Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(49274)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49275)
			}
			__antithesis_instrumentation__.Notify(49271)

			if count != 1 {
				__antithesis_instrumentation__.Notify(49276)
				return errors.Newf("expected to find 1 node with decommissioning=true, found %d", count)
			} else {
				__antithesis_instrumentation__.Notify(49277)
			}
			__antithesis_instrumentation__.Notify(49272)

			var nodeID int
			if err := db.QueryRow(
				`select node_id from crdb_internal.gossip_liveness where decommissioning = true;`).Scan(&nodeID); err != nil {
				__antithesis_instrumentation__.Notify(49278)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49279)
			}
			__antithesis_instrumentation__.Notify(49273)
			t.L().Printf("n%d decommissioning=true", nodeID)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(49280)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49281)
		}
	}
}

func checkNoDecommissioning(from int) versionStep {
	__antithesis_instrumentation__.Notify(49282)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49283)
		if err := retry.ForDuration(testutils.DefaultSucceedsSoonDuration, func() error {
			__antithesis_instrumentation__.Notify(49284)
			db := u.conn(ctx, t, from)
			var count int
			if err := db.QueryRow(
				`select count(*) from crdb_internal.gossip_liveness where decommissioning = true;`).Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(49287)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49288)
			}
			__antithesis_instrumentation__.Notify(49285)

			if count != 0 {
				__antithesis_instrumentation__.Notify(49289)
				return errors.Newf("expected to find 0 nodes with decommissioning=false, found %d", count)
			} else {
				__antithesis_instrumentation__.Notify(49290)
			}
			__antithesis_instrumentation__.Notify(49286)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(49291)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49292)
		}
	}
}

func checkOneMembership(from int, membership string) versionStep {
	__antithesis_instrumentation__.Notify(49293)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49294)
		if err := retry.ForDuration(testutils.DefaultSucceedsSoonDuration, func() error {
			__antithesis_instrumentation__.Notify(49295)
			db := u.conn(ctx, t, from)
			var count int
			if err := db.QueryRow(
				`select count(*) from crdb_internal.gossip_liveness where membership = $1;`, membership).Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(49299)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49300)
			}
			__antithesis_instrumentation__.Notify(49296)

			if count != 1 {
				__antithesis_instrumentation__.Notify(49301)
				return errors.Newf("expected to find 1 node with membership=%s, found %d", membership, count)
			} else {
				__antithesis_instrumentation__.Notify(49302)
			}
			__antithesis_instrumentation__.Notify(49297)

			var nodeID int
			if err := db.QueryRow(
				`select node_id from crdb_internal.gossip_liveness where decommissioning = true;`).Scan(&nodeID); err != nil {
				__antithesis_instrumentation__.Notify(49303)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49304)
			}
			__antithesis_instrumentation__.Notify(49298)
			t.L().Printf("n%d membership=%s", nodeID, membership)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(49305)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49306)
		}
	}
}

func checkAllMembership(from int, membership string) versionStep {
	__antithesis_instrumentation__.Notify(49307)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49308)
		if err := retry.ForDuration(testutils.DefaultSucceedsSoonDuration, func() error {
			__antithesis_instrumentation__.Notify(49309)
			db := u.conn(ctx, t, from)
			var count int
			if err := db.QueryRow(
				`select count(*) from crdb_internal.gossip_liveness where membership != $1;`, membership).Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(49312)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49313)
			}
			__antithesis_instrumentation__.Notify(49310)

			if count != 0 {
				__antithesis_instrumentation__.Notify(49314)
				return errors.Newf("expected to find 0 nodes with membership!=%s, found %d", membership, count)
			} else {
				__antithesis_instrumentation__.Notify(49315)
			}
			__antithesis_instrumentation__.Notify(49311)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(49316)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49317)
		}
	}
}

func uploadVersionStep(nodes option.NodeListOption, version string) versionStep {
	__antithesis_instrumentation__.Notify(49318)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49319)

		uploadVersion(ctx, t, u.c, nodes, version)
	}
}

func startVersion(nodes option.NodeListOption, version string) versionStep {
	__antithesis_instrumentation__.Notify(49320)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(49321)
		settings := install.MakeClusterSettings(install.BinaryOption(cockroachBinaryPath(version)))
		startOpts := option.DefaultStartOpts()
		u.c.Start(ctx, t.L(), startOpts, settings, nodes)
	}
}
