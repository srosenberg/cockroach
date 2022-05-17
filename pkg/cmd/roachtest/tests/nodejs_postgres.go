package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

var repoOwner = "richardjcai"
var supportedBranch = "allowing_passing_certs_through_pg_env"

func registerNodeJSPostgres(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49581)
	runNodeJSPostgres := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(49583)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(49586)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(49587)
		}
		__antithesis_instrumentation__.Notify(49584)
		node := c.Node(1)
		t.Status("setting up cockroach")
		err := c.PutE(ctx, t.L(), t.Cockroach(), "./cockroach", c.All())
		require.NoError(t, err)
		settings := install.MakeClusterSettings(install.SecureOption(true))
		err = c.StartE(ctx, t.L(), option.DefaultStartOpts(), settings)
		require.NoError(t, err)

		const user = "testuser"

		err = repeatRunE(ctx, t, c, node, "create sql user",
			fmt.Sprintf(
				`./cockroach sql --certs-dir certs -e "CREATE USER %s CREATEDB"`, user,
			))
		require.NoError(t, err)

		err = repeatRunE(ctx, t, c, node, "create test database",
			`./cockroach sql --certs-dir certs -e "CREATE DATABASE postgres_node_test"`,
		)
		require.NoError(t, err)

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		require.NoError(t, err)

		err = alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0])
		require.NoError(t, err)

		err = repeatRunE(
			ctx,
			t,
			c,
			node,
			"add nodesource repository",
			`sudo apt install ca-certificates && curl -sL https://deb.nodesource.com/setup_12.x | sudo -E bash -`,
		)
		require.NoError(t, err)

		err = repeatRunE(
			ctx, t, c, node, "install nodejs and npm", `sudo apt-get -qq install nodejs`,
		)
		require.NoError(t, err)

		err = repeatRunE(
			ctx, t, c, node, "update npm", `sudo npm i -g npm`,
		)
		require.NoError(t, err)

		err = repeatRunE(
			ctx, t, c, node, "install yarn", `sudo npm i -g yarn`,
		)
		require.NoError(t, err)

		err = repeatRunE(
			ctx, t, c, node, "install lerna", `sudo npm i --g lerna`,
		)
		require.NoError(t, err)

		err = repeatRunE(
			ctx, t, c, node, "remove old node-postgres", `sudo rm -rf /mnt/data1/node-postgres`,
		)
		require.NoError(t, err)

		err = repeatGitCloneE(
			ctx,
			t,
			c,
			fmt.Sprintf("https://github.com/%s/node-postgres.git", repoOwner),
			"/mnt/data1/node-postgres",
			supportedBranch,
			node,
		)
		require.NoError(t, err)

		err = repeatRunE(
			ctx, t, c, node, "configure git to avoid unauthenticated protocol",
			`cd /mnt/data1/node-postgres && sudo git config --global url."https://github".insteadOf "git://github"`,
		)
		require.NoError(t, err)

		err = repeatRunE(
			ctx,
			t,
			c,
			node,
			"building node-postgres",
			`cd /mnt/data1/node-postgres/ && sudo yarn && sudo yarn lerna bootstrap`,
		)
		require.NoError(t, err)

		t.Status("running node-postgres tests")
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
			fmt.Sprintf(
				`cd /mnt/data1/node-postgres/ && sudo \
PGPORT=26257 PGUSER=%s PGSSLMODE=require PGDATABASE=postgres_node_test \
PGSSLCERT=$HOME/certs/client.%s.crt PGSSLKEY=$HOME/certs/client.%s.key PGSSLROOTCERT=$HOME/certs/ca.crt yarn test`,
				user, user, user,
			),
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(49588)

			commandError := (*install.NonZeroExitCode)(nil)
			if !errors.As(err, &commandError) {
				__antithesis_instrumentation__.Notify(49589)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49590)
			}
		} else {
			__antithesis_instrumentation__.Notify(49591)
		}
		__antithesis_instrumentation__.Notify(49585)

		rawResultsStr := result.Stdout + result.Stderr
		t.L().Printf("Test Results: %s", rawResultsStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(49592)

			if strings.Contains(rawResultsStr, "1 failing") && func() bool {
				__antithesis_instrumentation__.Notify(49594)
				return strings.Contains(rawResultsStr, "1) pool size of 1") == true
			}() == true {
				__antithesis_instrumentation__.Notify(49595)
				err = nil
			} else {
				__antithesis_instrumentation__.Notify(49596)
			}
			__antithesis_instrumentation__.Notify(49593)
			if err != nil {
				__antithesis_instrumentation__.Notify(49597)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49598)
			}
		} else {
			__antithesis_instrumentation__.Notify(49599)
		}
	}
	__antithesis_instrumentation__.Notify(49582)

	r.Add(registry.TestSpec{
		Name:    "node-postgres",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `driver`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49600)
			runNodeJSPostgres(ctx, t, c)
		},
	})
}
